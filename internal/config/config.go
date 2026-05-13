package config

import (
	"fmt"
	"os"
	"strings"
)

const defaultScheduleCron = "0 */5 * * * *"

// Config holds runtime settings (env-based, Docker-friendly).
type Config struct {
	WebIOURL         string
	ThingSpeakAPIURL string
	ThingSpeakAPIKey string
	PegelAPIBaseURL  string
	PegelAPIKey      string
	ScheduleCron     string
}

// Load reads configuration from environment variables.
func Load() (Config, error) {
	c := Config{
		WebIOURL:         strings.TrimSpace(os.Getenv("WEBIO_URL")),
		ThingSpeakAPIURL: strings.TrimSpace(os.Getenv("THINGSPEAK_API_URL")),
		ThingSpeakAPIKey: strings.TrimSpace(os.Getenv("THINGSPEAK_API_KEY")),
		PegelAPIBaseURL:  strings.TrimSpace(os.Getenv("PEGEL_API_BASE_URL")),
		PegelAPIKey:      strings.TrimSpace(os.Getenv("PEGEL_API_KEY")),
		ScheduleCron:     strings.TrimSpace(os.Getenv("SCHEDULE_CRON")),
	}
	if c.ScheduleCron == "" {
		c.ScheduleCron = defaultScheduleCron
	}
	var missing []string
	if c.WebIOURL == "" {
		missing = append(missing, "WEBIO_URL")
	}
	if c.ThingSpeakAPIURL == "" {
		missing = append(missing, "THINGSPEAK_API_URL")
	}
	if c.ThingSpeakAPIKey == "" {
		missing = append(missing, "THINGSPEAK_API_KEY")
	}
	if c.PegelAPIBaseURL == "" {
		missing = append(missing, "PEGEL_API_BASE_URL")
	}
	if c.PegelAPIKey == "" {
		missing = append(missing, "PEGEL_API_KEY")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return c, nil
}
