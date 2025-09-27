package circuit

import (
	"fmt"
	"time"
)

// State represents the current state of a circuit breaker.
type State int32

const (
	// Closed indicates the circuit is closed and requests are passing through normally.
	// This is the initial state and the state when the service is healthy.
	Closed State = iota

	// Open indicates the circuit is open and requests are failing fast.
	// This state is entered when the failure threshold is exceeded.
	Open

	// HalfOpen indicates the circuit is in recovery mode, allowing limited requests
	// to test if the downstream service has recovered.
	HalfOpen
)

// String returns the string representation of the circuit state.
func (s State) String() string {
	switch s {
	case Closed:
		return "Closed"
	case Open:
		return "Open"
	case HalfOpen:
		return "HalfOpen"
	default:
		return fmt.Sprintf("State(%d)", int(s))
	}
}

// CircuitMetrics holds metrics for a circuit breaker instance.
type CircuitMetrics struct {
	// Name is the name of the circuit breaker
	Name string

	// State is the current state of the circuit
	State State

	// TotalRequests is the total number of requests processed
	TotalRequests int64

	// TotalFailures is the total number of failed requests
	TotalFailures int64

	// TotalSuccesses is the total number of successful requests
	TotalSuccesses int64

	// ConsecutiveFails is the current count of consecutive failures
	ConsecutiveFails int64

	// StateChanges is the total number of state transitions
	StateChanges int64

	// LastFailure is the timestamp of the last failure
	LastFailure time.Time

	// LastSuccess is the timestamp of the last success
	LastSuccess time.Time

	// LastStateChange is the timestamp of the last state change
	LastStateChange time.Time
}

// FailureRate returns the failure rate as a percentage (0.0 to 1.0).
func (m CircuitMetrics) FailureRate() float64 {
	if m.TotalRequests == 0 {
		return 0.0
	}
	return float64(m.TotalFailures) / float64(m.TotalRequests)
}

// SuccessRate returns the success rate as a percentage (0.0 to 1.0).
func (m CircuitMetrics) SuccessRate() float64 {
	return 1.0 - m.FailureRate()
}

// IsHealthy returns true if the circuit appears to be healthy based on recent activity.
func (m CircuitMetrics) IsHealthy() bool {
	return m.State == Closed && m.ConsecutiveFails == 0
}

// Config holds configuration for a circuit breaker.
type Config struct {
	// FailureThreshold is the number of consecutive failures required to trip the circuit.
	// Default: 5
	FailureThreshold int64

	// RecoveryTimeout is the duration to wait in the open state before transitioning
	// to half-open for recovery testing.
	// Default: 30 seconds
	RecoveryTimeout time.Duration

	// HalfOpenMaxRequests is the maximum number of requests allowed in half-open state.
	// Default: 3
	HalfOpenMaxRequests int64

	// HalfOpenSuccessThreshold is the number of successful requests required in
	// half-open state to transition back to closed.
	// Default: 2
	HalfOpenSuccessThreshold int64

	// IsFailure is a predicate function that determines if an error should be
	// counted as a failure for circuit breaker purposes. If nil, all non-nil
	// errors are considered failures.
	IsFailure func(error) bool

	// OnStateChange is called whenever the circuit breaker changes state.
	// This is useful for logging or metrics collection.
	OnStateChange func(from, to State)
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		FailureThreshold:         5,
		RecoveryTimeout:          30 * time.Second,
		HalfOpenMaxRequests:      3,
		HalfOpenSuccessThreshold: 2,
		IsFailure:                nil, // nil means all errors are failures
		OnStateChange:            nil, // nil means no callback
	}
}

// Validate checks if the configuration is valid and returns an error if not.
func (c *Config) Validate() error {
	if c.FailureThreshold <= 0 {
		return fmt.Errorf("failure threshold must be positive, got %d", c.FailureThreshold)
	}

	if c.RecoveryTimeout <= 0 {
		return fmt.Errorf("recovery timeout must be positive, got %v", c.RecoveryTimeout)
	}

	if c.HalfOpenMaxRequests <= 0 {
		return fmt.Errorf("half-open max requests must be positive, got %d", c.HalfOpenMaxRequests)
	}

	if c.HalfOpenSuccessThreshold <= 0 {
		return fmt.Errorf("half-open success threshold must be positive, got %d", c.HalfOpenSuccessThreshold)
	}

	if c.HalfOpenSuccessThreshold > c.HalfOpenMaxRequests {
		return fmt.Errorf("half-open success threshold (%d) cannot exceed max requests (%d)",
			c.HalfOpenSuccessThreshold, c.HalfOpenMaxRequests)
	}

	return nil
}
