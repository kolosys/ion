// Package ratelimit provides local process rate limiters for controlling function and I/O throughput.
// It includes token bucket and leaky bucket implementations with configurable options.
package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/kolosys/ion/shared"
)

// Limiter represents a rate limiter that controls the rate at which events are allowed to occur.
type Limiter interface {
	// AllowN reports whether n events may happen at time now.
	// It returns true if the events are allowed, false otherwise.
	// This method never blocks.
	AllowN(now time.Time, n int) bool

	// WaitN blocks until n events can be allowed or the context is canceled.
	// It returns an error if the context is canceled or times out.
	WaitN(ctx context.Context, n int) error
}

// Rate represents the rate at which tokens are added to the bucket.
type Rate struct {
	TokensPerSec float64
}

// NewRate creates a new Rate from the given number of tokens per time duration.
func NewRate(tokens int, duration time.Duration) Rate {
	return Rate{
		TokensPerSec: float64(tokens) / duration.Seconds(),
	}
}

// Per is a convenience function for creating rates.
// For example: Per(100, time.Second) creates a rate of 100 tokens per second.
func Per(tokens int, duration time.Duration) Rate {
	return NewRate(tokens, duration)
}

// PerSecond creates a rate of the given number of tokens per second.
func PerSecond(tokens int) Rate {
	return Rate{TokensPerSec: float64(tokens)}
}

// PerMinute creates a rate of the given number of tokens per minute.
func PerMinute(tokens int) Rate {
	return Per(tokens, time.Minute)
}

// PerHour creates a rate of the given number of tokens per hour.
func PerHour(tokens int) Rate {
	return Per(tokens, time.Hour)
}

// String returns a string representation of the rate.
func (r Rate) String() string {
	if r.TokensPerSec == 0 {
		return "0/s"
	}
	if r.TokensPerSec < 1 {
		return "1/" + time.Duration(1/r.TokensPerSec*float64(time.Second)).String()
	}
	return fmt.Sprintf("%.1f/s", r.TokensPerSec)
}

// Clock abstracts time operations for testability.
type Clock interface {
	Now() time.Time
	Sleep(time.Duration)
	AfterFunc(time.Duration, func()) Timer
}

// Timer represents a timer that can be stopped.
type Timer interface {
	Stop() bool
}

// realClock implements Clock using the real time functions.
type realClock struct{}

func (realClock) Now() time.Time                             { return time.Now() }
func (realClock) Sleep(d time.Duration)                      { time.Sleep(d) }
func (realClock) AfterFunc(d time.Duration, f func()) Timer { return &realTimer{time.AfterFunc(d, f)} }

// realTimer wraps time.Timer to implement our Timer interface.
type realTimer struct{ *time.Timer }

func (t *realTimer) Stop() bool { return t.Timer.Stop() }

// Option configures rate limiter behavior.
type Option func(*config)

type config struct {
	name     string
	clock    Clock
	jitter   float64
	obs      *shared.Observability
}

// WithName sets the rate limiter name for observability and error reporting.
func WithName(name string) Option {
	return func(c *config) {
		c.name = name
	}
}

// WithClock sets a custom clock implementation (useful for testing).
func WithClock(clock Clock) Option {
	return func(c *config) {
		c.clock = clock
	}
}

// WithJitter sets the jitter factor for WaitN operations (0.0 to 1.0).
// Jitter helps prevent thundering herd problems by randomizing wait times.
func WithJitter(jitter float64) Option {
	return func(c *config) {
		if jitter < 0 {
			jitter = 0
		}
		if jitter > 1 {
			jitter = 1
		}
		c.jitter = jitter
	}
}

// WithLogger sets the logger for observability.
func WithLogger(logger shared.Logger) Option {
	return func(c *config) {
		c.obs = c.obs.WithLogger(logger)
	}
}

// WithMetrics sets the metrics recorder for observability.
func WithMetrics(metrics shared.Metrics) Option {
	return func(c *config) {
		c.obs = c.obs.WithMetrics(metrics)
	}
}

// WithTracer sets the tracer for observability.
func WithTracer(tracer shared.Tracer) Option {
	return func(c *config) {
		c.obs = c.obs.WithTracer(tracer)
	}
}

// newConfig creates a config with default values.
func newConfig(opts ...Option) *config {
	cfg := &config{
		name:   "",
		clock:  realClock{},
		jitter: 0.0,
		obs:    shared.NewObservability(),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}
