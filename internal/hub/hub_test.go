package hub_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	domain "study-case/internal/domain/notification"
	"study-case/internal/hub"
)

func TestSubscribe_ReturnsChannel(t *testing.T) {
	h := hub.NewHub()
	ch := h.Subscribe("notif-1")
	assert.NotNil(t, ch)
}

func TestPublish_DeliverToSubscriber(t *testing.T) {
	h := hub.NewHub()
	ch := h.Subscribe("notif-1")

	h.Publish("notif-1", domain.StatusSent)

	update := <-ch
	assert.Equal(t, "notif-1", update.ID)
	assert.Equal(t, domain.StatusSent, update.Status)
}

func TestPublish_NoSubscriber_NoPanic(t *testing.T) {
	h := hub.NewHub()
	assert.NotPanics(t, func() {
		h.Publish("notif-999", domain.StatusSent)
	})
}

func TestPublish_MultipleSubscribers_AllReceive(t *testing.T) {
	h := hub.NewHub()
	ch1 := h.Subscribe("notif-1")
	ch2 := h.Subscribe("notif-1")

	h.Publish("notif-1", domain.StatusSent)

	assert.Equal(t, domain.StatusSent, (<-ch1).Status)
	assert.Equal(t, domain.StatusSent, (<-ch2).Status)
}

func TestPublish_OnlyTargetedSubscriberReceives(t *testing.T) {
	h := hub.NewHub()
	ch1 := h.Subscribe("notif-1")
	ch2 := h.Subscribe("notif-2")

	h.Publish("notif-1", domain.StatusSent)

	assert.Equal(t, domain.StatusSent, (<-ch1).Status)
	select {
	case <-ch2:
		t.Fatal("notif-2 subscriber should not have received anything")
	default:
	}
}

func TestUnsubscribe_ChannelClosed(t *testing.T) {
	h := hub.NewHub()
	ch := h.Subscribe("notif-1")
	h.Unsubscribe("notif-1", ch)

	_, open := <-ch
	assert.False(t, open)
}

func TestUnsubscribe_PublishAfterUnsubscribe_NotDelivered(t *testing.T) {
	h := hub.NewHub()
	ch := h.Subscribe("notif-1")
	h.Unsubscribe("notif-1", ch)

	// channel closed, publish should not panic
	assert.NotPanics(t, func() {
		h.Publish("notif-1", domain.StatusSent)
	})
}

func TestUnsubscribe_RemovesKeyWhenEmpty(t *testing.T) {
	h := hub.NewHub()
	ch := h.Subscribe("notif-1")
	h.Unsubscribe("notif-1", ch)

	// subscribing again should still work (key was cleaned up)
	ch2 := h.Subscribe("notif-1")
	h.Publish("notif-1", domain.StatusSent)
	assert.Equal(t, domain.StatusSent, (<-ch2).Status)
}

func TestConcurrent_SubscribePublish(t *testing.T) {
	h := hub.NewHub()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch := h.Subscribe("notif-1")
			h.Publish("notif-1", domain.StatusSent)
			h.Unsubscribe("notif-1", ch)
		}()
	}

	wg.Wait()
}
