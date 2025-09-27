# RateLimit

[![Go Reference](https://pkg.go.dev/badge/github.com/kolosys/ion/ratelimit.svg)](https://pkg.go.dev/github.com/kolosys/ion/ratelimit)

Local process rate limiters for controlling function and I/O throughput with token bucket, leaky bucket, and multi-tier rate limiting.

## Features

- **Token Bucket**: Burst-friendly rate limiting with configurable refill rates
- **Leaky Bucket**: Smooth traffic shaping with controlled processing rates
- **Multi-Tier Limiting**: Global, per-route, and per-resource rate limiting
- **Context-Aware**: All blocking operations respect context cancellation
- **Zero Dependencies**: No external dependencies beyond the Go standard library
- **Observability**: Built-in metrics, logging, and tracing support
- **API Integration**: Header-based rate limit updates for external APIs

## Quick Start

### Token Bucket - Burst Traffic

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kolosys/ion/ratelimit"
)

func main() {
    // Allow 10 requests per second with burst of 20
    limiter := ratelimit.NewTokenBucket(ratelimit.PerSecond(10), 20)

    // Immediate burst usage
    for i := 0; i < 25; i++ {
        if limiter.AllowN(time.Now(), 1) {
            fmt.Printf("Request %d: allowed\n", i+1)
        } else {
            fmt.Printf("Request %d: rate limited\n", i+1)
        }
    }

    fmt.Printf("Remaining tokens: %.1f\n", limiter.Tokens())
}
```

### Leaky Bucket - Smooth Processing

```go
// Process requests at steady 5/second rate with queue capacity of 10
processor := ratelimit.NewLeakyBucket(ratelimit.PerSecond(5), 10)

// Queue requests for processing
for i := 0; i < 12; i++ {
    if processor.AllowN(time.Now(), 1) {
        fmt.Printf("Request %d: queued (level: %.1f)\n", i+1, processor.Level())
    } else {
        fmt.Printf("Request %d: rejected (queue full)\n", i+1)
    }
}
```

### Multi-Tier API Gateway

```go
// Create sophisticated API gateway rate limiting
config := ratelimit.DefaultMultiTierConfig()
config.GlobalRate = ratelimit.PerSecond(1000)    // Global limit
config.DefaultRouteRate = ratelimit.PerSecond(100) // Per-route limit
config.DefaultResourceRate = ratelimit.PerSecond(50) // Per-resource limit

// Define specific route patterns
config.RoutePatterns = map[string]ratelimit.RouteConfig{
    "POST:/api/v1/users": {
        Rate:  ratelimit.PerSecond(2),  // User creation: limited
        Burst: 2,
    },
    "GET:/api/v1/users/{id}": {
        Rate:  ratelimit.PerSecond(30), // User lookup: higher limit
        Burst: 30,
    },
}

limiter := ratelimit.NewMultiTierLimiter(config, ratelimit.WithName("api-gateway"))

// Check rate limits for requests
req := &ratelimit.Request{
    Method:     "POST",
    Endpoint:   "/api/v1/users",
    ResourceID: "org123",  // Per-organization limits
    Context:    ctx,
}

if limiter.Allow(req) {
    // Process request
    handleUserCreation(req)
} else {
    // Return 429 Too Many Requests
    sendRateLimitError(w)
}
```

## API Reference

### Token Bucket

```go
func NewTokenBucket(rate Rate, burst int, opts ...Option) *TokenBucket

func (tb *TokenBucket) AllowN(now time.Time, n int) bool
func (tb *TokenBucket) WaitN(ctx context.Context, n int) error
func (tb *TokenBucket) Tokens() float64
```

**Best for:** API rate limiting, burst traffic handling, client-side throttling

### Leaky Bucket

```go
func NewLeakyBucket(rate Rate, capacity int, opts ...Option) *LeakyBucket

func (lb *LeakyBucket) AllowN(now time.Time, n int) bool
func (lb *LeakyBucket) WaitN(ctx context.Context, n int) error
func (lb *LeakyBucket) Level() float64
func (lb *LeakyBucket) Available() int
```

**Best for:** Queue management, traffic shaping, smooth request processing

### Multi-Tier Limiter

```go
func NewMultiTierLimiter(config *MultiTierConfig, opts ...Option) *MultiTierLimiter

func (mtl *MultiTierLimiter) Allow(req *Request) bool
func (mtl *MultiTierLimiter) Wait(req *Request) error
func (mtl *MultiTierLimiter) GetMetrics() *MultiTierMetrics
```

**Best for:** API gateways, microservices, multi-tenant applications

## Rate Specifications

### Convenience Functions

```go
ratelimit.PerSecond(100)                    // 100 per second
ratelimit.PerMinute(60)                     // 1 per second
ratelimit.PerHour(3600)                     // 1 per second
ratelimit.Per(5, 2*time.Second)             // 2.5 per second
```

### Custom Rates

```go
rate := ratelimit.Rate{TokensPerSec: 10.5}  // 10.5 per second
```

## Configuration Options

### Basic Options

```go
ratelimit.WithName("api-limiter")           // Set limiter name for observability
ratelimit.WithClock(customClock)            // Custom clock (useful for testing)
ratelimit.WithJitter(0.1)                  // Add 10% jitter to wait times
```

### Observability

```go
ratelimit.WithLogger(logger)                // Custom logger
ratelimit.WithMetrics(metrics)              // Custom metrics recorder
ratelimit.WithTracer(tracer)                // Custom tracer
```

## Use Cases

### API Client Rate Limiting

```go
// Respect third-party API rate limits
authLimiter := ratelimit.NewTokenBucket(ratelimit.PerMinute(100), 10)
dataLimiter := ratelimit.NewTokenBucket(ratelimit.PerSecond(10), 20)

func makeAPIRequest(endpoint string, limiter ratelimit.Limiter) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := limiter.WaitN(ctx, 1); err != nil {
        return fmt.Errorf("rate limit timeout: %w", err)
    }

    // Make API request
    return callAPI(endpoint)
}
```

### Background Job Processing

```go
// Control job processing rate to avoid overwhelming downstream services
jobProcessor := ratelimit.NewLeakyBucket(ratelimit.PerSecond(5), 100)

func processJobs(jobs <-chan Job) {
    for job := range jobs {
        // Wait for processing slot
        if err := jobProcessor.WaitN(context.Background(), 1); err != nil {
            log.Printf("Job processing canceled: %v", err)
            continue
        }

        go handleJob(job)
    }
}
```

### Multi-Tenant SaaS Applications

```go
// Different rate limits per customer tier
func createCustomerLimiter(tier string) *ratelimit.MultiTierLimiter {
    config := ratelimit.DefaultMultiTierConfig()

    switch tier {
    case "premium":
        config.GlobalRate = ratelimit.PerSecond(1000)
        config.DefaultResourceRate = ratelimit.PerSecond(100)
    case "standard":
        config.GlobalRate = ratelimit.PerSecond(500)
        config.DefaultResourceRate = ratelimit.PerSecond(50)
    case "basic":
        config.GlobalRate = ratelimit.PerSecond(100)
        config.DefaultResourceRate = ratelimit.PerSecond(10)
    }

    return ratelimit.NewMultiTierLimiter(config)
}
```

### HTTP Middleware

```go
func rateLimitMiddleware(limiter ratelimit.Limiter) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if !limiter.AllowN(time.Now(), 1) {
            http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
            return
        }

        // Continue to next handler
        next.ServeHTTP(w, r)
    }
}
```

## Algorithm Comparison

### Token Bucket vs Leaky Bucket

| Feature             | Token Bucket                               | Leaky Bucket                      |
| ------------------- | ------------------------------------------ | --------------------------------- |
| **Burst Handling**  | Excellent - allows burst up to bucket size | Limited - smooth processing only  |
| **Traffic Shaping** | Minimal - allows bursts                    | Excellent - enforces steady rate  |
| **Memory Usage**    | Low - tracks token count                   | Low - tracks queue level          |
| **Use Case**        | API rate limiting, client throttling       | Queue management, traffic shaping |

### When to Use Each

**Token Bucket:**

- API rate limiting with burst allowance
- Client-side request throttling
- Interactive applications needing responsive bursts

**Leaky Bucket:**

- Queue processing with controlled output rate
- Traffic shaping for downstream services
- Smooth resource utilization

**Multi-Tier:**

- API gateways with complex routing
- Multi-tenant applications
- Enterprise applications with resource isolation

## Multi-Tier Configuration

### Route Patterns

```go
config.RoutePatterns = map[string]ratelimit.RouteConfig{
    "GET:/api/v1/users/{id}": {
        Rate:  ratelimit.PerSecond(50),
        Burst: 50,
    },
    "POST:/api/v1/webhooks": {
        Rate:  ratelimit.PerSecond(5),   // Webhook creation is expensive
        Burst: 5,
    },
    "GET:/api/v1/health": {
        Rate:  ratelimit.PerSecond(1000), // Health checks are cheap
        Burst: 1000,
    },
}
```

### Resource-Based Limiting

```go
req := &ratelimit.Request{
    Method:     "GET",
    Endpoint:   "/api/v1/data",
    ResourceID: "organization-123",  // Per-organization limits
    UserID:     "user-456",         // Per-user limits
    Context:    ctx,
}

// Will apply global, route, and resource limits
allowed := limiter.Allow(req)
```

### API Integration

```go
// Process rate limit headers from external APIs
headers := map[string]string{
    "X-RateLimit-Limit":     "100",
    "X-RateLimit-Remaining": "95",
    "X-RateLimit-Reset":     "1640995200",
    "X-RateLimit-Bucket":    "api-bucket-123",
}

err := limiter.UpdateRateLimitFromHeaders(req, headers)
```

## Examples

- [Basic Usage](../examples/ratelimit/main.go) - Token and leaky bucket examples
- [Multi-Tier Demo](../examples/ratelimit/multitier_demo.go) - API gateway rate limiting
- [API Client](../examples/ratelimit/main.go) - Third-party API integration

## Performance

Benchmark results on modern hardware:

- **AllowN**: <100ns (uncontended), <500ns (high contention)
- **WaitN**: <1ms for immediate grants, accurate timing for waits
- **Memory**: 0 allocations for steady-state operations
- **Throughput**: 10M+ checks/second per limiter

## Thread Safety

All rate limiter implementations are safe for concurrent use across multiple goroutines.

## Testing Support

Built-in test clock for deterministic testing:

```go
func TestRateLimit(t *testing.T) {
    clock := &testClock{now: time.Now()}
    limiter := ratelimit.NewTokenBucket(
        ratelimit.PerSecond(10),
        5,
        ratelimit.WithClock(clock),
    )

    // Control time for deterministic tests
    clock.Advance(time.Second)
    assert.True(t, limiter.AllowN(clock.Now(), 10))
}
```

## Contributing

See the main [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

## License

Licensed under the [MIT License](../LICENSE).
