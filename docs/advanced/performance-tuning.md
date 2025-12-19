# Performance Tuning

This guide covers performance optimization strategies for Ion components.

## Performance Characteristics

Ion is designed for high performance:

- **<200ns hot paths** for critical operations
- **Zero allocations** in steady state
- **Lock-free algorithms** where possible
- **Minimal overhead** when not configured

## Benchmarking

### Running Benchmarks

Use Go's built-in benchmarking:

```bash
# Run all benchmarks
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkCircuitBreaker_Execute -benchmem ./circuit

# Run with race detector
go test -bench=. -race ./...
```

### Interpreting Results

```bash
BenchmarkCircuitBreaker_Execute-8    10000000    150 ns/op    0 B/op    0 allocs/op
```

- **ns/op**: Nanoseconds per operation
- **B/op**: Bytes allocated per operation
- **allocs/op**: Number of allocations per operation

## Circuit Breaker Performance

### Minimize State Checks

Circuit breakers use atomic operations for state checks, which are very fast. However, you can optimize further:

```go
// Good: Fast path check
if !cb.allowRequest() {
    return nil, NewCircuitOpenError(cb.name)
}
```

### Use Failure Predicates Efficiently

Keep failure predicates simple:

```go
// Good: Simple predicate
circuit.WithFailurePredicate(func(err error) bool {
    return err != nil && isServerError(err)
})

// Bad: Complex predicate with allocations
circuit.WithFailurePredicate(func(err error) bool {
    return strings.Contains(err.Error(), "timeout") // Allocates
})
```

### Avoid Unnecessary Metrics

Disable metrics if not needed:

```go
// For high-throughput scenarios, use no-op metrics
obs := observe.New() // No-op by default
cb := circuit.New("service", circuit.WithObservability(obs))
```

## Rate Limiter Performance

### Token Bucket Optimization

Token bucket uses mutex for synchronization. For very high throughput, consider:

```go
// Use multiple limiters for sharding
limiters := make([]ratelimit.Limiter, shardCount)
for i := range limiters {
    limiters[i] = ratelimit.NewTokenBucket(rate/shardCount, burst/shardCount)
}

// Route requests to different limiters
limiter := limiters[hash(key)%shardCount]
```

### Leaky Bucket Optimization

Leaky bucket is optimized for steady-state processing:

```go
// Good: Appropriate capacity
limiter := ratelimit.NewLeakyBucket(
    ratelimit.PerSecond(1000),
    2000, // 2x rate for burst capacity
)
```

### Multi-Tier Optimization

Multi-tier limiter uses sync.Map for route/resource lookups:

```go
// Pre-warm routes for better performance
for route, config := range routeConfigs {
    limiter.GetRouteLimiter(route) // Pre-create
}
```

## Semaphore Performance

### Fairness Mode Impact

Fairness mode affects performance:

```go
// Maximum performance: None fairness
sem := semaphore.NewWeighted(100, semaphore.WithFairness(semaphore.None))

// Balanced: FIFO fairness (default)
sem := semaphore.NewWeighted(100, semaphore.WithFairness(semaphore.FIFO))
```

**Performance order**: None > LIFO > FIFO

### Reduce Contention

Use multiple semaphores for different resource types:

```go
// Instead of one large semaphore
dbSem := semaphore.NewWeighted(10)
fileSem := semaphore.NewWeighted(5)
apiSem := semaphore.NewWeighted(20)
```

## Worker Pool Performance

### Optimal Worker Count

Size workers based on workload:

```go
// CPU-bound: Number of CPUs
pool := workerpool.New(runtime.NumCPU(), 100)

// I/O-bound: More workers
pool := workerpool.New(runtime.NumCPU()*2, 200)
```

### Queue Size Optimization

Size queue based on expected backlog:

```go
// Too small: Frequent rejections
pool := workerpool.New(10, 5) // Bad

// Too large: Memory waste
pool := workerpool.New(10, 10000) // Bad

// Balanced: Based on workload
pool := workerpool.New(10, 100) // Good
```

### Task Wrapper Overhead

Minimize task wrapper overhead:

```go
// Good: Minimal wrapper
pool := workerpool.New(10, 100,
    workerpool.WithTaskWrapper(func(task workerpool.Task) workerpool.Task {
        return task // Minimal overhead
    }),
)

// Bad: Heavy wrapper
pool := workerpool.New(10, 100,
    workerpool.WithTaskWrapper(func(task workerpool.Task) workerpool.Task {
        return func(ctx context.Context) error {
            // Heavy instrumentation
            start := time.Now()
            defer func() {
                // Complex logging/metrics
            }()
            return task(ctx)
        }
    }),
)
```

## Observability Performance

### No-Op Overhead

No-op implementations have zero overhead:

```go
// Zero overhead when not configured
obs := observe.New() // No-op logger, metrics, tracer
```

### Efficient Logging

Use structured logging efficiently:

```go
// Good: Minimal allocations
logger.Info("message", "key", value)

// Bad: String formatting allocates
logger.Info(fmt.Sprintf("message: %v", value))
```

### Batch Metrics

Batch metrics when possible:

```go
// Instead of individual increments
for i := 0; i < 100; i++ {
    metrics.Inc("counter") // 100 calls
}

// Batch if your metrics system supports it
metrics.Add("counter", 100) // 1 call
```

## Memory Optimization

### Reuse Components

Reuse components instead of creating new ones:

```go
// Good: Reuse
var globalCB circuit.CircuitBreaker

func init() {
    globalCB = circuit.New("service")
}

// Bad: Create new each time
func handler() {
    cb := circuit.New("service") // Allocates
}
```

### Pool Objects

Use object pooling for high-frequency allocations:

```go
var requestPool = sync.Pool{
    New: func() any {
        return &Request{}
    },
}

func getRequest() *Request {
    return requestPool.Get().(*Request)
}

func putRequest(r *Request) {
    r.Reset()
    requestPool.Put(r)
}
```

## CPU Optimization

### Reduce Lock Contention

Use multiple components to reduce contention:

```go
// Instead of one large component
limiters := make([]ratelimit.Limiter, shardCount)
for i := range limiters {
    limiters[i] = ratelimit.NewTokenBucket(rate/shardCount, burst/shardCount)
}
```

### Avoid Unnecessary Work

Check conditions before expensive operations:

```go
// Good: Fast path check
if !limiter.AllowN(time.Now(), 1) {
    return errors.New("rate limited")
}
// Expensive operation only if allowed

// Bad: Always do expensive work
result := expensiveOperation()
if !limiter.AllowN(time.Now(), 1) {
    return errors.New("rate limited")
}
```

## Profiling

### CPU Profiling

Identify CPU bottlenecks:

```bash
go test -cpuprofile=cpu.prof -bench=. ./...
go tool pprof cpu.prof
```

### Memory Profiling

Identify memory allocations:

```bash
go test -memprofile=mem.prof -bench=. ./...
go tool pprof mem.prof
```

### Trace Analysis

Analyze execution traces:

```bash
go test -trace=trace.out -bench=. ./...
go tool trace trace.out
```

## Optimization Checklist

- [ ] Run benchmarks to establish baseline
- [ ] Profile to identify bottlenecks
- [ ] Use appropriate component sizes
- [ ] Minimize allocations in hot paths
- [ ] Use no-op observability when not needed
- [ ] Reduce lock contention
- [ ] Reuse components when possible
- [ ] Choose appropriate fairness modes
- [ ] Size queues appropriately
- [ ] Monitor performance metrics

## Real-World Optimization Example

### High-Throughput API Gateway

```go
// Optimized configuration for high throughput
type OptimizedGateway struct {
    // Sharded rate limiters
    limiters []ratelimit.Limiter

    // Per-route circuit breakers (reused)
    circuits map[string]circuit.CircuitBreaker

    // Worker pool for request processing
    pool *workerpool.Pool
}

func NewOptimizedGateway() *OptimizedGateway {
    // Shard rate limiters
    shardCount := runtime.NumCPU()
    limiters := make([]ratelimit.Limiter, shardCount)
    for i := range limiters {
        limiters[i] = ratelimit.NewTokenBucket(
            ratelimit.PerSecond(1000/shardCount),
            2000/shardCount,
        )
    }

    // Reuse circuit breakers
    circuits := make(map[string]circuit.CircuitBreaker)

    // Sized worker pool
    pool := workerpool.New(runtime.NumCPU()*2, 200)

    return &OptimizedGateway{
        limiters: limiters,
        circuits: circuits,
        pool:     pool,
    }
}

func (g *OptimizedGateway) HandleRequest(ctx context.Context, req *Request) error {
    // Shard rate limiting
    limiter := g.limiters[hash(req.UserID)%len(g.limiters)]
    if !limiter.AllowN(time.Now(), 1) {
        return errors.New("rate limited")
    }

    // Get or create circuit breaker
    cb := g.getCircuitBreaker(req.Route)

    // Process in worker pool
    return g.pool.Submit(ctx, func(ctx context.Context) error {
        _, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
            return g.processRequest(ctx, req)
        })
        return err
    })
}
```

## Further Reading

- [Best Practices](./best-practices.md) - Recommended patterns
- [API Reference](../api-reference/) - Complete API documentation
- [Go Performance Best Practices](https://github.com/dgryski/go-perfbook)
- [Go Blog: Profiling Go Programs](https://go.dev/blog/pprof)
