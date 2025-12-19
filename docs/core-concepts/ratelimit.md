# Rate Limiting

**Import Path:** `github.com/kolosys/ion/ratelimit`

Rate limiting controls the rate at which operations execute, protecting your services from being overwhelmed and respecting external API limits.

## Overview

Ion provides three rate limiting algorithms:

1. **Token Bucket**: Allows bursts while maintaining average rate
2. **Leaky Bucket**: Smooths out bursts for steady processing
3. **Multi-Tier**: Sophisticated multi-level rate limiting for API gateways

### When to Use Rate Limiting

- **API Protection**: Prevent APIs from being overwhelmed
- **External API Limits**: Respect third-party API rate limits
- **Resource Control**: Limit resource consumption
- **API Gateways**: Multi-tier rate limiting per user, route, or resource
- **Background Jobs**: Control processing throughput

## Architecture

### Token Bucket

Tokens are added to the bucket at a fixed rate. Requests consume tokens, and if no tokens are available, requests must wait or are denied.

```
┌─────────────┐
│   Bucket    │
│  [Tokens]   │  ← Tokens added at rate R
│             │
└─────────────┘
      │
      │ Requests consume tokens
      ▼
```

**Characteristics:**

- Allows bursts up to bucket capacity
- Maintains average rate over time
- Good for handling traffic spikes

### Leaky Bucket

Requests are added to the bucket, and the bucket leaks at a constant rate. If the bucket is full, requests are denied.

```
┌─────────────┐
│   Bucket    │
│  [Requests] │  → Leaks at rate R
│             │
└─────────────┘
```

**Characteristics:**

- Smooths out bursts
- Steady processing rate
- Good for queue processing

### Multi-Tier Limiter

Supports global, per-route, and per-resource rate limiting with intelligent bucket management.

```
Global Limiter
    │
    ├─ Route Limiters (/api/users, /api/orders)
    │      │
    │      └─ Resource Limiters (user-123, org-456)
    │
    └─ Bucket Mapping (API-style rate limit buckets)
```

## Core Concepts

### Rate Definition

Rates are defined using the `Rate` type:

```go
// 10 requests per second
rate := ratelimit.PerSecond(10)

// 100 requests per minute
rate := ratelimit.PerMinute(100)

// Custom rate: 5 requests per 2 seconds
rate := ratelimit.Per(5, 2*time.Second)
```

### Non-Blocking Check

Check if a request is allowed without blocking:

```go
if limiter.AllowN(time.Now(), 1) {
    // Process request
} else {
    // Rate limited
}
```

### Blocking Wait

Wait for rate limit with context support:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := limiter.WaitN(ctx, 1); err != nil {
    // Context canceled or timeout
    return err
}
// Request allowed
```

## Real-World Scenarios

### Scenario 1: API Client Rate Limiting

Respect external API rate limits:

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kolosys/ion/ratelimit"
)

type APIClient struct {
    limiter ratelimit.Limiter
}

func NewAPIClient() *APIClient {
    // API allows 100 requests per minute with burst of 10
    return &APIClient{
        limiter: ratelimit.NewTokenBucket(
            ratelimit.PerMinute(100),
            10,
            ratelimit.WithName("external-api"),
        ),
    }
}

func (c *APIClient) CallAPI(ctx context.Context, endpoint string) error {
    // Wait for rate limit
    if err := c.limiter.WaitN(ctx, 1); err != nil {
        return fmt.Errorf("rate limit exceeded: %w", err)
    }

    // Make API call
    fmt.Printf("Calling %s\n", endpoint)
    return nil
}

func main() {
    client := NewAPIClient()
    ctx := context.Background()

    // Make multiple API calls - rate limiter controls the rate
    for i := 0; i < 20; i++ {
        if err := client.CallAPI(ctx, fmt.Sprintf("/api/v1/users/%d", i)); err != nil {
            fmt.Printf("Error: %v\n", err)
        }
        time.Sleep(100 * time.Millisecond)
    }
}
```

### Scenario 2: API Gateway with Multi-Tier Rate Limiting

Implement sophisticated rate limiting for an API gateway:

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kolosys/ion/ratelimit"
)

type APIGateway struct {
    limiter *ratelimit.MultiTierLimiter
}

func NewAPIGateway() *APIGateway {
    config := &ratelimit.MultiTierConfig{
        // Global rate limit: 1000 requests per second
        GlobalRate:  ratelimit.PerSecond(1000),
        GlobalBurst: 2000,

        // Default route limits
        DefaultRouteRate:  ratelimit.PerSecond(100),
        DefaultRouteBurst: 200,

        // Default resource limits
        DefaultResourceRate:  ratelimit.PerSecond(10),
        DefaultResourceBurst: 20,

        // Route-specific patterns
        RoutePatterns: map[string]ratelimit.RouteConfig{
            "/api/v1/auth/login": {
                Rate:  ratelimit.PerMinute(5), // Stricter for login
                Burst: 5,
            },
            "/api/v1/users/*": {
                Rate:  ratelimit.PerSecond(50),
                Burst: 100,
                MajorParameters: []string{"user_id"}, // Per-user limiting
            },
        },
    }

    limiter := ratelimit.NewMultiTier(config)

    return &APIGateway{
        limiter: limiter,
    }
}

func (g *APIGateway) HandleRequest(ctx context.Context, method, path, userID string) error {
    req := &ratelimit.Request{
        Method:   method,
        Endpoint: path,
        UserID:   userID,
        MajorParameters: map[string]string{
            "user_id": userID,
        },
    }

    allowed, err := g.limiter.Allow(ctx, req)
    if err != nil {
        return err
    }

    if !allowed {
        return fmt.Errorf("rate limit exceeded")
    }

    // Process request
    fmt.Printf("Processing %s %s for user %s\n", method, path, userID)
    return nil
}
```

### Scenario 3: Background Job Processing with Leaky Bucket

Control processing rate for background jobs:

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kolosys/ion/ratelimit"
)

type JobProcessor struct {
    limiter ratelimit.Limiter
    jobs    chan Job
}

type Job struct {
    ID   string
    Data string
}

func NewJobProcessor() *JobProcessor {
    // Process 10 jobs per second, queue capacity of 50
    return &JobProcessor{
        limiter: ratelimit.NewLeakyBucket(
            ratelimit.PerSecond(10),
            50,
            ratelimit.WithName("job-processor"),
        ),
        jobs: make(chan Job, 100),
    }
}

func (p *JobProcessor) Process(ctx context.Context) error {
    for {
        select {
        case job := <-p.jobs:
            // Check if we can process this job
            if !p.limiter.AllowN(time.Now(), 1) {
                fmt.Printf("Job %s: rate limited, requeuing\n", job.ID)
                // Requeue or handle appropriately
                continue
            }

            // Process job
            if err := p.processJob(ctx, job); err != nil {
                return err
            }

        case <-ctx.Done():
            return ctx.Err()
        }
    }
}

func (p *JobProcessor) processJob(ctx context.Context, job Job) error {
    fmt.Printf("Processing job %s: %s\n", job.ID, job.Data)
    time.Sleep(100 * time.Millisecond)
    return nil
}

func (p *JobProcessor) Submit(job Job) {
    p.jobs <- job
}
```

### Scenario 4: Per-User Rate Limiting

Implement per-user rate limiting:

```go
package main

import (
    "context"
    "sync"
    "time"

    "github.com/kolosys/ion/ratelimit"
)

type UserRateLimiter struct {
    limiters sync.Map // map[string]ratelimit.Limiter
    baseRate ratelimit.Rate
    burst    int
}

func NewUserRateLimiter() *UserRateLimiter {
    return &UserRateLimiter{
        baseRate: ratelimit.PerSecond(10), // 10 requests per second per user
        burst:    20,
    }
}

func (u *UserRateLimiter) GetLimiter(userID string) ratelimit.Limiter {
    if limiter, ok := u.limiters.Load(userID); ok {
        return limiter.(ratelimit.Limiter)
    }

    // Create new limiter for user
    limiter := ratelimit.NewTokenBucket(u.baseRate, u.burst,
        ratelimit.WithName("user-"+userID),
    )

    u.limiters.Store(userID, limiter)
    return limiter
}

func (u *UserRateLimiter) Allow(ctx context.Context, userID string) (bool, error) {
    limiter := u.GetLimiter(userID)
    return limiter.AllowN(time.Now(), 1), nil
}

func main() {
    limiter := NewUserRateLimiter()
    ctx := context.Background()

    // Different users have separate rate limits
    limiter.Allow(ctx, "user-1")
    limiter.Allow(ctx, "user-2")
}
```

### Scenario 5: Header-Based Rate Limit Updates

Update rate limits based on API response headers:

```go
package main

import (
    "context"
    "net/http"
    "strconv"
    "time"

    "github.com/kolosys/ion/ratelimit"
)

type AdaptiveRateLimiter struct {
    limiter *ratelimit.MultiTierLimiter
}

func NewAdaptiveRateLimiter() *AdaptiveRateLimiter {
    config := &ratelimit.MultiTierConfig{
        GlobalRate:  ratelimit.PerSecond(100),
        GlobalBurst: 200,
    }

    return &AdaptiveRateLimiter{
        limiter: ratelimit.NewMultiTier(config),
    }
}

func (a *AdaptiveRateLimiter) HandleResponse(resp *http.Response) {
    // Update rate limits based on response headers
    if limit := resp.Header.Get("X-RateLimit-Limit"); limit != "" {
        if limitInt, err := strconv.Atoi(limit); err == nil {
            // Update global rate limit
            newRate := ratelimit.PerSecond(limitInt)
            // Note: MultiTierLimiter would need UpdateConfig method
            // This is a conceptual example
        }
    }

    if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
        // Track remaining requests
    }

    if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
        // Track reset time
    }
}
```

## Algorithm Comparison

### Token Bucket vs Leaky Bucket

**Token Bucket:**

- ✅ Allows bursts
- ✅ Good for handling traffic spikes
- ✅ Maintains average rate
- ❌ Can have variable processing rate

**Leaky Bucket:**

- ✅ Smooth, steady processing
- ✅ Predictable rate
- ✅ Good for queue processing
- ❌ No burst capacity

**When to use Token Bucket:**

- API clients that need to handle bursts
- User-facing APIs with variable traffic
- Operations that benefit from burst capacity

**When to use Leaky Bucket:**

- Background job processing
- Queue processing systems
- Operations requiring steady rate

## Configuration Options

### Token Bucket

```go
limiter := ratelimit.NewTokenBucket(
    ratelimit.PerSecond(10), // Rate
    20,                       // Burst capacity
    ratelimit.WithName("api-client"),
    ratelimit.WithJitter(0.1), // 10% jitter for WaitN
)
```

### Leaky Bucket

```go
limiter := ratelimit.NewLeakyBucket(
    ratelimit.PerSecond(10), // Leak rate
    50,                      // Bucket capacity
    ratelimit.WithName("processor"),
)
```

### Multi-Tier

```go
config := &ratelimit.MultiTierConfig{
    GlobalRate:  ratelimit.PerSecond(1000),
    GlobalBurst: 2000,
    DefaultRouteRate:  ratelimit.PerSecond(100),
    DefaultRouteBurst: 200,
    RoutePatterns: map[string]ratelimit.RouteConfig{
        "/api/v1/*": {
            Rate:  ratelimit.PerSecond(50),
            Burst: 100,
        },
    },
}
limiter := ratelimit.NewMultiTier(config)
```

## Best Practices

1. **Choose Appropriate Rates**: Balance between protection and usability
2. **Set Reasonable Bursts**: Allow for traffic spikes without overwhelming services
3. **Use Context Timeouts**: Always use context with timeouts for WaitN
4. **Monitor Rate Limits**: Track rate limit hits in your observability system
5. **Provide Clear Errors**: Return user-friendly rate limit error messages
6. **Consider Multi-Tier**: Use multi-tier for complex rate limiting needs

## Common Pitfalls

### Pitfall 1: Not Using Context

**Problem**: WaitN blocks indefinitely

```go
// Bad: No timeout
limiter.WaitN(context.Background(), 1)
```

**Solution**: Always use context with timeout

```go
// Good: With timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
if err := limiter.WaitN(ctx, 1); err != nil {
    return err
}
```

### Pitfall 2: Too Restrictive Rates

**Problem**: Legitimate users are rate limited

**Solution**: Set rates based on actual usage patterns

```go
// Bad: Too restrictive
limiter := ratelimit.NewTokenBucket(ratelimit.PerMinute(1), 1)

// Good: Based on usage patterns
limiter := ratelimit.NewTokenBucket(ratelimit.PerSecond(10), 20)
```

### Pitfall 3: Not Handling Rate Limit Errors

**Problem**: Users see technical rate limit errors

**Solution**: Provide user-friendly error messages

```go
// Good
if !limiter.AllowN(time.Now(), 1) {
    return errors.New("too many requests, please try again later")
}
```

## Integration Guide

### With HTTP Middleware

```go
func RateLimitMiddleware(limiter ratelimit.Limiter) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if !limiter.AllowN(time.Now(), 1) {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        // Continue to next handler
    }
}
```

### With gRPC Interceptor

```go
func RateLimitUnaryInterceptor(limiter ratelimit.Limiter) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
        if !limiter.AllowN(time.Now(), 1) {
            return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
        }
        return handler(ctx, req)
    }
}
```

## Further Reading

- [API Reference](../api-reference/ratelimit.md) - Complete API documentation
- [Examples](../examples/ratelimit/) - Practical examples
- [Best Practices](../advanced/best-practices.md) - Recommended patterns
- [Performance Tuning](../advanced/performance-tuning.md) - Optimization strategies
