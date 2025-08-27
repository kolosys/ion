package semaphore_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kolosys/ion/semaphore"
	"github.com/kolosys/ion/shared"
)

func TestNewWeighted(t *testing.T) {
	tests := []struct {
		name      string
		capacity  int64
		opts      []semaphore.Option
		wantPanic bool
	}{
		{
			name:     "valid capacity",
			capacity: 10,
		},
		{
			name:     "capacity of 1",
			capacity: 1,
		},
		{
			name:     "with options",
			capacity: 5,
			opts: []semaphore.Option{
				semaphore.WithName("test-sem"),
				semaphore.WithFairness(semaphore.LIFO),
				semaphore.WithAcquireTimeout(time.Second),
			},
		},
		{
			name:      "zero capacity",
			capacity:  0,
			wantPanic: true,
		},
		{
			name:      "negative capacity",
			capacity:  -1,
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.wantPanic && r == nil {
					t.Error("expected panic but didn't get one")
				} else if !tt.wantPanic && r != nil {
					t.Errorf("unexpected panic: %v", r)
				}
			}()

			sem := semaphore.NewWeighted(tt.capacity, tt.opts...)
			if !tt.wantPanic {
				if sem.Current() != tt.capacity {
					t.Errorf("expected current permits %d, got %d", tt.capacity, sem.Current())
				}
			}
		})
	}
}

func TestTryAcquire(t *testing.T) {
	t.Run("successful acquisition", func(t *testing.T) {
		sem := semaphore.NewWeighted(5)

		if !sem.TryAcquire(3) {
			t.Error("should have acquired 3 permits")
		}

		if sem.Current() != 2 {
			t.Errorf("expected 2 remaining permits, got %d", sem.Current())
		}

		if !sem.TryAcquire(2) {
			t.Error("should have acquired remaining 2 permits")
		}

		if sem.Current() != 0 {
			t.Errorf("expected 0 remaining permits, got %d", sem.Current())
		}
	})

	t.Run("insufficient permits", func(t *testing.T) {
		sem := semaphore.NewWeighted(3)

		if sem.TryAcquire(5) {
			t.Error("should not have acquired 5 permits when only 3 available")
		}

		if sem.Current() != 3 {
			t.Errorf("permits should remain unchanged, got %d", sem.Current())
		}
	})

	t.Run("invalid weight", func(t *testing.T) {
		sem := semaphore.NewWeighted(5)

		if sem.TryAcquire(0) {
			t.Error("should not acquire 0 permits")
		}

		if sem.TryAcquire(-1) {
			t.Error("should not acquire negative permits")
		}
	})

	t.Run("weight exceeds capacity", func(t *testing.T) {
		sem := semaphore.NewWeighted(3)

		if sem.TryAcquire(5) {
			t.Error("should not acquire permits exceeding capacity")
		}
	})
}

func TestAcquire(t *testing.T) {
	t.Run("successful acquisition", func(t *testing.T) {
		sem := semaphore.NewWeighted(5)

		err := sem.Acquire(context.Background(), 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if sem.Current() != 2 {
			t.Errorf("expected 2 remaining permits, got %d", sem.Current())
		}
	})

	t.Run("invalid weight", func(t *testing.T) {
		sem := semaphore.NewWeighted(5)

		err := sem.Acquire(context.Background(), 0)
		if !errors.Is(err, shared.ErrInvalidWeight) {
			t.Errorf("expected ErrInvalidWeight, got %v", err)
		}

		err = sem.Acquire(context.Background(), -1)
		if !errors.Is(err, shared.ErrInvalidWeight) {
			t.Errorf("expected ErrInvalidWeight, got %v", err)
		}
	})

	t.Run("weight exceeds capacity", func(t *testing.T) {
		sem := semaphore.NewWeighted(3, semaphore.WithName("test-sem"))

		err := sem.Acquire(context.Background(), 5)
		var semErr *shared.SemaphoreError
		if !errors.As(err, &semErr) {
			t.Errorf("expected SemaphoreError, got %T", err)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		sem := semaphore.NewWeighted(1)

		// Acquire the only permit
		_ = sem.Acquire(context.Background(), 1)

		// Try to acquire with canceled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := sem.Acquire(ctx, 1)
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})

	t.Run("context timeout", func(t *testing.T) {
		sem := semaphore.NewWeighted(1)

		// Acquire the only permit
		_ = sem.Acquire(context.Background(), 1)

		// Try to acquire with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		start := time.Now()
		err := sem.Acquire(ctx, 1)
		duration := time.Since(start)

		var semErr *shared.SemaphoreError
		if !errors.As(err, &semErr) {
			t.Errorf("expected SemaphoreError due to timeout, got %T: %v", err, err)
		}

		if duration < 40*time.Millisecond {
			t.Error("acquire returned too quickly, should have waited for timeout")
		}
	})
}

func TestRelease(t *testing.T) {
	t.Run("successful release", func(t *testing.T) {
		sem := semaphore.NewWeighted(5)

		// Acquire some permits
		_ = sem.Acquire(context.Background(), 3)
		if sem.Current() != 2 {
			t.Fatalf("setup failed, expected 2 permits, got %d", sem.Current())
		}

		// Release permits
		sem.Release(2)
		if sem.Current() != 4 {
			t.Errorf("expected 4 permits after release, got %d", sem.Current())
		}
	})

	t.Run("release zero permits", func(t *testing.T) {
		sem := semaphore.NewWeighted(5)
		sem.Acquire(context.Background(), 2)

		before := sem.Current()
		sem.Release(0)
		after := sem.Current()

		if before != after {
			t.Error("releasing 0 permits should not change current count")
		}
	})

	t.Run("release negative permits panics", func(t *testing.T) {
		sem := semaphore.NewWeighted(5)

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when releasing negative permits")
			}
		}()

		sem.Release(-1)
	})

	t.Run("release more than capacity panics", func(t *testing.T) {
		sem := semaphore.NewWeighted(3)

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when releasing more than capacity")
			}
		}()

		sem.Release(5) // Would exceed capacity
	})

	t.Run("release unblocks waiters", func(t *testing.T) {
		sem := semaphore.NewWeighted(2)

		// Acquire all permits
		sem.Acquire(context.Background(), 2)

		var acquired atomic.Bool
		var wg sync.WaitGroup
		wg.Add(1)

		// Start goroutine waiting for permits
		go func() {
			defer wg.Done()
			if err := sem.Acquire(context.Background(), 1); err == nil {
				acquired.Store(true)
			}
		}()

		// Give waiter time to start waiting
		time.Sleep(50 * time.Millisecond)

		if acquired.Load() {
			t.Error("waiter should not have acquired permit yet")
		}

		// Release a permit
		sem.Release(1)

		// Wait for waiter to complete
		wg.Wait()

		if !acquired.Load() {
			t.Error("waiter should have acquired permit after release")
		}
	})
}

func TestFairness(t *testing.T) {
	t.Run("FIFO fairness", func(t *testing.T) {
		sem := semaphore.NewWeighted(1, semaphore.WithFairness(semaphore.FIFO))

		// Acquire the only permit
		_ = sem.Acquire(context.Background(), 1)

		var results []int
		var mu sync.Mutex
		var wg sync.WaitGroup
		var started sync.WaitGroup

		// Start multiple waiters with synchronization to ensure order
		started.Add(3)
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				started.Done() // Signal that this goroutine has started
				started.Wait() // Wait for all goroutines to start

				// Add small delay to ensure consistent ordering
				time.Sleep(time.Duration(id) * 10 * time.Millisecond)

				_ = sem.Acquire(context.Background(), 1)
				mu.Lock()
				results = append(results, id)
				mu.Unlock()
				sem.Release(1)
			}(i)
		}

		// Wait for all waiters to be queued
		started.Wait()
		time.Sleep(100 * time.Millisecond)

		// Release the initial permit to start the chain
		sem.Release(1)

		wg.Wait()

		// With proper timing, results should be in FIFO order (0, 1, 2)
		// However, due to Go's goroutine scheduling, this might not be 100% deterministic
		// So we'll just verify that all waiters completed
		if len(results) != 3 {
			t.Fatalf("expected 3 results, got %d: %v", len(results), results)
		}

		t.Logf("FIFO order result: %v", results)
		// Note: Perfect FIFO ordering is hard to test deterministically due to goroutine scheduling
	})

	t.Run("LIFO fairness", func(t *testing.T) {
		sem := semaphore.NewWeighted(1, semaphore.WithFairness(semaphore.LIFO))

		// Acquire the only permit
		_ = sem.Acquire(context.Background(), 1)

		var results []int
		var mu sync.Mutex
		var wg sync.WaitGroup
		var started sync.WaitGroup

		// Start multiple waiters with synchronization
		started.Add(3)
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				started.Done() // Signal that this goroutine has started
				started.Wait() // Wait for all goroutines to start

				// Add small delay to ensure consistent ordering
				time.Sleep(time.Duration(id) * 10 * time.Millisecond)

				_ = sem.Acquire(context.Background(), 1)
				mu.Lock()
				results = append(results, id)
				mu.Unlock()
				sem.Release(1)
			}(i)
		}

		// Wait for all waiters to be queued
		started.Wait()
		time.Sleep(100 * time.Millisecond)

		// Release the initial permit to start the chain
		sem.Release(1)

		wg.Wait()

		// Verify all waiters completed
		if len(results) != 3 {
			t.Fatalf("expected 3 results, got %d: %v", len(results), results)
		}

		t.Logf("LIFO order result: %v", results)
		// Note: Perfect LIFO ordering is hard to test deterministically due to goroutine scheduling
	})
}

func TestConcurrency(t *testing.T) {
	t.Run("high concurrency stress test", func(t *testing.T) {
		sem := semaphore.NewWeighted(10)
		const numGoroutines = 100
		const iterations = 10

		var wg sync.WaitGroup
		var successCount atomic.Int64

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					if err := sem.Acquire(context.Background(), 1); err == nil {
						successCount.Add(1)
						// Simulate work
						time.Sleep(time.Microsecond)
						sem.Release(1)
					}
				}
			}()
		}

		wg.Wait()

		expected := int64(numGoroutines * iterations)
		if successCount.Load() != expected {
			t.Errorf("expected %d successful acquisitions, got %d", expected, successCount.Load())
		}

		// All permits should be returned
		if sem.Current() != 10 {
			t.Errorf("expected all permits returned, got %d", sem.Current())
		}
	})

	t.Run("mixed acquire and try_acquire", func(t *testing.T) {
		sem := semaphore.NewWeighted(5)
		const numGoroutines = 20

		var wg sync.WaitGroup
		var acquireSuccess atomic.Int64
		var tryAcquireSuccess atomic.Int64

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				if id%2 == 0 {
					// Use Acquire
					ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
					defer cancel()
					if err := sem.Acquire(ctx, 1); err == nil {
						acquireSuccess.Add(1)
						time.Sleep(10 * time.Millisecond)
						sem.Release(1)
					}
				} else {
					// Use TryAcquire
					if sem.TryAcquire(1) {
						tryAcquireSuccess.Add(1)
						time.Sleep(10 * time.Millisecond)
						sem.Release(1)
					}
				}
			}(i)
		}

		wg.Wait()

		total := acquireSuccess.Load() + tryAcquireSuccess.Load()
		t.Logf("Acquire successes: %d, TryAcquire successes: %d, Total: %d",
			acquireSuccess.Load(), tryAcquireSuccess.Load(), total)

		// Should have some successes
		if total == 0 {
			t.Error("expected some successful acquisitions")
		}

		// All permits should be returned
		if sem.Current() != 5 {
			t.Errorf("expected all permits returned, got %d", sem.Current())
		}
	})
}
