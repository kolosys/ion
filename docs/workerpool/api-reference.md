# workerpool API Reference
## Functions
### TestMetrics



```go
func TestMetrics
```
### TestNew



```go
func TestNew
```
### TestPoolLifecycle



```go
func TestPoolLifecycle
```
### TestSubmit



```go
func TestSubmit
```
### TestTaskPanicRecovery



```go
func TestTaskPanicRecovery
```
### TestTrySubmit



```go
func TestTrySubmit
```
## Types
### Option

Option configures pool behavior


```go
type Option
```
### Pool

Pool represents a bounded worker pool that executes tasks with controlled
concurrency and queue management.


```go
type Pool
```
#### Methods
##### Close

Close immediately stops accepting new tasks and signals all workers to stop.
It waits for currently running tasks to complete unless the provided context
is canceled or times out. If the context expires, workers are asked to stop
via task context cancellation.


```go
func Close
```
##### Drain

Drain prevents new task submissions and waits for the queue to empty and all
currently running tasks to complete. Unlike Close, Drain allows queued tasks
to continue being processed until the queue is empty.


```go
func Drain
```
##### IsClosed

IsClosed returns true if the pool has been closed or is in the process of closing


```go
func IsClosed
```
##### IsDraining

IsDraining returns true if the pool is in draining mode (not accepting new tasks
but still processing queued tasks)


```go
func IsDraining
```
##### Metrics

Metrics returns a snapshot of the current pool metrics


```go
func Metrics
```
##### Submit

Submit submits a task to the pool for execution. It respects the provided context
for cancellation and timeouts. If the context is canceled before the task can be
queued, it returns the context error wrapped. If the pool is closed or draining,
it returns an appropriate error.


```go
func Submit
```
##### TrySubmit

TrySubmit attempts to submit a task to the pool without blocking.
It returns true if the task was successfully queued, false if the queue is full
or the pool is closed/draining. It does not respect context cancellation since
it returns immediately.


```go
func TrySubmit
```
##### executeTask

executeTask executes a single task with proper error handling and metrics


```go
func executeTask
```
##### worker

worker runs the main worker loop


```go
func worker
```
### PoolMetrics

PoolMetrics holds runtime metrics for the pool


```go
type PoolMetrics
```
### Task

Task represents a unit of work to be executed by the worker pool.
Tasks receive a context that will be canceled if either the submission
context or the pool's base context is canceled.


```go
type Task
```
### config



```go
type config
```
### taskSubmission

taskSubmission wraps a task with its submission context


```go
type taskSubmission
```
