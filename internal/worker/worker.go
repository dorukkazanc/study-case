package worker

import (
	"context"
	domain "study-case/internal/domain/notification"
	"study-case/internal/metrics"
	"study-case/internal/queue"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

type Config struct {
	Concurrency  int
	MaxRetries   int
	PollInterval time.Duration
	RateLimitRPS int
}

type Worker struct {
	cfg      Config
	queue    queue.Queue
	repo     Repository
	provider Provider
	limiters map[domain.Channel]*rate.Limiter
	log      zerolog.Logger
}

func NewWorker(cfg Config, q queue.Queue, repo Repository, provider Provider, log zerolog.Logger) *Worker {
	channels := []domain.Channel{domain.ChannelSMS, domain.ChannelEmail, domain.ChannelPush}
	limiters := make(map[domain.Channel]*rate.Limiter, len(channels))
	rps := cfg.RateLimitRPS
	if rps <= 0 {
		rps = 100
	}
	for _, ch := range channels {
		limiters[ch] = rate.NewLimiter(rate.Limit(rps), rps)
	}
	return &Worker{
		cfg:      cfg,
		queue:    q,
		repo:     repo,
		provider: provider,
		limiters: limiters,
		log:      log,
	}
}

func (w *Worker) Run(ctx context.Context) {
	channels := []domain.Channel{domain.ChannelEmail, domain.ChannelPush, domain.ChannelSMS}

	var wg sync.WaitGroup
	for _, channel := range channels {
		for i := 0; i < w.cfg.Concurrency; i++ {
			wg.Add(1)
			go func(channel domain.Channel) {
				defer wg.Done()
				w.runLoop(ctx, channel)
			}(channel)
		}
	}
	wg.Wait()
}

func (w *Worker) runLoop(ctx context.Context, channel domain.Channel) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n, err := w.queue.Dequeue(ctx, channel)
		if err != nil || n == nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(w.cfg.PollInterval):
			}
			continue
		}

		if err := w.limiters[n.Channel].Wait(ctx); err != nil {
			return
		}
		w.Process(ctx, n)
	}
}

func (w *Worker) Process(ctx context.Context, n *domain.Notification) {
	if err := w.repo.UpdateStatus(ctx, n.ID, domain.StatusProcessing); err != nil {
		return
	}

	if n.BatchID != nil {
		_ = w.repo.UpdateBatchCounters(ctx, *n.BatchID, n.Status, domain.StatusProcessing)
	}

	id, err := w.provider.Send(ctx, n)
	if err != nil {
		w.log.Error().Err(err).Str("id", n.ID).Msg("provider send failed")
		w.handleFailure(ctx, n, err)
		return
	}
	n.MarkSent(id)

	if err := w.repo.UpdateStatus(ctx, n.ID, domain.StatusSent, domain.WithProviderID(id)); err != nil {
		return
	}

	if n.BatchID != nil {
		_ = w.repo.UpdateBatchCounters(ctx, *n.BatchID, domain.StatusProcessing, n.Status)
	}

	metrics.NotificationsSent.WithLabelValues(string(n.Channel)).Inc()
	w.log.Info().Str("id", n.ID).Str("provider_id", id).Msg("notification sent")
}

func (w *Worker) handleFailure(ctx context.Context, n *domain.Notification, err error) {
	if n.RetryCount >= w.cfg.MaxRetries {
		e := w.queue.MoveToDead(ctx, n)
		if e != nil {
			return
		}
		if err := w.repo.UpdateStatus(ctx, n.ID, domain.StatusFailed); err != nil {
			return
		}
		n.MarkFailed("retry limit reached")
		metrics.NotificationsFailed.WithLabelValues(string(n.Channel)).Inc()
		w.log.Warn().Str("id", n.ID).Int("retry_count", n.RetryCount).Msg("notification moved to dead letter queue")

		if n.BatchID != nil {
			_ = w.repo.UpdateBatchCounters(ctx, *n.BatchID, domain.StatusProcessing, n.Status)
		}
		return
	}
	if err := w.repo.UpdateStatus(ctx, n.ID, domain.StatusQueued, domain.WithErrorMessage(err.Error())); err != nil {
		return
	}
	n.RetryCount++
	n.MarkQueued()

	if n.BatchID != nil {
		_ = w.repo.UpdateBatchCounters(ctx, *n.BatchID, domain.StatusProcessing, domain.StatusQueued)
	}

	delay := 5 * time.Second * (1 << n.RetryCount)
	w.log.Warn().Str("id", n.ID).Int("retry_count", n.RetryCount).Dur("delay", delay).Msg("notification scheduled for retry")
	if err := w.queue.EnqueueDelayed(ctx, n, delay); err != nil {
		return
	}

}
