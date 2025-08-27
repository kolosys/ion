# Ion - Concurrency and Scheduling Primitives for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/kolosys/ion.svg)](https://pkg.go.dev/github.com/kolosys/ion)
[![Go Report Card](https://goreportcard.com/badge/github.com/kolosys/ion)](https://goreportcard.com/report/github.com/kolosys/ion)

Ion provides a collection of robust, context-aware concurrency and scheduling primitives for Go applications. It focuses on deterministic behavior, safe cancellation, and pluggable observability without heavy dependencies.

## Features

### v0.1 (Current)

- **[workerpool](./workerpool)** - Bounded worker pool with context-aware submission and graceful shutdown
- **[semaphore](./semaphore)** - Weighted fair semaphore with configurable fairness modes (FIFO/LIFO/None)
- **[ratelimit](./ratelimit)** - Token bucket, leaky bucket, and multi-tier rate limiters with configurable options
- **[shared](./shared)** - Common error types and observability interfaces

### Planned (v0.2+)

- **circuit** - Circuit breaker with retry and jitter backoff
- **pipeline** - Fan-in/fan-out helpers with bounded channels
- **scheduler** - In-process delayed and periodic task scheduling

## Quick Start

### Installation

```bash
go get github.com/kolosys/ion@latest
```

### Worker Pool

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/kolosys/ion/workerpool"
)

func main() {
    // Create a pool with 4 workers and queue size of 20
    pool := workerpool.New(4, 20,
        workerpool.WithName("my-pool"),
        workerpool.WithDrainTimeout(10*time.Second),
    )
    defer pool.Close(context.Background())

    // Submit tasks
    for i := 0; i < 10; i++ {
        taskID := i
        task := func(ctx context.Context) error {
            // Your work here
            fmt.Printf("Processing task %d\\n", taskID)
            time.Sleep(100 * time.Millisecond)
            return nil
        }

        if err := pool.Submit(context.Background(), task); err != nil {
            log.Printf("Failed to submit task: %v", err)
        }
    }

    // Gracefully drain all tasks
    if err := pool.Drain(context.Background()); err != nil {
        log.Printf("Drain failed: %v", err)
    }

    fmt.Printf("Completed: %d tasks\\n", pool.Metrics().Completed)
}
```

### Rate Limiting

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kolosys/ion/ratelimit"
)

func main() {
    // Basic token bucket: 10 requests per second, burst of 20
    limiter := ratelimit.NewTokenBucket(ratelimit.PerSecond(10), 20)

    if limiter.AllowN(time.Now(), 1) {
        fmt.Println("Request allowed")
    } else {
        fmt.Println("Request denied")
    }

    // Multi-tier rate limiting for API gateways
    config := ratelimit.DefaultMultiTierConfig()
    config.GlobalRate = ratelimit.PerSecond(100)      // Global limit
    config.DefaultRouteRate = ratelimit.PerSecond(20)  // Per-route limit
    config.DefaultResourceRate = ratelimit.PerSecond(5) // Per-resource limit

    mtl := ratelimit.NewMultiTierLimiter(config)

    req := &ratelimit.Request{
        Method:     "GET",
        Endpoint:   "/api/v1/users",
        ResourceID: "org123",
        Context:    context.Background(),
    }

    if mtl.Allow(req) {
        fmt.Println("Multi-tier request allowed")
    }
}
```

### Weighted Semaphore

```go
package main

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/kolosys/ion/semaphore"
)

func main() {
    // Create a semaphore with capacity of 3 (e.g., database connections)
    sem := semaphore.NewWeighted(3,
        semaphore.WithName("db-pool"),
        semaphore.WithFairness(semaphore.FIFO),
    )

    var wg sync.WaitGroup

    // Simulate 5 clients accessing the resource pool
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(clientID int) {
            defer wg.Done()

            // Acquire a permit with timeout
            ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
            defer cancel()

            if err := sem.Acquire(ctx, 1); err != nil {
                log.Printf("Client %d failed: %v", clientID, err)
                return
            }

            fmt.Printf("Client %d: using resource\\n", clientID)
            time.Sleep(500 * time.Millisecond) // Simulate work

            // Release the permit
            sem.Release(1)
            fmt.Printf("Client %d: released resource\\n", clientID)
        }(i)
    }

    wg.Wait()
    fmt.Printf("Available permits: %d\\n", sem.Current())
}
```

### Rate Limiting

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/kolosys/ion/ratelimit"
)

func main() {
    // Create a token bucket: 100 requests/second, burst of 20
    limiter := ratelimit.NewTokenBucket(
        ratelimit.PerSecond(100),
        20,
        ratelimit.WithName("api-limiter"),
    )

    // Fast path: check if request is allowed immediately
    if limiter.AllowN(time.Now(), 1) {
        fmt.Println("Request allowed immediately")
    }

    // Blocking wait: wait for rate limit if needed
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()

    if err := limiter.WaitN(ctx, 5); err != nil {
        log.Printf("Rate limit exceeded: %v", err)
        return
    }

    fmt.Printf("5 requests processed. Tokens remaining: %.1f\\n", limiter.Tokens())

    // Alternative: Leaky bucket for smooth rate limiting
    leaky := ratelimit.NewLeakyBucket(ratelimit.PerSecond(50), 10)
    if leaky.AllowN(time.Now(), 3) {
        fmt.Printf("3 requests queued. Queue level: %.1f/%d\\n",
            leaky.Level(), leaky.Capacity())
    }
}
```

## Design Principles

- **Context-First**: All long-lived operations accept context for cancellation/timeouts
- **No Panics**: Library code returns errors instead of panicking
- **Minimal Dependencies**: Core functionality has zero external dependencies
- **Pluggable Observability**: Optional logging, metrics, and tracing hooks
- **Deterministic Behavior**: Predictable semantics under load and shutdown
- **Thread-Safe**: All public APIs are safe for concurrent use

## Performance Targets

- **workerpool**: <200ns Submit hot path, 0 allocations per Submit in steady state
- **semaphore**: Acquire/Release <150ns uncontended, fairness overhead <20%
- **ratelimit**: Performance within 10-20% of golang.org/x/time/rate

## Architecture

Ion uses a single-module architecture with clearly separated packages:

```
github.com/kolosys/ion/
├── go.mod           # Single module for all components
├── workerpool/      # Worker pool implementation
├── semaphore/       # Weighted semaphore with fairness modes
├── ratelimit/       # Token bucket and leaky bucket rate limiters
├── shared/          # Common types and interfaces
├── circuit/         # Circuit breaker (planned)
├── pipeline/        # Pipeline helpers (planned)
└── examples/        # Complete examples for each component
```

This provides a cohesive concurrency toolkit while keeping components well-organized.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

Licensed under the [MIT License](LICENSE).
