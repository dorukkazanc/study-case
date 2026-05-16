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
	if err := w.repo.UpdateStatus(ctx, n.ID, domain.StatusProcessing); err != nil {
		return
	}

	id, err := w.provider.Send(ctx, n)
	if err != nil {
		w.handleFailure(ctx, n, err)
		return
	}
	n.MarkSent(id)

	if err := w.repo.UpdateStatus(ctx, n.ID, domain.StatusSent, domain.WithProviderID(id)); err != nil {
		return
	}
}

func (w *Worker) handleFailure(ctx context.Context, notification *domain.Notification, err error) {
	if notification.RetryCount >= w.cfg.MaxRetries {
		e := w.queue.MoveToDead(ctx, notification)
		if e != nil {
			return
		}
		if err := w.repo.UpdateStatus(ctx, notification.ID, domain.StatusFailed); err != nil {
			return
		}
		return
	}
	notification.MarkFailed(err.Error())
	if err := w.repo.UpdateStatus(ctx, notification.ID, domain.StatusQueued, domain.WithErrorMessage(err.Error())); err != nil {
		return
	}
	delay := 5 * time.Second * (1 << notification.RetryCount)
	if err := w.queue.EnqueueDelayed(ctx, notification, delay); err != nil {
		return
	}

}
