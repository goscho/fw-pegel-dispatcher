package webio_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goscho/fw-pegel-dispatcher/internal/httpclient"
	"github.com/goscho/fw-pegel-dispatcher/internal/webio"
)

func TestRequester_RequestCurrentValues_ok(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method %s", r.Method)
		}
		_, _ = w.Write([]byte("0,209 m;3,400 l/m²"))
	}))
	t.Cleanup(srv.Close)

	r := webio.New(httpclient.New(), srv.URL)
	v, err := r.RequestCurrentValues()
	if err != nil {
		t.Fatal(err)
	}
	if v.Port1.Value != 0.209 || v.Port1.Unit != "m" {
		t.Fatalf("port1 %+v", v.Port1)
	}
	if v.Port2.Value != 3.4 || v.Port2.Unit != "l/m²" {
		t.Fatalf("port2 %+v", v.Port2)
	}
}

func TestParseResponseBody_invalid(t *testing.T) {
	t.Parallel()
	for _, body := range []string{"some response", "", "onlyone"} {
		t.Run(body, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(body))
			}))
			t.Cleanup(srv.Close)
			r := webio.New(httpclient.New(), srv.URL)
			_, err := r.RequestCurrentValues()
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestRequester_badStatus(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("0,209 m;3,400 l/m²"))
	}))
	t.Cleanup(srv.Close)
	r := webio.New(httpclient.New(), srv.URL)
	_, err := r.RequestCurrentValues()
	if err == nil {
		t.Fatal("expected error")
	}
}
