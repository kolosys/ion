package ratelimit_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kolosys/ion/ratelimit"
)

func TestRate(t *testing.T) {
	tests := []struct {
		name     string
		rate     ratelimit.Rate
		expected string
	}{
		{"zero rate", ratelimit.Rate{0}, "0/s"},
		{"per second", ratelimit.PerSecond(10), "10.0/s"},
		{"per minute", ratelimit.PerMinute(60), "1.0/s"},
		{"per hour", ratelimit.PerHour(3600), "1.0/s"},
		{"custom rate", ratelimit.Per(5, 2*time.Second), "2.5/s"},
		{"fractional", ratelimit.Rate{0.5}, "1/2s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rate.String(); got != tt.expected {
				t.Errorf("Rate.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTokenBucketNew(t *testing.T) {
	t.Run("valid parameters", func(t *testing.T) {
		tb := ratelimit.NewTokenBucket(ratelimit.PerSecond(10), 5)
		if tb.Rate().TokensPerSec != 10 {
			t.Errorf("expected rate 10, got %v", tb.Rate().TokensPerSec)
		}
		if tb.Burst() != 5 {
			t.Errorf("expected burst 5, got %v", tb.Burst())
		}
	})

	t.Run("zero burst panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for zero burst")
			}
		}()
		ratelimit.NewTokenBucket(ratelimit.PerSecond(10), 0)
	})

	t.Run("negative rate panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for negative rate")
			}
		}()
		ratelimit.NewTokenBucket(ratelimit.Rate{-1}, 5)
	})
}

func TestTokenBucketAllowN(t *testing.T) {
	t.Run("initial burst available", func(t *testing.T) {
		clock := newTestClock(time.Now())
		tb := ratelimit.NewTokenBucket(ratelimit.PerSecond(10), 5, ratelimit.WithClock(clock))

		if !tb.AllowN(clock.Now(), 5) {
			t.Error("should allow initial burst")
		}
		if tb.AllowN(clock.Now(), 1) {
			t.Error("should not allow more than burst")
		}
	})

	t.Run("refill over time", func(t *testing.T) {
		clock := newTestClock(time.Now())
		tb := ratelimit.NewTokenBucket(ratelimit.PerSecond(10), 5, ratelimit.WithClock(clock))

		// Use all initial tokens
		if !tb.AllowN(clock.Now(), 5) {
			t.Error("should allow initial burst")
		}

		// Advance time by 1 second to add 10 tokens (limited to 5 by burst)
		clock.Advance(time.Second)

		if !tb.AllowN(clock.Now(), 5) {
			t.Error("should allow 5 tokens after refill")
		}

		// Should have 0 tokens remaining after using all refilled tokens
		if tb.AllowN(clock.Now(), 1) {
			t.Error("should not have tokens after using all")
		}
	})

	t.Run("zero and negative requests", func(t *testing.T) {
		clock := newTestClock(time.Now())
		tb := ratelimit.NewTokenBucket(ratelimit.PerSecond(10), 5, ratelimit.WithClock(clock))

		if !tb.AllowN(clock.Now(), 0) {
			t.Error("should allow 0 tokens")
		}
		if !tb.AllowN(clock.Now(), -1) {
			t.Error("should allow negative tokens")
		}
	})
}

func TestTokenBucketWaitN(t *testing.T) {
	clock := newTestClock(time.Now())
	tb := ratelimit.NewTokenBucket(ratelimit.PerSecond(10), 5, ratelimit.WithClock(clock))

	t.Run("immediate success", func(t *testing.T) {
		err := tb.WaitN(context.Background(), 3)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("wait for refill", func(t *testing.T) {
		// Use remaining tokens
		tb.AllowN(clock.Now(), 2)

		ctx := context.Background()
		done := make(chan error, 1)

		// Start waiting for 1 token
		go func() {
			done <- tb.WaitN(ctx, 1)
		}()

		// Advance time to add tokens
		time.Sleep(10 * time.Millisecond)     // Let goroutine start
		clock.Advance(100 * time.Millisecond) // Add 1 token

		select {
		case err := <-done:
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("WaitN should have completed")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		// Empty the bucket
		tb.AllowN(clock.Now(), 5)

		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)

		go func() {
			done <- tb.WaitN(ctx, 1)
		}()

		time.Sleep(10 * time.Millisecond)
		cancel()

		select {
		case err := <-done:
			if err != context.Canceled {
				t.Errorf("expected context.Canceled, got %v", err)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("WaitN should have been canceled")
		}
	})

	t.Run("request exceeds burst", func(t *testing.T) {
		err := tb.WaitN(context.Background(), 10) // Burst is 5
		if err == nil {
			t.Error("expected error for request exceeding burst")
		}
	})
}

func TestLeakyBucketNew(t *testing.T) {
	t.Run("valid parameters", func(t *testing.T) {
		lb := ratelimit.NewLeakyBucket(ratelimit.PerSecond(10), 5)
		if lb.Rate().TokensPerSec != 10 {
			t.Errorf("expected rate 10, got %v", lb.Rate().TokensPerSec)
		}
		if lb.Capacity() != 5 {
			t.Errorf("expected capacity 5, got %v", lb.Capacity())
		}
	})

	t.Run("zero capacity panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for zero capacity")
			}
		}()
		ratelimit.NewLeakyBucket(ratelimit.PerSecond(10), 0)
	})
}

func TestLeakyBucketAllowN(t *testing.T) {
	clock := newTestClock(time.Now())
	lb := ratelimit.NewLeakyBucket(ratelimit.PerSecond(10), 5, ratelimit.WithClock(clock))

	t.Run("fill bucket", func(t *testing.T) {
		if !lb.AllowN(clock.Now(), 5) {
			t.Error("should allow filling bucket")
		}
		if lb.AllowN(clock.Now(), 1) {
			t.Error("should not allow overfilling bucket")
		}
	})

	t.Run("leak over time", func(t *testing.T) {
		// Advance time by 0.5 second to leak 5 requests
		clock.Advance(500 * time.Millisecond)

		if !lb.AllowN(clock.Now(), 5) {
			t.Error("should allow 5 requests after leak")
		}

		if lb.AllowN(clock.Now(), 1) {
			t.Error("should not allow more requests")
		}
	})

	t.Run("available space", func(t *testing.T) {
		clock.Advance(200 * time.Millisecond) // Leak 2 requests
		available := lb.Available()
		if available != 2 {
			t.Errorf("expected 2 available, got %v", available)
		}
	})
}

func TestLeakyBucketWaitN(t *testing.T) {
	clock := newTestClock(time.Now())
	lb := ratelimit.NewLeakyBucket(ratelimit.PerSecond(10), 5, ratelimit.WithClock(clock))

	t.Run("immediate success", func(t *testing.T) {
		err := lb.WaitN(context.Background(), 3)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("wait for leak", func(t *testing.T) {
		// Fill remaining space
		lb.AllowN(clock.Now(), 2)

		ctx := context.Background()
		done := make(chan error, 1)

		// Start waiting for 1 slot
		go func() {
			done <- lb.WaitN(ctx, 1)
		}()

		// Advance time to create space
		time.Sleep(10 * time.Millisecond)     // Let goroutine start
		clock.Advance(100 * time.Millisecond) // Leak 1 request

		select {
		case err := <-done:
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("WaitN should have completed")
		}
	})

	t.Run("request exceeds capacity", func(t *testing.T) {
		err := lb.WaitN(context.Background(), 10) // Capacity is 5
		if err == nil {
			t.Error("expected error for request exceeding capacity")
		}
	})
}

func TestConcurrency(t *testing.T) {
	t.Run("token bucket concurrency", func(t *testing.T) {
		tb := ratelimit.NewTokenBucket(ratelimit.PerSecond(100), 10)
		const numGoroutines = 50
		const requestsPerGoroutine = 10

		var allowed atomic.Int64
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < requestsPerGoroutine; j++ {
					if tb.AllowN(time.Now(), 1) {
						allowed.Add(1)
					}
				}
			}()
		}

		wg.Wait()

		// Should have allowed at least the burst amount
		if allowed.Load() < int64(tb.Burst()) {
			t.Errorf("expected at least %d allowed, got %d", tb.Burst(), allowed.Load())
		}
	})

	t.Run("leaky bucket concurrency", func(t *testing.T) {
		lb := ratelimit.NewLeakyBucket(ratelimit.PerSecond(100), 10)
		const numGoroutines = 50
		const requestsPerGoroutine = 10

		var allowed atomic.Int64
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < requestsPerGoroutine; j++ {
					if lb.AllowN(time.Now(), 1) {
						allowed.Add(1)
					}
				}
			}()
		}

		wg.Wait()

		// Should have allowed at most the capacity
		if allowed.Load() > int64(lb.Capacity()) {
			t.Errorf("expected at most %d allowed, got %d", lb.Capacity(), allowed.Load())
		}
	})
}

func TestZeroRate(t *testing.T) {
	t.Run("token bucket with zero rate", func(t *testing.T) {
		clock := newTestClock(time.Now())
		tb := ratelimit.NewTokenBucket(ratelimit.Rate{0}, 5, ratelimit.WithClock(clock))

		// Should allow initial burst
		if !tb.AllowN(clock.Now(), 5) {
			t.Error("should allow initial burst")
		}

		// Should not refill after time passes
		clock.Advance(time.Hour)
		if tb.AllowN(clock.Now(), 1) {
			t.Error("should not refill with zero rate")
		}
	})

	t.Run("leaky bucket with zero rate", func(t *testing.T) {
		clock := newTestClock(time.Now())
		lb := ratelimit.NewLeakyBucket(ratelimit.Rate{0}, 5, ratelimit.WithClock(clock))

		// Should allow filling bucket
		if !lb.AllowN(clock.Now(), 5) {
			t.Error("should allow filling bucket")
		}

		// Should not leak after time passes
		clock.Advance(time.Hour)
		if lb.AllowN(clock.Now(), 1) {
			t.Error("should not leak with zero rate")
		}
	})
}
