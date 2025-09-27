// Package semaphore provides a weighted semaphore with configurable fairness modes.
package semaphore

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kolosys/ion/observe"
)

// Fairness defines the ordering behavior for semaphore waiters
type Fairness int

const (
	// FIFO processes waiters in first-in-first-out order (default)
	FIFO Fairness = iota
	// LIFO processes waiters in last-in-first-out order
	LIFO
	// None provides no fairness guarantees, allowing maximum performance
	None
)

// String returns the string representation of the fairness mode
func (f Fairness) String() string {
	switch f {
	case FIFO:
		return "FIFO"
	case LIFO:
		return "LIFO"
	case None:
		return "None"
	default:
		return fmt.Sprintf("Fairness(%d)", int(f))
	}
}

// Semaphore represents a weighted semaphore that controls access to a resource
// with a fixed capacity. It supports configurable fairness modes and observability.
type Semaphore interface {
	// Acquire blocks until n permits are available or the context is canceled.
	// Returns an error if the context is canceled or if n exceeds the semaphore capacity.
	Acquire(ctx context.Context, n int64) error

	// TryAcquire attempts to acquire n permits without blocking.
	// Returns true if the permits were acquired, false otherwise.
	TryAcquire(n int64) bool

	// Release returns n permits to the semaphore, potentially unblocking waiters.
	// Panics if n is negative or if more permits are released than were acquired.
	Release(n int64)

	// Current returns the number of permits currently available.
	Current() int64
}

// weightedSemaphore implements the Semaphore interface with weighted permits and fairness
type weightedSemaphore struct {
	// Configuration
	name           string
	capacity       int64
	fairness       Fairness
	acquireTimeout time.Duration

	// Observability
	obs *observe.Observability

	// Synchronization
	mu      sync.Mutex
	current int64
	waiters waiterQueue
	closed  bool
}

// waiter represents a goroutine waiting to acquire permits
type waiter struct {
	weight   int64
	ready    chan struct{}
	ctx      context.Context
	acquired bool
}

// waiterQueue manages the queue of waiting goroutines based on fairness mode
type waiterQueue struct {
	fairness Fairness
	waiters  []*waiter
}

// push adds a waiter to the queue according to fairness policy
func (q *waiterQueue) push(w *waiter) {
	q.waiters = append(q.waiters, w)
}

// popReady removes and returns the first waiter that can be satisfied
func (q *waiterQueue) popReady(available int64) *waiter {
	if len(q.waiters) == 0 {
		return nil
	}

	var index int = -1

	switch q.fairness {
	case FIFO:
		// Find first waiter that can be satisfied
		for i, w := range q.waiters {
			if w.weight <= available {
				index = i
				break
			}
		}
	case LIFO:
		// Find last waiter that can be satisfied
		for i := len(q.waiters) - 1; i >= 0; i-- {
			if q.waiters[i].weight <= available {
				index = i
				break
			}
		}
	case None:
		// Find any waiter that can be satisfied (first match for simplicity)
		for i, w := range q.waiters {
			if w.weight <= available {
				index = i
				break
			}
		}
	}

	if index == -1 {
		return nil
	}

	waiter := q.waiters[index]
	// Remove waiter from slice
	q.waiters = append(q.waiters[:index], q.waiters[index+1:]...)
	return waiter
}

// removeWaiter removes a specific waiter from the queue (for cancellation)
func (q *waiterQueue) removeWaiter(target *waiter) bool {
	for i, w := range q.waiters {
		if w == target {
			q.waiters = append(q.waiters[:i], q.waiters[i+1:]...)
			return true
		}
	}
	return false
}

// len returns the number of waiters in the queue
func (q *waiterQueue) len() int {
	return len(q.waiters)
}

// Option configures semaphore behavior
type Option func(*config)

type config struct {
	name           string
	fairness       Fairness
	acquireTimeout time.Duration
	obs            *observe.Observability
}

// WithName sets the semaphore name for observability and error reporting
func WithName(name string) Option {
	return func(c *config) {
		c.name = name
	}
}

// WithFairness sets the fairness mode for waiter ordering
func WithFairness(fairness Fairness) Option {
	return func(c *config) {
		c.fairness = fairness
	}
}

// WithAcquireTimeout sets the default timeout for Acquire operations
func WithAcquireTimeout(timeout time.Duration) Option {
	return func(c *config) {
		c.acquireTimeout = timeout
	}
}

// WithLogger sets the logger for observability
func WithLogger(logger observe.Logger) Option {
	return func(c *config) {
		c.obs = c.obs.WithLogger(logger)
	}
}

// WithMetrics sets the metrics recorder for observability
func WithMetrics(metrics observe.Metrics) Option {
	return func(c *config) {
		c.obs = c.obs.WithMetrics(metrics)
	}
}

// WithTracer sets the tracer for observability
func WithTracer(tracer observe.Tracer) Option {
	return func(c *config) {
		c.obs = c.obs.WithTracer(tracer)
	}
}

// NewWeighted creates a new weighted semaphore with the specified capacity.
// The semaphore starts with all permits available.
func NewWeighted(capacity int64, opts ...Option) Semaphore {
	if capacity <= 0 {
		panic("semaphore: capacity must be positive")
	}

	cfg := &config{
		name:           "",
		fairness:       FIFO,
		acquireTimeout: 0, // no default timeout
		obs:            observe.New(),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	s := &weightedSemaphore{
		name:           cfg.name,
		capacity:       capacity,
		current:        capacity,
		fairness:       cfg.fairness,
		acquireTimeout: cfg.acquireTimeout,
		obs:            cfg.obs,
		waiters: waiterQueue{
			fairness: cfg.fairness,
			waiters:  make([]*waiter, 0),
		},
	}

	s.obs.Logger.Info("semaphore created",
		"name", s.name,
		"capacity", capacity,
		"fairness", cfg.fairness.String(),
	)

	return s
}

// todo_write to mark setup complete and move to implementation
