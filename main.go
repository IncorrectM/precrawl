package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IncorrectM/precrawl/internal/browser"
	"github.com/IncorrectM/precrawl/internal/server"
	"github.com/IncorrectM/precrawl/internal/task"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	queue := task.NewQueue()
	pool, err := browser.NewPool(ctx, 2)
	if err != nil {
		log.Fatalf("failed to create browser pool: %v", err)
	}
	defer pool.Close()

	baseTargetURL := os.Getenv("PRECRAWL_BASE_TARGET_URL")
	if baseTargetURL == "" {
		log.Fatal("PRECRAWL_BASE_TARGET_URL is required")
	}

	defaultSelector := os.Getenv("PRECRAWL_DEFAULT_SELECTOR")

	defaultWaitTimeout := 5 * time.Second
	if rawTimeout := os.Getenv("PRECRAWL_RENDER_TIMEOUT"); rawTimeout != "" {
		parsed, err := time.ParseDuration(rawTimeout)
		if err != nil {
			log.Fatalf("invalid PRECRAWL_RENDER_TIMEOUT: %v", err)
		}
		defaultWaitTimeout = parsed
	}
	if defaultWaitTimeout < 0 {
		log.Fatal("PRECRAWL_RENDER_TIMEOUT must be non-negative")
	}

	if err := server.Run(ctx, server.Config{Queue: queue, Pool: pool, WorkerCount: 2, BaseTargetURL: baseTargetURL, DefaultSelector: defaultSelector, DefaultWaitTimeout: defaultWaitTimeout}); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}
