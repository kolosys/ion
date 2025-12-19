# Semaphore

**Import Path:** `github.com/kolosys/ion/semaphore`

Semaphores control access to shared resources with weighted permits and configurable fairness modes. They're essential for managing concurrent access to limited resources like database connections, file handles, or memory.

## Overview

Semaphores allow a fixed number of concurrent operations. When all permits are acquired, additional requests must wait until permits are released.

### When to Use Semaphores

- **Database Connection Pools**: Limit concurrent database connections
- **File Operations**: Control concurrent file access
- **Memory Management**: Limit memory-intensive operations
- **External Resource Access**: Control access to external APIs or services
- **CPU-Bound Tasks**: Limit CPU-intensive operations

## Architecture

```
┌─────────────────────┐
│   Semaphore         │
│   Capacity: 10      │
├─────────────────────┤
│ Available: 7        │
│ Acquired: 3         │
│ Waiting: 2          │
└─────────────────────┘
```

### Fairness Modes

1. **FIFO (First-In-First-Out)**: Waiters are processed in order of arrival (default)
2. **LIFO (Last-In-First-Out)**: Most recent waiters are processed first
3. **None**: No fairness guarantees, maximum performance

## Core Concepts

### Weighted Permits

Semaphores support weighted permits, allowing operations to acquire multiple permits:

```go
sem := semaphore.NewWeighted(10) // Total capacity: 10

// Acquire 1 permit
sem.Acquire(ctx, 1)

// Acquire 3 permits (for larger operations)
sem.Acquire(ctx, 3)

// Release permits
sem.Release(1)
sem.Release(3)
```

### Non-Blocking Acquisition

Try to acquire permits without blocking:

```go
if sem.TryAcquire(1) {
    // Permit acquired
    defer sem.Release(1)
    // Use resource
} else {
    // No permits available
}
```

### Blocking Acquisition

Wait for permits with context support:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := sem.Acquire(ctx, 1); err != nil {
    // Context canceled or timeout
    return err
}
defer sem.Release(1)

// Use resource
```

## Real-World Scenarios

### Scenario 1: Database Connection Pool

Limit concurrent database connections:

```go
package main

import (
    "context"
    "database/sql"
    "fmt"

    "github.com/kolosys/ion/semaphore"
    _ "github.com/lib/pq"
)

type PooledDB struct {
    db  *sql.DB
    sem semaphore.Semaphore
}

func NewPooledDB(dsn string, maxConnections int) (*PooledDB, error) {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }

    db.SetMaxOpenConns(maxConnections)

    return &PooledDB{
        db: db,
        sem: semaphore.NewWeighted(int64(maxConnections),
            semaphore.WithName("db-pool"),
            semaphore.WithFairness(semaphore.FIFO),
        ),
    }, nil
}

func (p *PooledDB) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
    // Acquire a connection permit
    if err := p.sem.Acquire(ctx, 1); err != nil {
        return nil, fmt.Errorf("failed to acquire connection: %w", err)
    }
    defer p.sem.Release(1)

    // Use the connection
    return p.db.QueryContext(ctx, query, args...)
}

func (p *PooledDB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
    if err := p.sem.Acquire(ctx, 1); err != nil {
        return nil, fmt.Errorf("failed to acquire connection: %w", err)
    }
    defer p.sem.Release(1)

    return p.db.ExecContext(ctx, query, args...)
}

func main() {
    db, err := NewPooledDB("postgres://...", 10)
    if err != nil {
        panic(err)
    }

    ctx := context.Background()
    rows, err := db.Query(ctx, "SELECT * FROM users")
    if err != nil {
        panic(err)
    }
    defer rows.Close()
}
```

### Scenario 2: File Operation Limiting

Control concurrent file operations:

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/kolosys/ion/semaphore"
)

type FileProcessor struct {
    sem semaphore.Semaphore
}

func NewFileProcessor(maxConcurrent int) *FileProcessor {
    return &FileProcessor{
        sem: semaphore.NewWeighted(int64(maxConcurrent),
            semaphore.WithName("file-processor"),
        ),
    }
}

func (fp *FileProcessor) ProcessFile(ctx context.Context, path string) error {
    // Acquire permit for file operation
    if err := fp.sem.Acquire(ctx, 1); err != nil {
        return fmt.Errorf("failed to acquire file permit: %w", err)
    }
    defer fp.sem.Release(1)

    // Process file
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()

    // Read and process file
    fmt.Printf("Processing %s\n", path)
    return nil
}
```

### Scenario 3: Memory-Intensive Operation Control

Limit memory-intensive operations:

```go
package main

import (
    "context"
    "fmt"

    "github.com/kolosys/ion/semaphore"
)

type ImageProcessor struct {
    sem semaphore.Semaphore
}

func NewImageProcessor(maxConcurrent int) *ImageProcessor {
    // Each image processing operation uses significant memory
    // Limit concurrent operations to prevent OOM
    return &ImageProcessor{
        sem: semaphore.NewWeighted(int64(maxConcurrent),
            semaphore.WithName("image-processor"),
        ),
    }
}

func (ip *ImageProcessor) ProcessImage(ctx context.Context, imageData []byte) error {
    // Acquire permit (each image uses ~50MB)
    // With capacity 4, max memory usage is ~200MB
    if err := ip.sem.Acquire(ctx, 1); err != nil {
        return fmt.Errorf("failed to acquire processing permit: %w", err)
    }
    defer ip.sem.Release(1)

    // Process image (memory-intensive)
    fmt.Printf("Processing image (%d bytes)\n", len(imageData))
    return nil
}
```

### Scenario 4: Weighted Resource Access

Use weighted permits for operations with different resource requirements:

```go
package main

import (
    "context"
    "fmt"

    "github.com/kolosys/ion/semaphore"
)

type ResourceManager struct {
    sem semaphore.Semaphore
}

func NewResourceManager() *ResourceManager {
    // Total capacity: 10 units
    return &ResourceManager{
        sem: semaphore.NewWeighted(10,
            semaphore.WithName("resource-manager"),
        ),
    }
}

func (rm *ResourceManager) ProcessSmall(ctx context.Context) error {
    // Small operation: 1 unit
    if err := rm.sem.Acquire(ctx, 1); err != nil {
        return err
    }
    defer rm.sem.Release(1)

    fmt.Println("Processing small operation")
    return nil
}

func (rm *ResourceManager) ProcessLarge(ctx context.Context) error {
    // Large operation: 5 units
    if err := rm.sem.Acquire(ctx, 5); err != nil {
        return err
    }
    defer rm.sem.Release(5)

    fmt.Println("Processing large operation")
    return nil
}

func main() {
    rm := NewResourceManager()
    ctx := context.Background()

    // Can run 10 small operations concurrently
    // Or 2 large operations concurrently
    // Or mix: 5 small + 1 large = 10 units
    rm.ProcessSmall(ctx)
    rm.ProcessLarge(ctx)
}
```

### Scenario 5: API Rate Limiting with Semaphore

Use semaphore to limit concurrent API calls:

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kolosys/ion/semaphore"
)

type APIClient struct {
    sem semaphore.Semaphore
}

func NewAPIClient(maxConcurrent int) *APIClient {
    return &APIClient{
        sem: semaphore.NewWeighted(int64(maxConcurrent),
            semaphore.WithName("api-client"),
            semaphore.WithFairness(semaphore.FIFO),
        ),
    }
}

func (c *APIClient) CallAPI(ctx context.Context, endpoint string) error {
    // Limit concurrent API calls
    if err := c.sem.Acquire(ctx, 1); err != nil {
        return fmt.Errorf("failed to acquire API permit: %w", err)
    }
    defer c.sem.Release(1)

    // Make API call
    fmt.Printf("Calling %s\n", endpoint)
    time.Sleep(100 * time.Millisecond)
    return nil
}

func main() {
    client := NewAPIClient(5) // Max 5 concurrent API calls
    ctx := context.Background()

    // Make multiple API calls - semaphore limits concurrency
    for i := 0; i < 10; i++ {
        go func(id int) {
            if err := client.CallAPI(ctx, fmt.Sprintf("/api/v1/users/%d", id)); err != nil {
                fmt.Printf("Error: %v\n", err)
            }
        }(i)
    }

    time.Sleep(2 * time.Second)
}
```

## Fairness Modes

### FIFO (First-In-First-Out)

Waiters are processed in order of arrival:

```go
sem := semaphore.NewWeighted(10,
    semaphore.WithFairness(semaphore.FIFO), // Default
)
```

**Use when:**

- Fairness is important
- Operations have similar duration
- Preventing starvation is critical

### LIFO (Last-In-First-Out)

Most recent waiters are processed first:

```go
sem := semaphore.NewWeighted(10,
    semaphore.WithFairness(semaphore.LIFO),
)
```

**Use when:**

- Recent requests are more important
- Cache locality matters
- Throughput is prioritized over fairness

### None

No fairness guarantees, maximum performance:

```go
sem := semaphore.NewWeighted(10,
    semaphore.WithFairness(semaphore.None),
)
```

**Use when:**

- Maximum performance is critical
- Fairness is not a concern
- Operations are very short-lived

## Configuration Options

```go
sem := semaphore.NewWeighted(10,
    semaphore.WithName("my-semaphore"),
    semaphore.WithFairness(semaphore.FIFO),
    semaphore.WithAcquireTimeout(5*time.Second),
    semaphore.WithLogger(myLogger),
    semaphore.WithMetrics(myMetrics),
)
```

## Best Practices

1. **Always Release Permits**: Use `defer` to ensure permits are released
2. **Use Context Timeouts**: Always use context with timeouts for Acquire
3. **Choose Appropriate Capacity**: Balance between resource usage and throughput
4. **Monitor Semaphore Metrics**: Track wait times and permit usage
5. **Use Weighted Permits**: Use weighted permits for operations with different resource needs
6. **Consider Fairness**: Choose fairness mode based on your use case

## Common Pitfalls

### Pitfall 1: Not Releasing Permits

**Problem**: Deadlock when all permits are acquired but never released

```go
// Bad: Permit never released
sem.Acquire(ctx, 1)
// ... operation ...
// Forgot to release!
```

**Solution**: Always use defer

```go
// Good: Always released
if err := sem.Acquire(ctx, 1); err != nil {
    return err
}
defer sem.Release(1)
```

### Pitfall 2: Releasing More Than Acquired

**Problem**: Panic when releasing more permits than acquired

```go
// Bad
sem.Acquire(ctx, 1)
sem.Release(2) // Panic!
```

**Solution**: Always release exactly what you acquired

```go
// Good
weight := int64(1)
sem.Acquire(ctx, weight)
defer sem.Release(weight)
```

### Pitfall 3: Not Using Context

**Problem**: Acquire blocks indefinitely

```go
// Bad: No timeout
sem.Acquire(context.Background(), 1)
```

**Solution**: Always use context with timeout

```go
// Good: With timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
if err := sem.Acquire(ctx, 1); err != nil {
    return err
}
```

### Pitfall 4: Wrong Capacity

**Problem**: Too small capacity causes contention, too large wastes resources

**Solution**: Size based on actual resource constraints

```go
// Good: Based on actual database connection limit
sem := semaphore.NewWeighted(int64(db.MaxOpenConns()))
```

## Integration Guide

### With Worker Pools

```go
pool := workerpool.New(10, 100)
sem := semaphore.NewWeighted(5) // Limit resource access

pool.Submit(ctx, func(ctx context.Context) error {
    if err := sem.Acquire(ctx, 1); err != nil {
        return err
    }
    defer sem.Release(1)
    // Use limited resource
    return nil
})
```

### With Circuit Breakers

```go
cb := circuit.New("service")
sem := semaphore.NewWeighted(10)

_, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
    if err := sem.Acquire(ctx, 1); err != nil {
        return nil, err
    }
    defer sem.Release(1)
    // Protected operation
    return operation(ctx)
})
```

## Further Reading

- [API Reference](../api-reference/semaphore.md) - Complete API documentation
- [Examples](../examples/semaphore/) - Practical examples
- [Best Practices](../advanced/best-practices.md) - Recommended patterns
- [Performance Tuning](../advanced/performance-tuning.md) - Optimization strategies
