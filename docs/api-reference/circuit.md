# circuit API

Complete API documentation for the circuit package.

**Import Path:** `github.com/kolosys/ion/circuit`

## Package Documentation

Package circuit provides circuit breaker functionality for resilient microservices.
Circuit breakers prevent cascading failures by temporarily blocking requests to failing services,
allowing them time to recover while providing fast-fail behavior to callers.

The circuit breaker implements a three-state machine:
- Closed: Normal operation, requests pass through
- Open: Circuit is tripped, requests fail fast
- Half-Open: Testing recovery, limited requests allowed

Usage:

	cb := circuit.New("payment-service",
		circuit.WithFailureThreshold(5),
		circuit.WithRecoveryTimeout(30*time.Second),
	)

	result, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
		return paymentService.ProcessPayment(ctx, payment)
	})

The circuit breaker integrates with ION's observability system and supports
context cancellation, timeouts, and comprehensive metrics collection.


## Types

### CircuitBreaker
CircuitBreaker represents a circuit breaker that controls access to a potentially failing operation. It provides fast-fail behavior when the operation is failing and automatic recovery testing when appropriate.

#### Example Usage

```go
// Example implementation of CircuitBreaker
type MyCircuitBreaker struct {
    // Add your fields here
}

func (m MyCircuitBreaker) Execute(param1 context.Context, param2 func(context.Context) (any, error)) any {
    // Implement your logic here
    return
}

func (m MyCircuitBreaker) Call(param1 context.Context, param2 func(context.Context) error) error {
    // Implement your logic here
    return
}

func (m MyCircuitBreaker) State() State {
    // Implement your logic here
    return
}

func (m MyCircuitBreaker) Metrics() CircuitMetrics {
    // Implement your logic here
    return
}

func (m MyCircuitBreaker) Reset()  {
    // Implement your logic here
    return
}

func (m MyCircuitBreaker) Close() error {
    // Implement your logic here
    return
}


```

#### Type Definition

```go
type CircuitBreaker interface {
    Execute(ctx context.Context, fn func(context.Context) (any, error)) (any, error)
    Call(ctx context.Context, fn func(context.Context) error) error
    State() State
    Metrics() CircuitMetrics
    Reset()
    Close() error
}
```

## Methods

| Method | Description |
| ------ | ----------- |

### Constructor Functions

### New

New creates a new circuit breaker with the given name and options.

```go
func New(name string, options ...Option) CircuitBreaker
```

**Parameters:**
- `name` (string)
- `options` (...Option)

**Returns:**
- CircuitBreaker

### CircuitError
CircuitError represents circuit breaker specific errors with context

#### Example Usage

```go
// Create a new CircuitError
circuiterror := CircuitError{
    Op: "example",
    CircuitName: "example",
    State: "example",
    Err: error{},
}
```

#### Type Definition

```go
type CircuitError struct {
    Op string
    CircuitName string
    State string
    Err error
}
```

### Fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| Op | `string` | operation that failed |
| CircuitName | `string` | name of the circuit breaker |
| State | `string` | current state of the circuit |
| Err | `error` | underlying error |

## Methods

### Error



```go
func (*CircuitError) Error() string
```

**Parameters:**
  None

**Returns:**
- string

### IsCircuitOpen

IsCircuitOpen returns true if the error is due to an open circuit.

```go
func (*CircuitError) IsCircuitOpen() bool
```

**Parameters:**
  None

**Returns:**
- bool

### Unwrap



```go
func (*CircuitError) Unwrap() error
```

**Parameters:**
  None

**Returns:**
- error

### CircuitMetrics
CircuitMetrics holds metrics for a circuit breaker instance.

#### Example Usage

```go
// Create a new CircuitMetrics
circuitmetrics := CircuitMetrics{
    Name: "example",
    State: State{},
    TotalRequests: 42,
    TotalFailures: 42,
    TotalSuccesses: 42,
    ConsecutiveFails: 42,
    StateChanges: 42,
    LastFailure: /* value */,
    LastSuccess: /* value */,
    LastStateChange: /* value */,
}
```

#### Type Definition

```go
type CircuitMetrics struct {
    Name string
    State State
    TotalRequests int64
    TotalFailures int64
    TotalSuccesses int64
    ConsecutiveFails int64
    StateChanges int64
    LastFailure time.Time
    LastSuccess time.Time
    LastStateChange time.Time
}
```

### Fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| Name | `string` | Name is the name of the circuit breaker |
| State | `State` | State is the current state of the circuit |
| TotalRequests | `int64` | TotalRequests is the total number of requests processed |
| TotalFailures | `int64` | TotalFailures is the total number of failed requests |
| TotalSuccesses | `int64` | TotalSuccesses is the total number of successful requests |
| ConsecutiveFails | `int64` | ConsecutiveFails is the current count of consecutive failures |
| StateChanges | `int64` | StateChanges is the total number of state transitions |
| LastFailure | `time.Time` | LastFailure is the timestamp of the last failure |
| LastSuccess | `time.Time` | LastSuccess is the timestamp of the last success |
| LastStateChange | `time.Time` | LastStateChange is the timestamp of the last state change |

## Methods

### FailureRate

FailureRate returns the failure rate as a percentage (0.0 to 1.0).

```go
func (CircuitMetrics) FailureRate() float64
```

**Parameters:**
  None

**Returns:**
- float64

### IsHealthy

IsHealthy returns true if the circuit appears to be healthy based on recent activity.

```go
func (CircuitMetrics) IsHealthy() bool
```

**Parameters:**
  None

**Returns:**
- bool

### SuccessRate

SuccessRate returns the success rate as a percentage (0.0 to 1.0).

```go
func (CircuitMetrics) SuccessRate() float64
```

**Parameters:**
  None

**Returns:**
- float64

### Config
Config holds configuration for a circuit breaker.

#### Example Usage

```go
// Create a new Config
config := Config{
    FailureThreshold: 42,
    RecoveryTimeout: /* value */,
    HalfOpenMaxRequests: 42,
    HalfOpenSuccessThreshold: 42,
    IsFailure: /* value */,
    OnStateChange: /* value */,
}
```

#### Type Definition

```go
type Config struct {
    FailureThreshold int64
    RecoveryTimeout time.Duration
    HalfOpenMaxRequests int64
    HalfOpenSuccessThreshold int64
    IsFailure func(error) bool
    OnStateChange func(from, to State)
}
```

### Fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| FailureThreshold | `int64` | FailureThreshold is the number of consecutive failures required to trip the circuit. Default: 5 |
| RecoveryTimeout | `time.Duration` | RecoveryTimeout is the duration to wait in the open state before transitioning to half-open for recovery testing. Default: 30 seconds |
| HalfOpenMaxRequests | `int64` | HalfOpenMaxRequests is the maximum number of requests allowed in half-open state. Default: 3 |
| HalfOpenSuccessThreshold | `int64` | HalfOpenSuccessThreshold is the number of successful requests required in half-open state to transition back to closed. Default: 2 |
| IsFailure | `func(error) bool` | IsFailure is a predicate function that determines if an error should be counted as a failure for circuit breaker purposes. If nil, all non-nil errors are considered failures. |
| OnStateChange | `func(from, to State)` | OnStateChange is called whenever the circuit breaker changes state. This is useful for logging or metrics collection. |

### Constructor Functions

### DefaultConfig

DefaultConfig returns a Config with sensible defaults.

```go
func DefaultConfig() *Config
```

**Parameters:**
  None

**Returns:**
- *Config

## Methods

### Validate

Validate checks if the configuration is valid and returns an error if not.

```go
func (*Config) Validate() error
```

**Parameters:**
  None

**Returns:**
- error

### Option
Option is a function that configures a circuit breaker.

#### Example Usage

```go
// Example usage of Option
var value Option
// Initialize with appropriate value
```

#### Type Definition

```go
type Option func(*Config, *observe.Observability)
```

### Constructor Functions

### Aggressive

Aggressive returns options for a circuit breaker that trips quickly and takes time to recover. Suitable for protecting against cascading failures.

```go
func Aggressive() []Option
```

**Parameters:**
  None

**Returns:**
- []Option

### Conservative

Conservative returns options for a circuit breaker that is slow to trip and slow to recover. Suitable for critical operations.

```go
func Conservative() []Option
```

**Parameters:**
  None

**Returns:**
- []Option

### QuickFailover

QuickFailover returns options for a circuit breaker that fails over quickly but also recovers quickly. Suitable for non-critical operations.

```go
func QuickFailover() []Option
```

**Parameters:**
  None

**Returns:**
- []Option

### WithFailurePredicate

WithFailurePredicate sets a custom predicate to determine what constitutes a failure. If not set, all non-nil errors are considered failures.

```go
func WithFailurePredicate(isFailure func(error) bool) Option
```

**Parameters:**
- `isFailure` (func(error) bool)

**Returns:**
- Option

### WithFailureThreshold

WithFailureThreshold sets the number of consecutive failures required to trip the circuit.

```go
func WithFailureThreshold(threshold int64) Option
```

**Parameters:**
- `threshold` (int64)

**Returns:**
- Option

### WithHalfOpenMaxRequests

WithHalfOpenMaxRequests sets the maximum number of requests allowed in half-open state.

```go
func WithHalfOpenMaxRequests(maxRequests int64) Option
```

**Parameters:**
- `maxRequests` (int64)

**Returns:**
- Option

### WithHalfOpenSuccessThreshold

WithHalfOpenSuccessThreshold sets the number of successful requests required in half-open state to transition back to closed.

```go
func WithHalfOpenSuccessThreshold(threshold int64) Option
```

**Parameters:**
- `threshold` (int64)

**Returns:**
- Option

### WithLogger

WithLogger sets the logger for the circuit breaker.

```go
func WithLogger(logger observe.Logger) Option
```

**Parameters:**
- `logger` (observe.Logger)

**Returns:**
- Option

### WithMetrics

WithMetrics sets the metrics recorder for the circuit breaker.

```go
func WithMetrics(metrics observe.Metrics) Option
```

**Parameters:**
- `metrics` (observe.Metrics)

**Returns:**
- Option

### WithName

WithName is a convenience option that adds the circuit breaker name to log and metric tags. This is automatically handled by the New function, but can be useful for testing.

```go
func WithName(name string) Option
```

**Parameters:**
- `name` (string)

**Returns:**
- Option

### WithObservability

WithObservability sets the observability hooks for logging, metrics, and tracing.

```go
func WithObservability(observability *observe.Observability) Option
```

**Parameters:**
- `observability` (*observe.Observability)

**Returns:**
- Option

### WithRecoveryTimeout

WithRecoveryTimeout sets the duration to wait in open state before attempting recovery.

```go
func WithRecoveryTimeout(timeout time.Duration) Option
```

**Parameters:**
- `timeout` (time.Duration)

**Returns:**
- Option

### WithStateChangeCallback

WithStateChangeCallback sets a callback to be invoked on state changes.

```go
func WithStateChangeCallback(callback func(from, to State)) Option
```

**Parameters:**
- `callback` (func(from, to State))

**Returns:**
- Option

### WithTracer

WithTracer sets the tracer for the circuit breaker.

```go
func WithTracer(tracer observe.Tracer) Option
```

**Parameters:**
- `tracer` (observe.Tracer)

**Returns:**
- Option

### State
State represents the current state of a circuit breaker.

#### Example Usage

```go
// Example usage of State
var value State
// Initialize with appropriate value
```

#### Type Definition

```go
type State int32
```

## Methods

### String

String returns the string representation of the circuit state.

```go
func (State) String() string
```

**Parameters:**
  None

**Returns:**
- string

## Functions

### NewCircuitOpenError
NewCircuitOpenError creates an error indicating the circuit is open

```go
func NewCircuitOpenError(circuitName string) error
```

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `circuitName` | `string` | |

**Returns:**
| Type | Description |
|------|-------------|
| `error` | |

**Example:**

```go
// Example usage of NewCircuitOpenError
result := NewCircuitOpenError(/* parameters */)
```

### NewCircuitTimeoutError
NewCircuitTimeoutError creates an error indicating a circuit operation timed out

```go
func NewCircuitTimeoutError(circuitName string) error
```

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `circuitName` | `string` | |

**Returns:**
| Type | Description |
|------|-------------|
| `error` | |

**Example:**

```go
// Example usage of NewCircuitTimeoutError
result := NewCircuitTimeoutError(/* parameters */)
```

## External Links

- [Package Overview](../packages/circuit.md)
- [pkg.go.dev Documentation](https://pkg.go.dev/github.com/kolosys/ion/circuit)
- [Source Code](https://github.com/kolosys/ion/tree/main/circuit)
