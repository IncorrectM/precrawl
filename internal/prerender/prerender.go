package prerender

import (
	"context"
	"errors"
	"time"

	"github.com/chromedp/chromedp"

	"github.com/IncorrectM/precrawl/internal/browser"
)

var (
	ErrInvalidPool  = errors.New("invalid browser pool")
	ErrEmptyURL     = errors.New("target url is empty")
	ErrNegativeWait = errors.New("wait duration must be non-negative")
)

// Render navigates to a URL, waits, and returns the full HTML document.
func RenderUntil(
	ctx context.Context,
	pool *browser.Pool,
	targetURL string,
	wait time.Duration,
	querySelector string,
) (html string, err error) {
	if pool == nil {
		return "", ErrInvalidPool
	}
	if targetURL == "" {
		return "", ErrEmptyURL
	}
	if wait < 0 {
		return "", ErrNegativeWait
	}
	if ctx == nil {
		ctx = context.Background()
	}

	page, err := pool.AcquireBlank(ctx)
	if err != nil {
		return "", err
	}
	defer func() {
		releaseErr := pool.Release(page)
		if err == nil && releaseErr != nil {
			err = releaseErr
		}
	}()

	runCtx, cancel := context.WithCancel(page.Ctx)
	defer cancel()

	go func() {
		select {
		case <-ctx.Done():
			cancel()
		case <-runCtx.Done():
		}
	}()

	if err := chromedp.Run(
		runCtx,
		chromedp.Navigate(targetURL),
		chromedp.WaitVisible(querySelector, chromedp.ByQuery),
		chromedp.Sleep(wait),
		chromedp.OuterHTML("html", &html, chromedp.ByQuery),
	); err != nil {
		return "", err
	}

	return html, nil
}

func Render(
	ctx context.Context,
	pool *browser.Pool,
	targetURL string,
	wait time.Duration,
) (string, error) {
	return RenderUntil(ctx, pool, targetURL, wait, "body")
}
