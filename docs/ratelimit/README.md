# ratelimit

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
}
```

## API Reference
### Types
- [Clock](api-reference.md#clock) - Clock abstracts time operations for testability.

- [LeakyBucket](api-reference.md#leakybucket) - LeakyBucket implements a leaky bucket rate limiter.
Requests are added to the bucket, and the buc...
- [Limiter](api-reference.md#limiter) - Limiter represents a rate limiter that controls the rate at which events are allowed to occur.

- [MultiTierConfig](api-reference.md#multitierconfig) - MultiTierConfig holds configuration for multi-tier rate limiting.

- [MultiTierLimiter](api-reference.md#multitierlimiter) - MultiTierLimiter implements a sophisticated multi-tier rate limiting system.
It supports global, ...
- [MultiTierMetrics](api-reference.md#multitiermetrics) - MultiTierMetrics tracks metrics for multi-tier rate limiting.

- [Option](api-reference.md#option) - Option configures rate limiter behavior.

- [Rate](api-reference.md#rate) - Rate represents the rate at which tokens are added to the bucket.

- [Request](api-reference.md#request) - Request represents a request for rate limiting evaluation.

- [RouteConfig](api-reference.md#routeconfig) - RouteConfig defines rate limiting for specific route patterns.

- [Timer](api-reference.md#timer) - Timer represents a timer that can be stopped.

- [TokenBucket](api-reference.md#tokenbucket) - TokenBucket implements a token bucket rate limiter.
Tokens are added to the bucket at a fixed rat...
- [config](api-reference.md#config) - 
- [realClock](api-reference.md#realclock) - realClock implements Clock using the real time functions.

- [realTimer](api-reference.md#realtimer) - realTimer wraps time.Timer to implement our Timer interface.


## Examples

See [examples](examples.md) for detailed usage examples.
