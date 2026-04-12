package app

import (
	"testing"

	"github.com/dujiao-next/internal/config"
)

func TestShouldStartWorker(t *testing.T) {
	t.Run("skip worker when queue disabled in all mode", func(t *testing.T) {
		cfg := &config.Config{
			Queue: config.QueueConfig{
				Enabled: false,
			},
		}

		if shouldStartWorker(cfg, ModeAll) {
			t.Fatal("expected worker to be skipped when queue is disabled")
		}
	})

	t.Run("start worker when queue enabled in all mode", func(t *testing.T) {
		cfg := &config.Config{
			Queue: config.QueueConfig{
				Enabled: true,
			},
		}

		if !shouldStartWorker(cfg, ModeAll) {
			t.Fatal("expected worker to start when queue is enabled")
		}
	})

	t.Run("start worker in worker mode when queue enabled", func(t *testing.T) {
		cfg := &config.Config{
			Queue: config.QueueConfig{
				Enabled: true,
			},
		}

		if !shouldStartWorker(cfg, ModeWorker) {
			t.Fatal("expected worker mode to start worker when queue is enabled")
		}
	})
}
