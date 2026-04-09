package scheduler_test

import (
	"bytes"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/goscho/fw-pegel-dispatcher/internal/scheduler"
	"github.com/goscho/fw-pegel-dispatcher/internal/webio"
)

type fakeWebIO struct {
	v   webio.Values
	err error
}

func (f *fakeWebIO) RequestCurrentValues() (webio.Values, error) {
	return f.v, f.err
}

type fakeThingSpeak struct {
	id  int64
	err error
	got []float32
}

func (f *fakeThingSpeak) AddEntry(fields ...float32) (int64, error) {
	f.got = append([]float32(nil), fields...)
	return f.id, f.err
}

type fakeWebsite struct {
	err    error
	called bool
	got    float32
}

func (f *fakeWebsite) UpdateWebsite(gaugeLevel float32) error {
	f.called = true
	f.got = gaugeLevel
	return f.err
}

func TestScheduler_allSuccess(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	wio := &fakeWebIO{v: webio.Values{
		Port1: webio.Value{Value: 0.123, Unit: "m"},
		Port2: webio.Value{Value: 0.5, Unit: "x"},
	}}
	ts := &fakeThingSpeak{id: 123}
	ws := &fakeWebsite{}
	s := scheduler.New(log, wio, ts, ws)
	s.UpdateValues()
	if len(ts.got) != 2 || ts.got[0] != 0.12 || ts.got[1] != 1 {
		t.Fatalf("thingspeak got %#v", ts.got)
	}
	if ws.got != 0.12 {
		t.Fatalf("website got %#v", ws.got)
	}
	out := buf.String()
	if !strings.Contains(out, "update started") || !strings.Contains(out, "update successful") {
		t.Fatalf("log: %s", out)
	}
}

func TestScheduler_webioError(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	wio := &fakeWebIO{err: io.EOF}
	ts := &fakeThingSpeak{}
	ws := &fakeWebsite{}
	s := scheduler.New(log, wio, ts, ws)
	s.UpdateValues()
	if ts.got != nil {
		t.Fatalf("thingSpeak called: %#v", ts.got)
	}
	if ws.called {
		t.Fatalf("website called: %#v", ws.got)
	}
}

func TestScheduler_thingSpeakError_stillCallsWebsite(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	wio := &fakeWebIO{v: webio.Values{
		Port1: webio.Value{Value: 0.123, Unit: "m"},
		Port2: webio.Value{Value: 0.6, Unit: "x"},
	}}
	ts := &fakeThingSpeak{err: io.EOF}
	ws := &fakeWebsite{}
	s := scheduler.New(log, wio, ts, ws)
	s.UpdateValues()
	if ws.got != 0.12 {
		t.Fatalf("website got %#v", ws.got)
	}
}

func TestScheduler_websiteError(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	wio := &fakeWebIO{v: webio.Values{
		Port1: webio.Value{Value: 0.123, Unit: "m"},
		Port2: webio.Value{Value: 0.4, Unit: "x"},
	}}
	ts := &fakeThingSpeak{id: 1}
	ws := &fakeWebsite{err: io.EOF}
	s := scheduler.New(log, wio, ts, ws)
	s.UpdateValues()
	if len(ts.got) != 2 || ts.got[0] != 0.12 || ts.got[1] != 0 {
		t.Fatalf("thingspeak got %#v", ts.got)
	}
	if ws.got != 0.12 {
		t.Fatalf("website got %#v", ws.got)
	}
}
