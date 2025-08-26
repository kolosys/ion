# semaphore API Reference
## Functions
### TestAcquire



```go
func TestAcquire
```
### TestConcurrency



```go
func TestConcurrency
```
### TestFairness



```go
func TestFairness
```
### TestNewWeighted



```go
func TestNewWeighted
```
### TestRelease



```go
func TestRelease
```
### TestTryAcquire



```go
func TestTryAcquire
```
## Types
### Fairness

Fairness defines the ordering behavior for semaphore waiters


```go
type Fairness
```
#### Methods
##### String

String returns the string representation of the fairness mode


```go
func String
```
### Option

Option configures semaphore behavior


```go
type Option
```
### Semaphore

Semaphore represents a weighted semaphore that controls access to a resource
with a fixed capacity. It supports configurable fairness modes and observability.


```go
type Semaphore
```
### config



```go
type config
```
### waiter

waiter represents a goroutine waiting to acquire permits


```go
type waiter
```
### waiterQueue

waiterQueue manages the queue of waiting goroutines based on fairness mode


```go
type waiterQueue
```
#### Methods
##### len

len returns the number of waiters in the queue


```go
func len
```
##### popReady

popReady removes and returns the first waiter that can be satisfied


```go
func popReady
```
##### push

push adds a waiter to the queue according to fairness policy


```go
func push
```
##### removeWaiter

removeWaiter removes a specific waiter from the queue (for cancellation)


```go
func removeWaiter
```
### weightedSemaphore

weightedSemaphore implements the Semaphore interface with weighted permits and fairness


```go
type weightedSemaphore
```
#### Methods
##### Acquire

Acquire blocks until n permits are available or the context is canceled.
Returns an error if n is invalid, exceeds capacity, or if the context is canceled.


```go
func Acquire
```
##### Current

Current returns the number of permits currently available


```go
func Current
```
##### Release

Release returns n permits to the semaphore, potentially unblocking waiters.
Panics if n is negative or if releasing would exceed the semaphore capacity.


```go
func Release
```
##### TryAcquire

TryAcquire attempts to acquire n permits without blocking.
Returns true if successful, false otherwise.


```go
func TryAcquire
```
##### acquireSlow

acquireSlow handles the blocking acquisition path


```go
func acquireSlow
```
##### notifyWaiters

notifyWaiters attempts to satisfy waiting acquire requests
Must be called with s.mu held


```go
func notifyWaiters
```
##### tryAcquireFast

tryAcquireFast attempts to acquire permits without blocking


```go
func tryAcquireFast
```
