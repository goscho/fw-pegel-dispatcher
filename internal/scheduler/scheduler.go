package scheduler

import (
	"log/slog"

	"github.com/goscho/fw-pegel-dispatcher/internal/round"
	"github.com/goscho/fw-pegel-dispatcher/internal/webio"
)

// WebIOReader fetches current Web-IO values.
type WebIOReader interface {
	RequestCurrentValues() (webio.Values, error)
}

// ThingSpeakWriter posts to ThingSpeak.
type ThingSpeakWriter interface {
	AddEntry(fields ...float32) (int64, error)
}

// WebsiteUpdater pushes stream gauge to the website API.
type WebsiteUpdater interface {
	UpdateWebsite(gaugeLevel float32) error
}

// Scheduler orchestrates Web-IO → ThingSpeak → website (same order as Java UpdateScheduler).
type Scheduler struct {
	Log        *slog.Logger
	WebIO      WebIOReader
	ThingSpeak ThingSpeakWriter
	Website    WebsiteUpdater
}

// New returns a new Scheduler.
func New(log *slog.Logger, webIO WebIOReader, thingSpeak ThingSpeakWriter, website WebsiteUpdater) *Scheduler {
	return &Scheduler{
		Log:        log,
		WebIO:      webIO,
		ThingSpeak: thingSpeak,
		Website:    website,
	}
}

// UpdateValues runs one full update cycle (invoked on cron).
func (s *Scheduler) UpdateValues() {
	s.Log.Info("update started")
	values, err := s.WebIO.RequestCurrentValues()
	if err != nil {
		s.Log.Error("Updating failed. Exception while requesting Web IO", "err", err)
		return
	}

	succeeded := s.updateThingSpeak(values)
	succeeded = s.updateWebsite(values) && succeeded

	if succeeded {
		s.Log.Info("update successful")
	} else {
		s.Log.Error("update failed")
	}
}

func (s *Scheduler) updateWebsite(v webio.Values) bool {
	streamGauge := round.Float32(v.Port1.Value, 2)
	if err := s.Website.UpdateWebsite(streamGauge); err != nil {
		s.Log.Error("Updating failed. Exception while pushing data to Website API", "err", err)
		return false
	}
	return true
}

func (s *Scheduler) updateThingSpeak(v webio.Values) bool {
	streamGauge := round.Float32(v.Port1.Value, 2)
	rainfall := round.Float32(v.Port2.Value, 0)
	if _, err := s.ThingSpeak.AddEntry(streamGauge, rainfall); err != nil {
		s.Log.Error("Updating failed. Exception while pushing data to ThingSpeak", "err", err)
		return false
	}
	return true
}
