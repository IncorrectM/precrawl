package prerender

import (
	"context"
	"errors"
	"time"

	"github.com/chromedp/chromedp"

	"github.com/IncorrectM/precrawl/internal/browser"
)

var (
	ErrInvalidPool         = errors.New("invalid browser pool")
	ErrEmptyURL            = errors.New("target url is empty")
	ErrNegativeWait        = errors.New("wait duration must be non-negative")
	ErrNegativeWaitTimeout = errors.New("wait timeout must be non-negative")
	ErrWaitTimeout         = errors.New("wait timeout exceeded")
)

// Render navigates to a URL, waits, and returns the full HTML document.
func RenderUntil(
	ctx context.Context,
	pool *browser.Pool,
	targetURL string,
	wait time.Duration,
	querySelector string,
	waitTimeout time.Duration,
) (html string, err error) {
	// validate inputs
	if pool == nil {
		return "", ErrInvalidPool
	}
	if targetURL == "" {
		return "", ErrEmptyURL
	}
	if wait < 0 {
		return "", ErrNegativeWait
	}
	if waitTimeout < 0 {
		return "", ErrNegativeWaitTimeout
	}
	if ctx == nil {
		ctx = context.Background()
	}

	// acquire a browser page from the pool
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
		// gracefully handle context cancellation to ensure the browser page is released
		case <-ctx.Done():
			cancel()
		case <-runCtx.Done():
		}
	}()

	if err := chromedp.Run(
		runCtx,
		chromedp.Navigate(targetURL),
	); err != nil {
		return "", err
	}

	// wait for the specified element to become visible
	waitTimedOut := false
	if waitTimeout > 0 {
		waitCtx, waitCancel := context.WithTimeout(runCtx, waitTimeout)
		waitErr := chromedp.Run(waitCtx, chromedp.WaitVisible(querySelector, chromedp.ByQuery))
		waitCancel()
		if waitErr != nil {
			if errors.Is(waitErr, context.DeadlineExceeded) && errors.Is(waitCtx.Err(), context.DeadlineExceeded) {
				waitTimedOut = true
			} else {
				return "", waitErr
			}
		}
	} else {
		if err := chromedp.Run(runCtx, chromedp.WaitVisible(querySelector, chromedp.ByQuery)); err != nil {
			return "", err
		}
	}

	if err := chromedp.Run(
		runCtx,
		chromedp.Sleep(wait), // ensure any additional content has time to load
		chromedp.OuterHTML("html", &html, chromedp.ByQuery),
	); err != nil {
		return "", err
	}

	if waitTimedOut {
		return html, ErrWaitTimeout
	}

	return html, nil
}

func Render(
	ctx context.Context,
	pool *browser.Pool,
	targetURL string,
	wait time.Duration,
) (string, error) {
	return RenderUntil(ctx, pool, targetURL, wait, "body", 0)
}
