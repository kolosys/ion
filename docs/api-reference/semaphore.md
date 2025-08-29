# semaphore API
## Types
### Fairness

Fairness defines the ordering behavior for semaphore waiters


```go
type Fairness int
```
#### Underlying Type

```go
int
```
#### Methods
##### String

String returns the string representation of the fairness mode


```go
func (f Fairness) String() string
```
### Option

Option configures semaphore behavior


```go
type Option func(*config)
```
#### Underlying Type

```go
func(*config)
```
### Semaphore

Semaphore represents a weighted semaphore that controls access to a resource
with a fixed capacity. It supports configurable fairness modes and observability.


```go
type Semaphore interface {
	// Acquire blocks until n permits are available or the context is canceled.
	// Returns an error if the context is canceled or if n exceeds the semaphore capacity.
	Acquire(ctx context.Context, n int64) error

	// TryAcquire attempts to acquire n permits without blocking.
	// Returns true if the permits were acquired, false otherwise.
	TryAcquire(n int64) bool

	// Release returns n permits to the semaphore, potentially unblocking waiters.
	// Panics if n is negative or if more permits are released than were acquired.
	Release(n int64)

	// Current returns the number of permits currently available.
	Current() int64
}
```
### config



```go
type config struct {
	name           string
	fairness       Fairness
	acquireTimeout time.Duration
	obs            *shared.Observability
}
```
#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | `string` |  |
| `fairness` | `Fairness` |  |
| `acquireTimeout` | `time.Duration` |  |
| `obs` | `*shared.Observability` |  |
### waiter

waiter represents a goroutine waiting to acquire permits


```go
type waiter struct {
	weight   int64
	ready    chan struct{}
	ctx      context.Context
	acquired bool
}
```
#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `weight` | `int64` |  |
| `ready` | `chan struct{}` |  |
| `ctx` | `context.Context` |  |
| `acquired` | `bool` |  |
### waiterQueue

waiterQueue manages the queue of waiting goroutines based on fairness mode


```go
type waiterQueue struct {
	fairness Fairness
	waiters  []*waiter
}
```
#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `fairness` | `Fairness` |  |
| `waiters` | `[]*waiter` |  |
#### Methods
##### len

len returns the number of waiters in the queue


```go
func (q *waiterQueue) len() int
```
##### popReady

popReady removes and returns the first waiter that can be satisfied


```go
func (q *waiterQueue) popReady(available int64) *waiter
```
##### push

push adds a waiter to the queue according to fairness policy


```go
func (q *waiterQueue) push(w *waiter)
```
##### removeWaiter

removeWaiter removes a specific waiter from the queue (for cancellation)


```go
func (q *waiterQueue) removeWaiter(target *waiter) bool
```
### weightedSemaphore

weightedSemaphore implements the Semaphore interface with weighted permits and fairness


```go
type weightedSemaphore struct {
	// Configuration
	name           string
	capacity       int64
	fairness       Fairness
	acquireTimeout time.Duration

	// Observability
	obs *shared.Observability

	// Synchronization
	mu      sync.Mutex
	current int64
	waiters waiterQueue
	closed  bool
}
```
#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | `string` | Configuration |
| `capacity` | `int64` |  |
| `fairness` | `Fairness` |  |
| `acquireTimeout` | `time.Duration` |  |
| `obs` | `*shared.Observability` | Observability |
| `mu` | `sync.Mutex` | Synchronization |
| `current` | `int64` |  |
| `waiters` | `waiterQueue` |  |
| `closed` | `bool` |  |
#### Methods
##### Acquire

Acquire blocks until n permits are available or the context is canceled.
Returns an error if n is invalid, exceeds capacity, or if the context is canceled.


```go
func (s *weightedSemaphore) Acquire(ctx context.Context, n int64) error
```
##### Current

Current returns the number of permits currently available


```go
func (s *weightedSemaphore) Current() int64
```
##### Release

Release returns n permits to the semaphore, potentially unblocking waiters.
Panics if n is negative or if releasing would exceed the semaphore capacity.


```go
func (s *weightedSemaphore) Release(n int64)
```
##### TryAcquire

TryAcquire attempts to acquire n permits without blocking.
Returns true if successful, false otherwise.


```go
func (s *weightedSemaphore) TryAcquire(n int64) bool
```
##### acquireSlow

acquireSlow handles the blocking acquisition path


```go
func (s *weightedSemaphore) acquireSlow(ctx context.Context, n int64) error
```
##### notifyWaiters

notifyWaiters attempts to satisfy waiting acquire requests
Must be called with s.mu held


```go
func (s *weightedSemaphore) notifyWaiters()
```
##### tryAcquireFast

tryAcquireFast attempts to acquire permits without blocking


```go
func (s *weightedSemaphore) tryAcquireFast(n int64) bool
```
