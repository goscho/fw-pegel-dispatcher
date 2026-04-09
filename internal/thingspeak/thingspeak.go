package thingspeak

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	paramAPIKey    = "api_key"
	paramCreatedAt = "created_at"
	paramField     = "field"
	errorResponse  = "0"
)

// Client posts channel updates to ThingSpeak.
type Client struct {
	HTTP *http.Client
	URL  string
	Key  string
}

// New returns a new Client.
func New(client *http.Client, url, key string) *Client {
	return &Client{
		HTTP: client,
		URL:  url,
		Key:  key,
	}
}

// AddEntry sends field1..fieldN (up to 8) as form fields.
func (c *Client) AddEntry(fields ...float32) (int64, error) {
	if len(fields) < 1 {
		return 0, fmt.Errorf("at least one field value must be provided")
	}
	return c.executePost(c.URL, c.Key, fields)
}

func (c *Client) executePost(apiURL, apiKey string, fields []float32) (int64, error) {
	form := url.Values{}
	form.Set(paramAPIKey, apiKey)
	form.Set(paramCreatedAt, utcTimestampMinute())
	for i := 1; i <= len(fields) && i <= 8; i++ {
		form.Set(paramField+strconv.Itoa(i), formatFloat(fields[i-1]))
	}
	req, err := http.NewRequest(http.MethodPost, apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return 0, fmt.Errorf("thingspeak request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return 0, fmt.Errorf("thingspeak request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("thingspeak read body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("thingspeak: HTTP %d", resp.StatusCode)
	}
	bodyStr := strings.TrimSpace(string(body))
	if bodyStr == "" || bodyStr == errorResponse {
		return 0, fmt.Errorf("received %s instead of entry ID", errorResponse)
	}
	id, err := strconv.ParseInt(bodyStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("thingspeak parse entry id: %w", err)
	}
	return id, nil
}

// utcTimestampMinute matches Java: ZonedDateTime.now(UTC).truncatedTo(MINUTES).format(ISO_INSTANT).
func utcTimestampMinute() string {
	t := time.Now().UTC().Truncate(time.Minute)
	return t.Format(time.RFC3339)
}

func formatFloat(f float32) string {
	return strconv.FormatFloat(float64(f), 'f', -1, 32)
}
