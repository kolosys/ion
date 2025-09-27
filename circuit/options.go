package circuit

import (
	"time"

	"github.com/kolosys/ion/observe"
)

// Option is a function that configures a circuit breaker.
type Option func(*Config, *observe.Observability)

// WithFailureThreshold sets the number of consecutive failures required to trip the circuit.
func WithFailureThreshold(threshold int64) Option {
	return func(config *Config, obs *observe.Observability) {
		config.FailureThreshold = threshold
	}
}

// WithRecoveryTimeout sets the duration to wait in open state before attempting recovery.
func WithRecoveryTimeout(timeout time.Duration) Option {
	return func(config *Config, obs *observe.Observability) {
		config.RecoveryTimeout = timeout
	}
}

// WithHalfOpenMaxRequests sets the maximum number of requests allowed in half-open state.
func WithHalfOpenMaxRequests(maxRequests int64) Option {
	return func(config *Config, obs *observe.Observability) {
		config.HalfOpenMaxRequests = maxRequests
	}
}

// WithHalfOpenSuccessThreshold sets the number of successful requests required
// in half-open state to transition back to closed.
func WithHalfOpenSuccessThreshold(threshold int64) Option {
	return func(config *Config, obs *observe.Observability) {
		config.HalfOpenSuccessThreshold = threshold
	}
}

// WithFailurePredicate sets a custom predicate to determine what constitutes a failure.
// If not set, all non-nil errors are considered failures.
func WithFailurePredicate(isFailure func(error) bool) Option {
	return func(config *Config, obs *observe.Observability) {
		config.IsFailure = isFailure
	}
}

// WithStateChangeCallback sets a callback to be invoked on state changes.
func WithStateChangeCallback(callback func(from, to State)) Option {
	return func(config *Config, obs *observe.Observability) {
		config.OnStateChange = callback
	}
}

// WithObservability sets the observability hooks for logging, metrics, and tracing.
func WithObservability(observability *observe.Observability) Option {
	return func(config *Config, obs *observe.Observability) {
		*obs = *observability
	}
}

// WithLogger sets the logger for the circuit breaker.
func WithLogger(logger observe.Logger) Option {
	return func(config *Config, obs *observe.Observability) {
		obs.Logger = logger
	}
}

// WithMetrics sets the metrics recorder for the circuit breaker.
func WithMetrics(metrics observe.Metrics) Option {
	return func(config *Config, obs *observe.Observability) {
		obs.Metrics = metrics
	}
}

// WithTracer sets the tracer for the circuit breaker.
func WithTracer(tracer observe.Tracer) Option {
	return func(config *Config, obs *observe.Observability) {
		obs.Tracer = tracer
	}
}

// WithName is a convenience option that adds the circuit breaker name to log and metric tags.
// This is automatically handled by the New function, but can be useful for testing.
func WithName(name string) Option {
	return func(config *Config, obs *observe.Observability) {
		// Name is handled by the circuit breaker itself
		// This option exists for consistency with other ION components
	}
}

// Preset configurations for common use cases

// QuickFailover returns options for a circuit breaker that fails over quickly
// but also recovers quickly. Suitable for non-critical operations.
func QuickFailover() []Option {
	return []Option{
		WithFailureThreshold(3),
		WithRecoveryTimeout(10 * time.Second),
		WithHalfOpenMaxRequests(2),
		WithHalfOpenSuccessThreshold(1),
	}
}

// Conservative returns options for a circuit breaker that is slow to trip
// and slow to recover. Suitable for critical operations.
func Conservative() []Option {
	return []Option{
		WithFailureThreshold(10),
		WithRecoveryTimeout(60 * time.Second),
		WithHalfOpenMaxRequests(5),
		WithHalfOpenSuccessThreshold(3),
	}
}

// Aggressive returns options for a circuit breaker that trips quickly
// and takes time to recover. Suitable for protecting against cascading failures.
func Aggressive() []Option {
	return []Option{
		WithFailureThreshold(2),
		WithRecoveryTimeout(45 * time.Second),
		WithHalfOpenMaxRequests(1),
		WithHalfOpenSuccessThreshold(1),
	}
}
