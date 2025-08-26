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

- [Option](api-reference.md#option) - Option configures rate limiter behavior.

- [Rate](api-reference.md#rate) - Rate represents the rate at which tokens are added to the bucket.

- [Timer](api-reference.md#timer) - Timer represents a timer that can be stopped.

- [TokenBucket](api-reference.md#tokenbucket) - TokenBucket implements a token bucket rate limiter.
Tokens are added to the bucket at a fixed rat...
- [config](api-reference.md#config) - 
- [realClock](api-reference.md#realclock) - realClock implements Clock using the real time functions.

- [realTimer](api-reference.md#realtimer) - realTimer wraps time.Timer to implement our Timer interface.

- [testClock](api-reference.md#testclock) - testClock is a controllable clock implementation for testing.

- [testTimer](api-reference.md#testtimer) - 

## Examples

See [examples](examples.md) for detailed usage examples.
