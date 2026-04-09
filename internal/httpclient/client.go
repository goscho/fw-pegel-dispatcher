package httpclient

import (
	"net"
	"net/http"
	"time"
)

// New returns an *http.Client with 10s connect timeout and 10s response header timeout,
// matching the Spring RestTemplateBuilder settings in the former Java app.
func New() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext:           (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
			ResponseHeaderTimeout: 10 * time.Second,
			ForceAttemptHTTP2:     true,
		},
	}
}
