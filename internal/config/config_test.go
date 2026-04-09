package config_test

import (
	"testing"

	"github.com/goscho/fw-pegel-dispatcher/internal/config"
)

func TestLoad_missingVars(t *testing.T) {
	t.Setenv("WEBIO_URL", "")
	t.Setenv("THINGSPEAK_API_URL", "")
	t.Setenv("THINGSPEAK_API_KEY", "")
	t.Setenv("PEGEL_API_BASE_URL", "")
	t.Setenv("PEGEL_API_KEY", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_ok(t *testing.T) {
	for k, v := range map[string]string{
		"WEBIO_URL":          "http://1.2.3.4/Single",
		"THINGSPEAK_API_URL": "https://api.thingspeak.com/update",
		"THINGSPEAK_API_KEY": "k",
		"PEGEL_API_BASE_URL": "https://example",
		"PEGEL_API_KEY":      "s",
	} {
		t.Setenv(k, v)
	}
	c, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.WebIOURL != "http://1.2.3.4/Single" || c.LogDir != "." {
		t.Fatalf("%+v", c)
	}
	if c.ScheduleCron != "0 */5 * * * *" {
		t.Fatalf("ScheduleCron %q", c.ScheduleCron)
	}
}

func TestLoad_scheduleCron_custom(t *testing.T) {
	for k, v := range map[string]string{
		"WEBIO_URL":          "http://1.2.3.4/Single",
		"THINGSPEAK_API_URL": "https://api.thingspeak.com/update",
		"THINGSPEAK_API_KEY": "k",
		"PEGEL_API_BASE_URL": "https://example",
		"PEGEL_API_KEY":      "s",
		"SCHEDULE_CRON":      "0 */10 * * * *",
	} {
		t.Setenv(k, v)
	}
	c, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.ScheduleCron != "0 */10 * * * *" {
		t.Fatalf("ScheduleCron %q", c.ScheduleCron)
	}
}
