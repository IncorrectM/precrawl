package prerender

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/IncorrectM/precrawl/internal/browser"
)

func TestRenderValidation(t *testing.T) {
	t.Parallel()

	if _, err := Render(context.Background(), nil, "https://example.com", 0); !errors.Is(err, ErrInvalidPool) {
		t.Fatalf("expected ErrInvalidPool, got %v", err)
	}

	pool, err := browser.NewPool(context.Background(), 1)
	if err != nil {
		t.Fatalf("NewPool error: %v", err)
	}
	defer pool.Close()

	if _, err := Render(context.Background(), pool, "", 0); !errors.Is(err, ErrEmptyURL) {
		t.Fatalf("expected ErrEmptyURL, got %v", err)
	}

	if _, err := Render(context.Background(), pool, "https://example.com", -time.Millisecond); !errors.Is(err, ErrNegativeWait) {
		t.Fatalf("expected ErrNegativeWait, got %v", err)
	}
}

func TestRenderAcquireTimeout(t *testing.T) {
	t.Parallel()

	pool, err := browser.NewPool(context.Background(), 1)
	if err != nil {
		t.Fatalf("NewPool error: %v", err)
	}
	defer pool.Close()

	page, err := pool.AcquireBlank(context.Background())
	if err != nil {
		t.Fatalf("Acquire error: %v", err)
	}
	defer func() {
		if err := pool.Release(page); err != nil {
			t.Fatalf("Release error: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err = Render(ctx, pool, "https://example.com", 0)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}

func TestRenderSuccess(t *testing.T) {
	t.Parallel()

	pool, err := browser.NewPool(context.Background(), 1)
	if err != nil {
		t.Fatalf("NewPool error: %v", err)
	}
	defer pool.Close()

	html, err := Render(context.Background(), pool, "https://example.com", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if len(html) == 0 {
		t.Fatal("expected non-empty HTML")
	} else {
		t.Logf("Rendered HTML length: %d", len(html))
	}
}
