# Semaphore

[![Go Reference](https://pkg.go.dev/badge/github.com/kolosys/ion/semaphore.svg)](https://pkg.go.dev/github.com/kolosys/ion/semaphore)

Weighted semaphores with configurable fairness modes for controlling access to limited resources.

## Features

- **Weighted Permits**: Support for variable-weight resource acquisition
- **Fairness Modes**: FIFO, LIFO, and no-fairness ordering policies
- **Context-Aware**: All operations respect context cancellation and timeouts
- **Non-Blocking Operations**: TryAcquire for immediate resource availability checks
- **Observability**: Built-in metrics, logging, and tracing support
- **Zero Dependencies**: No external dependencies beyond the Go standard library

## Quick Start

### Basic Resource Pool

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kolosys/ion/semaphore"
)

func main() {
    // Database connection pool with 10 connections
    dbSem := semaphore.NewWeighted(10,
        semaphore.WithName("postgres-pool"),
        semaphore.WithFairness(semaphore.FIFO),
    )

    // Acquire a connection
    if err := dbSem.Acquire(context.Background(), 1); err != nil {
        fmt.Printf("Failed to get connection: %v\n", err)
        return
    }
    defer dbSem.Release(1)

    fmt.Printf("Got connection, %d remaining\n", dbSem.Current())

    // Use database connection...
    time.Sleep(100 * time.Millisecond)
}
```

### Weighted Resource Management

```go
// CPU scheduler: different tasks require different core counts
cpuSem := semaphore.NewWeighted(8) // 8 CPU cores available

// Small task needs 1 core
go func() {
    if err := cpuSem.Acquire(ctx, 1); err != nil {
        return
    }
    defer cpuSem.Release(1)

    // Run lightweight task
    processSmallJob()
}()

// Large task needs 4 cores
go func() {
    if err := cpuSem.Acquire(ctx, 4); err != nil {
        return
    }
    defer cpuSem.Release(4)

    // Run compute-intensive task
    processLargeJob()
}()
```

### Non-Blocking Resource Checks

```go
// Try to acquire resource without blocking
if sem.TryAcquire(2) {
    defer sem.Release(2)

    // Got resources immediately
    fmt.Println("Processing with 2 units")
} else {
    // Resources not available, handle gracefully
    fmt.Println("Resources busy, trying later")
}
```

## API Reference

### Semaphore Creation

```go
func NewWeighted(capacity int64, opts ...Option) Semaphore
```

Creates a new weighted semaphore with the specified capacity.

**Parameters:**

- `capacity`: Maximum number of permits available
- `opts`: Configuration options

### Resource Acquisition

```go
func (s Semaphore) Acquire(ctx context.Context, n int64) error
func (s Semaphore) TryAcquire(n int64) bool
```

**Acquire** blocks until n permits are available or context is canceled.
**TryAcquire** returns immediately with success/failure status.

### Resource Release

```go
func (s Semaphore) Release(n int64)
func (s Semaphore) Current() int64
```

**Release** returns n permits to the semaphore.
**Current** returns the number of currently available permits.

## Configuration Options

### Basic Options

```go
semaphore.WithName("resource-pool")              // Set semaphore name for observability
semaphore.WithFairness(semaphore.FIFO)          // Set ordering policy
semaphore.WithAcquireTimeout(5*time.Second)     // Default timeout for acquisitions
```

### Fairness Modes

```go
semaphore.FIFO    // First-in-first-out (default)
semaphore.LIFO    // Last-in-first-out
semaphore.None    // No fairness guarantees (highest performance)
```

### Observability

```go
semaphore.WithLogger(logger)                    // Custom logger
semaphore.WithMetrics(metrics)                  // Custom metrics recorder
semaphore.WithTracer(tracer)                    // Custom tracer
```

## Use Cases

### Database Connection Pools

```go
// Limit concurrent database connections
dbPool := semaphore.NewWeighted(maxConnections,
    semaphore.WithName("database-pool"),
    semaphore.WithFairness(semaphore.FIFO),
)

func queryDatabase(ctx context.Context, query string) error {
    if err := dbPool.Acquire(ctx, 1); err != nil {
        return fmt.Errorf("connection timeout: %w", err)
    }
    defer dbPool.Release(1)

    // Execute database query
    return db.Query(ctx, query)
}
```

### Rate Limiting by Resource

```go
// Different rate limits per organization
orgLimits := make(map[string]semaphore.Semaphore)

func getOrgSemaphore(orgID string) semaphore.Semaphore {
    if sem, exists := orgLimits[orgID]; exists {
        return sem
    }

    // Create per-org semaphore
    sem := semaphore.NewWeighted(100, // 100 req/sec per org
        semaphore.WithName("org-"+orgID),
    )
    orgLimits[orgID] = sem
    return sem
}
```

### Memory Management

```go
// Limit memory-intensive operations
memSem := semaphore.NewWeighted(totalMemoryGB,
    semaphore.WithName("memory-limiter"),
)

func processLargeFile(ctx context.Context, file string, sizeGB int64) error {
    if err := memSem.Acquire(ctx, sizeGB); err != nil {
        return fmt.Errorf("insufficient memory: %w", err)
    }
    defer memSem.Release(sizeGB)

    // Process file using sizeGB of memory
    return process(file)
}
```

### CPU Core Allocation

```go
// Allocate CPU cores for different workload types
cpuSem := semaphore.NewWeighted(int64(runtime.NumCPU()),
    semaphore.WithName("cpu-scheduler"),
)

func runTask(ctx context.Context, task Task) error {
    cores := task.RequiredCores()

    if err := cpuSem.Acquire(ctx, cores); err != nil {
        return fmt.Errorf("CPU unavailable: %w", err)
    }
    defer cpuSem.Release(cores)

    // Run task with allocated cores
    return task.Execute()
}
```

## Error Handling

The semaphore package defines specific error types:

```go
import "github.com/kolosys/ion/semaphore"

err := sem.Acquire(ctx, 5)
if err != nil {
    var semErr *semaphore.SemaphoreError
    if errors.As(err, &semErr) {
        // Handle semaphore-specific errors
        fmt.Printf("Semaphore error: %v", semErr)
    }
}
```

**Common Errors:**

- `semaphore.ErrInvalidWeight`: Negative or zero weight requested
- `semaphore.NewWeightExceedsCapacityError()`: Requested weight exceeds semaphore capacity
- `semaphore.NewAcquireTimeoutError()`: Acquisition timed out

## Best Practices

### Resource Sizing

- **Database Pools**: Start with 2x CPU cores, adjust based on connection latency
- **Memory Limits**: Leave 20-30% headroom for system overhead
- **CPU Allocation**: Consider hyperthreading when setting core counts

### Error Handling

```go
// Always handle acquisition errors
if err := sem.Acquire(ctx, weight); err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        return fmt.Errorf("resource timeout: %w", err)
    }
    return fmt.Errorf("resource unavailable: %w", err)
}
defer sem.Release(weight)
```

### Context Usage

```go
// Use timeouts for bounded waiting
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := sem.Acquire(ctx, 1); err != nil {
    // Handle timeout or cancellation
    return err
}
```

### Fairness Considerations

- **FIFO**: Best for ensuring fair access across all callers
- **LIFO**: Useful for cache-like access patterns
- **None**: Maximum performance when fairness isn't required

## Fairness Examples

### FIFO Fairness

```go
// Requests are served in order of arrival
sem := semaphore.NewWeighted(1, semaphore.WithFairness(semaphore.FIFO))

// First request will be served first, even if later requests
// require fewer resources
```

### LIFO Fairness

```go
// Most recent requests are prioritized
sem := semaphore.NewWeighted(5, semaphore.WithFairness(semaphore.LIFO))

// Useful for stack-like processing where recent requests
// might be more relevant
```

### No Fairness

```go
// Requests are served based on resource availability
sem := semaphore.NewWeighted(10, semaphore.WithFairness(semaphore.None))

// Highest performance, but no ordering guarantees
// Smaller requests might be served before larger ones
```

## Examples

- [Basic Usage](../examples/semaphore/main.go) - Database connection pool simulation
- [Weighted Resources](../examples/semaphore/main.go) - CPU core allocation
- [Fairness Demo](../examples/semaphore/main.go) - Different fairness modes

## Performance

Benchmark results on modern hardware:

- **Acquire/Release**: <150ns (uncontended)
- **TryAcquire**: <50ns
- **Memory**: 0 allocations for acquire/release operations
- **Fairness Overhead**: <10% for FIFO/LIFO vs None

## Thread Safety

All Semaphore methods are safe for concurrent use. The implementation uses atomic operations and fine-grained locking for optimal performance.

## Contributing

See the main [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

## License

Licensed under the [MIT License](../LICENSE).
