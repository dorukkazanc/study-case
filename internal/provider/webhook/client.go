package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	domain "study-case/internal/domain/notification"
)

type Client struct {
	url        string
	httpClient *http.Client
}

func NewClient(url string, timeout time.Duration) *Client {
	return &Client{
		url:        url,
		httpClient: &http.Client{Timeout: timeout},
	}
}

type payload struct {
	ID        string `json:"id"`
	Recipient string `json:"recipient"`
	Channel   string `json:"channel"`
	Content   string `json:"content"`
	Priority  string `json:"priority"`
}

func (c *Client) Send(ctx context.Context, n *domain.Notification) (string, error) {
	body, err := json.Marshal(payload{
		ID:        n.ID,
		Recipient: n.Recipient,
		Channel:   string(n.Channel),
		Content:   n.Content,
		Priority:  string(n.Priority),
	})
	if err != nil {
		return "", fmt.Errorf("webhook marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("webhook new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("webhook send: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return resp.Header.Get("X-Request-Id"), nil
}
