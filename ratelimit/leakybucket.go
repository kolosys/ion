package ratelimit

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// LeakyBucket implements a leaky bucket rate limiter.
// Requests are added to the bucket, and the bucket leaks at a constant rate.
// If the bucket is full, requests are denied or must wait.
type LeakyBucket struct {
	// Configuration
	rate     Rate
	capacity int
	cfg      *config

	// State
	mu           sync.Mutex
	level        float64 // Current level in the bucket
	lastLeak     time.Time
	initialized  bool
}

// NewLeakyBucket creates a new leaky bucket rate limiter.
// rate determines how fast the bucket leaks (processes requests).
// capacity is the maximum number of requests the bucket can hold.
func NewLeakyBucket(rate Rate, capacity int, opts ...Option) *LeakyBucket {
	if capacity <= 0 {
		panic("ratelimit: capacity must be positive")
	}
	if rate.TokensPerSec < 0 {
		panic("ratelimit: rate cannot be negative")
	}

	cfg := newConfig(opts...)

	lb := &LeakyBucket{
		rate:     rate,
		capacity: capacity,
		cfg:      cfg,
		level:    0, // Start with empty bucket
	}

	lb.cfg.obs.Logger.Info("leaky bucket created",
		"name", cfg.name,
		"rate", rate.String(),
		"capacity", capacity,
	)

	return lb
}

// AllowN reports whether n requests can be added to the bucket at time now.
// It returns true if the requests were accepted, false otherwise.
func (lb *LeakyBucket) AllowN(now time.Time, n int) bool {
	if n <= 0 {
		return true
	}

	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.leakLocked(now)

	// Check if we can add n requests to the bucket
	if lb.level+float64(n) <= float64(lb.capacity) {
		lb.level += float64(n)
		lb.cfg.obs.Metrics.Inc("ion_ratelimit_requests_total",
			"limiter_name", lb.cfg.name, "result", "allowed")
		lb.cfg.obs.Metrics.Gauge("ion_ratelimit_bucket_level",
			lb.level, "limiter_name", lb.cfg.name)
		return true
	}

	lb.cfg.obs.Metrics.Inc("ion_ratelimit_requests_total",
		"limiter_name", lb.cfg.name, "result", "denied")
	return false
}

// WaitN blocks until n requests can be added to the bucket or the context is canceled.
func (lb *LeakyBucket) WaitN(ctx context.Context, n int) error {
	if n <= 0 {
		return nil
	}

	// Fast path: try to add requests immediately
	now := lb.cfg.clock.Now()
	if lb.AllowN(now, n) {
		return nil
	}

	// Slow path: wait for space in bucket
	return lb.waitSlow(ctx, n, now)
}

// waitSlow handles the blocking wait for bucket space.
func (lb *LeakyBucket) waitSlow(ctx context.Context, n int, now time.Time) error {
	lb.mu.Lock()
	lb.leakLocked(now)

	// Check if request can ever be satisfied
	if n > lb.capacity {
		lb.mu.Unlock()
		return fmt.Errorf("ratelimit: requested %d requests exceeds bucket capacity %d", n, lb.capacity)
	}

	// Calculate wait time
	needed := lb.level + float64(n) - float64(lb.capacity)
	var waitDuration time.Duration
	if needed > 0 && lb.rate.TokensPerSec > 0 {
		waitDuration = time.Duration(needed/lb.rate.TokensPerSec*float64(time.Second))
	} else if lb.rate.TokensPerSec <= 0 {
		// Rate is zero, bucket never leaks
		lb.mu.Unlock()
		<-ctx.Done()
		return ctx.Err()
	}

	// Apply jitter if configured
	if lb.cfg.jitter > 0 && waitDuration > 0 {
		jitter := rand.Float64() * lb.cfg.jitter * waitDuration.Seconds()
		waitDuration += time.Duration(jitter * float64(time.Second))
	}

	lb.mu.Unlock()

	if waitDuration <= 0 {
		// Should be able to add requests now
		return lb.WaitN(ctx, n)
	}

	lb.cfg.obs.Logger.Debug("leaky bucket waiting",
		"limiter_name", lb.cfg.name,
		"requested", n,
		"wait_duration", waitDuration,
	)

	start := lb.cfg.clock.Now()

	// Wait for the calculated duration or context cancellation
	timer := lb.cfg.clock.AfterFunc(waitDuration, func() {})
	defer timer.Stop()

	select {
	case <-ctx.Done():
		lb.cfg.obs.Metrics.Inc("ion_ratelimit_requests_total",
			"limiter_name", lb.cfg.name, "result", "canceled")
		return ctx.Err()

	case <-time.After(waitDuration):
		// Try to add requests again
		now = lb.cfg.clock.Now()
		if lb.AllowN(now, n) {
			duration := lb.cfg.clock.Now().Sub(start)
			lb.cfg.obs.Metrics.Histogram("ion_ratelimit_wait_duration_seconds",
				duration.Seconds(), "limiter_name", lb.cfg.name)
			return nil
		}

		// Shouldn't happen with correct implementation, but handle gracefully
		return fmt.Errorf("ratelimit: bucket space not available after wait")
	}
}

// leakLocked removes requests from the bucket based on elapsed time.
// Must be called with lb.mu held.
func (lb *LeakyBucket) leakLocked(now time.Time) {
	if !lb.initialized {
		lb.lastLeak = now
		lb.initialized = true
		return
	}

	if lb.rate.TokensPerSec <= 0 {
		return // No leak for zero rate
	}

	elapsed := now.Sub(lb.lastLeak)
	if elapsed <= 0 {
		return // Time hasn't advanced or went backwards
	}

	// Calculate how much should leak out
	leakAmount := lb.rate.TokensPerSec * elapsed.Seconds()
	lb.level = math.Max(0, lb.level-leakAmount)
	lb.lastLeak = now

	lb.cfg.obs.Metrics.Gauge("ion_ratelimit_bucket_level",
		lb.level, "limiter_name", lb.cfg.name)
}

// Level returns the current level of the bucket.
func (lb *LeakyBucket) Level() float64 {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.leakLocked(lb.cfg.clock.Now())
	return lb.level
}

// Rate returns the current leak rate.
func (lb *LeakyBucket) Rate() Rate {
	return lb.rate
}

// Capacity returns the bucket capacity.
func (lb *LeakyBucket) Capacity() int {
	return lb.capacity
}

// Available returns the number of requests that can be immediately accepted.
func (lb *LeakyBucket) Available() int {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.leakLocked(lb.cfg.clock.Now())
	return int(math.Max(0, float64(lb.capacity)-lb.level))
}
