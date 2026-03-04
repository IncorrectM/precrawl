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
	"github.com/IncorrectM/precrawl/internal/config"
	"github.com/IncorrectM/precrawl/internal/server"
	"github.com/IncorrectM/precrawl/internal/task"
	"github.com/IncorrectM/precrawl/internal/transformer"
)

func main() {
	// Graceful shutdown on interrupt signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// initialize task queue and browser pool
	queue := task.NewQueue()
	pool, err := browser.NewPool(ctx, 2)
	if err != nil {
		log.Fatalf("failed to create browser pool: %v", err)
	}
	defer pool.Close()

	// read configuration from environment variables
	baseTargetURL := os.Getenv("PRECRAWL_BASE_TARGET_URL")

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

	// by default, use all transformers
	transformers := transformer.DefaultTransformers()

	// by default, use 2 workers to process the queue
	workerCount := 2

	// read configuration from config.yml
	// this overrides environment variables if present
	configData, err := os.ReadFile("config.yml")
	if err != nil {
		log.Printf("warning: failed to read config.yml: %v", err)
	} else {
		config, err := config.LoadConfig(configData)
		if err != nil {
			log.Printf("warning: failed to parse config.yml: %v", err)
		}
		if config.BaseTargetURL != nil {
			baseTargetURL = *config.BaseTargetURL
		}
		if config.DefaultSelector != nil {
			defaultSelector = *config.DefaultSelector
		}
		if config.DefaultWaitTimeout != nil {
			parsed, err := time.ParseDuration(*config.DefaultWaitTimeout)
			if err != nil {
				log.Fatalf("invalid PRECRAWL_RENDER_TIMEOUT in config.yml: %v", err)
			}
			defaultWaitTimeout = parsed
		}
		if config.Transformers != nil {
			transformers = transformer.FromNames(*config.Transformers...)
		}
		if config.WorkerCount != nil && *config.WorkerCount > 0 {
			log.Printf("overriding worker count to %d from config.yml", *config.WorkerCount)
			workerCount = *config.WorkerCount
		}
	}

	// validate configuration
	if baseTargetURL == "" {
		log.Fatal("PRECRAWL_BASE_TARGET_URL is required")
	}

	// start the server
	if err := server.Run(ctx, server.Config{
		Queue:              queue,
		Pool:               pool,
		WorkerCount:        workerCount,
		BaseTargetURL:      baseTargetURL,
		DefaultSelector:    defaultSelector,
		DefaultWaitTimeout: defaultWaitTimeout,
	}, transformers); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}
