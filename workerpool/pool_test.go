package workerpool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kolosys/ion/shared"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		size      int
		queueSize int
		opts      []Option
		wantSize  int
	}{
		{
			name:      "default size",
			size:      0,
			queueSize: 10,
			wantSize:  8, // GOMAXPROCS typically returns number of CPU cores
		},
		{
			name:      "custom size",
			size:      2,
			queueSize: 5,
			wantSize:  2,
		},
		{
			name:      "with options",
			size:      3,
			queueSize: 8,
			opts: []Option{
				WithName("test-pool"),
				WithDrainTimeout(10 * time.Second),
			},
			wantSize: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := New(tt.size, tt.queueSize, tt.opts...)
			defer pool.Close(context.Background())

			if pool.size != tt.wantSize {
				t.Errorf("expected size %d, got %d", tt.wantSize, pool.size)
			}
			if pool.queueSize != tt.queueSize {
				t.Errorf("expected queueSize %d, got %d", tt.queueSize, pool.queueSize)
			}
		})
	}
}

func TestSubmit(t *testing.T) {
	t.Run("successful submission", func(t *testing.T) {
		pool := New(2, 5, WithName("test-pool"))
		defer pool.Close(context.Background())

		var executed atomic.Bool
		task := func(ctx context.Context) error {
			executed.Store(true)
			return nil
		}

		err := pool.Submit(context.Background(), task)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Wait for task execution
		time.Sleep(100 * time.Millisecond)

		if !executed.Load() {
			t.Error("task was not executed")
		}
	})

	t.Run("nil task", func(t *testing.T) {
		pool := New(1, 1)
		defer pool.Close(context.Background())

		err := pool.Submit(context.Background(), nil)
		if err == nil || err.Error() != "ion: nil task" {
			t.Errorf("expected nil task error, got %v", err)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		pool := New(1, 0) // zero queue to force blocking
		defer pool.Close(context.Background())

		// Submit a long-running task to fill the single worker
		longTask := func(ctx context.Context) error {
			select {
			case <-time.After(1 * time.Second):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Submit the blocking task
		go func() { _ = pool.Submit(context.Background(), longTask) }()
		time.Sleep(10 * time.Millisecond) // Let it start

		// Try to submit another task with canceled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		task := func(ctx context.Context) error { return nil }
		err := pool.Submit(ctx, task)

		if err == nil || err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})

	t.Run("closed pool", func(t *testing.T) {
		pool := New(1, 1)
		pool.Close(context.Background())

		task := func(ctx context.Context) error { return nil }
		err := pool.Submit(context.Background(), task)

		if err == nil {
			t.Error("expected error when submitting to closed pool")
		}

		var poolErr *shared.PoolError
		if !errors.As(err, &poolErr) {
			t.Errorf("expected PoolError, got %T", err)
		}
	})
}

func TestTrySubmit(t *testing.T) {
	t.Run("successful submission", func(t *testing.T) {
		pool := New(2, 5)
		defer pool.Close(context.Background())

		var executed atomic.Bool
		task := func(ctx context.Context) error {
			executed.Store(true)
			return nil
		}

		err := pool.TrySubmit(task)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Wait for task execution
		time.Sleep(100 * time.Millisecond)

		if !executed.Load() {
			t.Error("task was not executed")
		}
	})

	t.Run("queue full", func(t *testing.T) {
		pool := New(1, 1)
		defer pool.Close(context.Background())

		// Fill the queue
		blockingTask := func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		}

		// Submit tasks to fill worker + queue
		_ = pool.TrySubmit(blockingTask) // fills worker
		_ = pool.TrySubmit(blockingTask) // fills queue

		// This should fail
		quickTask := func(ctx context.Context) error { return nil }
		err := pool.TrySubmit(quickTask)

		if err == nil {
			t.Error("expected error when queue is full")
		}

		var poolErr *shared.PoolError
		if !errors.As(err, &poolErr) {
			t.Errorf("expected PoolError, got %T", err)
		}
	})
}

func TestPoolLifecycle(t *testing.T) {
	t.Run("close waits for running tasks", func(t *testing.T) {
		pool := New(1, 0)

		var taskStarted, taskFinished atomic.Bool
		task := func(ctx context.Context) error {
			taskStarted.Store(true)
			time.Sleep(200 * time.Millisecond)
			taskFinished.Store(true)
			return nil
		}

		// Submit task
		go func() { _ = pool.Submit(context.Background(), task) }()

		// Wait for task to start
		for !taskStarted.Load() {
			time.Sleep(10 * time.Millisecond)
		}

		// Close the pool
		start := time.Now()
		err := pool.Close(context.Background())
		duration := time.Since(start)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if !taskFinished.Load() {
			t.Error("task should have finished before close returned")
		}

		if duration < 100*time.Millisecond {
			t.Error("close returned too quickly, should have waited for task")
		}
	})

	t.Run("drain waits for queue to empty", func(t *testing.T) {
		pool := New(1, 2)

		var completedTasks atomic.Int64
		task := func(ctx context.Context) error {
			time.Sleep(50 * time.Millisecond)
			completedTasks.Add(1)
			return nil
		}

		// Submit multiple tasks
		_ = pool.Submit(context.Background(), task) // starts immediately
		_ = pool.Submit(context.Background(), task) // queued
		_ = pool.Submit(context.Background(), task) // queued

		// Start draining
		start := time.Now()
		err := pool.Drain(context.Background())
		duration := time.Since(start)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if completedTasks.Load() != 3 {
			t.Errorf("expected 3 completed tasks, got %d", completedTasks.Load())
		}

		// Should have taken time to process all tasks
		if duration < 100*time.Millisecond {
			t.Error("drain returned too quickly, should have waited for all tasks")
		}
	})
}

func TestMetrics(t *testing.T) {
	pool := New(2, 3, WithName("metrics-test"))
	defer pool.Close(context.Background())

	metrics := pool.Metrics()
	if metrics.Size != 2 {
		t.Errorf("expected size 2, got %d", metrics.Size)
	}

	// Submit a task and check metrics
	var wg sync.WaitGroup
	wg.Add(1)
	task := func(ctx context.Context) error {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond)
		return nil
	}

	pool.Submit(context.Background(), task)

	// Check that queued/running counts change
	time.Sleep(10 * time.Millisecond) // Let task start
	metrics = pool.Metrics()

	// Either running or completed should be > 0
	if metrics.Running == 0 && metrics.Completed == 0 {
		t.Error("expected either running or completed tasks")
	}

	wg.Wait() // Wait for task completion

	// Check final metrics
	time.Sleep(10 * time.Millisecond)
	metrics = pool.Metrics()
	if metrics.Completed == 0 {
		t.Error("expected at least one completed task")
	}
}

func TestTaskPanicRecovery(t *testing.T) {
	var panicValue any
	var panicMutex sync.Mutex
	var panicReceived bool

	pool := New(1, 1, WithPanicRecovery(func(r any) {
		panicMutex.Lock()
		panicValue = r
		panicReceived = true
		panicMutex.Unlock()
	}))
	defer pool.Close(context.Background())

	task := func(ctx context.Context) error {
		panic("test panic")
	}

	err := pool.Submit(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Wait for panic recovery with proper synchronization
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("timeout waiting for panic recovery")
		case <-ticker.C:
			panicMutex.Lock()
			if panicReceived {
				expected := "test panic"
				if panicValue != expected {
					t.Errorf("expected panic value %q, got %v", expected, panicValue)
				}
				panicMutex.Unlock()
				goto checkMetrics
			}
			panicMutex.Unlock()
		}
	}

checkMetrics:
	metrics := pool.Metrics()
	if metrics.Panicked == 0 {
		t.Error("expected panic count > 0")
	}
}
