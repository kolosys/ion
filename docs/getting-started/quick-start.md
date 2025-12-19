# Quick Start

This guide will help you get started with Ion quickly with practical examples for each package.

## Basic Usage

Here's a simple example demonstrating all Ion packages:

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kolosys/ion/circuit"
    "github.com/kolosys/ion/ratelimit"
    "github.com/kolosys/ion/semaphore"
    "github.com/kolosys/ion/workerpool"
)

func main() {
    ctx := context.Background()

    // Circuit Breaker: Protect external service calls
    cb := circuit.New("payment-service",
        circuit.WithFailureThreshold(5),
        circuit.WithRecoveryTimeout(30*time.Second),
    )

    // Rate Limiter: Control request rate
    limiter := ratelimit.NewTokenBucket(ratelimit.PerSecond(10), 20)

    // Semaphore: Limit concurrent operations
    sem := semaphore.NewWeighted(5)

    // Worker Pool: Process tasks concurrently
    pool := workerpool.New(4, 20)
    defer pool.Close(ctx)

    // Use them together
    for i := 0; i < 10; i++ {
        if limiter.AllowN(time.Now(), 1) {
            pool.Submit(ctx, func(ctx context.Context) error {
                if err := sem.Acquire(ctx, 1); err != nil {
                    return err
                }
                defer sem.Release(1)

                _, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
                    return processPayment(ctx, i)
                })
                return err
            })
        }
    }
}

func processPayment(ctx context.Context, id int) (string, error) {
    // Simulate payment processing
    return fmt.Sprintf("payment-%d", id), nil
}
```

## Circuit Breaker

Protect your services from cascading failures.

### Basic Example

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/kolosys/ion/circuit"
)

func main() {
    // Create a circuit breaker for a payment service
    cb := circuit.New("payment-service",
        circuit.WithFailureThreshold(5),        // Trip after 5 failures
        circuit.WithRecoveryTimeout(30*time.Second), // Wait 30s before retry
    )

    ctx := context.Background()

    // Execute operations with circuit breaker protection
    result, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
        // Your actual service call here
        return callPaymentService(ctx)
    })

    if err != nil {
        if circuit.IsCircuitOpen(err) {
            fmt.Println("Circuit is open - service unavailable")
        } else {
            fmt.Printf("Operation failed: %v\n", err)
        }
        return
    }

    fmt.Printf("Success: %v\n", result)
}

func callPaymentService(ctx context.Context) (string, error) {
    // Simulate service call
    return "payment-id-123", nil
}
```

### Real-World Scenario: HTTP Client Protection

```go
package main

import (
    "context"
    "net/http"
    "time"

    "github.com/kolosys/ion/circuit"
)

type ProtectedHTTPClient struct {
    client  *http.Client
    circuit circuit.CircuitBreaker
}

func NewProtectedHTTPClient() *ProtectedHTTPClient {
    return &ProtectedHTTPClient{
        client: &http.Client{
            Timeout: 5 * time.Second,
        },
        circuit: circuit.New("http-client",
            circuit.WithFailureThreshold(3),
            circuit.WithRecoveryTimeout(15*time.Second),
            circuit.WithFailurePredicate(func(err error) bool {
                // Only count 5xx errors and timeouts as failures
                // 4xx errors (client errors) should not trip the circuit
                return err != nil
            }),
        ),
    }
}

func (c *ProtectedHTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
    result, err := c.circuit.Execute(ctx, func(ctx context.Context) (any, error) {
        req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
        if err != nil {
            return nil, err
        }
        return c.client.Do(req)
    })

    if err != nil {
        return nil, err
    }

    return result.(*http.Response), nil
}
```

## Rate Limiting

Control the rate at which operations execute.

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kolosys/ion/ratelimit"
)

func main() {
    // Create a token bucket: 10 requests per second, burst of 20
    limiter := ratelimit.NewTokenBucket(ratelimit.PerSecond(10), 20)

    ctx := context.Background()

    // Check if request is allowed (non-blocking)
    if limiter.AllowN(time.Now(), 1) {
        fmt.Println("Request allowed")
        // Process request
    } else {
        fmt.Println("Request rate limited")
    }

    // Wait for rate limit (blocking)
    if err := limiter.WaitN(ctx, 1); err != nil {
        fmt.Printf("Rate limit wait failed: %v\n", err)
        return
    }
    fmt.Println("Request allowed after waiting")
}
```

### Real-World Scenario: API Client Rate Limiting

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kolosys/ion/ratelimit"
)

type RateLimitedAPIClient struct {
    limiter ratelimit.Limiter
}

func NewRateLimitedAPIClient() *RateLimitedAPIClient {
    // Respect API limits: 100 requests per minute with burst of 10
    return &RateLimitedAPIClient{
        limiter: ratelimit.NewTokenBucket(
            ratelimit.PerMinute(100),
            10,
            ratelimit.WithName("api-client"),
        ),
    }
}

func (c *RateLimitedAPIClient) CallAPI(ctx context.Context, endpoint string) error {
    // Wait for rate limit
    if err := c.limiter.WaitN(ctx, 1); err != nil {
        return fmt.Errorf("rate limit exceeded: %w", err)
    }

    // Make API call
    fmt.Printf("Calling %s\n", endpoint)
    return nil
}

func main() {
    client := NewRateLimitedAPIClient()
    ctx := context.Background()

    // Make multiple API calls - rate limiter will control the rate
    for i := 0; i < 5; i++ {
        if err := client.CallAPI(ctx, fmt.Sprintf("/api/v1/users/%d", i)); err != nil {
            fmt.Printf("Error: %v\n", err)
        }
        time.Sleep(100 * time.Millisecond)
    }
}
```

## Semaphore

Control access to shared resources with weighted semaphores.

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kolosys/ion/semaphore"
)

func main() {
    // Create a semaphore with capacity of 5
    sem := semaphore.NewWeighted(5,
        semaphore.WithName("db-connections"),
        semaphore.WithFairness(semaphore.FIFO),
    )

    ctx := context.Background()

    // Acquire a permit
    if err := sem.Acquire(ctx, 1); err != nil {
        fmt.Printf("Failed to acquire: %v\n", err)
        return
    }
    defer sem.Release(1)

    // Use the resource
    fmt.Println("Using shared resource")
    time.Sleep(100 * time.Millisecond)
}
```

### Real-World Scenario: Database Connection Pool

```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "time"

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
```

## Worker Pool

Execute tasks with bounded concurrency and graceful shutdown.

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kolosys/ion/workerpool"
)

func main() {
    // Create a worker pool: 4 workers, queue size 20
    pool := workerpool.New(4, 20,
        workerpool.WithName("task-processor"),
    )
    defer pool.Close(context.Background())

    ctx := context.Background()

    // Submit tasks
    for i := 0; i < 10; i++ {
        taskID := i
        pool.Submit(ctx, func(ctx context.Context) error {
            fmt.Printf("Processing task %d\n", taskID)
            time.Sleep(100 * time.Millisecond)
            return nil
        })
    }

    // Wait for tasks to complete
    time.Sleep(2 * time.Second)
}
```

### Real-World Scenario: Image Processing Pipeline

```go
package main

import (
    "context"
    "fmt"
    "image"
    "time"

    "github.com/kolosys/ion/workerpool"
)

type ImageProcessor struct {
    pool *workerpool.Pool
}

func NewImageProcessor() *ImageProcessor {
    return &ImageProcessor{
        pool: workerpool.New(4, 50, // 4 workers, 50 image queue
            workerpool.WithName("image-processor"),
        ),
    }
}

func (p *ImageProcessor) ProcessImage(ctx context.Context, img image.Image) error {
    return p.pool.Submit(ctx, func(ctx context.Context) error {
        // Process image: resize, compress, etc.
        fmt.Println("Processing image...")
        time.Sleep(500 * time.Millisecond)
        return nil
    })
}

func (p *ImageProcessor) Close(ctx context.Context) error {
    return p.pool.Close(ctx)
}

func main() {
    processor := NewImageProcessor()
    defer processor.Close(context.Background())

    ctx := context.Background()

    // Process multiple images
    for i := 0; i < 10; i++ {
        if err := processor.ProcessImage(ctx, nil); err != nil {
            fmt.Printf("Failed to submit image: %v\n", err)
        }
    }

    time.Sleep(5 * time.Second)
}
```

## Combining Components

Here's a complete example combining all Ion components:

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kolosys/ion/circuit"
    "github.com/kolosys/ion/ratelimit"
    "github.com/kolosys/ion/semaphore"
    "github.com/kolosys/ion/workerpool"
)

func main() {
    // Setup components
    cb := circuit.New("external-api",
        circuit.WithFailureThreshold(5),
        circuit.WithRecoveryTimeout(30*time.Second),
    )

    limiter := ratelimit.NewTokenBucket(ratelimit.PerSecond(10), 20)
    sem := semaphore.NewWeighted(5)
    pool := workerpool.New(4, 20)

    ctx := context.Background()
    defer pool.Close(ctx)

    // Process requests with all protections
    for i := 0; i < 20; i++ {
        requestID := i

        // Rate limit check
        if !limiter.AllowN(time.Now(), 1) {
            fmt.Printf("Request %d: rate limited\n", requestID)
            continue
        }

        // Submit to worker pool
        pool.Submit(ctx, func(ctx context.Context) error {
            // Acquire semaphore
            if err := sem.Acquire(ctx, 1); err != nil {
                return err
            }
            defer sem.Release(1)

            // Execute with circuit breaker
            _, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
                return processRequest(ctx, requestID)
            })

            return err
        })
    }

    time.Sleep(5 * time.Second)
}

func processRequest(ctx context.Context, id int) (string, error) {
    fmt.Printf("Processing request %d\n", id)
    time.Sleep(100 * time.Millisecond)
    return fmt.Sprintf("result-%d", id), nil
}
```

## Error Handling

Always handle errors appropriately:

```go
result, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
    return operation(ctx)
})

if err != nil {
    if circuit.IsCircuitOpen(err) {
        // Circuit is open - handle gracefully
        return handleCircuitOpen()
    }
    // Other error - handle normally
    return fmt.Errorf("operation failed: %w", err)
}
```

## Configuration Options

Ion uses functional options for flexible configuration:

```go
cb := circuit.New("service",
    circuit.WithFailureThreshold(5),
    circuit.WithRecoveryTimeout(30*time.Second),
    circuit.WithLogger(myLogger),
    circuit.WithMetrics(myMetrics),
)
```

## Next Steps

Now that you've seen the basics, explore:

- **[Core Concepts](../core-concepts/)** - Understanding each package in depth
- **[API Reference](../api-reference/)** - Complete API documentation
- **[Examples](../examples/)** - More detailed examples
- **[Advanced Topics](../advanced/)** - Performance tuning and advanced patterns

## Getting Help

If you run into issues:

1. Check the [API Reference](../api-reference/)
2. Browse the [Examples](../examples/)
3. Visit the [GitHub Issues](https://github.com/kolosys/ion/issues) page
