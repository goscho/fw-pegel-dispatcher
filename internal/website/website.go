package website

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type addPegelRequest struct {
	Value      float32 `json:"value"`
	RecordedAt string  `json:"recorded_at"`
}

// Client posts gauge values to the pegel API.
type Client struct {
	HTTP    *http.Client
	BaseURL string
	APIKey  string
}

// New returns a new Client.
func New(client *http.Client, baseURL, apiKey string) *Client {
	return &Client{
		HTTP:    client,
		BaseURL: baseURL,
		APIKey:  apiKey,
	}
}

// UpdateWebsite POSTs one gauge measurement to /api/pegel.
func (c *Client) UpdateWebsite(waterLevel float32) error {
	payload := addPegelRequest{
		Value:      waterLevel,
		RecordedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("website marshal payload: %w", err)
	}
	endpoint := strings.TrimRight(c.BaseURL, "/") + "/api/pegel"

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("website request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.APIKey)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("website request: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("website: HTTP %d", resp.StatusCode)
	}
	return nil
}
