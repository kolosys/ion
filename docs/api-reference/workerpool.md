# workerpool API

Complete API documentation for the workerpool package.

**Import Path:** `github.com/kolosys/ion/workerpool`

## Package Documentation

Package workerpool provides a bounded worker pool with context-aware submission,
graceful shutdown, and observability hooks.


## Types

### Option
Option configures pool behavior

#### Example Usage

```go
// Example usage of Option
var value Option
// Initialize with appropriate value
```

#### Type Definition

```go
type Option func(*config)
```

### Constructor Functions

### WithBaseContext

WithBaseContext sets the base context for the pool. All task contexts will be derived from this context.

```go
func WithBaseContext(ctx context.Context) Option
```

**Parameters:**
- `ctx` (context.Context)

**Returns:**
- Option

### WithDrainTimeout

WithDrainTimeout sets the default timeout for Drain operations

```go
func WithDrainTimeout(timeout time.Duration) Option
```

**Parameters:**
- `timeout` (time.Duration)

**Returns:**
- Option

### WithLogger

WithLogger sets the logger for observability

```go
func WithLogger(logger observe.Logger) Option
```

**Parameters:**
- `logger` (observe.Logger)

**Returns:**
- Option

### WithMetrics

WithMetrics sets the metrics recorder for observability

```go
func WithMetrics(metrics observe.Metrics) Option
```

**Parameters:**
- `metrics` (observe.Metrics)

**Returns:**
- Option

### WithName

WithName sets the pool name for observability and error reporting

```go
func WithName(name string) Option
```

**Parameters:**
- `name` (string)

**Returns:**
- Option

### WithPanicRecovery

WithPanicRecovery sets a custom panic handler for task execution. If not set, panics are recovered and counted in metrics.

```go
func WithPanicRecovery(handler func(any)) Option
```

**Parameters:**
- `handler` (func(any))

**Returns:**
- Option

### WithTaskWrapper

WithTaskWrapper sets a function to wrap tasks for instrumentation. The wrapper is applied to every submitted task.

```go
func WithTaskWrapper(wrapper func(Task) Task) Option
```

**Parameters:**
- `wrapper` (func(Task) Task)

**Returns:**
- Option

### WithTracer

WithTracer sets the tracer for observability

```go
func WithTracer(tracer observe.Tracer) Option
```

**Parameters:**
- `tracer` (observe.Tracer)

**Returns:**
- Option

### Pool
Pool represents a bounded worker pool that executes tasks with controlled concurrency and queue management.

#### Example Usage

```go
// Create a new Pool
pool := Pool{

}
```

#### Type Definition

```go
type Pool struct {
}
```

### Constructor Functions

### New

New creates a new worker pool with the specified size and queue capacity. size determines the number of worker goroutines. queueSize determines the maximum number of queued tasks.

```go
func New(size, queueSize int, opts ...Option) *Pool
```

**Parameters:**
- `size` (int)
- `queueSize` (int)
- `opts` (...Option)

**Returns:**
- *Pool

## Methods

### Close

Close immediately stops accepting new tasks and signals all workers to stop. It waits for currently running tasks to complete unless the provided context is canceled or times out. If the context expires, workers are asked to stop via task context cancellation.

```go
func (*Pool) Close(ctx context.Context) error
```

**Parameters:**
- `ctx` (context.Context)

**Returns:**
- error

### Drain

Drain prevents new task submissions and waits for the queue to empty and all currently running tasks to complete. Unlike Close, Drain allows queued tasks to continue being processed until the queue is empty.

```go
func (*Pool) Drain(ctx context.Context) error
```

**Parameters:**
- `ctx` (context.Context)

**Returns:**
- error

### GetName

GetName returns the name of the pool

```go
func (*Pool) GetName() string
```

**Parameters:**
  None

**Returns:**
- string

### GetQueueSize

GetQueueSize returns the queue size of the pool

```go
func (*Pool) GetQueueSize() int
```

**Parameters:**
  None

**Returns:**
- int

### GetSize

GetSize returns the size of the pool

```go
func (*Pool) GetSize() int
```

**Parameters:**
  None

**Returns:**
- int

### IsClosed

IsClosed returns true if the pool has been closed or is in the process of closing

```go
func (*Pool) IsClosed() bool
```

**Parameters:**
  None

**Returns:**
- bool

### IsDraining

IsDraining returns true if the pool is in draining mode (not accepting new tasks but still processing queued tasks)

```go
func (*Pool) IsDraining() bool
```

**Parameters:**
  None

**Returns:**
- bool

### Metrics

Metrics returns a snapshot of the current pool metrics

```go
func (*Pool) Metrics() PoolMetrics
```

**Parameters:**
  None

**Returns:**
- PoolMetrics

### Submit

Submit submits a task to the pool for execution. It respects the provided context for cancellation and timeouts. If the context is canceled before the task can be queued, it returns the context error wrapped. If the pool is closed or draining, it returns an appropriate error.

```go
func (*Pool) Submit(ctx context.Context, task Task) error
```

**Parameters:**
- `ctx` (context.Context)
- `task` (Task)

**Returns:**
- error

### TrySubmit

TrySubmit attempts to submit a task to the pool without blocking. It returns true if the task was successfully queued, false if the queue is full or the pool is closed/draining. It does not respect context cancellation since it returns immediately.

```go
func (*Pool) TrySubmit(task Task) error
```

**Parameters:**
- `task` (Task)

**Returns:**
- error

### PoolError
PoolError represents workerpool-specific errors with context

#### Example Usage

```go
// Create a new PoolError
poolerror := PoolError{
    Op: "example",
    PoolName: "example",
    Err: error{},
}
```

#### Type Definition

```go
type PoolError struct {
    Op string
    PoolName string
    Err error
}
```

### Fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| Op | `string` | operation that failed |
| PoolName | `string` | name of the pool |
| Err | `error` | underlying error |

## Methods

### Error



```go
func (*PoolError) Error() string
```

**Parameters:**
  None

**Returns:**
- string

### Unwrap



```go
func (*PoolError) Unwrap() error
```

**Parameters:**
  None

**Returns:**
- error

### PoolMetrics
PoolMetrics holds runtime metrics for the pool

#### Example Usage

```go
// Create a new PoolMetrics
poolmetrics := PoolMetrics{
    Size: 42,
    Queued: 42,
    Running: 42,
    Completed: 42,
    Failed: 42,
    Panicked: 42,
}
```

#### Type Definition

```go
type PoolMetrics struct {
    Size int
    Queued int64
    Running int64
    Completed uint64
    Failed uint64
    Panicked uint64
}
```

### Fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| Size | `int` | configured pool size |
| Queued | `int64` | current queue length |
| Running | `int64` | currently running tasks |
| Completed | `uint64` | total completed tasks |
| Failed | `uint64` | total failed tasks |
| Panicked | `uint64` | total panicked tasks |

### Task
Task represents a unit of work to be executed by the worker pool. Tasks receive a context that will be canceled if either the submission context or the pool's base context is canceled.

#### Example Usage

```go
// Example usage of Task
var value Task
// Initialize with appropriate value
```

#### Type Definition

```go
type Task func(ctx context.Context) error
```

## Functions

### NewPoolClosedError
NewPoolClosedError creates an error indicating the pool is closed

```go
func NewPoolClosedError(poolName string) error
```

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `poolName` | `string` | |

**Returns:**
| Type | Description |
|------|-------------|
| `error` | |

**Example:**

```go
// Example usage of NewPoolClosedError
result := NewPoolClosedError(/* parameters */)
```

### NewQueueFullError
NewQueueFullError creates an error indicating the queue is full

```go
func NewQueueFullError(poolName string, queueSize int) error
```

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `poolName` | `string` | |
| `queueSize` | `int` | |

**Returns:**
| Type | Description |
|------|-------------|
| `error` | |

**Example:**

```go
// Example usage of NewQueueFullError
result := NewQueueFullError(/* parameters */)
```

## External Links

- [Package Overview](../packages/workerpool.md)
- [pkg.go.dev Documentation](https://pkg.go.dev/github.com/kolosys/ion/workerpool)
- [Source Code](https://github.com/kolosys/ion/tree/main/workerpool)
