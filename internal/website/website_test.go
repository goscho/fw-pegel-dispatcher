package website_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/goscho/fw-pegel-dispatcher/internal/httpclient"
	"github.com/goscho/fw-pegel-dispatcher/internal/website"
)

type addPegelRequest struct {
	Value      float32 `json:"value"`
	RecordedAt string  `json:"recorded_at"`
}

func TestClient_UpdateWebsite_success(t *testing.T) {
	t.Parallel()
	const apiKey = "123secret"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method %s", r.Method)
		}
		if r.URL.Path != "/api/pegel" {
			t.Fatalf("path %s", r.URL.Path)
		}
		if got := r.Header.Get("X-API-Key"); got != apiKey {
			t.Fatalf("X-API-Key %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type %q", got)
		}
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		var payload addPegelRequest
		if err := json.Unmarshal(b, &payload); err != nil {
			t.Fatal(err)
		}
		if payload.Value != 1.43 {
			t.Fatalf("value %v", payload.Value)
		}
		parsed, err := time.Parse(time.RFC3339Nano, payload.RecordedAt)
		if err != nil {
			t.Fatalf("recorded_at parse: %v", err)
		}
		if parsed.Location() != time.UTC {
			t.Fatalf("recorded_at not UTC: %v", parsed.Location())
		}
		w.WriteHeader(http.StatusCreated)
	}))
	t.Cleanup(srv.Close)

	c := &website.Client{HTTP: httpclient.New(), BaseURL: srv.URL, APIKey: apiKey}
	if err := c.UpdateWebsite(1.43); err != nil {
		t.Fatal(err)
	}
}

func TestClient_UpdateWebsite_badStatus(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)
	c := &website.Client{HTTP: httpclient.New(), BaseURL: srv.URL, APIKey: "s"}
	err := c.UpdateWebsite(1)
	if err == nil {
		t.Fatal("expected error")
	}
}
