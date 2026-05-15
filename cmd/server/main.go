package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"study-case/internal/config"
	pgrepo "study-case/internal/repository/postgres"
)

func main() {
	cfg := config.Load()

	db := mustConnectDB(cfg)
	mustMigrate(cfg)

	_ = pgrepo.NewNotificationRepository(db)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	<-ctx.Done()
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
