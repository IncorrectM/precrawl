package task

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrEmptyQueue   = errors.New("task queue is empty")
	ErrEmptyURL     = errors.New("target url is empty")
	ErrEmptyQuery   = errors.New("query selector is empty")
	ErrNegativeWait = errors.New("wait duration must be non-negative")
)

// Task holds parameters required by prerender.RenderUntil.
type Task struct {
	TargetURL     string
	Wait          time.Duration
	QuerySelector string
	ResultCh      chan Result
}

// Result represents the outcome of executing a task.
type Result struct {
	HTML string
	Err  error
}

// TaskQueue stores tasks in FIFO order.
type TaskQueue struct {
	mu       sync.Mutex
	items    []Task
	notEmpty *sync.Cond
}

// NewQueue returns an empty task queue.
func NewQueue() *TaskQueue {
	queue := &TaskQueue{items: make([]Task, 0)}
	queue.notEmpty = sync.NewCond(&queue.mu)
	return queue
}

// Enqueue appends a task to the queue.
func (q *TaskQueue) Enqueue(task Task) error {
	if task.TargetURL == "" {
		return ErrEmptyURL
	}
	if task.QuerySelector == "" {
		return ErrEmptyQuery
	}
	if task.Wait < 0 {
		return ErrNegativeWait
	}

	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, task)
	q.notEmpty.Signal()
	return nil
}

// Dequeue removes and returns the next task.
func (q *TaskQueue) Dequeue() (Task, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return Task{}, ErrEmptyQueue
	}

	item := q.items[0]
	q.items[0] = Task{}
	q.items = q.items[1:]
	return item, nil
}

// WaitDequeue blocks until a task is available or ctx is done.
func (q *TaskQueue) WaitDequeue(ctx context.Context) (Task, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			q.mu.Lock()
			q.notEmpty.Broadcast()
			q.mu.Unlock()
		case <-done:
		}
	}()
	defer close(done)

	for len(q.items) == 0 {
		if err := ctx.Err(); err != nil {
			return Task{}, err
		}
		q.notEmpty.Wait()
	}

	item := q.items[0]
	q.items[0] = Task{}
	q.items = q.items[1:]
	return item, nil
}

// Len returns the number of queued tasks.
func (q *TaskQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

// Peek returns the next task without removing it.
func (q *TaskQueue) Peek() (Task, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return Task{}, ErrEmptyQueue
	}

	return q.items[0], nil
}
