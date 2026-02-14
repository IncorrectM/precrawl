package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/IncorrectM/precrawl/internal/browser"
	"github.com/IncorrectM/precrawl/internal/prerender"
	"github.com/IncorrectM/precrawl/internal/task"
)

const (
	defaultAddr     = ":8080"
	defaultSelector = "body"
	selectorHeader  = "X-Render-Selector"
	waitHeader      = "X-Render-Wait"
	waitMsHeader    = "X-Render-Wait-Ms"
)

var (
	ErrInvalidConfig        = errors.New("invalid server config")
	ErrInvalidBaseTargetURL = errors.New("invalid base target url")
)

type Config struct {
	Addr            string
	BaseTargetURL   string
	DefaultSelector string
	Queue           *task.TaskQueue
	Pool            *browser.Pool
	WorkerCount     int
}

func Run(ctx context.Context, cfg Config) error {
	if cfg.Queue == nil || cfg.Pool == nil {
		return ErrInvalidConfig
	}
	baseURL, err := parseBaseTargetURL(cfg.BaseTargetURL)
	if err != nil {
		return err
	}
	if cfg.Addr == "" {
		cfg.Addr = defaultAddr
	}
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 1
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(cfg.DefaultSelector) == "" {
		cfg.DefaultSelector = defaultSelector
	}

	log.Printf("server starting addr=%s baseTargetURL=%s workers=%d defaultSelector=%s", cfg.Addr, baseURL.String(), cfg.WorkerCount, cfg.DefaultSelector)

	workerCtx, cancelWorkers := context.WithCancel(ctx)
	defer cancelWorkers()

	for i := 0; i < cfg.WorkerCount; i++ {
		go workerLoop(workerCtx, i+1, cfg.Queue, cfg.Pool)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleRender(w, r, cfg.Queue, baseURL, cfg.DefaultSelector)
	})

	server := &http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.Printf("server shutting down: %v", ctx.Err())
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
		err := <-errCh
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case err := <-errCh:
		log.Printf("server stopped: %v", err)
		return err
	}
}

func handleRender(w http.ResponseWriter, r *http.Request, queue *task.TaskQueue, baseURL *url.URL, defaultSelectorValue string) {
	start := time.Now()
	if r.Method != http.MethodGet {
		log.Printf("reject method=%s path=%s", r.Method, r.URL.Path)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	targetURL, err := buildTargetURL(baseURL, r.URL)
	if err != nil {
		log.Printf("invalid target url path=%s query=%s err=%v", r.URL.Path, r.URL.RawQuery, err)
		http.Error(w, fmt.Sprintf("invalid target url: %v", err), http.StatusBadRequest)
		return
	}

	selector := strings.TrimSpace(r.Header.Get(selectorHeader))
	if selector == "" {
		selector = defaultSelectorValue
	}

	wait, err := parseWaitHeaders(r)
	if err != nil {
		log.Printf("invalid wait header path=%s err=%v", r.URL.Path, err)
		http.Error(w, fmt.Sprintf("invalid wait: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("request path=%s query=%s target=%s selector=%s wait=%s remote=%s", r.URL.Path, r.URL.RawQuery, targetURL, selector, wait, r.RemoteAddr)

	resultCh := make(chan task.Result, 1)
	taskItem := task.Task{
		TargetURL:     targetURL,
		Wait:          wait,
		QuerySelector: selector,
		ResultCh:      resultCh,
	}

	if err := queue.Enqueue(taskItem); err != nil {
		log.Printf("enqueue failed target=%s err=%v", targetURL, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	select {
	case result := <-resultCh:
		if result.Err != nil {
			log.Printf("render failed target=%s err=%v duration=%s", targetURL, result.Err, time.Since(start))
			http.Error(w, result.Err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(result.HTML))
		log.Printf("render ok target=%s bytes=%d duration=%s", targetURL, len(result.HTML), time.Since(start))
	case <-r.Context().Done():
		log.Printf("request canceled target=%s err=%v duration=%s", targetURL, r.Context().Err(), time.Since(start))
		http.Error(w, "request canceled", http.StatusRequestTimeout)
	}
}

func parseWaitHeaders(r *http.Request) (time.Duration, error) {
	if waitValue := strings.TrimSpace(r.Header.Get(waitHeader)); waitValue != "" {
		return time.ParseDuration(waitValue)
	}
	if waitMs := strings.TrimSpace(r.Header.Get(waitMsHeader)); waitMs != "" {
		value, err := strconv.ParseInt(waitMs, 10, 64)
		if err != nil {
			return 0, err
		}
		return time.Duration(value) * time.Millisecond, nil
	}
	return 0, nil
}

func parseBaseTargetURL(raw string) (*url.URL, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, ErrInvalidBaseTargetURL
	}
	baseURL, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, ErrInvalidBaseTargetURL
	}
	return baseURL, nil
}

func buildTargetURL(baseURL *url.URL, reqURL *url.URL) (string, error) {
	if baseURL == nil || reqURL == nil {
		return "", ErrInvalidBaseTargetURL
	}
	ref := &url.URL{Path: reqURL.Path, RawQuery: reqURL.RawQuery}
	return baseURL.ResolveReference(ref).String(), nil
}

func workerLoop(ctx context.Context, id int, queue *task.TaskQueue, pool *browser.Pool) {
	log.Printf("worker started id=%d", id)
	for {
		item, err := queue.WaitDequeue(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				log.Printf("worker stopped id=%d err=%v", id, err)
				return
			}
			log.Printf("worker dequeue error id=%d err=%v", id, err)
			continue
		}

		start := time.Now()
		html, renderErr := prerender.RenderUntil(context.Background(), pool, item.TargetURL, item.Wait, item.QuerySelector)
		if item.ResultCh != nil {
			item.ResultCh <- task.Result{HTML: html, Err: renderErr}
			close(item.ResultCh)
		}
		if renderErr != nil {
			log.Printf("worker render failed id=%d target=%s err=%v duration=%s", id, item.TargetURL, renderErr, time.Since(start))
			continue
		}
		log.Printf("worker render ok id=%d target=%s bytes=%d duration=%s", id, item.TargetURL, len(html), time.Since(start))
	}
}
