# ratelimit Best Practices

Best practices and recommended patterns for using the ratelimit package effectively.

## Overview

Package ratelimit provides local process rate limiters for controlling function and I/O throughput.
It includes token bucket and leaky bucket implementations with configurable options.


## General Best Practices

### Import and Setup

```go
import "github.com/kolosys/ion/ratelimit"

// Always check for errors when initializing
config, err := ratelimit.New()
if err != nil {
    log.Fatal(err)
}
```

### Error Handling

Always handle errors returned by ratelimit functions:

```go
result, err := ratelimit.DoSomething()
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

### ratelimit Package

#### Using Types

**Clock**

Clock abstracts time operations for testability.

```go
// Example usage of Clock
// Example implementation of Clock
type MyClock struct {
    // Add your fields here
}

func (m MyClock) Now() time.Time {
    // Implement your logic here
    return
}

func (m MyClock) Sleep(param1 time.Duration)  {
    // Implement your logic here
    return
}

func (m MyClock) AfterFunc(param1 time.Duration, param2 func()) Timer {
    // Implement your logic here
    return
}


```

**LeakyBucket**

LeakyBucket implements a leaky bucket rate limiter. Requests are added to the bucket, and the bucket leaks at a constant rate. If the bucket is full, requests are denied or must wait.

```go
// Example usage of LeakyBucket
// Create a new LeakyBucket
leakybucket := LeakyBucket{
    rate: Rate{},
    capacity: 42,
    cfg: &config{}{},
    mu: /* value */,
    level: 3.14,
    lastLeak: /* value */,
    initialized: true,
}
```

**Limiter**

Limiter represents a rate limiter that controls the rate at which events are allowed to occur.

```go
// Example usage of Limiter
// Example implementation of Limiter
type MyLimiter struct {
    // Add your fields here
}

func (m MyLimiter) AllowN(param1 time.Time, param2 int) bool {
    // Implement your logic here
    return
}

func (m MyLimiter) WaitN(param1 context.Context, param2 int) error {
    // Implement your logic here
    return
}


```

**MultiTierConfig**

MultiTierConfig holds configuration for multi-tier rate limiting.

```go
// Example usage of MultiTierConfig
// Create a new MultiTierConfig
multitierconfig := MultiTierConfig{
    GlobalRate: Rate{},
    GlobalBurst: 42,
    DefaultRouteRate: Rate{},
    DefaultRouteBurst: 42,
    DefaultResourceRate: Rate{},
    DefaultResourceBurst: 42,
    QueueSize: 42,
    EnablePreemptive: true,
    EnableBucketMapping: true,
    BucketTTL: /* value */,
    RoutePatterns: map[],
}
```

**MultiTierLimiter**

MultiTierLimiter implements a sophisticated multi-tier rate limiting system. It supports global, per-route, and per-resource rate limiting with intelligent bucket management and flexible API compatibility.

```go
// Example usage of MultiTierLimiter
// Create a new MultiTierLimiter
multitierlimiter := MultiTierLimiter{
    mu: /* value */,
    global: Limiter{},
    routes: /* value */,
    resources: /* value */,
    bucketMap: /* value */,
    config: &MultiTierConfig{}{},
    cfg: &config{}{},
    metrics: &MultiTierMetrics{}{},
}
```

**MultiTierMetrics**

MultiTierMetrics tracks metrics for multi-tier rate limiting.

```go
// Example usage of MultiTierMetrics
// Create a new MultiTierMetrics
multitiermetrics := MultiTierMetrics{
    mu: /* value */,
    TotalRequests: 42,
    GlobalLimitHits: 42,
    RouteLimitHits: 42,
    ResourceLimitHits: 42,
    QueuedRequests: 42,
    DroppedRequests: 42,
    AvgWaitTime: /* value */,
    MaxWaitTime: /* value */,
    BucketsActive: 42,
}
```

**Option**

Option configures rate limiter behavior.

```go
// Example usage of Option
// Example usage of Option
var value Option
// Initialize with appropriate value
```

**Rate**

Rate represents the rate at which tokens are added to the bucket.

```go
// Example usage of Rate
// Create a new Rate
rate := Rate{
    TokensPerSec: 3.14,
}
```

**RateLimitError**

RateLimitError represents rate limiting specific errors with context

```go
// Example usage of RateLimitError
// Create a new RateLimitError
ratelimiterror := RateLimitError{
    Op: "example",
    LimiterName: "example",
    Err: error{},
    RetryAfter: /* value */,
    Global: true,
    Bucket: "example",
    Remaining: 42,
    Limit: 42,
}
```

**Request**

Request represents a request for rate limiting evaluation.

```go
// Example usage of Request
// Create a new Request
request := Request{
    Method: "example",
    Endpoint: "example",
    ResourceID: "example",
    SubResourceID: "example",
    UserID: "example",
    MajorParameters: map[],
    Priority: 42,
    Context: /* value */,
}
```

**RouteConfig**

RouteConfig defines rate limiting for specific route patterns.

```go
// Example usage of RouteConfig
// Create a new RouteConfig
routeconfig := RouteConfig{
    Rate: Rate{},
    Burst: 42,
    MajorParameters: [],
}
```

**Timer**

Timer represents a timer that can be stopped.

```go
// Example usage of Timer
// Example implementation of Timer
type MyTimer struct {
    // Add your fields here
}

func (m MyTimer) Stop() bool {
    // Implement your logic here
    return
}


```

**TokenBucket**

TokenBucket implements a token bucket rate limiter. Tokens are added to the bucket at a fixed rate, and requests consume tokens. If no tokens are available, requests must wait or are denied.

```go
// Example usage of TokenBucket
// Create a new TokenBucket
tokenbucket := TokenBucket{
    rate: Rate{},
    burst: 42,
    cfg: &config{}{},
    mu: /* value */,
    tokens: 3.14,
    lastRefill: /* value */,
    initialized: true,
}
```

#### Using Functions

**NewBucketLimitError**

NewBucketLimitError creates an error for bucket-specific rate limits

```go
// Example usage of NewBucketLimitError
result := NewBucketLimitError(/* parameters */)
```

**NewGlobalRateLimitError**

NewGlobalRateLimitError creates an error for global rate limit hits

```go
// Example usage of NewGlobalRateLimitError
result := NewGlobalRateLimitError(/* parameters */)
```

**NewRateLimitExceededError**

NewRateLimitExceededError creates an error indicating rate limit was exceeded

```go
// Example usage of NewRateLimitExceededError
result := NewRateLimitExceededError(/* parameters */)
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
func TestratelimitFunction(t *testing.T) {
    // Test setup
    input := "test input"

    // Execute function
    result, err := ratelimit.Function(input)

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
func TestratelimitIntegration(t *testing.T) {
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

When upgrading ratelimit:

1. Check the changelog for breaking changes
2. Update your code to use new APIs
3. Test thoroughly after upgrades
4. Review deprecated functions and types

## Additional Resources

- [API Reference](../../api-reference/ratelimit.md)
