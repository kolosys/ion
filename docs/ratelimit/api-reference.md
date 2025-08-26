# ratelimit API Reference
## Functions
### BenchmarkComparison_AllowN

Benchmark comparing with golang.org/x/time/rate for reference


```go
func BenchmarkComparison_AllowN
```
### BenchmarkHighContention

High contention benchmark


```go
func BenchmarkHighContention
```
### BenchmarkLeakyBucketAlloc



```go
func BenchmarkLeakyBucketAlloc
```
### BenchmarkLeakyBucketAllowN



```go
func BenchmarkLeakyBucketAllowN
```
### BenchmarkLeakyBucketAllowN_Uncontended



```go
func BenchmarkLeakyBucketAllowN_Uncontended
```
### BenchmarkLeakyBucketAllowN_WithLeak



```go
func BenchmarkLeakyBucketAllowN_WithLeak
```
### BenchmarkLeakyBucketWaitN



```go
func BenchmarkLeakyBucketWaitN
```
### BenchmarkScalability

Benchmark different burst/capacity sizes


```go
func BenchmarkScalability
```
### BenchmarkTokenBucketAlloc

Memory allocation benchmarks


```go
func BenchmarkTokenBucketAlloc
```
### BenchmarkTokenBucketAllowN



```go
func BenchmarkTokenBucketAllowN
```
### BenchmarkTokenBucketAllowN_Uncontended



```go
func BenchmarkTokenBucketAllowN_Uncontended
```
### BenchmarkTokenBucketAllowN_WithRefill



```go
func BenchmarkTokenBucketAllowN_WithRefill
```
### BenchmarkTokenBucketWaitN



```go
func BenchmarkTokenBucketWaitN
```
### TestConcurrency



```go
func TestConcurrency
```
### TestLeakyBucketAllowN



```go
func TestLeakyBucketAllowN
```
### TestLeakyBucketNew



```go
func TestLeakyBucketNew
```
### TestLeakyBucketWaitN



```go
func TestLeakyBucketWaitN
```
### TestRate



```go
func TestRate
```
### TestTokenBucketAllowN



```go
func TestTokenBucketAllowN
```
### TestTokenBucketNew



```go
func TestTokenBucketNew
```
### TestTokenBucketWaitN



```go
func TestTokenBucketWaitN
```
### TestZeroRate



```go
func TestZeroRate
```
## Types
### Clock

Clock abstracts time operations for testability.


```go
type Clock
```
### LeakyBucket

LeakyBucket implements a leaky bucket rate limiter.
Requests are added to the bucket, and the bucket leaks at a constant rate.
If the bucket is full, requests are denied or must wait.


```go
type LeakyBucket
```
#### Methods
##### AllowN

AllowN reports whether n requests can be added to the bucket at time now.
It returns true if the requests were accepted, false otherwise.


```go
func AllowN
```
##### Available

Available returns the number of requests that can be immediately accepted.


```go
func Available
```
##### Capacity

Capacity returns the bucket capacity.


```go
func Capacity
```
##### Level

Level returns the current level of the bucket.


```go
func Level
```
##### Rate

Rate returns the current leak rate.


```go
func Rate
```
##### WaitN

WaitN blocks until n requests can be added to the bucket or the context is canceled.


```go
func WaitN
```
##### leakLocked

leakLocked removes requests from the bucket based on elapsed time.
Must be called with lb.mu held.


```go
func leakLocked
```
##### waitSlow

waitSlow handles the blocking wait for bucket space.


```go
func waitSlow
```
### Limiter

Limiter represents a rate limiter that controls the rate at which events are allowed to occur.


```go
type Limiter
```
### Option

Option configures rate limiter behavior.


```go
type Option
```
### Rate

Rate represents the rate at which tokens are added to the bucket.


```go
type Rate
```
#### Methods
##### String

String returns a string representation of the rate.


```go
func String
```
### Timer

Timer represents a timer that can be stopped.


```go
type Timer
```
### TokenBucket

TokenBucket implements a token bucket rate limiter.
Tokens are added to the bucket at a fixed rate, and requests consume tokens.
If no tokens are available, requests must wait or are denied.


```go
type TokenBucket
```
#### Methods
##### AllowN

AllowN reports whether n tokens are available at time now.
It returns true if the tokens were consumed, false otherwise.


```go
func AllowN
```
##### Burst

Burst returns the bucket capacity.


```go
func Burst
```
##### Rate

Rate returns the current token refill rate.


```go
func Rate
```
##### Tokens

Tokens returns the current number of available tokens.


```go
func Tokens
```
##### WaitN

WaitN blocks until n tokens are available or the context is canceled.


```go
func WaitN
```
##### refillLocked

refillLocked adds tokens to the bucket based on elapsed time.
Must be called with tb.mu held.


```go
func refillLocked
```
##### waitSlow

waitSlow handles the blocking wait for tokens.


```go
func waitSlow
```
### config



```go
type config
```
### realClock

realClock implements Clock using the real time functions.


```go
type realClock
```
#### Methods
##### AfterFunc



```go
func AfterFunc
```
##### Now



```go
func Now
```
##### Sleep



```go
func Sleep
```
### realTimer

realTimer wraps time.Timer to implement our Timer interface.


```go
type realTimer
```
#### Methods
##### Stop



```go
func Stop
```
### testClock

testClock is a controllable clock implementation for testing.


```go
type testClock
```
#### Methods
##### Advance

Advance advances the clock by the given duration and fires any timers.


```go
func Advance
```
##### AfterFunc



```go
func AfterFunc
```
##### Now



```go
func Now
```
##### Set

Set sets the clock to a specific time.


```go
func Set
```
##### Sleep



```go
func Sleep
```
### testTimer



```go
type testTimer
```
#### Methods
##### Stop



```go
func Stop
```
