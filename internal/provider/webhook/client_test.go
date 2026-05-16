package webhook_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "study-case/internal/domain/notification"
	"study-case/internal/provider/webhook"
)

func newNotification() *domain.Notification {
	return domain.NewNotification("+905551234567", domain.ChannelSMS, "test message", domain.PriorityNormal)
}

func TestSend_Success_MessageIDFromBody(t *testing.T) {
	messageID := uuid.NewString()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"messageId": messageID})
	}))
	defer srv.Close()

	client := webhook.NewClient(srv.URL, 5*time.Second)
	id, err := client.Send(context.Background(), newNotification())

	require.NoError(t, err)
	assert.Equal(t, messageID, id)
}

func TestSend_Success_MessageIDFromHeader(t *testing.T) {
	headerID := uuid.NewString()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-Id", headerID)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	client := webhook.NewClient(srv.URL, 5*time.Second)
	id, err := client.Send(context.Background(), newNotification())

	require.NoError(t, err)
	assert.Equal(t, headerID, id)
}

func TestSend_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := webhook.NewClient(srv.URL, 5*time.Second)
	_, err := client.Send(context.Background(), newNotification())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestSend_BadRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	client := webhook.NewClient(srv.URL, 5*time.Second)
	_, err := client.Send(context.Background(), newNotification())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "400")
}

func TestSend_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	client := webhook.NewClient(srv.URL, 50*time.Millisecond)
	_, err := client.Send(context.Background(), newNotification())

	require.Error(t, err)
}

func TestSend_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := webhook.NewClient(srv.URL, 5*time.Second)
	_, err := client.Send(ctx, newNotification())

	require.Error(t, err)
}

func TestSend_SendsCorrectPayload(t *testing.T) {
	n := newNotification()
	var received map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	client := webhook.NewClient(srv.URL, 5*time.Second)
	client.Send(context.Background(), n)

	assert.Equal(t, n.Recipient, received["recipient"])
	assert.Equal(t, string(n.Channel), received["channel"])
	assert.Equal(t, n.Content, received["content"])
}
