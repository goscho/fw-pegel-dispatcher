package scheduler

import (
	"log/slog"

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

// WebsiteUpdater pushes stream level and rainfall to the website API.
type WebsiteUpdater interface {
	UpdateLevel(streamLevel float32) error
	UpdateRainfall(rainfall float32) error
}

// Scheduler orchestrates dispatching from Web-IO → website & ThingSpeak
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
		s.Log.Error("update aborted. Exception while requesting Web IO", "err", err)
		return
	}

	if s.updateStreamLevel(values) {
		s.Log.Info("pegel update successful")
	}
	if s.updateRainfall(values) {
		s.Log.Info("Rainfall update successful")
	}
	if s.updateThingSpeak(values) {
		s.Log.Info("ThingSpeak update successful")
	}

	s.Log.Info("update finished")
}

func (s *Scheduler) updateStreamLevel(v webio.Values) bool {
	if err := s.Website.UpdateLevel(v.Port1.Value); err != nil {
		s.Log.Error("Updating failed. Exception while pushing data to Website Pegel API", "err", err)
		return false
	}
	return true
}

func (s *Scheduler) updateRainfall(v webio.Values) bool {
	if err := s.Website.UpdateRainfall(v.Port2.Value); err != nil {
		s.Log.Error("Updating failed. Exception while pushing data to Website Rainfall API", "err", err)
		return false
	}
	return true
}

func (s *Scheduler) updateThingSpeak(v webio.Values) bool {
	streamGauge := v.Port1.Value
	rainfall := v.Port2.Value
	if _, err := s.ThingSpeak.AddEntry(streamGauge, rainfall); err != nil {
		s.Log.Error("Updating failed. Exception while pushing data to ThingSpeak", "err", err)
		return false
	}
	return true
}
