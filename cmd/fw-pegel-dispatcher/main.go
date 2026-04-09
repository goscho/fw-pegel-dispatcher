package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/goscho/fw-pegel-dispatcher/internal/config"
	"github.com/goscho/fw-pegel-dispatcher/internal/httpclient"
	"github.com/goscho/fw-pegel-dispatcher/internal/logging"
	"github.com/goscho/fw-pegel-dispatcher/internal/scheduler"
	"github.com/goscho/fw-pegel-dispatcher/internal/thingspeak"
	"github.com/goscho/fw-pegel-dispatcher/internal/webio"
	"github.com/goscho/fw-pegel-dispatcher/internal/website"
	"github.com/robfig/cron/v3"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	logger, err := logging.New(cfg.LogDir)
	if err != nil {
		slog.Error("logging", "err", err)
		os.Exit(1)
	}

	httpc := httpclient.New()
	sched := scheduler.New(
		logger,
		webio.New(httpc, cfg.WebIOURL),
		thingspeak.New(httpc, cfg.ThingSpeakAPIURL, cfg.ThingSpeakAPIKey),
		website.New(httpc, cfg.PegelAPIBaseURL, cfg.PegelAPIKey),
	)

	c := cron.New(cron.WithSeconds())
	if _, err := c.AddFunc(cfg.ScheduleCron, sched.UpdateValues); err != nil {
		logger.Error("cron", "err", err)
		os.Exit(1)
	}
	c.Start()
	logger.Info("fw-pegel-dispatcher started", "cron", cfg.ScheduleCron)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx := c.Stop()
	<-ctx.Done()
	logger.Info("fw-pegel-dispatcher stopped")
}
