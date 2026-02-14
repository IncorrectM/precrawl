package task

import (
	"errors"
	"testing"
	"time"
)

func TestQueueEnqueueValidation(t *testing.T) {
	t.Parallel()

	queue := NewQueue()

	if err := queue.Enqueue(Task{TargetURL: "", QuerySelector: "body"}); !errors.Is(err, ErrEmptyURL) {
		t.Fatalf("expected ErrEmptyURL, got %v", err)
	}

	if err := queue.Enqueue(Task{TargetURL: "https://example.com", QuerySelector: ""}); !errors.Is(err, ErrEmptyQuery) {
		t.Fatalf("expected ErrEmptyQuery, got %v", err)
	}

	if err := queue.Enqueue(Task{TargetURL: "https://example.com", QuerySelector: "body", Wait: -time.Millisecond}); !errors.Is(err, ErrNegativeWait) {
		t.Fatalf("expected ErrNegativeWait, got %v", err)
	}
}

func TestQueueFIFOAndLen(t *testing.T) {
	t.Parallel()

	queue := NewQueue()

	first := Task{TargetURL: "https://a.example", QuerySelector: "body", Wait: 10 * time.Millisecond}
	second := Task{TargetURL: "https://b.example", QuerySelector: "#main", Wait: 20 * time.Millisecond}
	third := Task{TargetURL: "https://c.example", QuerySelector: ".content", Wait: 30 * time.Millisecond}

	if err := queue.Enqueue(first); err != nil {
		t.Fatalf("enqueue first error: %v", err)
	}
	if err := queue.Enqueue(second); err != nil {
		t.Fatalf("enqueue second error: %v", err)
	}
	if err := queue.Enqueue(third); err != nil {
		t.Fatalf("enqueue third error: %v", err)
	}

	if got := queue.Len(); got != 3 {
		t.Fatalf("expected length 3, got %d", got)
	}

	peek, err := queue.Peek()
	if err != nil {
		t.Fatalf("peek error: %v", err)
	}
	if peek != first {
		t.Fatalf("expected peek to return first task, got %+v", peek)
	}

	item, err := queue.Dequeue()
	if err != nil {
		t.Fatalf("dequeue first error: %v", err)
	}
	if item != first {
		t.Fatalf("expected first task, got %+v", item)
	}

	item, err = queue.Dequeue()
	if err != nil {
		t.Fatalf("dequeue second error: %v", err)
	}
	if item != second {
		t.Fatalf("expected second task, got %+v", item)
	}

	item, err = queue.Dequeue()
	if err != nil {
		t.Fatalf("dequeue third error: %v", err)
	}
	if item != third {
		t.Fatalf("expected third task, got %+v", item)
	}

	if got := queue.Len(); got != 0 {
		t.Fatalf("expected length 0, got %d", got)
	}
}

func TestQueueEmptyErrors(t *testing.T) {
	t.Parallel()

	queue := NewQueue()

	if _, err := queue.Dequeue(); !errors.Is(err, ErrEmptyQueue) {
		t.Fatalf("expected ErrEmptyQueue from dequeue, got %v", err)
	}

	if _, err := queue.Peek(); !errors.Is(err, ErrEmptyQueue) {
		t.Fatalf("expected ErrEmptyQueue from peek, got %v", err)
	}
}
