package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	domain "study-case/internal/domain/notification"
)

const (
	keyDelayed = "queue:delayed"
	keyDead    = "queue:dead"
)

var priorityScore = map[domain.Priority]float64{
	domain.PriorityHigh:   1,
	domain.PriorityNormal: 2,
	domain.PriorityLow:    3,
}

type PriorityQueue struct {
	client *goredis.Client
}

func NewPriorityQueue(client *goredis.Client) *PriorityQueue {
	return &PriorityQueue{client: client}
}

func channelKey(ch domain.Channel) string {
	return fmt.Sprintf("queue:%s", ch)
}

func (q *PriorityQueue) Enqueue(ctx context.Context, n *domain.Notification) error {
	data, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("enqueue marshal: %w", err)
	}
	return q.client.ZAdd(ctx, channelKey(n.Channel), goredis.Z{
		Score:  priorityScore[n.Priority],
		Member: string(data),
	}).Err()
}

func (q *PriorityQueue) Dequeue(ctx context.Context, channel domain.Channel) (*domain.Notification, error) {
	results, err := q.client.ZPopMin(ctx, channelKey(channel), 1).Result()
	if err != nil {
		return nil, fmt.Errorf("dequeue: %w", err)
	}
	if len(results) == 0 {
		return nil, nil
	}
	var n domain.Notification
	if err := json.Unmarshal([]byte(results[0].Member.(string)), &n); err != nil {
		return nil, fmt.Errorf("dequeue unmarshal: %w", err)
	}
	return &n, nil
}

func (q *PriorityQueue) EnqueueDelayed(ctx context.Context, n *domain.Notification, delay time.Duration) error {
	data, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("enqueue delayed marshal: %w", err)
	}
	score := float64(time.Now().Add(delay).Unix())
	return q.client.ZAdd(ctx, keyDelayed, goredis.Z{
		Score:  score,
		Member: string(data),
	}).Err()
}

func (q *PriorityQueue) MoveToDead(ctx context.Context, n *domain.Notification) error {
	data, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("move to dead marshal: %w", err)
	}
	return q.client.LPush(ctx, keyDead, string(data)).Err()
}
