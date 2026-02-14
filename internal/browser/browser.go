package browser

import (
	"context"
	"errors"
	"sync"

	"github.com/chromedp/chromedp"
)

var (
	ErrPoolClosed   = errors.New("browser pool is closed")
	ErrInvalidSize  = errors.New("browser pool size must be positive")
	ErrInvalidPage  = errors.New("invalid page")
	ErrDoubleReturn = errors.New("page returned more than once")
)

const BlankURL = "about:blank"

// Page represents a single tab context managed by the pool.
type Page struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}

// Pool manages a fixed number of chromedp page contexts.
type Pool struct {
	size        int
	allocCtx    context.Context
	allocCancel context.CancelFunc
	pages       chan *Page
	mu          sync.Mutex
	closed      bool
}

// NewPool creates a pool with a shared browser allocator and N page contexts.
func NewPool(parent context.Context, size int, opts ...chromedp.ExecAllocatorOption) (*Pool, error) {
	if size <= 0 {
		return nil, ErrInvalidSize
	}

	allocatorOpts := append([]chromedp.ExecAllocatorOption{}, chromedp.DefaultExecAllocatorOptions[:]...)
	allocatorOpts = append(allocatorOpts, opts...)

	allocCtx, allocCancel := chromedp.NewExecAllocator(parent, allocatorOpts...)
	pages := make(chan *Page, size)

	for range size {
		ctx, cancel := chromedp.NewContext(allocCtx)
		pages <- &Page{Ctx: ctx, Cancel: cancel}
	}

	return &Pool{
		size:        size,
		allocCtx:    allocCtx,
		allocCancel: allocCancel,
		pages:       pages,
	}, nil
}

// Acquire waits for a free page or returns when ctx is done.
func (p *Pool) Acquire(ctx context.Context, initialURL string) (*Page, error) {
	p.mu.Lock()
	closed := p.closed
	p.mu.Unlock()

	if closed {
		return nil, ErrPoolClosed
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case page := <-p.pages:
		// navigate to initial URL
		chromedp.Run(page.Ctx, chromedp.Navigate(initialURL))
		return page, nil
	}
}

func (p *Pool) AcquireBlank(ctx context.Context) (*Page, error) {
	return p.Acquire(ctx, BlankURL)
}

// Release returns a page to the pool.
func (p *Pool) Release(page *Page) error {
	if page == nil {
		return ErrInvalidPage
	}

	p.mu.Lock()
	closed := p.closed
	p.mu.Unlock()

	if closed {
		page.Cancel()
		return nil
	}

	select {
	case p.pages <- page:
		return nil
	default:
		page.Cancel()
		return ErrDoubleReturn
	}
}

// Close closes the pool and cancels all idle pages and the allocator.
func (p *Pool) Close() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	p.mu.Unlock()

	for {
		select {
		case page := <-p.pages:
			page.Cancel()
		default:
			p.allocCancel()
			return
		}
	}
}

// Size returns the configured pool size.
func (p *Pool) Size() int {
	return p.size
}

// Available returns the number of idle pages in the pool.
func (p *Pool) Available() int {
	return len(p.pages)
}
