# WorkerPool

[![Go Reference](https://pkg.go.dev/badge/github.com/kolosys/ion/workerpool.svg)](https://pkg.go.dev/github.com/kolosys/ion/workerpool)

Production-grade bounded worker pools with context-aware submission, graceful shutdown, and comprehensive observability.

## Features

- **Bounded Execution**: Configurable worker count and queue size for predictable resource usage
- **Context-Aware**: All operations respect context cancellation and timeouts
- **Graceful Shutdown**: Clean shutdown with `Close()` and `Drain()` methods
- **Panic Recovery**: Built-in panic handling with optional custom recovery handlers
- **Observability**: Comprehensive metrics, logging, and tracing support
- **Task Wrapping**: Optional task instrumentation and middleware support
- **Zero Dependencies**: No external dependencies beyond the Go standard library

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kolosys/ion/workerpool"
)

func main() {
    // Create a pool with 4 workers and queue size of 20
    pool := workerpool.New(4, 20, workerpool.WithName("image-processor"))
    defer pool.Close(context.Background())

    // Submit work with context cancellation
    for i := 0; i < 100; i++ {
        taskID := i
        err := pool.Submit(context.Background(), func(ctx context.Context) error {
            // Simulate work
            time.Sleep(100 * time.Millisecond)
            fmt.Printf("Processed task %d\n", taskID)
            return nil
        })

        if err != nil {
            fmt.Printf("Failed to submit task %d: %v\n", taskID, err)
        }
    }

    // Graceful shutdown waits for completion
    pool.Drain(context.Background())
    fmt.Printf("Completed: %d tasks\n", pool.Metrics().Completed)
}
```

### Non-Blocking Submission

```go
// TrySubmit returns immediately if the queue is full
err := pool.TrySubmit(func(ctx context.Context) error {
    return processData(ctx)
})

if err != nil {
    // Handle queue full or pool closed
    fmt.Printf("Submission failed: %v\n", err)
}
```

### Error Handling and Observability

```go
// Custom logger
logger := &customLogger{}

pool := workerpool.New(2, 5,
    workerpool.WithName("api-processor"),
    workerpool.WithLogger(logger),
    workerpool.WithPanicRecovery(func(r any) {
        log.Printf("Task panicked: %v", r)
    }),
)

// Tasks that return errors are logged automatically
pool.Submit(ctx, func(ctx context.Context) error {
    if rand.Float64() < 0.1 {
        return errors.New("simulated error")
    }
    return nil
})
```

## API Reference

### Pool Creation

```go
func New(size, queueSize int, opts ...Option) *Pool
```

Creates a new worker pool with the specified worker count and queue capacity.

**Parameters:**

- `size`: Number of worker goroutines (0 = GOMAXPROCS)
- `queueSize`: Maximum queued tasks (0 = unbounded)
- `opts`: Configuration options

### Task Submission

```go
func (p *Pool) Submit(ctx context.Context, task Task) error
func (p *Pool) TrySubmit(task Task) error
```

**Submit** blocks until the task is queued or context is canceled.
**TrySubmit** returns immediately if the queue is full.

### Lifecycle Management

```go
func (p *Pool) Close(ctx context.Context) error
func (p *Pool) Drain(ctx context.Context) error
```

**Close** immediately stops accepting new tasks and waits for workers to finish.
**Drain** stops accepting new tasks and waits for the queue to empty.

### Monitoring

```go
func (p *Pool) Metrics() PoolMetrics
func (p *Pool) IsClosed() bool
func (p *Pool) IsDraining() bool
```

## Configuration Options

### Basic Options

```go
workerpool.WithName("my-pool")                    // Set pool name for observability
workerpool.WithBaseContext(ctx)                  // Set base context for all tasks
workerpool.WithDrainTimeout(30*time.Second)      // Default timeout for Drain operations
```

### Observability

```go
workerpool.WithLogger(logger)                    // Custom logger
workerpool.WithMetrics(metrics)                  // Custom metrics recorder
workerpool.WithTracer(tracer)                    // Custom tracer
```

### Advanced Features

```go
workerpool.WithPanicRecovery(func(r any) {       // Custom panic handler
    log.Printf("Panic recovered: %v", r)
})

workerpool.WithTaskWrapper(func(task Task) Task { // Task instrumentation
    return func(ctx context.Context) error {
        start := time.Now()
        err := task(ctx)
        log.Printf("Task took %v", time.Since(start))
        return err
    }
})
```

## Metrics

The pool provides comprehensive runtime metrics:

```go
type PoolMetrics struct {
    Size      int    // configured pool size
    Queued    int64  // current queue length
    Running   int64  // currently running tasks
    Completed uint64 // total completed tasks
    Failed    uint64 // total failed tasks
    Panicked  uint64 // total panicked tasks
}
```

## Error Handling

The workerpool package defines several error types for different failure scenarios:

- **Pool Closed**: Task submission to a closed pool
- **Queue Full**: Non-blocking submission when queue is full
- **Context Canceled**: Task submission canceled by context

```go
import "github.com/kolosys/ion/workerpool"

err := pool.Submit(ctx, task)
if err != nil {
    var poolErr *workerpool.PoolError
    if errors.As(err, &poolErr) {
        // Handle pool-specific errors
        fmt.Printf("Pool error: %v", poolErr)
    }
}
```

## Best Practices

### Sizing Guidelines

- **Workers**: Start with `runtime.GOMAXPROCS(0)` and adjust based on workload
- **Queue Size**: 2-5x worker count for CPU-bound tasks, higher for I/O-bound
- **Task Granularity**: Aim for 1-100ms task duration for optimal throughput

### Resource Management

```go
// Always ensure graceful shutdown
defer func() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := pool.Drain(ctx); err != nil {
        log.Printf("Drain timeout: %v", err)
        pool.Close(context.Background())
    }
}()
```

### Context Usage

```go
// Use context for task coordination
pool.Submit(ctx, func(taskCtx context.Context) error {
    select {
    case <-taskCtx.Done():
        return taskCtx.Err() // Respect cancellation
    case <-time.After(workDuration):
        return nil
    }
})
```

## Examples

- [Basic Usage](../examples/workerpool/main.go) - Simple task processing
- [HTTP Request Processing](../examples/workerpool/main.go) - API endpoint with worker pool
- [Batch Processing](../examples/workerpool/main.go) - Large dataset processing

## Performance

Benchmark results on modern hardware:

- **Submit**: <200ns (uncontended), <1μs (high contention)
- **Throughput**: 1M+ tasks/second
- **Memory**: 0 allocations in steady state
- **Latency**: <1ms p99 under load

## Thread Safety

All Pool methods are safe for concurrent use. Tasks execute concurrently in separate goroutines with proper synchronization.

## Contributing

See the main [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

## License

Licensed under the [MIT License](../LICENSE).
