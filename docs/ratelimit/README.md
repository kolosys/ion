# ratelimit

Package ratelimit provides local process rate limiters for controlling function and I/O throughput.
It includes token bucket, leaky bucket, and multi-tier rate limiting implementations with configurable options.

## Features

- **Token Bucket**: Allows bursts while maintaining sustained rate limits
- **Leaky Bucket**: Smooths out traffic bursts for consistent processing
- **Multi-Tier Rate Limiting**: Sophisticated rate limiting with global, per-route, and per-resource limits
- **Route Pattern Matching**: Support for custom route patterns and API gateway scenarios
- **Resource Isolation**: Per-resource rate limiting for multi-tenant applications
- **Header Integration**: Support for external API rate limit headers
- **Metrics & Observability**: Comprehensive metrics and logging
- **Thread-Safe**: Designed for concurrent access


## Installation

```bash
go get github.com/kolosys/ion/ratelimit
```

## Quick Start

```go
package main

import "github.com/kolosys/ion/ratelimit"

func main() {
    // Your code here
}
```

## Quick Start

### Basic Rate Limiting

```go
package main

import "github.com/kolosys/ion/ratelimit"

func main() {
    // Token bucket: 10 requests per second, burst of 20
    limiter := ratelimit.NewTokenBucket(ratelimit.PerSecond(10), 20)
    
    if limiter.AllowN(time.Now(), 1) {
        // Request allowed
    } else {
        // Request denied
    }
}
```

### Multi-Tier Rate Limiting

```go
package main

import "github.com/kolosys/ion/ratelimit"

func main() {
    config := ratelimit.DefaultMultiTierConfig()
    config.GlobalRate = ratelimit.PerSecond(100)      // Global limit
    config.DefaultRouteRate = ratelimit.PerSecond(20)  // Per-route limit
    config.DefaultResourceRate = ratelimit.PerSecond(5) // Per-resource limit
    
    limiter := ratelimit.NewMultiTierLimiter(config)
    
    req := &ratelimit.Request{
        Method:     "GET",
        Endpoint:   "/api/v1/users",
        ResourceID: "org123",
        Context:    context.Background(),
    }
    
    if limiter.Allow(req) {
        // Request allowed
    }
}
```

## API Reference
### Core Types
- [Limiter](api-reference.md#limiter) - Rate limiter interface
- [Rate](api-reference.md#rate) - Rate configuration
- [Option](api-reference.md#option) - Configuration options

### Implementations
- [TokenBucket](api-reference.md#tokenbucket) - Token bucket rate limiter
- [LeakyBucket](api-reference.md#leakybucket) - Leaky bucket rate limiter
- [MultiTierLimiter](api-reference.md#multitierlimiter) - Multi-tier rate limiter

### Supporting Types
- [Clock](api-reference.md#clock) - Time abstraction for testing
- [Timer](api-reference.md#timer) - Timer interface
- [Request](api-reference.md#request) - Rate limit request
- [MultiTierConfig](api-reference.md#multitierconfig) - Multi-tier configuration
- [RouteConfig](api-reference.md#routeconfig) - Route-specific configuration
- [MultiTierMetrics](api-reference.md#multitiermetrics) - Metrics and observability

## Examples

See [examples](examples.md) for detailed usage examples.
