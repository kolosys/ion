package ratelimit

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// TokenBucket implements a token bucket rate limiter.
// Tokens are added to the bucket at a fixed rate, and requests consume tokens.
// If no tokens are available, requests must wait or are denied.
type TokenBucket struct {
	// Configuration
	rate  Rate
	burst int
	cfg   *config

	// State
	mu          sync.Mutex
	tokens      float64
	lastRefill  time.Time
	initialized bool
}

// NewTokenBucket creates a new token bucket rate limiter.
// rate determines how fast tokens are added to the bucket.
// burst is the maximum number of tokens the bucket can hold.
func NewTokenBucket(rate Rate, burst int, opts ...Option) *TokenBucket {
	if burst <= 0 {
		panic("ratelimit: burst must be positive")
	}
	if rate.TokensPerSec < 0 {
		panic("ratelimit: rate cannot be negative")
	}

	cfg := newConfig(opts...)

	tb := &TokenBucket{
		rate:   rate,
		burst:  burst,
		cfg:    cfg,
		tokens: float64(burst), // Start with full bucket
	}

	tb.cfg.obs.Logger.Info("token bucket created",
		"name", cfg.name,
		"rate", rate.String(),
		"burst", burst,
	)

	return tb
}

// AllowN reports whether n tokens are available at time now.
// It returns true if the tokens were consumed, false otherwise.
func (tb *TokenBucket) AllowN(now time.Time, n int) bool {
	if n <= 0 {
		return true
	}

	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refillLocked(now)

	if float64(n) <= tb.tokens {
		tb.tokens -= float64(n)
		tb.cfg.obs.Metrics.Inc("ion_ratelimit_requests_total",
			"limiter_name", tb.cfg.name, "result", "allowed")
		tb.cfg.obs.Metrics.Gauge("ion_ratelimit_tokens_available",
			tb.tokens, "limiter_name", tb.cfg.name)
		return true
	}

	tb.cfg.obs.Metrics.Inc("ion_ratelimit_requests_total",
		"limiter_name", tb.cfg.name, "result", "denied")
	return false
}

// WaitN blocks until n tokens are available or the context is canceled.
func (tb *TokenBucket) WaitN(ctx context.Context, n int) error {
	if n <= 0 {
		return nil
	}

	// Fast path: try to get tokens immediately
	now := tb.cfg.clock.Now()
	if tb.AllowN(now, n) {
		return nil
	}

	// Slow path: wait for tokens
	return tb.waitSlow(ctx, n, now)
}

// waitSlow handles the blocking wait for tokens.
func (tb *TokenBucket) waitSlow(ctx context.Context, n int, now time.Time) error {
	tb.mu.Lock()
	tb.refillLocked(now)

	// Check if request can ever be satisfied
	if n > tb.burst {
		tb.mu.Unlock()
		return fmt.Errorf("ratelimit: requested %d tokens exceeds burst limit %d", n, tb.burst)
	}

	// Calculate wait time
	deficit := float64(n) - tb.tokens
	var waitDuration time.Duration
	if tb.rate.TokensPerSec > 0 {
		waitDuration = time.Duration(deficit / tb.rate.TokensPerSec * float64(time.Second))
	} else {
		// Rate is zero, wait indefinitely
		tb.mu.Unlock()
		<-ctx.Done()
		return ctx.Err()
	}

	// Apply jitter if configured
	if tb.cfg.jitter > 0 {
		jitter := rand.Float64() * tb.cfg.jitter * waitDuration.Seconds()
		waitDuration += time.Duration(jitter * float64(time.Second))
	}

	tb.mu.Unlock()

	tb.cfg.obs.Logger.Debug("rate limiter waiting",
		"limiter_name", tb.cfg.name,
		"requested", n,
		"wait_duration", waitDuration,
	)

	start := tb.cfg.clock.Now()

	// Wait for the calculated duration or context cancellation
	timer := tb.cfg.clock.AfterFunc(waitDuration, func() {})
	defer timer.Stop()

	select {
	case <-ctx.Done():
		tb.cfg.obs.Metrics.Inc("ion_ratelimit_requests_total",
			"limiter_name", tb.cfg.name, "result", "canceled")
		return ctx.Err()

	case <-time.After(waitDuration):
		// Try to acquire tokens again
		now = tb.cfg.clock.Now()
		if tb.AllowN(now, n) {
			duration := tb.cfg.clock.Now().Sub(start)
			tb.cfg.obs.Metrics.Histogram("ion_ratelimit_wait_duration_seconds",
				duration.Seconds(), "limiter_name", tb.cfg.name)
			return nil
		}

		// Shouldn't happen with correct implementation, but handle gracefully
		return fmt.Errorf("ratelimit: tokens not available after wait")
	}
}

// refillLocked adds tokens to the bucket based on elapsed time.
// Must be called with tb.mu held.
func (tb *TokenBucket) refillLocked(now time.Time) {
	if !tb.initialized {
		tb.lastRefill = now
		tb.initialized = true
		return
	}

	if tb.rate.TokensPerSec <= 0 {
		return // No refill for zero rate
	}

	elapsed := now.Sub(tb.lastRefill)
	if elapsed <= 0 {
		return // Time hasn't advanced or went backwards
	}

	// Calculate tokens to add
	tokensToAdd := tb.rate.TokensPerSec * elapsed.Seconds()
	tb.tokens = math.Min(tb.tokens+tokensToAdd, float64(tb.burst))
	tb.lastRefill = now

	tb.cfg.obs.Metrics.Gauge("ion_ratelimit_tokens_available",
		tb.tokens, "limiter_name", tb.cfg.name)
}

// Tokens returns the current number of available tokens.
func (tb *TokenBucket) Tokens() float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refillLocked(tb.cfg.clock.Now())
	return tb.tokens
}

// Rate returns the current token refill rate.
func (tb *TokenBucket) Rate() Rate {
	return tb.rate
}

// Burst returns the bucket capacity.
func (tb *TokenBucket) Burst() int {
	return tb.burst
}
