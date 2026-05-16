package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	goredis "github.com/redis/go-redis/v9"

	domain "study-case/internal/domain/notification"
)

type Promoter struct {
	client   *goredis.Client
	queue    *PriorityQueue
	interval time.Duration
}

func NewPromoter(client *goredis.Client, queue *PriorityQueue, interval time.Duration) *Promoter {
	return &Promoter{client: client, queue: queue, interval: interval}
}

func (p *Promoter) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(p.interval):
			if err := p.promote(ctx); err != nil {
				log.Printf("promoter error: %v", err)
			}
		}
	}
}

func (p *Promoter) promote(ctx context.Context) error {
	now := float64(time.Now().Unix())

	results, err := p.client.ZRangeArgs(ctx, goredis.ZRangeArgs{
		Key:     keyDelayed,
		Start:   "-inf",
		Stop:    fmt.Sprintf("%f", now),
		ByScore: true,
	}).Result()
	if err != nil {
		return fmt.Errorf("promoter range: %w", err)
	}

	for _, member := range results {
		removed, err := p.client.ZRem(ctx, keyDelayed, member).Result()
		if err != nil || removed == 0 {
			continue
		}

		var n domain.Notification
		if err := json.Unmarshal([]byte(member), &n); err != nil {
			log.Printf("promoter unmarshal error: %v", err)
			continue
		}

		if err := p.queue.Enqueue(ctx, &n); err != nil {
			log.Printf("promoter enqueue error: %v", err)
		}
	}

	return nil
}
