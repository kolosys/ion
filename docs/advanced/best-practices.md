# Best Practices

This guide covers best practices for using Ion effectively in production systems.

## General Principles

### 1. Always Use Context

All Ion components support `context.Context` for cancellation and timeouts. Always use it:

```go
// Good: With timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

result, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
    return operation(ctx)
})
```

**Why**: Context provides cancellation, timeouts, and request-scoped values essential for production systems.

### 2. Handle Errors Explicitly

Always check and handle errors returned by Ion components:

```go
// Good: Explicit error handling
if err := pool.Submit(ctx, task); err != nil {
    if workerpool.IsQueueFull(err) {
        // Handle queue full
        return handleQueueFull()
    }
    return fmt.Errorf("failed to submit task: %w", err)
}
```

**Why**: Errors provide important information about component state and failures.

### 3. Use Defer for Cleanup

Always use `defer` for releasing resources:

```go
// Good: Always released
if err := sem.Acquire(ctx, 1); err != nil {
    return err
}
defer sem.Release(1)
```

**Why**: Ensures resources are always released, even if errors occur.

### 4. Configure Observability

Always configure observability in production:

```go
// Good: Observability configured
obs := observe.New().
    WithLogger(myLogger).
    WithMetrics(myMetrics).
    WithTracer(myTracer)

cb := circuit.New("service", circuit.WithObservability(obs))
```

**Why**: Observability is essential for monitoring, debugging, and understanding system behavior.

## Circuit Breaker Best Practices

### Choose Appropriate Thresholds

Balance between sensitivity and false positives:

```go
// For critical services: higher threshold
cb := circuit.New("payment-service",
    circuit.WithFailureThreshold(10), // More tolerant
    circuit.WithRecoveryTimeout(60*time.Second),
)

// For non-critical services: lower threshold
cb := circuit.New("analytics-service",
    circuit.WithFailureThreshold(3), // More sensitive
    circuit.WithRecoveryTimeout(10*time.Second),
)
```

### Use Failure Predicates

Distinguish between transient and permanent failures:

```go
cb := circuit.New("http-client",
    circuit.WithFailurePredicate(func(err error) bool {
        // Only count 5xx errors and timeouts as failures
        // 4xx errors (client errors) should not trip the circuit
        if err == nil {
            return false
        }
        // Check HTTP status code or error type
        return isServerError(err) || isTimeout(err)
    }),
)
```

### Monitor Circuit State

Log state changes for debugging:

```go
cb := circuit.New("service",
    circuit.WithStateChangeCallback(func(from, to circuit.State) {
        logger.Info("circuit state changed",
            "name", "service",
            "from", from,
            "to", to,
        )
    }),
)
```

### Provide User-Friendly Errors

Don't expose circuit breaker internals to users:

```go
result, err := cb.Execute(ctx, fn)
if err != nil {
    if circuit.IsCircuitOpen(err) {
        // User-friendly error
        return errors.New("service temporarily unavailable, please try again later")
    }
    return err
}
```

## Rate Limiting Best Practices

### Size Rates Appropriately

Set rates based on actual usage patterns:

```go
// Analyze your traffic patterns first
// Then set rates accordingly
limiter := ratelimit.NewTokenBucket(
    ratelimit.PerSecond(100), // Based on actual capacity
    200,                      // Allow 2x burst
)
```

### Use Context Timeouts

Always use context with timeouts for `WaitN`:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := limiter.WaitN(ctx, 1); err != nil {
    if err == context.DeadlineExceeded {
        // Handle timeout
    }
    return err
}
```

### Choose the Right Algorithm

- **Token Bucket**: For APIs that need to handle bursts
- **Leaky Bucket**: For steady processing rates
- **Multi-Tier**: For complex rate limiting needs

```go
// Token bucket for API client
apiLimiter := ratelimit.NewTokenBucket(ratelimit.PerSecond(10), 20)

// Leaky bucket for job processing
jobLimiter := ratelimit.NewLeakyBucket(ratelimit.PerSecond(5), 50)
```

### Handle Rate Limit Errors Gracefully

Provide clear error messages:

```go
if !limiter.AllowN(time.Now(), 1) {
    return errors.New("rate limit exceeded, please try again later")
}
```

## Semaphore Best Practices

### Always Release Permits

Use `defer` to ensure permits are always released:

```go
if err := sem.Acquire(ctx, 1); err != nil {
    return err
}
defer sem.Release(1) // Always released
```

### Size Capacity Correctly

Base capacity on actual resource constraints:

```go
// Based on database connection limit
sem := semaphore.NewWeighted(int64(db.MaxOpenConns()))

// Based on memory constraints
sem := semaphore.NewWeighted(calculateMemoryCapacity())
```

### Choose Appropriate Fairness

- **FIFO**: When fairness is important
- **LIFO**: When recent requests are more important
- **None**: When maximum performance is critical

```go
// FIFO for database connections (fairness important)
dbSem := semaphore.NewWeighted(10, semaphore.WithFairness(semaphore.FIFO))

// None for high-performance operations
perfSem := semaphore.NewWeighted(100, semaphore.WithFairness(semaphore.None))
```

### Use Weighted Permits

Use weighted permits for operations with different resource needs:

```go
sem := semaphore.NewWeighted(10)

// Small operation: 1 unit
sem.Acquire(ctx, 1)

// Large operation: 5 units
sem.Acquire(ctx, 5)
```

## Worker Pool Best Practices

### Size Pools Appropriately

Balance workers and queue size:

```go
// Too few workers: underutilized
pool := workerpool.New(1, 100) // Bad

// Too many workers: resource contention
pool := workerpool.New(1000, 10) // Bad

// Balanced: based on workload
pool := workerpool.New(10, 100) // Good
```

### Always Close Pools

Use `defer` to ensure pools are closed:

```go
pool := workerpool.New(4, 20)
defer pool.Close(context.Background())
```

### Handle Queue Full

Check for queue full errors:

```go
if err := pool.Submit(ctx, task); err != nil {
    if workerpool.IsQueueFull(err) {
        // Handle: retry, reject, or backpressure
        return handleQueueFull()
    }
    return err
}
```

### Use Panic Recovery

Always use panic recovery in production:

```go
pool := workerpool.New(4, 20,
    workerpool.WithPanicRecovery(func(r any) {
        logger.Error("panic recovered", "panic", r)
        // Report to error tracking service
    }),
)
```

### Respect Context in Tasks

Always check context in tasks:

```go
pool.Submit(ctx, func(taskCtx context.Context) error {
    select {
    case <-taskCtx.Done():
        return taskCtx.Err()
    default:
        // Process task
    }
    return nil
})
```

## Observability Best Practices

### Use Structured Logging

Pass key-value pairs for better log analysis:

```go
logger.Info("circuit state changed",
    "name", "payment-service",
    "from", "Closed",
    "to", "Open",
    "failures", 5,
)
```

### Consistent Metric Naming

Use consistent naming conventions:

```go
// Good: Consistent prefix and naming
metrics.Inc("ion_circuit_requests_total", "name", "payment-service")
metrics.Inc("ion_circuit_requests_failed", "name", "payment-service")
metrics.Histogram("ion_circuit_request_duration", duration, "name", "payment-service")
```

### Propagate Context

Ensure context propagates through all operations:

```go
result, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
    // Context propagates to nested operations
    return service.Call(ctx, req)
})
```

### Monitor Key Metrics

Track important metrics:

- Circuit breaker: state changes, failure rates
- Rate limiter: limit hits, wait times
- Semaphore: wait times, permit usage
- Worker pool: queue depth, completion rates

## Error Handling Best Practices

### Distinguish Error Types

Check for specific error types:

```go
if err := pool.Submit(ctx, task); err != nil {
    switch {
    case workerpool.IsQueueFull(err):
        // Handle queue full
    case workerpool.IsPoolClosed(err):
        // Handle pool closed
    default:
        // Handle other errors
    }
}
```

### Wrap Errors with Context

Add context to errors:

```go
if err := sem.Acquire(ctx, 1); err != nil {
    return fmt.Errorf("failed to acquire semaphore permit: %w", err)
}
```

### Don't Expose Internal Errors

Provide user-friendly error messages:

```go
result, err := cb.Execute(ctx, fn)
if err != nil {
    if circuit.IsCircuitOpen(err) {
        // User-friendly error
        return errors.New("service temporarily unavailable")
    }
    // Log internal error
    logger.Error("circuit breaker error", "error", err)
    return errors.New("operation failed")
}
```

## Configuration Best Practices

### Use Functional Options

Ion uses functional options for flexible configuration:

```go
cb := circuit.New("service",
    circuit.WithFailureThreshold(5),
    circuit.WithRecoveryTimeout(30*time.Second),
    circuit.WithLogger(myLogger),
    circuit.WithMetrics(myMetrics),
)
```

### Create Configuration Helpers

Create helpers for common configurations:

```go
func NewPaymentServiceCircuit() circuit.CircuitBreaker {
    return circuit.New("payment-service",
        circuit.WithFailureThreshold(5),
        circuit.WithRecoveryTimeout(30*time.Second),
        circuit.WithHalfOpenMaxRequests(2),
        circuit.WithHalfOpenSuccessThreshold(1),
    )
}
```

### Validate Configuration

Validate configuration before use:

```go
config := circuit.DefaultConfig()
config.FailureThreshold = 5
if err := config.Validate(); err != nil {
    return fmt.Errorf("invalid configuration: %w", err)
}
```

## Testing Best Practices

### Test Error Paths

Always test error handling:

```go
func TestCircuitBreaker_OpenState(t *testing.T) {
    cb := circuit.New("test", circuit.WithFailureThreshold(1))

    // Trip the circuit
    cb.Execute(ctx, func(ctx context.Context) (any, error) {
        return nil, errors.New("failure")
    })

    // Circuit should be open
    _, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
        return "success", nil
    })

    if !circuit.IsCircuitOpen(err) {
        t.Error("expected circuit to be open")
    }
}
```

### Test Context Cancellation

Test context cancellation:

```go
func TestWorkerPool_ContextCancellation(t *testing.T) {
    pool := workerpool.New(1, 10)
    defer pool.Close(context.Background())

    ctx, cancel := context.WithCancel(context.Background())
    cancel() // Cancel immediately

    err := pool.Submit(ctx, func(ctx context.Context) error {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(time.Second):
            return nil
        }
    })

    if err == nil {
        t.Error("expected context cancellation error")
    }
}
```

### Use Test Clocks

Use test clocks for time-dependent tests:

```go
func TestRateLimiter_WaitN(t *testing.T) {
    clock := &testclock.FakeClock{}
    limiter := ratelimit.NewTokenBucket(
        ratelimit.PerSecond(1),
        1,
        ratelimit.WithClock(clock),
    )

    // Test time-dependent behavior
    clock.Advance(time.Second)
}
```

## Performance Best Practices

### Avoid Unnecessary Allocations

Reuse components when possible:

```go
// Good: Reuse circuit breaker
var paymentCB circuit.CircuitBreaker

func init() {
    paymentCB = circuit.New("payment-service")
}
```

### Use Appropriate Concurrency

Don't over-provision workers:

```go
// Good: Based on actual needs
pool := workerpool.New(runtime.NumCPU(), 100)
```

### Monitor Performance

Track performance metrics:

```go
start := time.Now()
result, err := cb.Execute(ctx, fn)
duration := time.Since(start)

metrics.Histogram("operation_duration", duration.Seconds())
```

## Security Best Practices

### Don't Expose Internal State

Don't expose component internals in errors:

```go
// Bad: Exposes internal state
return fmt.Errorf("circuit breaker error: %v", err)

// Good: User-friendly error
return errors.New("service temporarily unavailable")
```

### Validate Inputs

Validate inputs before using components:

```go
if capacity <= 0 {
    return fmt.Errorf("invalid capacity: %d", capacity)
}
sem := semaphore.NewWeighted(int64(capacity))
```

## Further Reading

- [Performance Tuning](./performance-tuning.md) - Optimization strategies
- [API Reference](../api-reference/) - Complete API documentation
- [Examples](../examples/) - Practical examples
