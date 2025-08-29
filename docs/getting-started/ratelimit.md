# Getting Started with ratelimit

Package ratelimit provides local process rate limiters for controlling function and I/O throughput.
It includes token bucket and leaky bucket implementations with configurable options.


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
    fmt.Println("Hello from ratelimit!")
}
```

## Basic Usage
### Types
- **Clock** - Clock abstracts time operations for testability.
- **LeakyBucket** - LeakyBucket implements a leaky bucket rate limiter.
- **Limiter** - Limiter represents a rate limiter that controls the rate at which events are allowed to occur.
- **MultiTierConfig** - MultiTierConfig holds configuration for multi-tier rate limiting.
- **MultiTierLimiter** - MultiTierLimiter implements a sophisticated multi-tier rate limiting system.
- **MultiTierMetrics** - MultiTierMetrics tracks metrics for multi-tier rate limiting.
- **Option** - Option configures rate limiter behavior.
- **Rate** - Rate represents the rate at which tokens are added to the bucket.
- **Request** - Request represents a request for rate limiting evaluation.
- **RouteConfig** - RouteConfig defines rate limiting for specific route patterns.
- **Timer** - Timer represents a timer that can be stopped.
- **TokenBucket** - TokenBucket implements a token bucket rate limiter.
- **config** - 
- **realClock** - realClock implements Clock using the real time functions.
- **realTimer** - realTimer wraps time.Timer to implement our Timer interface.

## Next Steps

- [Package Overview](../packages/ratelimit.md) - Complete package information
- [API Reference](../api-reference/ratelimit.md) - Detailed API documentation
- [Examples](../examples/ratelimit/README.md) - Working examples and tutorials  
- [Best Practices](../guides/ratelimit/best-practices.md) - Recommended usage patterns
- [Common Patterns](../guides/ratelimit/patterns.md) - Common implementation patterns
