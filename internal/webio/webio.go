package webio

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// Value is one port reading (value + unit).
type Value struct {
	Value float32
	Unit  string
}

// Values holds port1 and port2 readings from the Web-IO body.
type Values struct {
	Port1 Value
	Port2 Value
}

// Requester fetches the Web-IO endpoint.
type Requester struct {
	HTTP *http.Client
	URL  string
}

// New returns a new Requester.
func New(client *http.Client, url string) *Requester {
	return &Requester{
		HTTP: client,
		URL:  url,
	}
}

// RequestCurrentValues GETs the configured URL and parses the body (e.g. "0,209 m;0,000 l/m²").
func (r *Requester) RequestCurrentValues() (Values, error) {
	req, err := http.NewRequest(http.MethodGet, r.URL, nil)
	if err != nil {
		return Values{}, fmt.Errorf("webio request: %w", err)
	}
	resp, err := r.HTTP.Do(req)
	if err != nil {
		return Values{}, fmt.Errorf("webio request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Values{}, fmt.Errorf("webio: HTTP %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return Values{}, fmt.Errorf("webio read body: %w", err)
	}
	v, err := parseResponseBody(string(b))
	if err != nil {
		return Values{}, fmt.Errorf("webio parse: %w", err)
	}
	return v, nil
}

func parseResponseBody(valueString string) (Values, error) {
	parts := strings.Split(strings.TrimSpace(valueString), ";")
	if len(parts) != 2 {
		return Values{}, fmt.Errorf("expected 2 segments separated by ';', got %d", len(parts))
	}
	v1, err := parseValueUnitString(strings.TrimSpace(parts[0]))
	if err != nil {
		return Values{}, err
	}
	v2, err := parseValueUnitString(strings.TrimSpace(parts[1]))
	if err != nil {
		return Values{}, err
	}
	return Values{Port1: v1, Port2: v2}, nil
}

func parseValueUnitString(s string) (Value, error) {
	fields := strings.Fields(s)
	if len(fields) < 2 {
		return Value{}, fmt.Errorf("expected value and unit, got %q", s)
	}
	valStr := strings.ReplaceAll(fields[0], ",", ".")
	f64, parseErr := strconv.ParseFloat(valStr, 32)
	if parseErr != nil {
		return Value{}, fmt.Errorf("parse float %q: %w", valStr, parseErr)
	}
	v := float32(f64)
	unit := strings.TrimSpace(strings.Join(fields[1:], " "))
	return Value{Value: v, Unit: unit}, nil
}
