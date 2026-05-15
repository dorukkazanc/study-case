package worker

import (
	"study-case/internal/queue"
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
