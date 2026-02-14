package task

import (
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
}

// TaskQueue stores tasks in FIFO order.
type TaskQueue struct {
	mu    sync.Mutex
	items []Task
}

// NewQueue returns an empty task queue.
func NewQueue() *TaskQueue {
	return &TaskQueue{items: make([]Task, 0)}
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
