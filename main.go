package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	baseTargetURL := os.Getenv("BASE_TARGET_URL")
	if baseTargetURL == "" {
		log.Fatal("BASE_TARGET_URL is required")
	}

	defaultSelector := os.Getenv("DEFAULT_SELECTOR")

	if err := server.Run(ctx, server.Config{Queue: queue, Pool: pool, WorkerCount: 2, BaseTargetURL: baseTargetURL, DefaultSelector: defaultSelector}); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}
