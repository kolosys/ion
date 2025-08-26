# Getting Started with Ion

Ion provides high-performance concurrency primitives for Go applications. This guide will help you get up and running quickly.

## Installation

```bash
go get github.com/kolosys/ion@latest
```

## Quick Start

Here's a simple example using all of Ion's core components:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/kolosys/ion/workerpool"
    "github.com/kolosys/ion/ratelimit"
    "github.com/kolosys/ion/semaphore"
)

func main() {
    ctx := context.Background()

    // Create a worker pool
    pool := workerpool.New(4, 20, workerpool.WithName("example-pool"))
    defer pool.Close(ctx)

    // Create a rate limiter (10 requests per second)
    limiter := ratelimit.NewTokenBucket(ratelimit.PerSecond(10), 10)

    // Create a semaphore (max 3 concurrent operations)
    sem := semaphore.NewWeighted(3)

    // Submit some tasks
    for i := 0; i < 20; i++ {
        taskID := i
        
        // Rate limit the submissions
        if err := limiter.WaitN(ctx, 1); err != nil {
            log.Printf("Rate limit error: %v", err)
            continue
        }
        
        task := func(ctx context.Context) error {
            // Acquire semaphore
            if err := sem.Acquire(ctx, 1); err != nil {
                return err
            }
            defer sem.Release(1)
            
            // Simulate work
            fmt.Printf("Processing task %d\n", taskID)
            time.Sleep(100 * time.Millisecond)
            return nil
        }
        
        if err := pool.Submit(ctx, task); err != nil {
            log.Printf("Failed to submit task %d: %v", taskID, err)
        }
    }

    // Wait for all tasks to complete
    if err := pool.Drain(ctx); err != nil {
        log.Printf("Failed to drain pool: %v", err)
    }

    fmt.Printf("Completed %d tasks\n", pool.Metrics().Completed)
}
```

## Next Steps

- [Worker Pool Documentation](workerpool/)
- [Rate Limiting Documentation](ratelimit/)
- [Semaphore Documentation](semaphore/)
- [API Reference](reference/api.md)
- [Examples](reference/examples.md)
