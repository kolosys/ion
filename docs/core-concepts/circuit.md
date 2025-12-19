# Circuit Breaker

**Import Path:** `github.com/kolosys/ion/circuit`

The circuit breaker pattern prevents cascading failures in distributed systems by temporarily blocking requests to failing services, allowing them time to recover while providing fast-fail behavior to callers.

## Overview

Circuit breakers are essential for building resilient microservices. They automatically detect failures and "trip" to prevent overwhelming a failing service, then automatically test recovery when appropriate.

### When to Use Circuit Breakers

- **External Service Calls**: HTTP clients, gRPC services, third-party APIs
- **Database Operations**: Connection failures, query timeouts
- **Cache Access**: Redis, Memcached, or other cache failures
- **Microservice Communication**: Inter-service calls in distributed systems
- **Resource-Intensive Operations**: Operations that can fail under load

## Architecture

The circuit breaker implements a three-state state machine:

```
┌─────────┐  failures ≥ threshold   ┌─────────┐
│ Closed  │ ──────────────────────> │  Open   │
│         │ <─────────────────────  │         │
└─────────┘  recovery timeout       └─────────┘
     ▲                                    │
     │                                    │
     │  successes ≥ threshold             │
     └────────────────────────────────────┘
              ┌───────────┐
              │ Half-Open │
              └───────────┘
```

### States

1. **Closed**: Normal operation, requests pass through

   - Tracks consecutive failures
   - Transitions to Open when failure threshold is reached

2. **Open**: Circuit is tripped, requests fail fast

   - Immediately rejects requests
   - Transitions to Half-Open after recovery timeout

3. **Half-Open**: Testing recovery, limited requests allowed
   - Allows a small number of test requests
   - Transitions to Closed on success threshold
   - Transitions back to Open on any failure

## Core Concepts

### Failure Detection

The circuit breaker tracks consecutive failures and trips when the threshold is reached:

```go
cb := circuit.New("payment-service",
    circuit.WithFailureThreshold(5), // Trip after 5 consecutive failures
)
```

### Recovery Testing

After the recovery timeout, the circuit enters Half-Open state to test if the service has recovered:

```go
cb := circuit.New("payment-service",
    circuit.WithRecoveryTimeout(30*time.Second), // Wait 30s before testing
    circuit.WithHalfOpenMaxRequests(3),          // Allow 3 test requests
    circuit.WithHalfOpenSuccessThreshold(2),     // Need 2 successes to close
)
```

### Failure Predicates

Customize what counts as a failure:

```go
cb := circuit.New("http-client",
    circuit.WithFailurePredicate(func(err error) bool {
        // Only count 5xx errors and timeouts as failures
        // 4xx errors (client errors) should not trip the circuit
        if err == nil {
            return false
        }
        // In a real implementation, check HTTP status code
        return true
    }),
)
```

## Real-World Scenarios

### Scenario 1: Payment Service Protection

Protect a payment processing service from cascading failures:

```go
package main

import (
    "context"
    "errors"
    "time"

    "github.com/kolosys/ion/circuit"
)

type PaymentService struct {
    circuit circuit.CircuitBreaker
}

func NewPaymentService() *PaymentService {
    return &PaymentService{
        circuit: circuit.New("payment-service",
            circuit.WithFailureThreshold(5),
            circuit.WithRecoveryTimeout(30*time.Second),
            circuit.WithHalfOpenMaxRequests(2),
            circuit.WithHalfOpenSuccessThreshold(1),
            circuit.WithStateChangeCallback(func(from, to circuit.State) {
                logStateChange("payment-service", from, to)
            }),
        ),
    }
}

func (ps *PaymentService) ProcessPayment(ctx context.Context, amount float64) error {
    _, err := ps.circuit.Execute(ctx, func(ctx context.Context) (any, error) {
        // Call actual payment service
        return callPaymentAPI(ctx, amount)
    })

    if err != nil {
        if circuit.IsCircuitOpen(err) {
            // Circuit is open - return user-friendly error
            return errors.New("payment service temporarily unavailable, please try again later")
        }
        return err
    }

    return nil
}

func callPaymentAPI(ctx context.Context, amount float64) (string, error) {
    // Actual API call implementation
    return "payment-id", nil
}

func logStateChange(name string, from, to circuit.State) {
    // Log state changes for monitoring
    fmt.Printf("Circuit %s: %s -> %s\n", name, from, to)
}
```

### Scenario 2: Database Connection Protection

Protect database operations from connection pool exhaustion:

```go
package main

import (
    "context"
    "database/sql"
    "time"

    "github.com/kolosys/ion/circuit"
)

type ProtectedDB struct {
    db      *sql.DB
    circuit circuit.CircuitBreaker
}

func NewProtectedDB(dsn string) (*ProtectedDB, error) {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }

    return &ProtectedDB{
        db: db,
        circuit: circuit.New("database",
            circuit.WithFailureThreshold(3),
            circuit.WithRecoveryTimeout(10*time.Second),
            circuit.WithFailurePredicate(func(err error) bool {
                // Only count connection errors, not query errors
                return err == sql.ErrConnDone || err == context.DeadlineExceeded
            }),
        ),
    }, nil
}

func (pdb *ProtectedDB) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
    result, err := pdb.circuit.Execute(ctx, func(ctx context.Context) (any, error) {
        return pdb.db.QueryContext(ctx, query, args...)
    })

    if err != nil {
        return nil, err
    }

    return result.(*sql.Rows), nil
}
```

### Scenario 3: Multi-Service Circuit Breaker

Protect multiple services with separate circuit breakers:

```go
package main

import (
    "context"
    "time"

    "github.com/kolosys/ion/circuit"
)

type ServiceMesh struct {
    userService    circuit.CircuitBreaker
    orderService   circuit.CircuitBreaker
    paymentService circuit.CircuitBreaker
}

func NewServiceMesh() *ServiceMesh {
    return &ServiceMesh{
        userService: circuit.New("user-service",
            circuit.WithFailureThreshold(5),
            circuit.WithRecoveryTimeout(30*time.Second),
        ),
        orderService: circuit.New("order-service",
            circuit.WithFailureThreshold(3),
            circuit.WithRecoveryTimeout(20*time.Second),
        ),
        paymentService: circuit.New("payment-service",
            circuit.WithFailureThreshold(2), // More sensitive
            circuit.WithRecoveryTimeout(60*time.Second), // Longer recovery
        ),
    }
}

func (sm *ServiceMesh) ProcessOrder(ctx context.Context, orderID string) error {
    // Check user service
    _, err := sm.userService.Execute(ctx, func(ctx context.Context) (any, error) {
        return validateUser(ctx, orderID)
    })
    if err != nil {
        return err
    }

    // Check order service
    _, err = sm.orderService.Execute(ctx, func(ctx context.Context) (any, error) {
        return validateOrder(ctx, orderID)
    })
    if err != nil {
        return err
    }

    // Process payment
    _, err = sm.paymentService.Execute(ctx, func(ctx context.Context) (any, error) {
        return processPayment(ctx, orderID)
    })
    return err
}
```

### Scenario 4: HTTP Client with Retry Logic

Combine circuit breaker with retry logic:

```go
package main

import (
    "context"
    "net/http"
    "time"

    "github.com/kolosys/ion/circuit"
)

type ResilientHTTPClient struct {
    client  *http.Client
    circuit circuit.CircuitBreaker
}

func NewResilientHTTPClient() *ResilientHTTPClient {
    return &ResilientHTTPClient{
        client: &http.Client{
            Timeout: 5 * time.Second,
        },
        circuit: circuit.New("http-client",
            circuit.WithFailureThreshold(3),
            circuit.WithRecoveryTimeout(15*time.Second),
        ),
    }
}

func (c *ResilientHTTPClient) GetWithRetry(ctx context.Context, url string, maxRetries int) (*http.Response, error) {
    var lastErr error

    for i := 0; i < maxRetries; i++ {
        result, err := c.circuit.Execute(ctx, func(ctx context.Context) (any, error) {
            req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
            if err != nil {
                return nil, err
            }
            return c.client.Do(req)
        })

        if err == nil {
            return result.(*http.Response), nil
        }

        // If circuit is open, don't retry
        if circuit.IsCircuitOpen(err) {
            return nil, err
        }

        lastErr = err
        time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
    }

    return nil, lastErr
}
```

## Configuration Presets

Ion provides preset configurations for common scenarios:

### Quick Failover

Fast to trip, fast to recover - suitable for non-critical operations:

```go
cb := circuit.New("non-critical-service", circuit.QuickFailover()...)
```

### Conservative

Slow to trip, slow to recover - suitable for critical operations:

```go
cb := circuit.New("critical-service", circuit.Conservative()...)
```

### Aggressive

Quick to trip, slow to recover - suitable for protecting against cascading failures:

```go
cb := circuit.New("protected-service", circuit.Aggressive()...)
```

## Observability

Circuit breakers integrate with Ion's observability system:

```go
import (
    "github.com/kolosys/ion/circuit"
    "github.com/kolosys/ion/observe"
)

obs := observe.New().
    WithLogger(myLogger).
    WithMetrics(myMetrics).
    WithTracer(myTracer)

cb := circuit.New("service",
    circuit.WithObservability(obs),
)

// Metrics are automatically collected:
// - circuit.requests_total
// - circuit.requests_succeeded
// - circuit.requests_failed
// - circuit.requests_rejected
// - circuit.state_changes
// - circuit.request_duration
```

## Metrics and Monitoring

Access circuit breaker metrics:

```go
metrics := cb.Metrics()

fmt.Printf("State: %s\n", metrics.State)
fmt.Printf("Total Requests: %d\n", metrics.TotalRequests)
fmt.Printf("Successes: %d\n", metrics.TotalSuccesses)
fmt.Printf("Failures: %d\n", metrics.TotalFailures)
fmt.Printf("Failure Rate: %.2f%%\n", metrics.FailureRate()*100)
fmt.Printf("State Changes: %d\n", metrics.StateChanges)
```

## Best Practices

1. **Choose Appropriate Thresholds**: Balance between sensitivity and false positives
2. **Monitor State Changes**: Log state transitions for debugging
3. **Use Failure Predicates**: Distinguish between transient and permanent failures
4. **Set Reasonable Timeouts**: Recovery timeout should match service recovery time
5. **Combine with Retries**: Use circuit breakers with exponential backoff retries
6. **Monitor Metrics**: Track circuit breaker metrics in your observability system

## Common Pitfalls

### Pitfall 1: Too Sensitive Thresholds

**Problem**: Circuit trips on normal transient failures

```go
// Too sensitive
cb := circuit.New("service", circuit.WithFailureThreshold(1))
```

**Solution**: Use appropriate thresholds based on your service's failure characteristics

```go
// Better
cb := circuit.New("service", circuit.WithFailureThreshold(5))
```

### Pitfall 2: Ignoring Circuit State

**Problem**: Not handling circuit open errors appropriately

```go
// Bad
result, err := cb.Execute(ctx, fn)
if err != nil {
    return err // User sees circuit breaker error
}
```

**Solution**: Provide user-friendly error messages

```go
// Good
result, err := cb.Execute(ctx, fn)
if err != nil {
    if circuit.IsCircuitOpen(err) {
        return errors.New("service temporarily unavailable")
    }
    return err
}
```

### Pitfall 3: Not Using Failure Predicates

**Problem**: Client errors (4xx) trip the circuit

**Solution**: Use failure predicates to distinguish error types

```go
cb := circuit.New("http-client",
    circuit.WithFailurePredicate(func(err error) bool {
        // Only count server errors (5xx) and timeouts
        return isServerError(err) || isTimeout(err)
    }),
)
```

## Integration Guide

### With HTTP Clients

```go
type HTTPClient struct {
    client  *http.Client
    circuit circuit.CircuitBreaker
}

func (c *HTTPClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
    result, err := c.circuit.Execute(ctx, func(ctx context.Context) (any, error) {
        return c.client.Do(req.WithContext(ctx))
    })
    if err != nil {
        return nil, err
    }
    return result.(*http.Response), nil
}
```

### With gRPC

```go
func (c *gRPCClient) Call(ctx context.Context, method string, req, resp any) error {
    _, err := c.circuit.Execute(ctx, func(ctx context.Context) (any, error) {
        return c.conn.Invoke(ctx, method, req, resp)
    })
    return err
}
```

### With Database Operations

```go
func (db *DB) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
    result, err := db.circuit.Execute(ctx, func(ctx context.Context) (any, error) {
        return db.conn.QueryContext(ctx, query, args...)
    })
    if err != nil {
        return nil, err
    }
    return result.(*sql.Rows), nil
}
```

## Further Reading

- [API Reference](../api-reference/circuit.md) - Complete API documentation
- [Examples](../examples/circuit/) - Practical examples
- [Best Practices](../advanced/best-practices.md) - Recommended patterns
- [Performance Tuning](../advanced/performance-tuning.md) - Optimization strategies
