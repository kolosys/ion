# Quick Start

This guide will help you get started with ion quickly with a basic example.

## Basic Usage

Here's a simple example to get you started:

```go
package main

import (
    "fmt"
    "log"
    "github.com/kolosys/ion/circuit"
    "github.com/kolosys/ion/observe"
    "github.com/kolosys/ion/ratelimit"
    "github.com/kolosys/ion/semaphore"
    "github.com/kolosys/ion/workerpool"
)

func main() {
    // Basic usage example
    fmt.Println("Welcome to ion!")
    
    // TODO: Add your code here
}
```

## Common Use Cases

### Using circuit

**Import Path:** `github.com/kolosys/ion/circuit`

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


```go
package main

import (
    "fmt"
    "github.com/kolosys/ion/circuit"
)

func main() {
    // Example usage of circuit
    fmt.Println("Using circuit package")
}
```

#### Available Types
- **CircuitBreaker** - CircuitBreaker represents a circuit breaker that controls access to a potentially failing operation. It provides fast-fail behavior when the operation is failing and automatic recovery testing when appropriate.
- **CircuitError** - CircuitError represents circuit breaker specific errors with context
- **CircuitMetrics** - CircuitMetrics holds metrics for a circuit breaker instance.
- **Config** - Config holds configuration for a circuit breaker.
- **Option** - Option is a function that configures a circuit breaker.
- **State** - State represents the current state of a circuit breaker.

#### Available Functions
- **NewCircuitOpenError** - NewCircuitOpenError creates an error indicating the circuit is open
- **NewCircuitTimeoutError** - NewCircuitTimeoutError creates an error indicating a circuit operation timed out

For detailed API documentation, see the [circuit API Reference](../reference/api-reference/circuit.md).

### Using observe

**Import Path:** `github.com/kolosys/ion/observe`

Package observe provides observability interfaces and implementations
for logging, metrics, and tracing across all Ion components.


```go
package main

import (
    "fmt"
    "github.com/kolosys/ion/observe"
)

func main() {
    // Example usage of observe
    fmt.Println("Using observe package")
}
```

#### Available Types
- **Logger** - Logger provides a simple logging interface that components can use without depending on specific logging libraries.
- **Metrics** - Metrics provides a simple metrics interface for recording component behavior without depending on specific metrics libraries.
- **NopLogger** - NopLogger is a no-operation logger that discards all log messages
- **NopMetrics** - NopMetrics is a no-operation metrics recorder that discards all metrics
- **NopTracer** - NopTracer is a no-operation tracer that creates no spans
- **Observability** - Observability holds observability hooks for a component
- **Tracer** - Tracer provides a simple tracing interface for observing component operations without depending on specific tracing libraries.

For detailed API documentation, see the [observe API Reference](../reference/api-reference/observe.md).

### Using ratelimit

**Import Path:** `github.com/kolosys/ion/ratelimit`

Package ratelimit provides local process rate limiters for controlling function and I/O throughput.
It includes token bucket and leaky bucket implementations with configurable options.


```go
package main

import (
    "fmt"
    "github.com/kolosys/ion/ratelimit"
)

func main() {
    // Example usage of ratelimit
    fmt.Println("Using ratelimit package")
}
```

#### Available Types
- **Clock** - Clock abstracts time operations for testability.
- **LeakyBucket** - LeakyBucket implements a leaky bucket rate limiter. Requests are added to the bucket, and the bucket leaks at a constant rate. If the bucket is full, requests are denied or must wait.
- **Limiter** - Limiter represents a rate limiter that controls the rate at which events are allowed to occur.
- **MultiTierConfig** - MultiTierConfig holds configuration for multi-tier rate limiting.
- **MultiTierLimiter** - MultiTierLimiter implements a sophisticated multi-tier rate limiting system. It supports global, per-route, and per-resource rate limiting with intelligent bucket management and flexible API compatibility.
- **MultiTierMetrics** - MultiTierMetrics tracks metrics for multi-tier rate limiting.
- **Option** - Option configures rate limiter behavior.
- **Rate** - Rate represents the rate at which tokens are added to the bucket.
- **RateLimitError** - RateLimitError represents rate limiting specific errors with context
- **Request** - Request represents a request for rate limiting evaluation.
- **RouteConfig** - RouteConfig defines rate limiting for specific route patterns.
- **Timer** - Timer represents a timer that can be stopped.
- **TokenBucket** - TokenBucket implements a token bucket rate limiter. Tokens are added to the bucket at a fixed rate, and requests consume tokens. If no tokens are available, requests must wait or are denied.

#### Available Functions
- **NewBucketLimitError** - NewBucketLimitError creates an error for bucket-specific rate limits
- **NewGlobalRateLimitError** - NewGlobalRateLimitError creates an error for global rate limit hits
- **NewRateLimitExceededError** - NewRateLimitExceededError creates an error indicating rate limit was exceeded

For detailed API documentation, see the [ratelimit API Reference](../reference/api-reference/ratelimit.md).

### Using semaphore

**Import Path:** `github.com/kolosys/ion/semaphore`

Package semaphore provides a weighted semaphore with configurable fairness modes.


```go
package main

import (
    "fmt"
    "github.com/kolosys/ion/semaphore"
)

func main() {
    // Example usage of semaphore
    fmt.Println("Using semaphore package")
}
```

#### Available Types
- **Fairness** - Fairness defines the ordering behavior for semaphore waiters
- **Option** - Option configures semaphore behavior
- **Semaphore** - Semaphore represents a weighted semaphore that controls access to a resource with a fixed capacity. It supports configurable fairness modes and observability.
- **SemaphoreError** - SemaphoreError represents semaphore-specific errors with context

#### Available Functions
- **NewAcquireTimeoutError** - NewAcquireTimeoutError creates an error indicating an acquire operation timed out
- **NewWeightExceedsCapacityError** - NewWeightExceedsCapacityError creates an error indicating the requested weight exceeds capacity

For detailed API documentation, see the [semaphore API Reference](../reference/api-reference/semaphore.md).

### Using workerpool

**Import Path:** `github.com/kolosys/ion/workerpool`

Package workerpool provides a bounded worker pool with context-aware submission,
graceful shutdown, and observability hooks.


```go
package main

import (
    "fmt"
    "github.com/kolosys/ion/workerpool"
)

func main() {
    // Example usage of workerpool
    fmt.Println("Using workerpool package")
}
```

#### Available Types
- **Option** - Option configures pool behavior
- **Pool** - Pool represents a bounded worker pool that executes tasks with controlled concurrency and queue management.
- **PoolError** - PoolError represents workerpool-specific errors with context
- **PoolMetrics** - PoolMetrics holds runtime metrics for the pool
- **Task** - Task represents a unit of work to be executed by the worker pool. Tasks receive a context that will be canceled if either the submission context or the pool's base context is canceled.

#### Available Functions
- **NewPoolClosedError** - NewPoolClosedError creates an error indicating the pool is closed
- **NewQueueFullError** - NewQueueFullError creates an error indicating the queue is full

For detailed API documentation, see the [workerpool API Reference](../reference/api-reference/workerpool.md).

## Step-by-Step Tutorial

### Step 1: Import the Package

First, import the necessary packages in your Go file:

```go
import (
    "fmt"
    "github.com/kolosys/ion/circuit"
    "github.com/kolosys/ion/observe"
    "github.com/kolosys/ion/ratelimit"
    "github.com/kolosys/ion/semaphore"
    "github.com/kolosys/ion/workerpool"
)
```

### Step 2: Initialize

Set up the basic configuration:

```go
func main() {
    // Initialize your application
    fmt.Println("Initializing ion...")
}
```

### Step 3: Use the Library

Implement your specific use case:

```go
func main() {
    // Your implementation here
}
```

## Running Your Code

To run your Go program:

```bash
go run main.go
```

To build an executable:

```bash
go build -o myapp
./myapp
```

## Configuration Options

ion can be configured to suit your needs. Check the [Core Concepts](../core-concepts/) section for detailed information about configuration options.

## Error Handling

Always handle errors appropriately:

```go
result, err := someFunction()
if err != nil {
    log.Fatalf("Error: %v", err)
}
```

## Best Practices

- Always handle errors returned by library functions
- Check the API documentation for detailed parameter information
- Use meaningful variable and function names
- Add comments to document your code

## Complete Example

Here's a complete working example:

```go
package main

import (
    "fmt"
    "log"
    "github.com/kolosys/ion/circuit"
    "github.com/kolosys/ion/observe"
    "github.com/kolosys/ion/ratelimit"
    "github.com/kolosys/ion/semaphore"
    "github.com/kolosys/ion/workerpool"
)

func main() {
    fmt.Println("Starting ion application...")
    
    // Add your implementation here
    
    fmt.Println("Application completed successfully!")
}
```

## Next Steps

Now that you've seen the basics, explore:

- **[Core Concepts](../core-concepts/)** - Understanding the library architecture
- **[API Reference](../reference/api-reference/README.md)** - Complete API documentation
- **[Examples](../reference/examples/README.md)** - More detailed examples
- **[Advanced Topics](../advanced/)** - Performance tuning and advanced patterns

## Getting Help

If you run into issues:

1. Check the [API Reference](../reference/api-reference/README.md)
2. Browse the [Examples](../reference/examples/README.md)
3. Visit the [GitHub Issues](https://github.com/kolosys/ion/issues) page

