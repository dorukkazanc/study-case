package hub

import (
	"sync"

	domain "study-case/internal/domain/notification"
)

type StatusUpdate struct {
	ID     string        `json:"id"`
	Status domain.Status `json:"status"`
}

type Hub struct {
	mu          sync.RWMutex
	subscribers map[string][]chan StatusUpdate
}

func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[string][]chan StatusUpdate),
	}
}

func (h *Hub) Subscribe(notifID string) chan StatusUpdate {
	ch := make(chan StatusUpdate, 1)
	h.mu.Lock()
	h.subscribers[notifID] = append(h.subscribers[notifID], ch)
	h.mu.Unlock()
	return ch
}

func (h *Hub) Unsubscribe(notifID string, ch chan StatusUpdate) {
	h.mu.Lock()
	defer h.mu.Unlock()
	subs := h.subscribers[notifID]
	for i, s := range subs {
		if s == ch {
			h.subscribers[notifID] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	if len(h.subscribers[notifID]) == 0 {
		delete(h.subscribers, notifID)
	}
	close(ch)
}

func (h *Hub) Publish(notifID string, status domain.Status) {
	h.mu.RLock()
	subs := h.subscribers[notifID]
	h.mu.RUnlock()
	for _, ch := range subs {
		select {
		case ch <- StatusUpdate{ID: notifID, Status: status}:
		default:
		}
	}
}
