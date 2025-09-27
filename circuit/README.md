# Circuit

[![Go Reference](https://pkg.go.dev/badge/github.com/kolosys/ion/circuit.svg)](https://pkg.go.dev/github.com/kolosys/ion/circuit)

Circuit breakers with threshold-based state transitions and automatic failure detection for protecting external service calls.

## Features

- **State Management**: Closed, Open, and Half-Open states with automatic transitions
- **Failure Detection**: Configurable failure predicates and thresholds
- **Recovery Testing**: Controlled recovery with success thresholds
- **Context-Aware**: All operations respect context cancellation and timeouts
- **Observability**: Comprehensive metrics, logging, and state change callbacks
- **Zero Dependencies**: No external dependencies beyond the Go standard library
- **Preset Configurations**: Quick setup with common patterns

## Quick Start

### Basic Circuit Breaker

```go
package main

import (
    "context"
    "fmt"
    "errors"

    "github.com/kolosys/ion/circuit"
)

func main() {
    // Create circuit breaker for payment service
    cb := circuit.New("payment-service",
        circuit.WithFailureThreshold(5),
        circuit.WithRecoveryTimeout(30*time.Second),
        circuit.WithHalfOpenMaxRequests(3),
    )

    // Protect external service calls
    result, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
        return paymentService.ProcessPayment(ctx, payment)
    })

    if err != nil {
        var circuitErr *circuit.CircuitError
        if errors.As(err, &circuitErr) && circuitErr.IsCircuitOpen() {
            // Circuit is open - handle degraded service
            return handlePaymentUnavailable()
        }
        return handlePaymentError(err)
    }

    // Use successful result
    fmt.Printf("Payment processed: %v\n", result)
}
```

### HTTP Client Protection

```go
// Protect HTTP client with circuit breaker
httpCircuit := circuit.New("external-api",
    circuit.WithFailureThreshold(3),
    circuit.WithRecoveryTimeout(15*time.Second),
    circuit.WithFailurePredicate(func(err error) bool {
        // Only count 5xx errors and timeouts as failures
        // 4xx errors (client errors) should not trip the circuit
        if httpErr, ok := err.(*HTTPError); ok {
            return httpErr.StatusCode >= 500
        }
        return true // Network errors count as failures
    }),
)

func makeHTTPRequest(ctx context.Context, url string) (*http.Response, error) {
    result, err := httpCircuit.Execute(ctx, func(ctx context.Context) (any, error) {
        req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
        if err != nil {
            return nil, err
        }
        return http.DefaultClient.Do(req)
    })

    if err != nil {
        return nil, err
    }

    return result.(*http.Response), nil
}
```

### Database Connection Protection

```go
// Protect database operations
dbCircuit := circuit.New("database",
    circuit.WithFailureThreshold(10),
    circuit.WithRecoveryTimeout(60*time.Second),
    circuit.WithStateChangeCallback(func(from, to circuit.State) {
        log.Printf("Database circuit: %s -> %s", from, to)

        if to == circuit.Open {
            // Switch to read-only replica or cache
            enableDegradedMode()
        } else if to == circuit.Closed {
            // Resume normal operations
            disableDegradedMode()
        }
    }),
)

func queryDatabase(ctx context.Context, query string) (*Result, error) {
    result, err := dbCircuit.Execute(ctx, func(ctx context.Context) (any, error) {
        return db.Query(ctx, query)
    })

    if err != nil {
        return nil, err
    }

    return result.(*Result), nil
}
```

## API Reference

### Circuit Breaker Creation

```go
func New(name string, options ...Option) CircuitBreaker
```

Creates a new circuit breaker with the given name and configuration options.

### Core Operations

```go
func (cb CircuitBreaker) Execute(ctx context.Context, fn func(context.Context) (any, error)) (any, error)
func (cb CircuitBreaker) Call(ctx context.Context, fn func(context.Context) error) error
func (cb CircuitBreaker) State() State
func (cb CircuitBreaker) Metrics() CircuitMetrics
func (cb CircuitBreaker) Reset()
func (cb CircuitBreaker) Close() error
```

**Execute** runs a function with circuit breaker protection.
**Call** is a convenience method for functions that don't return values.
**State** returns the current circuit state.
**Metrics** provides comprehensive circuit statistics.
**Reset** manually resets the circuit to closed state.
**Close** gracefully shuts down the circuit breaker.

## Configuration Options

### Basic Configuration

```go
circuit.WithFailureThreshold(5)                 // Failures before opening
circuit.WithRecoveryTimeout(30*time.Second)     // Wait time before half-open
circuit.WithHalfOpenMaxRequests(3)              // Max requests in half-open
circuit.WithHalfOpenSuccessThreshold(2)         // Successes needed to close
```

### Advanced Configuration

```go
circuit.WithFailurePredicate(func(err error) bool {
    // Custom logic to determine what counts as a failure
    return err != nil && !isRetryableError(err)
})

circuit.WithStateChangeCallback(func(from, to circuit.State) {
    // React to state changes
    log.Printf("Circuit %s -> %s", from, to)
})

circuit.WithObservability(observability)        // Complete observability setup
circuit.WithLogger(logger)                      // Custom logger
circuit.WithMetrics(metrics)                    // Custom metrics
circuit.WithTracer(tracer)                      // Custom tracer
```

### Preset Configurations

```go
// Quick failover for responsive services
circuit.QuickFailover()
// Equivalent to:
// WithFailureThreshold(3)
// WithRecoveryTimeout(5*time.Second)
// WithHalfOpenMaxRequests(1)

// Conservative for stable services
circuit.Conservative()
// Equivalent to:
// WithFailureThreshold(10)
// WithRecoveryTimeout(60*time.Second)
// WithHalfOpenMaxRequests(5)

// Aggressive for unreliable services
circuit.Aggressive()
// Equivalent to:
// WithFailureThreshold(2)
// WithRecoveryTimeout(10*time.Second)
// WithHalfOpenMaxRequests(1)
```

## States and Transitions

### Circuit States

```go
circuit.Closed    // Normal operation - all requests allowed
circuit.Open      // Failure mode - all requests fail fast
circuit.HalfOpen  // Recovery testing - limited requests allowed
```

### State Transitions

```
Closed --[failure threshold]--> Open
Open --[recovery timeout]--> HalfOpen
HalfOpen --[success threshold]--> Closed
HalfOpen --[any failure]--> Open
```

### State Behavior

**Closed State:**

- All requests are allowed through
- Failures are counted
- Transitions to Open when failure threshold is reached

**Open State:**

- All requests fail immediately with circuit open error
- No requests reach the protected service
- Transitions to Half-Open after recovery timeout

**Half-Open State:**

- Limited number of requests are allowed through
- Transitions to Closed after sufficient successes
- Transitions back to Open on any failure

## Metrics and Monitoring

### Circuit Metrics

```go
type CircuitMetrics struct {
    Name              string    // Circuit breaker name
    State             State     // Current state
    TotalRequests     int64     // Total requests processed
    TotalFailures     int64     // Total failed requests
    TotalSuccesses    int64     // Total successful requests
    ConsecutiveFails  int64     // Current consecutive failures
    StateChanges      int64     // Number of state transitions
    LastFailure       time.Time // Timestamp of last failure
    LastSuccess       time.Time // Timestamp of last success
    LastStateChange   time.Time // Timestamp of last state change
}

// Helper methods
func (m CircuitMetrics) FailureRate() float64    // 0.0 to 1.0
func (m CircuitMetrics) SuccessRate() float64    // 0.0 to 1.0
func (m CircuitMetrics) IsHealthy() bool         // Based on recent success rate
```

### Real-time Monitoring

```go
// Monitor circuit health
ticker := time.NewTicker(30 * time.Second)
go func() {
    for range ticker.C {
        metrics := cb.Metrics()
        log.Printf("Circuit %s: state=%s, failure_rate=%.2f%%, requests=%d",
            metrics.Name, metrics.State, metrics.FailureRate()*100, metrics.TotalRequests)
    }
}()
```

## Use Cases

### Microservice Communication

```go
// Protect inter-service calls
userServiceCircuit := circuit.New("user-service", circuit.QuickFailover()...)

func getUserProfile(ctx context.Context, userID string) (*UserProfile, error) {
    result, err := userServiceCircuit.Execute(ctx, func(ctx context.Context) (any, error) {
        return userServiceClient.GetProfile(ctx, userID)
    })

    if err != nil {
        // Return cached profile or default profile on circuit open
        if circuitErr, ok := err.(*circuit.CircuitError); ok && circuitErr.IsCircuitOpen() {
            return getCachedProfile(userID)
        }
        return nil, err
    }

    return result.(*UserProfile), nil
}
```

### External API Integration

```go
// Protect third-party API calls with custom failure detection
paymentCircuit := circuit.New("payment-gateway",
    circuit.WithFailureThreshold(5),
    circuit.WithRecoveryTimeout(45*time.Second),
    circuit.WithFailurePredicate(func(err error) bool {
        // Don't count validation errors as circuit failures
        if paymentErr, ok := err.(*PaymentError); ok {
            return paymentErr.Type != "validation_error"
        }
        return true
    }),
)

func processPayment(ctx context.Context, payment *Payment) (*PaymentResult, error) {
    result, err := paymentCircuit.Execute(ctx, func(ctx context.Context) (any, error) {
        return paymentGateway.Charge(ctx, payment)
    })

    if err != nil {
        return nil, fmt.Errorf("payment processing failed: %w", err)
    }

    return result.(*PaymentResult), nil
}
```

### Database Failover

```go
// Automatic failover to read replica
primaryDBCircuit := circuit.New("primary-db", circuit.Conservative()...)

func executeQuery(ctx context.Context, query string) (*Result, error) {
    // Try primary database first
    result, err := primaryDBCircuit.Execute(ctx, func(ctx context.Context) (any, error) {
        return primaryDB.Query(ctx, query)
    })

    if err != nil {
        var circuitErr *circuit.CircuitError
        if errors.As(err, &circuitErr) && circuitErr.IsCircuitOpen() {
            // Primary is down, use read replica
            log.Warn("Primary DB circuit open, using read replica")
            return readReplicaDB.Query(ctx, query)
        }
        return nil, err
    }

    return result.(*Result), nil
}
```

### Cascading Failure Prevention

```go
// Prevent cascading failures in service chains
func handleRequest(ctx context.Context, req *Request) (*Response, error) {
    // Each service call is protected by its own circuit
    userInfo, err := getUserInfo(ctx, req.UserID)
    if err != nil {
        return nil, err
    }

    permissions, err := getPermissions(ctx, req.UserID)
    if err != nil {
        // Continue with default permissions if service is down
        if isCircuitOpenError(err) {
            permissions = getDefaultPermissions()
        } else {
            return nil, err
        }
    }

    return processRequest(ctx, req, userInfo, permissions)
}
```

## Error Handling

### Circuit-Specific Errors

```go
import "github.com/kolosys/ion/circuit"

_, err := cb.Execute(ctx, riskyOperation)
if err != nil {
    var circuitErr *circuit.CircuitError
    if errors.As(err, &circuitErr) {
        switch {
        case circuitErr.IsCircuitOpen():
            // Circuit is open - service unavailable
            return handleServiceUnavailable()
        default:
            // Other circuit error
            return handleCircuitError(circuitErr)
        }
    }

    // Original error from the protected function
    return handleOperationError(err)
}
```

### Graceful Degradation

```go
func getRecommendations(ctx context.Context, userID string) ([]Recommendation, error) {
    result, err := recommendationCircuit.Execute(ctx, func(ctx context.Context) (any, error) {
        return mlService.GetRecommendations(ctx, userID)
    })

    if err != nil {
        var circuitErr *circuit.CircuitError
        if errors.As(err, &circuitErr) && circuitErr.IsCircuitOpen() {
            // ML service is down, return popular items
            log.Info("Recommendation service unavailable, using fallback")
            return getPopularItems(), nil
        }
        return nil, err
    }

    return result.([]Recommendation), nil
}
```

## Best Practices

### Failure Threshold Tuning

- **Responsive services**: 3-5 failures
- **Stable services**: 5-10 failures
- **Batch services**: 10-20 failures
- **External APIs**: 3-5 failures (you have less control)

### Recovery Timeout Guidelines

- **Fast recovery**: 5-15 seconds (for transient issues)
- **Moderate recovery**: 30-60 seconds (for service restarts)
- **Slow recovery**: 60-300 seconds (for deployment/scaling)

### Half-Open Configuration

- **Max requests**: 1-5 (limit blast radius during recovery)
- **Success threshold**: 1-3 (balance between quick recovery and stability)

### State Change Callbacks

```go
circuit.WithStateChangeCallback(func(from, to circuit.State) {
    // Log state changes
    log.Printf("Circuit %s: %s -> %s", cb.Name(), from, to)

    // Update metrics
    circuitStateGauge.WithLabelValues(cb.Name()).Set(float64(to))

    // Send alerts
    if to == circuit.Open {
        alerting.SendAlert("Circuit breaker opened", cb.Name())
    }
})
```

## Examples

- [Basic Usage](../examples/circuit/main.go) - Payment service protection
- [HTTP Client](../examples/circuit/main.go) - External API integration
- [Configuration Examples](../examples/circuit/main.go) - Different preset configurations
- [Recovery Scenarios](../examples/circuit/main.go) - State transition examples

## Performance

Benchmark results on modern hardware:

- **Execute (Closed)**: <100ns overhead
- **Execute (Open)**: <50ns (fast-fail)
- **State Check**: <10ns
- **Memory**: Minimal allocation overhead
- **Throughput**: 10M+ operations/second

## Thread Safety

All CircuitBreaker methods are safe for concurrent use. The implementation uses atomic operations for optimal performance under contention.

## Testing

```go
func TestCircuitBreaker(t *testing.T) {
    cb := circuit.New("test-circuit",
        circuit.WithFailureThreshold(2),
        circuit.WithRecoveryTimeout(100*time.Millisecond),
    )

    // Trigger failures to open circuit
    for i := 0; i < 3; i++ {
        _, err := cb.Execute(context.Background(), func(ctx context.Context) (any, error) {
            return nil, errors.New("failure")
        })
        assert.Error(t, err)
    }

    // Verify circuit is open
    assert.Equal(t, circuit.Open, cb.State())

    // Test fast-fail behavior
    _, err := cb.Execute(context.Background(), func(ctx context.Context) (any, error) {
        t.Error("Should not execute when circuit is open")
        return nil, nil
    })

    var circuitErr *circuit.CircuitError
    assert.True(t, errors.As(err, &circuitErr))
    assert.True(t, circuitErr.IsCircuitOpen())
}
```

## Contributing

See the main [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

## License

Licensed under the [MIT License](../LICENSE).
