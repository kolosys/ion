# circuit Best Practices

Best practices and recommended patterns for using the circuit package effectively.

## Overview

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


## General Best Practices

### Import and Setup

```go
import "github.com/kolosys/ion/circuit"

// Always check for errors when initializing
config, err := circuit.New()
if err != nil {
    log.Fatal(err)
}
```

### Error Handling

Always handle errors returned by circuit functions:

```go
result, err := circuit.DoSomething()
if err != nil {
    // Handle the error appropriately
    log.Printf("Error: %v", err)
    return err
}
```

### Resource Management

Ensure proper cleanup of resources:

```go
// Use defer for cleanup
defer resource.Close()

// Or use context for cancellation
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
```

## Package-Specific Patterns

### circuit Package

#### Using Types

**CircuitBreaker**

CircuitBreaker represents a circuit breaker that controls access to a potentially failing operation. It provides fast-fail behavior when the operation is failing and automatic recovery testing when appropriate.

```go
// Example usage of CircuitBreaker
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

**CircuitError**

CircuitError represents circuit breaker specific errors with context

```go
// Example usage of CircuitError
// Create a new CircuitError
circuiterror := CircuitError{
    Op: "example",
    CircuitName: "example",
    State: "example",
    Err: error{},
}
```

**CircuitMetrics**

CircuitMetrics holds metrics for a circuit breaker instance.

```go
// Example usage of CircuitMetrics
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

**Config**

Config holds configuration for a circuit breaker.

```go
// Example usage of Config
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

**Option**

Option is a function that configures a circuit breaker.

```go
// Example usage of Option
// Example usage of Option
var value Option
// Initialize with appropriate value
```

**State**

State represents the current state of a circuit breaker.

```go
// Example usage of State
// Example usage of State
var value State
// Initialize with appropriate value
```

#### Using Functions

**NewCircuitOpenError**

NewCircuitOpenError creates an error indicating the circuit is open

```go
// Example usage of NewCircuitOpenError
result := NewCircuitOpenError(/* parameters */)
```

**NewCircuitTimeoutError**

NewCircuitTimeoutError creates an error indicating a circuit operation timed out

```go
// Example usage of NewCircuitTimeoutError
result := NewCircuitTimeoutError(/* parameters */)
```

## Performance Considerations

### Optimization Tips

- Use appropriate data structures for your use case
- Consider memory usage for large datasets
- Profile your code to identify bottlenecks

### Caching

When appropriate, implement caching to improve performance:

```go
// Example caching pattern
var cache = make(map[string]interface{})

func getCachedValue(key string) (interface{}, bool) {
    return cache[key], true
}
```

## Security Best Practices

### Input Validation

Always validate inputs:

```go
func processInput(input string) error {
    if input == "" {
        return errors.New("input cannot be empty")
    }
    // Process the input
    return nil
}
```

### Error Information

Be careful not to expose sensitive information in error messages:

```go
// Good: Generic error message
return errors.New("authentication failed")

// Bad: Exposing internal details
return fmt.Errorf("authentication failed: invalid token %s", token)
```

## Testing Best Practices

### Unit Tests

Write comprehensive unit tests:

```go
func TestcircuitFunction(t *testing.T) {
    // Test setup
    input := "test input"

    // Execute function
    result, err := circuit.Function(input)

    // Assertions
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }

    if result == nil {
        t.Error("Expected non-nil result")
    }
}
```

### Integration Tests

Test integration with other components:

```go
func TestcircuitIntegration(t *testing.T) {
    // Setup integration test environment
    // Run integration tests
    // Cleanup
}
```

## Common Pitfalls

### What to Avoid

1. **Ignoring errors**: Always check returned errors
2. **Not cleaning up resources**: Use defer or context cancellation
3. **Hardcoding values**: Use configuration instead
4. **Not testing edge cases**: Test boundary conditions

### Debugging Tips

1. Use logging to trace execution flow
2. Add debug prints for troubleshooting
3. Use Go's built-in profiling tools
4. Check the [FAQ](../faq.md) for common issues

## Migration and Upgrades

### Version Compatibility

When upgrading circuit:

1. Check the changelog for breaking changes
2. Update your code to use new APIs
3. Test thoroughly after upgrades
4. Review deprecated functions and types

## Additional Resources

- [API Reference](../../api-reference/circuit.md)
