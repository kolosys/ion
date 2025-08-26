# workerpool API Reference
## Functions
### TestMetrics



```go
func TestMetrics(t *testing.T)
```
### TestNew



```go
func TestNew(t *testing.T)
```
### TestPoolLifecycle



```go
func TestPoolLifecycle(t *testing.T)
```
### TestSubmit



```go
func TestSubmit(t *testing.T)
```
### TestTaskPanicRecovery



```go
func TestTaskPanicRecovery(t *testing.T)
```
### TestTrySubmit



```go
func TestTrySubmit(t *testing.T)
```
## Types
### Option

Option configures pool behavior


```go
type Option func(*config)
```
#### Underlying Type

```go
func(*config)
```
### Pool

Pool represents a bounded worker pool that executes tasks with controlled
concurrency and queue management.


```go
type Pool struct {
	// Configuration
	name         string
	size         int
	queueSize    int
	drainTimeout time.Duration

	// Observability
	obs *shared.Observability

	// Lifecycle management
	baseCtx   context.Context
	cancel    context.CancelFunc
	closed    chan struct{}
	draining  atomic.Bool
	closeOnce sync.Once
	drainOnce sync.Once

	// Task management
	taskCh   chan taskSubmission
	workerWg sync.WaitGroup

	// Metrics
	metrics PoolMetrics

	// Panic recovery
	panicHandler func(any)
	taskWrapper  func(Task) Task
}
```
#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | `string` | Configuration |
| `size` | `int` |  |
| `queueSize` | `int` |  |
| `drainTimeout` | `time.Duration` |  |
| `obs` | `*shared.Observability` | Observability |
| `baseCtx` | `context.Context` | Lifecycle management |
| `cancel` | `context.CancelFunc` |  |
| `closed` | `chan struct{}` |  |
| `draining` | `atomic.Bool` |  |
| `closeOnce` | `sync.Once` |  |
| `drainOnce` | `sync.Once` |  |
| `taskCh` | `chan taskSubmission` | Task management |
| `workerWg` | `sync.WaitGroup` |  |
| `metrics` | `PoolMetrics` | Metrics |
| `panicHandler` | `func(any)` | Panic recovery |
| `taskWrapper` | `func(Task) Task` |  |
#### Methods
##### Close

Close immediately stops accepting new tasks and signals all workers to stop.
It waits for currently running tasks to complete unless the provided context
is canceled or times out. If the context expires, workers are asked to stop
via task context cancellation.


```go
func (p *Pool) Close(ctx context.Context) error
```
##### Drain

Drain prevents new task submissions and waits for the queue to empty and all
currently running tasks to complete. Unlike Close, Drain allows queued tasks
to continue being processed until the queue is empty.


```go
func (p *Pool) Drain(ctx context.Context) error
```
##### IsClosed

IsClosed returns true if the pool has been closed or is in the process of closing


```go
func (p *Pool) IsClosed() bool
```
##### IsDraining

IsDraining returns true if the pool is in draining mode (not accepting new tasks
but still processing queued tasks)


```go
func (p *Pool) IsDraining() bool
```
##### Metrics

Metrics returns a snapshot of the current pool metrics


```go
func (p *Pool) Metrics() PoolMetrics
```
##### Submit

Submit submits a task to the pool for execution. It respects the provided context
for cancellation and timeouts. If the context is canceled before the task can be
queued, it returns the context error wrapped. If the pool is closed or draining,
it returns an appropriate error.


```go
func (p *Pool) Submit(ctx context.Context, task Task) error
```
##### TrySubmit

TrySubmit attempts to submit a task to the pool without blocking.
It returns true if the task was successfully queued, false if the queue is full
or the pool is closed/draining. It does not respect context cancellation since
it returns immediately.


```go
func (p *Pool) TrySubmit(task Task) error
```
##### executeTask

executeTask executes a single task with proper error handling and metrics


```go
func (p *Pool) executeTask(submission taskSubmission, workerID int)
```
##### worker

worker runs the main worker loop


```go
func (p *Pool) worker(id int)
```
### PoolMetrics

PoolMetrics holds runtime metrics for the pool


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
#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `Size` | `int` | configured pool size |
| `Queued` | `int64` | current queue length |
| `Running` | `int64` | currently running tasks |
| `Completed` | `uint64` | total completed tasks |
| `Failed` | `uint64` | total failed tasks |
| `Panicked` | `uint64` | total panicked tasks |
### Task

Task represents a unit of work to be executed by the worker pool.
Tasks receive a context that will be canceled if either the submission
context or the pool's base context is canceled.


```go
type Task func(ctx context.Context) error
```
#### Underlying Type

```go
func(ctx context.Context) error
```
### config



```go
type config struct {
	name         string
	baseCtx      context.Context
	drainTimeout time.Duration
	obs          *shared.Observability
	panicHandler func(any)
	taskWrapper  func(Task) Task
}
```
#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | `string` |  |
| `baseCtx` | `context.Context` |  |
| `drainTimeout` | `time.Duration` |  |
| `obs` | `*shared.Observability` |  |
| `panicHandler` | `func(any)` |  |
| `taskWrapper` | `func(Task) Task` |  |
### taskSubmission

taskSubmission wraps a task with its submission context


```go
type taskSubmission struct {
	task Task
	ctx  context.Context
}
```
#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `task` | `Task` |  |
| `ctx` | `context.Context` |  |
