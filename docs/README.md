# Ion - Concurrency Primitives for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/kolosys/ion.svg)](https://pkg.go.dev/github.com/kolosys/ion)
[![Go Report Card](https://goreportcard.com/badge/github.com/kolosys/ion)](https://goreportcard.com/report/github.com/kolosys/ion)

Ion provides a collection of robust, context-aware concurrency and scheduling primitives for Go applications. It focuses on deterministic behavior, safe cancellation, and pluggable observability without heavy dependencies.

## 🚀 Quick Start

```bash
go get github.com/kolosys/ion@latest
```

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kolosys/ion/workerpool"
    "github.com/kolosys/ion/ratelimit"
    "github.com/kolosys/ion/semaphore"
)

func main() {
    // Worker Pool: Process tasks concurrently
    pool := workerpool.New(4, 10) // 4 workers, queue size 10
    defer pool.Close(context.Background())

    task := func(ctx context.Context) error {
        fmt.Println("Processing task...")
        return nil
    }

    pool.Submit(context.Background(), task)

    // Rate Limiting: Control request rates
    limiter := ratelimit.NewTokenBucket(10, time.Second) // 10 req/sec

    if limiter.Allow() {
        fmt.Println("Request allowed")
    }

    // Semaphore: Manage resource access
    sem := semaphore.New(3) // Allow 3 concurrent operations
    defer sem.Close()

    sem.Acquire(context.Background(), 1)
    defer sem.Release(1)

    fmt.Println("Critical section")
}
```

## 📦 Components

- **[Worker Pool](workerpool/)** - Bounded worker pool with graceful shutdown
- **[Rate Limiting](ratelimit/)** - Token bucket and leaky bucket algorithms  
- **[Semaphore](semaphore/)** - Weighted semaphore with fairness modes
- **[Shared](shared/)** - Common utilities and observability interfaces

## 🎯 Design Principles

- **Context-First**: All long-lived operations accept context for cancellation/timeouts
- **No Panics**: Library code returns errors instead of panicking
- **Minimal Dependencies**: Core functionality has zero external dependencies
- **Pluggable Observability**: Optional logging, metrics, and tracing hooks
- **Deterministic Behavior**: Predictable semantics under load and shutdown
- **Thread-Safe**: All public APIs are safe for concurrent use

## 📈 Performance Targets

- **workerpool**: <200ns Submit hot path, 0 allocations per Submit in steady state
- **semaphore**: Acquire/Release <150ns uncontended, fairness overhead <20%
- **ratelimit**: Performance within 10-20% of golang.org/x/time/rate

## 🤝 Contributing

See [Contributing Guide](https://github.com/kolosys/ion/blob/main/CONTRIBUTING.md) for guidelines.

## 📄 License

Licensed under the [MIT License](https://github.com/kolosys/ion/blob/main/LICENSE).
