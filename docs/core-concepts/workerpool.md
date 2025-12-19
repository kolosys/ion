# Worker Pool

**Import Path:** `github.com/kolosys/ion/workerpool`

Worker pools execute tasks with bounded concurrency, graceful shutdown, and queue management. They're essential for processing background jobs, handling request queues, and managing concurrent operations.

## Overview

Worker pools provide a controlled way to execute tasks concurrently. They maintain a fixed number of worker goroutines and a queue for pending tasks, ensuring predictable resource usage and graceful shutdown.

### When to Use Worker Pools

- **Background Job Processing**: Process jobs from a queue
- **Request Handling**: Handle incoming requests with bounded concurrency
- **Task Processing**: Execute independent tasks concurrently
- **Resource Management**: Control resource usage with bounded workers
- **Graceful Shutdown**: Ensure in-flight tasks complete before shutdown

## Architecture

```
┌─────────────────────────────────┐
│         Worker Pool             │
├─────────────────────────────────┤
│  Workers: 4                     │
│  Queue Size: 20                 │
├─────────────────────────────────┤
│  ┌──────┐  ┌──────┐             │
│  │Worker│  │Worker│  ...        │
│  └──────┘  └──────┘             │
│     │         │                 │
│     └─────────┘                 │
│           │                     │
│     ┌──────────┐                │
│     │  Queue   │                │
│     └──────────┘                │
└─────────────────────────────────┘
```

### Components

1. **Workers**: Fixed number of goroutines that process tasks
2. **Queue**: Buffered channel holding pending tasks
3. **Task Context**: Each task receives a context that respects cancellation
4. **Metrics**: Built-in metrics for monitoring pool health

## Core Concepts

### Task Submission

Submit tasks to the pool with context support:

```go
pool := workerpool.New(4, 20) // 4 workers, queue size 20

ctx := context.Background()
err := pool.Submit(ctx, func(ctx context.Context) error {
    // Task implementation
    return processTask(ctx)
})
```

### Graceful Shutdown

Close the pool and wait for in-flight tasks:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := pool.Close(ctx); err != nil {
    // Timeout or error
}
```

### Context Propagation

Tasks receive a context that cancels when:

- The submission context is canceled
- The pool's base context is canceled
- The pool is closed

```go
pool.Submit(ctx, func(taskCtx context.Context) error {
    // taskCtx is canceled if submission ctx or pool ctx is canceled
    select {
    case <-taskCtx.Done():
        return taskCtx.Err()
    default:
        // Process task
    }
    return nil
})
```

## Real-World Scenarios

### Scenario 1: Image Processing Pipeline

Process images with bounded concurrency:

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

### Scenario 2: Email Sending Service

Send emails with rate limiting and graceful shutdown:

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kolosys/ion/workerpool"
)

type EmailService struct {
    pool *workerpool.Pool
}

type Email struct {
    To      string
    Subject string
    Body    string
}

func NewEmailService() *EmailService {
    return &EmailService{
        pool: workerpool.New(5, 100, // 5 workers, 100 email queue
            workerpool.WithName("email-service"),
        ),
    }
}

func (s *EmailService) SendEmail(ctx context.Context, email Email) error {
    return s.pool.Submit(ctx, func(ctx context.Context) error {
        // Send email
        fmt.Printf("Sending email to %s: %s\n", email.To, email.Subject)
        time.Sleep(200 * time.Millisecond)
        return nil
    })
}

func (s *EmailService) Shutdown(ctx context.Context) error {
    fmt.Println("Shutting down email service...")
    return s.pool.Close(ctx)
}

func main() {
    service := NewEmailService()

    ctx := context.Background()

    // Send multiple emails
    for i := 0; i < 20; i++ {
        email := Email{
            To:      fmt.Sprintf("user%d@example.com", i),
            Subject: "Test Email",
            Body:    "This is a test email",
        }
        if err := service.SendEmail(ctx, email); err != nil {
            fmt.Printf("Failed to queue email: %v\n", err)
        }
    }

    // Graceful shutdown
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := service.Shutdown(shutdownCtx); err != nil {
        fmt.Printf("Shutdown error: %v\n", err)
    }
}
```

### Scenario 3: Data Transformation Pipeline

Transform data with error handling and metrics:

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kolosys/ion/workerpool"
)

type DataTransformer struct {
    pool *workerpool.Pool
}

type DataRecord struct {
    ID   string
    Data string
}

func NewDataTransformer() *DataTransformer {
    return &DataTransformer{
        pool: workerpool.New(8, 200, // 8 workers, 200 record queue
            workerpool.WithName("data-transformer"),
            workerpool.WithPanicRecovery(func(r any) {
                fmt.Printf("Panic recovered: %v\n", r)
            }),
        ),
    }
}

func (dt *DataTransformer) Transform(ctx context.Context, record DataRecord) error {
    return dt.pool.Submit(ctx, func(ctx context.Context) error {
        // Transform data
        fmt.Printf("Transforming record %s\n", record.ID)
        time.Sleep(100 * time.Millisecond)
        return nil
    })
}

func (dt *DataTransformer) Metrics() workerpool.PoolMetrics {
    return dt.pool.Metrics()
}

func main() {
    transformer := NewDataTransformer()
    defer transformer.Close(context.Background())

    ctx := context.Background()

    // Transform multiple records
    for i := 0; i < 50; i++ {
        record := DataRecord{
            ID:   fmt.Sprintf("record-%d", i),
            Data: fmt.Sprintf("data-%d", i),
        }
        if err := transformer.Transform(ctx, record); err != nil {
            fmt.Printf("Failed to queue record: %v\n", err)
        }
    }

    time.Sleep(2 * time.Second)

    // Check metrics
    metrics := transformer.Metrics()
    fmt.Printf("Completed: %d, Failed: %d, Queued: %d\n",
        metrics.Completed, metrics.Failed, metrics.Queued)
}
```

### Scenario 4: HTTP Request Handler

Handle HTTP requests with worker pool:

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/kolosys/ion/workerpool"
)

type RequestHandler struct {
    pool *workerpool.Pool
}

func NewRequestHandler() *RequestHandler {
    return &RequestHandler{
        pool: workerpool.New(10, 100, // 10 workers, 100 request queue
            workerpool.WithName("http-handler"),
        ),
    }
}

func (h *RequestHandler) HandleRequest(ctx context.Context, req *http.Request) error {
    return h.pool.Submit(ctx, func(ctx context.Context) error {
        // Process HTTP request
        fmt.Printf("Handling request: %s %s\n", req.Method, req.URL.Path)
        time.Sleep(50 * time.Millisecond)
        return nil
    })
}

func (h *RequestHandler) HTTPHandler(w http.ResponseWriter, r *http.Request) {
    if err := h.HandleRequest(r.Context(), r); err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "Request queued")
}

func main() {
    handler := NewRequestHandler()
    defer handler.Close(context.Background())

    http.HandleFunc("/", handler.HTTPHandler)
    http.ListenAndServe(":8080", nil)
}
```

### Scenario 5: Event Processing System

Process events with task wrapping for instrumentation:

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kolosys/ion/workerpool"
)

type EventProcessor struct {
    pool *workerpool.Pool
}

type Event struct {
    Type string
    Data string
}

func NewEventProcessor() *EventProcessor {
    return &EventProcessor{
        pool: workerpool.New(6, 150,
            workerpool.WithName("event-processor"),
            workerpool.WithTaskWrapper(func(task workerpool.Task) workerpool.Task {
                // Wrap task for instrumentation
                return func(ctx context.Context) error {
                    start := time.Now()
                    defer func() {
                        duration := time.Since(start)
                        fmt.Printf("Event processed in %v\n", duration)
                    }()
                    return task(ctx)
                }
            }),
        ),
    }
}

func (ep *EventProcessor) ProcessEvent(ctx context.Context, event Event) error {
    return ep.pool.Submit(ctx, func(ctx context.Context) error {
        fmt.Printf("Processing event: %s - %s\n", event.Type, event.Data)
        time.Sleep(100 * time.Millisecond)
        return nil
    })
}

func main() {
    processor := NewEventProcessor()
    defer processor.Close(context.Background())

    ctx := context.Background()

    // Process events
    for i := 0; i < 20; i++ {
        event := Event{
            Type: "user_action",
            Data: fmt.Sprintf("action-%d", i),
        }
        if err := processor.ProcessEvent(ctx, event); err != nil {
            fmt.Printf("Failed to queue event: %v\n", err)
        }
    }

    time.Sleep(3 * time.Second)
}
```

## Configuration Options

```go
pool := workerpool.New(4, 20,
    workerpool.WithName("my-pool"),
    workerpool.WithBaseContext(ctx),
    workerpool.WithDrainTimeout(30*time.Second),
    workerpool.WithLogger(myLogger),
    workerpool.WithMetrics(myMetrics),
    workerpool.WithTracer(myTracer),
    workerpool.WithPanicRecovery(func(r any) {
        // Handle panic
    }),
    workerpool.WithTaskWrapper(func(task workerpool.Task) workerpool.Task {
        // Wrap task for instrumentation
        return task
    }),
)
```

## Best Practices

1. **Size Appropriately**: Balance workers and queue size based on workload
2. **Use Context Timeouts**: Always use context with timeouts for Close
3. **Handle Queue Full**: Check for queue full errors and handle appropriately
4. **Monitor Metrics**: Track pool metrics for health monitoring
5. **Graceful Shutdown**: Always use Close for graceful shutdown
6. **Error Handling**: Handle task errors appropriately
7. **Panic Recovery**: Use panic recovery for production systems

## Common Pitfalls

### Pitfall 1: Not Closing the Pool

**Problem**: Goroutines leak when pool is not closed

```go
// Bad: Pool never closed
pool := workerpool.New(4, 20)
// ... use pool ...
// Pool never closed, goroutines leak
```

**Solution**: Always close the pool

```go
// Good: Pool always closed
pool := workerpool.New(4, 20)
defer pool.Close(context.Background())
```

### Pitfall 2: Queue Too Small

**Problem**: Tasks rejected when queue is full

```go
// Bad: Queue too small
pool := workerpool.New(10, 5) // 10 workers, only 5 queue slots
```

**Solution**: Size queue appropriately

```go
// Good: Queue sized for workload
pool := workerpool.New(10, 100) // 10 workers, 100 queue slots
```

### Pitfall 3: Not Handling Queue Full

**Problem**: Tasks silently fail when queue is full

```go
// Bad: Ignore queue full error
pool.Submit(ctx, task)
```

**Solution**: Handle queue full errors

```go
// Good: Handle queue full
if err := pool.Submit(ctx, task); err != nil {
    if workerpool.IsQueueFull(err) {
        // Handle queue full: retry, reject, or backpressure
    }
    return err
}
```

### Pitfall 4: Not Using Context

**Problem**: Tasks don't respect cancellation

```go
// Bad: No context cancellation
pool.Submit(context.Background(), func(ctx context.Context) error {
    // Long-running operation that doesn't check ctx
    time.Sleep(10 * time.Second)
    return nil
})
```

**Solution**: Always respect context

```go
// Good: Respects context
pool.Submit(ctx, func(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-time.After(10 * time.Second):
        // Operation complete
    }
    return nil
})
```

## Integration Guide

### With Semaphores

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
pool := workerpool.New(10, 100)
cb := circuit.New("service")

pool.Submit(ctx, func(ctx context.Context) error {
    _, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
        return operation(ctx)
    })
    return err
})
```

### With Rate Limiters

```go
pool := workerpool.New(10, 100)
limiter := ratelimit.NewTokenBucket(ratelimit.PerSecond(10), 20)

pool.Submit(ctx, func(ctx context.Context) error {
    if !limiter.AllowN(time.Now(), 1) {
        return errors.New("rate limited")
    }
    // Process task
    return nil
})
```

## Metrics

Access pool metrics:

```go
metrics := pool.Metrics()

fmt.Printf("Size: %d\n", metrics.Size)
fmt.Printf("Queued: %d\n", metrics.Queued)
fmt.Printf("Running: %d\n", metrics.Running)
fmt.Printf("Completed: %d\n", metrics.Completed)
fmt.Printf("Failed: %d\n", metrics.Failed)
fmt.Printf("Panicked: %d\n", metrics.Panicked)
```

## Further Reading

- [API Reference](../api-reference/workerpool.md) - Complete API documentation
- [Examples](../examples/workerpool/) - Practical examples
- [Best Practices](../advanced/best-practices.md) - Recommended patterns
- [Performance Tuning](../advanced/performance-tuning.md) - Optimization strategies
