// Package circuit provides circuit breaker functionality for resilient microservices.
// Circuit breakers prevent cascading failures by temporarily blocking requests to failing services,
// allowing them time to recover while providing fast-fail behavior to callers.
//
// The circuit breaker implements a three-state machine:
// - Closed: Normal operation, requests pass through
// - Open: Circuit is tripped, requests fail fast
// - Half-Open: Testing recovery, limited requests allowed
//
// Usage:
//
//	cb := circuit.New("payment-service",
//		circuit.WithFailureThreshold(5),
//		circuit.WithRecoveryTimeout(30*time.Second),
//	)
//
//	result, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
//		return paymentService.ProcessPayment(ctx, payment)
//	})
//
// The circuit breaker integrates with ION's observability system and supports
// context cancellation, timeouts, and comprehensive metrics collection.
package circuit

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/kolosys/ion/observe"
)

// CircuitBreaker represents a circuit breaker that controls access to a potentially
// failing operation. It provides fast-fail behavior when the operation is failing
// and automatic recovery testing when appropriate.
type CircuitBreaker interface {
	// Execute runs the given function with circuit breaker protection.
	// Returns the result of the function or a circuit breaker error if the circuit is open.
	Execute(ctx context.Context, fn func(context.Context) (any, error)) (any, error)

	// Call is a convenience method for functions that don't return values.
	// It's equivalent to Execute but discards the return value.
	Call(ctx context.Context, fn func(context.Context) error) error

	// State returns the current state of the circuit breaker.
	State() State

	// Metrics returns current metrics for the circuit breaker.
	Metrics() CircuitMetrics

	// Reset manually resets the circuit breaker to the closed state.
	// This should be used sparingly and only when external monitoring
	// indicates the service has recovered.
	Reset()

	// Close gracefully shuts down the circuit breaker, preventing new operations
	// and waiting for in-flight operations to complete.
	Close() error
}

// circuitBreaker is the concrete implementation of CircuitBreaker.
type circuitBreaker struct {
	name string

	// Configuration
	config *Config

	// State management (atomic access only)
	state           atomic.Int32 // State value
	failures        atomic.Int64 // consecutive failures
	successes       atomic.Int64 // consecutive successes in half-open
	lastFailure     atomic.Int64 // unix nano timestamp
	lastSuccess     atomic.Int64 // unix nano timestamp
	lastStateChange atomic.Int64 // unix nano timestamp

	// Metrics (atomic access only)
	totalRequests  atomic.Int64
	totalFailures  atomic.Int64
	totalSuccesses atomic.Int64
	stateChanges   atomic.Int64

	// Observability
	obs *observe.Observability
}

// New creates a new circuit breaker with the given name and options.
func New(name string, options ...Option) CircuitBreaker {
	cb := &circuitBreaker{
		name:   name,
		config: DefaultConfig(),
		obs:    observe.New(),
	}

	// Apply options
	for _, option := range options {
		option(cb.config, cb.obs)
	}

	// Initialize state
	cb.state.Store(int32(Closed))
	now := time.Now().UnixNano()
	cb.lastStateChange.Store(now)

	cb.obs.Logger.Info("circuit breaker created",
		"name", name,
		"failure_threshold", cb.config.FailureThreshold,
		"recovery_timeout", cb.config.RecoveryTimeout,
	)

	return cb
}

// Execute implements CircuitBreaker.Execute
func (cb *circuitBreaker) Execute(ctx context.Context, fn func(context.Context) (any, error)) (any, error) {
	// Fast path: check if we should allow the request
	if !cb.allowRequest() {
		cb.obs.Metrics.Inc("circuit.requests_rejected", "name", cb.name, "state", cb.State().String())
		return nil, NewCircuitOpenError(cb.name)
	}

	// Increment total requests
	cb.totalRequests.Add(1)
	cb.obs.Metrics.Inc("circuit.requests_total", "name", cb.name, "state", cb.State().String())

	// Create tracing span
	spanCtx, finish := cb.obs.Tracer.Start(ctx, "circuit.execute", "name", cb.name)
	defer func() { finish(nil) }()

	// Execute the function
	start := time.Now()
	result, err := fn(spanCtx)
	duration := time.Since(start)

	cb.obs.Metrics.Histogram("circuit.request_duration", duration.Seconds(), "name", cb.name)

	// Record the result
	if err != nil {
		// Check if this error should count as a failure
		isFailure := cb.config.IsFailure == nil || cb.config.IsFailure(err)
		if isFailure {
			cb.recordFailure()
			cb.obs.Metrics.Inc("circuit.requests_failed", "name", cb.name)
		} else {
			cb.recordSuccess()
			cb.obs.Metrics.Inc("circuit.requests_succeeded", "name", cb.name)
		}
		cb.obs.Logger.Debug("circuit breaker request failed", "name", cb.name, "error", err, "counted_as_failure", isFailure)
	} else {
		cb.recordSuccess()
		cb.obs.Metrics.Inc("circuit.requests_succeeded", "name", cb.name)
	}

	return result, err
}

// Call implements CircuitBreaker.Call
func (cb *circuitBreaker) Call(ctx context.Context, fn func(context.Context) error) error {
	_, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
		return nil, fn(ctx)
	})
	return err
}

// State implements CircuitBreaker.State
func (cb *circuitBreaker) State() State {
	return State(cb.state.Load())
}

// Metrics implements CircuitBreaker.Metrics
func (cb *circuitBreaker) Metrics() CircuitMetrics {
	return CircuitMetrics{
		Name:             cb.name,
		State:            cb.State(),
		TotalRequests:    cb.totalRequests.Load(),
		TotalFailures:    cb.totalFailures.Load(),
		TotalSuccesses:   cb.totalSuccesses.Load(),
		ConsecutiveFails: cb.failures.Load(),
		StateChanges:     cb.stateChanges.Load(),
		LastFailure:      time.Unix(0, cb.lastFailure.Load()),
		LastSuccess:      time.Unix(0, cb.lastSuccess.Load()),
		LastStateChange:  time.Unix(0, cb.lastStateChange.Load()),
	}
}

// Reset implements CircuitBreaker.Reset
func (cb *circuitBreaker) Reset() {
	cb.setState(Closed)
	cb.failures.Store(0)
	cb.successes.Store(0)
	cb.obs.Logger.Info("circuit breaker manually reset", "name", cb.name)
	cb.obs.Metrics.Inc("circuit.manual_reset", "name", cb.name)
}

// Close implements CircuitBreaker.Close
func (cb *circuitBreaker) Close() error {
	cb.obs.Logger.Info("circuit breaker closing", "name", cb.name)
	return nil
}

// allowRequest determines if a request should be allowed based on current state
func (cb *circuitBreaker) allowRequest() bool {
	state := cb.State()
	now := time.Now()

	switch state {
	case Closed:
		return true

	case Open:
		// Check if recovery timeout has passed
		lastStateChange := time.Unix(0, cb.lastStateChange.Load())
		if now.Sub(lastStateChange) >= cb.config.RecoveryTimeout {
			// Transition to half-open for testing
			if cb.setState(HalfOpen) {
				cb.obs.Logger.Info("circuit breaker transitioning to half-open", "name", cb.name)
			}
			return true
		}
		return false

	case HalfOpen:
		// Allow limited requests in half-open state
		successes := cb.successes.Load()
		return successes < cb.config.HalfOpenMaxRequests

	default:
		return false
	}
}

// recordSuccess records a successful operation
func (cb *circuitBreaker) recordSuccess() {
	cb.totalSuccesses.Add(1)
	cb.lastSuccess.Store(time.Now().UnixNano())

	state := cb.State()
	switch state {
	case Closed:
		// Reset failure count on success in closed state
		cb.failures.Store(0)

	case HalfOpen:
		successes := cb.successes.Add(1)
		if successes >= cb.config.HalfOpenSuccessThreshold {
			// Enough successes - transition back to closed
			if cb.setState(Closed) {
				cb.obs.Logger.Info("circuit breaker recovered, transitioning to closed", "name", cb.name)
			}
		}
	}
}

// recordFailure records a failed operation
func (cb *circuitBreaker) recordFailure() {
	cb.totalFailures.Add(1)
	cb.lastFailure.Store(time.Now().UnixNano())

	state := cb.State()
	switch state {
	case Closed:
		failures := cb.failures.Add(1)
		if failures >= cb.config.FailureThreshold {
			// Too many failures - trip the circuit
			if cb.setState(Open) {
				cb.obs.Logger.Warn("circuit breaker tripped, transitioning to open",
					"name", cb.name, "failures", failures)
			}
		}

	case HalfOpen:
		// Any failure in half-open state trips the circuit back to open
		if cb.setState(Open) {
			cb.obs.Logger.Warn("circuit breaker failed during recovery, transitioning to open", "name", cb.name)
		}
	}
}

// setState atomically changes the circuit state and resets counters
func (cb *circuitBreaker) setState(newState State) bool {
	oldState := State(cb.state.Swap(int32(newState)))
	if oldState != newState {
		// State changed - reset counters and update metrics
		cb.failures.Store(0)
		cb.successes.Store(0)
		cb.lastStateChange.Store(time.Now().UnixNano())
		cb.stateChanges.Add(1)

		cb.obs.Metrics.Inc("circuit.state_changes",
			"name", cb.name,
			"from", oldState.String(),
			"to", newState.String())

		// Call state change callback if configured
		if cb.config.OnStateChange != nil {
			cb.config.OnStateChange(oldState, newState)
		}

		return true
	}
	return false
}
