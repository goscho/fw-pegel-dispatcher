package thingspeak_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/goscho/fw-pegel-dispatcher/internal/httpclient"
	"github.com/goscho/fw-pegel-dispatcher/internal/thingspeak"
)

func TestClient_AddEntry_success(t *testing.T) {
	t.Parallel()
	const wantID int64 = 123
	var gotForm url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/x-www-form-urlencoded") {
			t.Fatalf("content-type %q", ct)
		}
		b, _ := io.ReadAll(r.Body)
		var err error
		gotForm, err = url.ParseQuery(string(b))
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte("123"))
	}))
	t.Cleanup(srv.Close)

	c := &thingspeak.Client{HTTP: httpclient.New(), URL: srv.URL, Key: "123ABC"}
	id, err := c.AddEntry(1.456, 0.34)
	if err != nil {
		t.Fatal(err)
	}
	if id != wantID {
		t.Fatalf("id %d want %d", id, wantID)
	}
	if gotForm.Get("api_key") != "123ABC" {
		t.Fatalf("api_key %q", gotForm.Get("api_key"))
	}
	if gotForm.Get("field1") != "1.456" || gotForm.Get("field2") != "0.34" {
		t.Fatalf("fields %v %v", gotForm.Get("field1"), gotForm.Get("field2"))
	}
	ts := gotForm.Get("created_at")
	if ts == "" {
		t.Fatal("missing created_at")
	}
	parsed, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if parsed.Year() != now.Year() || parsed.Month() != now.Month() || parsed.Day() != now.Day() {
		t.Fatalf("date mismatch %v vs %v", parsed, now)
	}
	if parsed.Second() != 0 || parsed.Nanosecond() != 0 {
		t.Fatalf("expected minute truncation, got %v", parsed)
	}
}

func TestClient_AddEntry_noFields(t *testing.T) {
	t.Parallel()
	c := &thingspeak.Client{HTTP: httpclient.New(), URL: "http://x", Key: "k"}
	_, err := c.AddEntry()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClient_AddEntry_errorZero(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("0"))
	}))
	t.Cleanup(srv.Close)
	c := &thingspeak.Client{HTTP: httpclient.New(), URL: srv.URL, Key: "k"}
	_, err := c.AddEntry(0.1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClient_AddEntry_errorWhitespaceZero(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(" 0 "))
	}))
	t.Cleanup(srv.Close)
	c := &thingspeak.Client{HTTP: httpclient.New(), URL: srv.URL, Key: "k"}
	_, err := c.AddEntry(0.1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClient_AddEntry_badHTTP(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("123"))
	}))
	t.Cleanup(srv.Close)
	c := &thingspeak.Client{HTTP: httpclient.New(), URL: srv.URL, Key: "k"}
	_, err := c.AddEntry(1.4)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClient_AddEntry_nonNumericBody(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("12.3"))
	}))
	t.Cleanup(srv.Close)
	c := &thingspeak.Client{HTTP: httpclient.New(), URL: srv.URL, Key: "k"}
	_, err := c.AddEntry(1.4)
	if err == nil {
		t.Fatal("expected error")
	}
}
