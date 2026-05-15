package worker

import (
	"context"
	domain "study-case/internal/domain/notification"
	"study-case/internal/queue"
	"sync"
	"time"
)

type Config struct {
	Concurrency  int
	MaxRetries   int
	PollInterval time.Duration
}

type Worker struct {
	cfg      Config
	queue    queue.Queue
	repo     Repository
	provider Provider
}

func NewWorker(cfg Config, queue queue.Queue, repo Repository, provider Provider) *Worker {
	return &Worker{
		cfg:      cfg,
		queue:    queue,
		repo:     repo,
		provider: provider,
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

		w.Process(ctx, n)
	}
}

func (w *Worker) Process(ctx context.Context, n *domain.Notification) {
	panic("not implemented")
}
