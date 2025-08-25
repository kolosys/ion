# Contributing to Kolosys/Ion

Thank you for your interest in contributing to Ion! This document provides guidelines and information for contributors to the Ion concurrency toolkit.

## 🎯 Project Overview

Ion is a collection of robust, context-aware concurrency and scheduling primitives for Go applications. It focuses on deterministic behavior, safe cancellation, and pluggable observability without heavy dependencies.

## 🚀 Getting Started

### Prerequisites

- **Go 1.22.4+** (check with `go version`)
- **Git** for version control
- **golangci-lint** for comprehensive linting (optional but recommended)

### Development Setup

1. **Fork and Clone**

   ```bash
   git clone https://github.com/YOUR_USERNAME/ion.git
   cd ion
   ```

2. **Install Dependencies**

   ```bash
   go mod download
   go mod tidy
   ```

3. **Run All Tests**

   ```bash
   go test ./...
   go test -race ./...  # Test for race conditions
   ```

4. **Run Tests with Coverage**

   ```bash
   go test -v -cover ./...
   go test -coverprofile=coverage.out ./...
   go tool cover -html=coverage.out  # View coverage in browser
   ```

5. **Format and Lint**

   ```bash
   go fmt ./...
   go vet ./...
   golangci-lint run  # If available
   ```

6. **Run Examples**
   ```bash
   go run examples/workerpool/main.go
   go run examples/semaphore/main.go
   go run examples/ratelimit/main.go
   ```

## 📦 Project Structure

Ion uses a single-module architecture with clearly separated packages:

```
github.com/kolosys/ion/
├── go.mod              # Single module for all components
├── workerpool/         # Worker pool implementation
├── semaphore/          # Weighted semaphore with fairness modes
├── ratelimit/          # Token bucket and leaky bucket rate limiters
├── shared/             # Common types and interfaces
├── circuit/            # Circuit breaker (planned v0.2)
├── pipeline/           # Pipeline helpers (planned v0.2)
├── scheduler/          # Task scheduling (planned v0.2)
└── examples/           # Complete examples for each component
```

## 📝 Contributing Guidelines

### Issue Reporting

Before creating an issue, please:

1. **Search existing issues** to avoid duplicates
2. **Use a clear, descriptive title**
3. **Specify the component** (workerpool, semaphore, ratelimit, etc.)
4. **Provide reproduction steps** for bugs
5. **Include Go version, OS, and Ion version**
6. **Add relevant code samples** when applicable

**Bug Report Template:**

````markdown
## Component

[workerpool/semaphore/ratelimit/shared/other]

## Bug Description

Brief description of the issue

## Steps to Reproduce

1. Step 1
2. Step 2
3. Step 3

## Expected Behavior

What should happen

## Actual Behavior

What actually happens

## Environment

- Go version: `go version`
- OS:
- Ion version/commit:

## Code Sample

```go
// Minimal code to reproduce the issue
```
````

### Feature Requests

When requesting features:

1. **Check the [PRD](PRD)** to see if it's already planned
2. **Explain the use case** and problem being solved
3. **Provide examples** of the proposed API
4. **Consider consistency** with existing Ion patterns
5. **Think about observability** (metrics, logging, tracing)

### Pull Request Process

1. **Create a feature branch**

   ```bash
   git checkout -b feature/workerpool-metrics
   # or
   git checkout -b fix/semaphore-fairness-bug
   ```

2. **Make your changes**

   - Follow the [coding standards](#coding-standards)
   - Add comprehensive tests
   - Update documentation and examples
   - Ensure all tests pass including race tests

3. **Test thoroughly**

   ```bash
   go test ./...
   go test -race ./...
   go test -run Example ./...  # Test documentation examples
   ```

4. **Commit your changes**

   ```bash
   git add .
   git commit -m "feat(workerpool): add custom metrics support"
   # or
   git commit -m "fix(semaphore): resolve FIFO fairness edge case"
   ```

5. **Push and create PR**

   ```bash
   git push origin feature/workerpool-metrics
   ```

6. **Fill out the PR template** with:
   - Component(s) affected
   - Description of changes
   - Related issue links
   - Testing details
   - Breaking changes (if any)
   - Performance impact

## 📋 Coding Standards

### Ion Design Principles

All contributions must follow Ion's core design principles:

1. **Context-First**: All long-lived operations accept `context.Context`
2. **No Panics**: Library code returns errors instead of panicking
3. **Minimal Dependencies**: Core functionality has zero external dependencies
4. **Pluggable Observability**: Optional logging, metrics, and tracing hooks
5. **Thread Safety**: All public APIs must be safe for concurrent use
6. **Deterministic Behavior**: Predictable behavior under load and cancellation

### Code Style

- **Follow standard Go formatting**: Use `go fmt` and `go vet`
- **Use meaningful variable names**: `ctx context.Context`, `poolSize int`
- **Add comprehensive documentation**: All exported types and functions
- **Keep functions focused**: Single responsibility principle
- **Prefer composition over complex hierarchies**

### Context Handling

```go
// ✅ Good: Accept context and respect cancellation
func (p *Pool) Submit(ctx context.Context, fn func()) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case p.queue <- task{fn: fn, ctx: ctx}:
        return nil
    }
}

// ❌ Bad: No context handling
func (p *Pool) Submit(fn func()) error {
    p.queue <- task{fn: fn}
    return nil
}
```

### Error Handling

Use Ion's error patterns:

```go
// ✅ Good: Contextual errors with structured information
func NewQueueFullError(name string, queueSize int) error {
    return &shared.PoolError{
        Msg:  fmt.Sprintf("queue is full (size: %d)", queueSize),
        Name: name,
    }
}

// ✅ Good: Error wrapping with context
if err := p.storage.Store(key, value); err != nil {
    return fmt.Errorf("failed to store in pool %q: %w", p.name, err)
}
```

### Observability Integration

All components should support observability hooks:

```go
type Pool struct {
    // ...
    obs *shared.Observability
}

func (p *Pool) Submit(ctx context.Context, fn func()) error {
    start := time.Now()
    defer func() {
        p.obs.Metrics.Histogram("ion_workerpool_submit_duration_seconds",
            time.Since(start).Seconds(), "pool_name", p.name)
    }()

    p.obs.Logger.Debug("submitting task", "pool_name", p.name)

    // ... implementation

    p.obs.Metrics.Inc("ion_workerpool_tasks_submitted_total",
        "pool_name", p.name, "result", "success")
    return nil
}
```

### Function Documentation

```go
// Submit submits a function to be executed by the worker pool.
// The function will be executed asynchronously by one of the pool workers.
//
// The provided context is used for cancellation during submission only.
// Task execution uses the pool's base context, not the submission context.
//
// Returns ErrPoolClosed if the pool has been closed.
// Returns ErrQueueFull if the queue is at capacity and cannot accept more tasks.
//
// Example:
//   err := pool.Submit(ctx, func() {
//       fmt.Println("Hello from worker")
//   })
//   if err != nil {
//       log.Printf("Failed to submit task: %v", err)
//   }
func (p *Pool) Submit(ctx context.Context, fn func()) error {
    // implementation
}
```

## 🧪 Testing Standards

### Test Coverage Requirements

- **Minimum 90% coverage** for new code
- **100% coverage** for critical paths (context cancellation, error handling)
- **Race condition testing** with `-race` flag
- **Documentation examples** must be testable

### Test Structure

1. **Unit Tests**: Test individual functions and methods

   ```go
   func TestPoolSubmit(t *testing.T) {
       tests := []struct {
           name        string
           poolSize    int
           queueSize   int
           taskCount   int
           expectError bool
       }{
           {
               name:      "successful submission",
               poolSize:  2,
               queueSize: 5,
               taskCount: 3,
           },
           // more test cases...
       }

       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               // test implementation
           })
       }
   }
   ```

2. **Integration Tests**: Test component interactions
3. **Race Tests**: Ensure thread safety
4. **Stress Tests**: High-load scenarios
5. **Example Tests**: Validate documentation

### Context Testing

Always test context cancellation:

```go
func TestPoolSubmit_ContextCancellation(t *testing.T) {
    pool := workerpool.New(1, 1) // Small pool to force blocking
    defer pool.Close()

    ctx, cancel := context.WithCancel(context.Background())

    // Fill the pool
    pool.Submit(context.Background(), func() { time.Sleep(time.Hour) })

    // This should block, then fail when context is canceled
    go func() {
        time.Sleep(50 * time.Millisecond)
        cancel()
    }()

    err := pool.Submit(ctx, func() {})
    if err != context.Canceled {
        t.Errorf("expected context.Canceled, got %v", err)
    }
}
```

### Benchmark Tests

Performance-critical code needs benchmarks:

```go
func BenchmarkPoolSubmit(b *testing.B) {
    pool := workerpool.New(runtime.GOMAXPROCS(0), 1000)
    defer pool.Close()

    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            pool.Submit(context.Background(), func() {})
        }
    })
}
```

### Example Tests

All public APIs need example tests:

```go
func ExamplePool_Submit() {
    pool := workerpool.New(2, 10)
    defer pool.Close()

    err := pool.Submit(context.Background(), func() {
        fmt.Println("Hello from worker!")
    })

    if err != nil {
        log.Printf("Failed to submit: %v", err)
    }

    // Output: Hello from worker!
}
```

## 🏗️ Component Development

### Adding New Components

When adding new components (like v0.2 circuit breaker):

1. **Create package directory** under the Ion root
2. **Follow naming conventions**: `circuit/breaker.go`, `circuit/breaker_test.go`
3. **Implement core interfaces** with consistent patterns
4. **Add observability hooks** (metrics, logging, tracing)
5. **Create comprehensive tests** including race tests
6. **Add example usage** in `examples/circuit/`
7. **Update README** with the new component

### Component Structure

```
component/
├── component.go         # Main implementation
├── options.go          # Configuration options
├── component_test.go   # Unit tests
├── example_test.go     # Example tests for documentation
├── benchmark_test.go   # Performance benchmarks
└── testutil.go        # Test utilities (if needed)
```

### Interface Design

Keep interfaces small and focused:

```go
// ✅ Good: Focused interface
type Limiter interface {
    AllowN(now time.Time, n int) bool
    WaitN(ctx context.Context, n int) error
}

// ❌ Bad: Too many responsibilities
type LimiterWithEverything interface {
    AllowN(now time.Time, n int) bool
    WaitN(ctx context.Context, n int) error
    Configure(options ...Option)
    Close() error
    Stats() Statistics
    // ... too many methods
}
```

### Options Pattern

Use functional options for configuration:

```go
type config struct {
    name     string
    obs      *shared.Observability
    // ... other options
}

type Option func(*config)

func WithName(name string) Option {
    return func(c *config) { c.name = name }
}

func New(rate Rate, burst int, opts ...Option) *TokenBucket {
    cfg := &config{
        obs: shared.NewObservability(), // Default
    }
    for _, opt := range opts {
        opt(cfg)
    }
    // ... use config
}
```

## 🔧 Development Workflow

### Branch Naming

- **Features**: `feat/circuit-breaker`, `feat/workerpool-metrics`
- **Bug fixes**: `fix/semaphore-fairness`, `fix/ratelimit-race`
- **Documentation**: `docs/workerpool-examples`, `docs/api-reference`
- **Refactoring**: `refactor/shared-errors`, `refactor/options-pattern`
- **Performance**: `perf/ratelimit-optimization`, `perf/reduce-allocations`

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/) with component scope:

- `feat(workerpool): add custom task context derivation`
- `fix(semaphore): resolve FIFO fairness race condition`
- `docs(ratelimit): add token bucket vs leaky bucket comparison`
- `test(shared): add comprehensive error type tests`
- `perf(ratelimit): optimize token bucket refill calculation`

### Component Prefixes

- `workerpool`: Worker pool related changes
- `semaphore`: Semaphore related changes
- `ratelimit`: Rate limiting related changes
- `shared`: Shared utilities and types
- `circuit`: Circuit breaker (v0.2+)
- `pipeline`: Pipeline helpers (v0.2+)
- `scheduler`: Task scheduler (v0.2+)

## 📚 Documentation Requirements

### Code Documentation

Every exported type and function needs documentation:

```go
// TokenBucket implements a token bucket rate limiter.
// Tokens are added to the bucket at a fixed rate, and requests consume tokens.
// If no tokens are available, requests must wait or are denied.
//
// TokenBucket is safe for concurrent use.
type TokenBucket struct {
    // ...
}

// AllowN reports whether n tokens are available at time now.
// It returns true if the tokens were consumed, false otherwise.
// This method never blocks and does not respect context cancellation.
//
// For blocking behavior that respects context, use WaitN instead.
func (tb *TokenBucket) AllowN(now time.Time, n int) bool {
    // ...
}
```

### README Updates

When adding features:

- Update the feature list in the main README
- Add usage examples showing the new functionality
- Update the architecture diagram if needed
- Add the component to the quick start section

### Example Applications

Each component needs a runnable example:

```go
// examples/circuit/main.go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kolosys/ion/circuit"
)

func main() {
    fmt.Println("Ion Circuit Breaker Example")

    // Create circuit breaker
    cb := circuit.New(circuit.Config{
        FailureThreshold: 5,
        RecoveryTimeout:  30 * time.Second,
    })

    // Use circuit breaker
    err := cb.Do(ctx, func() error {
        return callExternalService()
    })

    if err != nil {
        fmt.Printf("Circuit breaker: %v\n", err)
    }
}
```

## 🎯 Priority Areas

We're especially interested in contributions for:

### v0.1 Enhancements

- **Performance optimizations** for existing components
- **Additional options** for fine-tuning behavior
- **More comprehensive examples** and tutorials
- **Observability improvements** (better metrics, tracing)

### v0.2 Components

- **Circuit Breaker**: Failure detection and recovery
- **Pipeline**: Fan-in/fan-out with bounded channels
- **Scheduler**: Delayed and periodic task execution

### v0.3+ Features

- **Pool Management**: Dynamic worker pool sizing
- **Advanced Rate Limiting**: Distributed rate limiting
- **Monitoring Tools**: CLI tools and dashboards

### Cross-Component

- **Integration examples** showing multiple components working together
- **Performance benchmarks** comparing with standard library alternatives
- **Error handling improvements** with better error types and messages

## 🔬 Testing and Validation

### Running the Full Test Suite

```bash
# All tests
go test ./...

# With race detection
go test -race ./...

# With coverage
go test -cover ./...

# Benchmarks
go test -bench=. ./...

# Specific component
go test ./workerpool -v
go test ./ratelimit -bench=. -v

# Example tests
go test -run Example ./...
```

### Manual Testing

```bash
# Test examples work correctly
cd examples/workerpool && go run main.go
cd examples/semaphore && go run main.go
cd examples/ratelimit && go run main.go

# Performance testing under load
go test -bench=BenchmarkPoolSubmit -benchtime=10s ./workerpool
```

## ❓ Getting Help

- **Documentation**: Check the [README](README.md), [PRD](PRD), and code comments
- **Issues**: Search [existing issues](https://github.com/kolosys/ion/issues)
- **Discussions**: Use [GitHub Discussions](https://github.com/kolosys/ion/discussions)
- **Component Questions**: Tag issues with the relevant component label

## 📄 License

By contributing to Ion, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to Ion! Your help makes Go concurrency safer and more powerful. 🚀
