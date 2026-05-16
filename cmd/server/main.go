package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"study-case/internal/provider/webhook"
	queue3 "study-case/internal/queue"
	"study-case/internal/queue/redis"
	"study-case/internal/worker"
	"study-case/pkg/logger"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	goredis "github.com/redis/go-redis/v9"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"study-case/internal/api"
	"study-case/internal/config"
	pgrepo "study-case/internal/repository/postgres"
	"study-case/internal/service"
)

func main() {
	cfg := config.Load()

	db := mustConnectDB(cfg)
	mustMigrate(cfg)

	rds := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	zeroLogger := logger.NewLogger(cfg.Log.Level)

	repo := pgrepo.NewNotificationRepository(db)
	pq := redis.NewPriorityQueue(rds)
	queue := queue3.Queue(pq)
	svc := service.NewNotificationService(repo, queue, zeroLogger)
	router := api.NewNotificationRouter(svc, zeroLogger)
	promoter := redis.NewPromoter(rds, pq, time.Minute)
	collector := redis.NewMetricsCollector(pq, 15*time.Second)
	provider := webhook.NewClient(cfg.Provider.WebhookURL, cfg.Provider.Timeout)

	w := worker.NewWorker(worker.Config{
		Concurrency:  3,
		MaxRetries:   3,
		PollInterval: 500 * time.Millisecond,
		RateLimitRPS: cfg.Provider.RateLimitRPS,
	}, pq, repo, provider, zeroLogger)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go promoter.Start(ctx)
	go collector.Start(ctx)
	go w.Run(ctx)

	go func() {
		zeroLogger.Info().Msgf("server started on :%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zeroLogger.Fatal().Err(err).Msg("server error")
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		zeroLogger.Error().Err(err).Msg("shutdown error")
	}
	zeroLogger.Info().Msg("server stopped")
}

func mustConnectDB(cfg *config.Config) *gorm.DB {
	db, err := gorm.Open(pgdriver.Open(cfg.Database.DSN), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	return db
}

func mustMigrate(cfg *config.Config) {
	m, err := migrate.New("file://migrations", cfg.Database.DSN)
	if err != nil {
		log.Fatalf("failed to create migrator: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed to run migrations: %v", err)
	}
	log.Println("migrations applied")
}
