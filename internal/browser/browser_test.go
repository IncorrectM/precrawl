package browser

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestNewPoolInvalidSize(t *testing.T) {
	t.Parallel()

	_, err := NewPool(context.Background(), 0)
	if !errors.Is(err, ErrInvalidSize) {
		t.Fatalf("expected ErrInvalidSize, got %v", err)
	}
}

func TestAcquireRelease(t *testing.T) {
	t.Parallel()

	pool, err := NewPool(context.Background(), 2)
	if err != nil {
		t.Fatalf("NewPool error: %v", err)
	}
	t.Cleanup(pool.Close)

	if got := pool.Available(); got != 2 {
		t.Fatalf("expected 2 available, got %d", got)
	}

	page, err := pool.AcquireBlank(context.Background())
	if err != nil {
		t.Fatalf("Acquire error: %v", err)
	}

	if got := pool.Available(); got != 1 {
		t.Fatalf("expected 1 available, got %d", got)
	}

	if err := pool.Release(page); err != nil {
		t.Fatalf("Release error: %v", err)
	}

	if got := pool.Available(); got != 2 {
		t.Fatalf("expected 2 available after release, got %d", got)
	}
}

func TestAcquireTimeout(t *testing.T) {
	t.Parallel()

	pool, err := NewPool(context.Background(), 1)
	if err != nil {
		t.Fatalf("NewPool error: %v", err)
	}
	t.Cleanup(pool.Close)

	page, err := pool.AcquireBlank(context.Background())
	if err != nil {
		t.Fatalf("Acquire error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err = pool.AcquireBlank(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}

	if err := pool.Release(page); err != nil {
		t.Fatalf("Release error: %v", err)
	}
}

func TestCloseAndRelease(t *testing.T) {
	t.Parallel()

	pool, err := NewPool(context.Background(), 1)
	if err != nil {
		t.Fatalf("NewPool error: %v", err)
	}

	page, err := pool.AcquireBlank(context.Background())
	if err != nil {
		t.Fatalf("Acquire error: %v", err)
	}

	pool.Close()

	if _, err := pool.AcquireBlank(context.Background()); !errors.Is(err, ErrPoolClosed) {
		t.Fatalf("expected ErrPoolClosed, got %v", err)
	}

	if err := pool.Release(page); err != nil {
		t.Fatalf("Release error: %v", err)
	}

	select {
	case <-page.Ctx.Done():
	default:
		t.Fatalf("expected page context to be canceled after release to closed pool")
	}
}

func TestDoubleReturn(t *testing.T) {
	t.Parallel()

	pool, err := NewPool(context.Background(), 1)
	if err != nil {
		t.Fatalf("NewPool error: %v", err)
	}
	t.Cleanup(pool.Close)

	page, err := pool.AcquireBlank(context.Background())
	if err != nil {
		t.Fatalf("Acquire error: %v", err)
	}

	if err := pool.Release(page); err != nil {
		t.Fatalf("Release error: %v", err)
	}

	if err := pool.Release(page); !errors.Is(err, ErrDoubleReturn) {
		t.Fatalf("expected ErrDoubleReturn, got %v", err)
	}

	select {
	case <-page.Ctx.Done():
	default:
		t.Fatalf("expected page context to be canceled after double return")
	}
}

func TestAccessExampleDotCom(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser integration test in short mode")
	}

	pool, err := NewPool(context.Background(), 1)
	if err != nil {
		t.Fatalf("NewPool error: %v", err)
	}
	defer pool.Close()
	t.Log("pool created")

	page, err := pool.AcquireBlank(context.Background())
	if err != nil {
		t.Fatalf("Acquire error: %v", err)
	}
	t.Log("page acquired")
	defer func() {
		if err := pool.Release(page); err != nil {
			t.Fatalf("Release error: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(page.Ctx, 15*time.Second)
	defer cancel()
	t.Log("navigating to https://example.com")

	var title string
	if err := chromedp.Run(
		ctx,
		chromedp.Navigate("https://example.com"),
		chromedp.Title(&title),
	); err != nil {
		t.Fatalf("chromedp run error: %v", err)
	}
	t.Logf("page title: %q", title)

	if title != "Example Domain" {
		t.Fatalf("unexpected title: %q", title)
	}
}
