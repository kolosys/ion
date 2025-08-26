# ratelimit API Reference
## Functions
### BenchmarkComparison_AllowN

Benchmark comparing with golang.org/x/time/rate for reference


```go
func BenchmarkComparison_AllowN(b *testing.B)
```
### BenchmarkHighContention

High contention benchmark


```go
func BenchmarkHighContention(b *testing.B)
```
### BenchmarkLeakyBucketAlloc



```go
func BenchmarkLeakyBucketAlloc(b *testing.B)
```
### BenchmarkLeakyBucketAllowN



```go
func BenchmarkLeakyBucketAllowN(b *testing.B)
```
### BenchmarkLeakyBucketAllowN_Uncontended



```go
func BenchmarkLeakyBucketAllowN_Uncontended(b *testing.B)
```
### BenchmarkLeakyBucketAllowN_WithLeak



```go
func BenchmarkLeakyBucketAllowN_WithLeak(b *testing.B)
```
### BenchmarkLeakyBucketWaitN



```go
func BenchmarkLeakyBucketWaitN(b *testing.B)
```
### BenchmarkScalability

Benchmark different burst/capacity sizes


```go
func BenchmarkScalability(b *testing.B)
```
### BenchmarkTokenBucketAlloc

Memory allocation benchmarks


```go
func BenchmarkTokenBucketAlloc(b *testing.B)
```
### BenchmarkTokenBucketAllowN



```go
func BenchmarkTokenBucketAllowN(b *testing.B)
```
### BenchmarkTokenBucketAllowN_Uncontended



```go
func BenchmarkTokenBucketAllowN_Uncontended(b *testing.B)
```
### BenchmarkTokenBucketAllowN_WithRefill



```go
func BenchmarkTokenBucketAllowN_WithRefill(b *testing.B)
```
### BenchmarkTokenBucketWaitN



```go
func BenchmarkTokenBucketWaitN(b *testing.B)
```
### TestConcurrency



```go
func TestConcurrency(t *testing.T)
```
### TestLeakyBucketAllowN



```go
func TestLeakyBucketAllowN(t *testing.T)
```
### TestLeakyBucketNew



```go
func TestLeakyBucketNew(t *testing.T)
```
### TestLeakyBucketWaitN



```go
func TestLeakyBucketWaitN(t *testing.T)
```
### TestRate



```go
func TestRate(t *testing.T)
```
### TestTokenBucketAllowN



```go
func TestTokenBucketAllowN(t *testing.T)
```
### TestTokenBucketNew



```go
func TestTokenBucketNew(t *testing.T)
```
### TestTokenBucketWaitN



```go
func TestTokenBucketWaitN(t *testing.T)
```
### TestZeroRate



```go
func TestZeroRate(t *testing.T)
```
## Types
### Clock

Clock abstracts time operations for testability.


```go
type Clock interface {
	Now() time.Time
	Sleep(time.Duration)
	AfterFunc(time.Duration, func()) Timer
}
```
### LeakyBucket

LeakyBucket implements a leaky bucket rate limiter.
Requests are added to the bucket, and the bucket leaks at a constant rate.
If the bucket is full, requests are denied or must wait.


```go
type LeakyBucket struct {
	// Configuration
	rate     Rate
	capacity int
	cfg      *config

	// State
	mu          sync.Mutex
	level       float64 // Current level in the bucket
	lastLeak    time.Time
	initialized bool
}
```
#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `rate` | `Rate` | Configuration |
| `capacity` | `int` |  |
| `cfg` | `*config` |  |
| `mu` | `sync.Mutex` | State |
| `level` | `float64` | Current level in the bucket |
| `lastLeak` | `time.Time` |  |
| `initialized` | `bool` |  |
#### Methods
##### AllowN

AllowN reports whether n requests can be added to the bucket at time now.
It returns true if the requests were accepted, false otherwise.


```go
func (lb *LeakyBucket) AllowN(now time.Time, n int) bool
```
##### Available

Available returns the number of requests that can be immediately accepted.


```go
func (lb *LeakyBucket) Available() int
```
##### Capacity

Capacity returns the bucket capacity.


```go
func (lb *LeakyBucket) Capacity() int
```
##### Level

Level returns the current level of the bucket.


```go
func (lb *LeakyBucket) Level() float64
```
##### Rate

Rate returns the current leak rate.


```go
func (lb *LeakyBucket) Rate() Rate
```
##### WaitN

WaitN blocks until n requests can be added to the bucket or the context is canceled.


```go
func (lb *LeakyBucket) WaitN(ctx context.Context, n int) error
```
##### leakLocked

leakLocked removes requests from the bucket based on elapsed time.
Must be called with lb.mu held.


```go
func (lb *LeakyBucket) leakLocked(now time.Time)
```
##### waitSlow

waitSlow handles the blocking wait for bucket space.


```go
func (lb *LeakyBucket) waitSlow(ctx context.Context, n int, now time.Time) error
```
### Limiter

Limiter represents a rate limiter that controls the rate at which events are allowed to occur.


```go
type Limiter interface {
	// AllowN reports whether n events may happen at time now.
	// It returns true if the events are allowed, false otherwise.
	// This method never blocks.
	AllowN(now time.Time, n int) bool

	// WaitN blocks until n events can be allowed or the context is canceled.
	// It returns an error if the context is canceled or times out.
	WaitN(ctx context.Context, n int) error
}
```
### Option

Option configures rate limiter behavior.


```go
type Option func(*config)
```
#### Underlying Type

```go
func(*config)
```
### Rate

Rate represents the rate at which tokens are added to the bucket.


```go
type Rate struct {
	TokensPerSec float64
}
```
#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `TokensPerSec` | `float64` |  |
#### Methods
##### String

String returns a string representation of the rate.


```go
func (r Rate) String() string
```
### Timer

Timer represents a timer that can be stopped.


```go
type Timer interface {
	Stop() bool
}
```
### TokenBucket

TokenBucket implements a token bucket rate limiter.
Tokens are added to the bucket at a fixed rate, and requests consume tokens.
If no tokens are available, requests must wait or are denied.


```go
type TokenBucket struct {
	// Configuration
	rate  Rate
	burst int
	cfg   *config

	// State
	mu          sync.Mutex
	tokens      float64
	lastRefill  time.Time
	initialized bool
}
```
#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `rate` | `Rate` | Configuration |
| `burst` | `int` |  |
| `cfg` | `*config` |  |
| `mu` | `sync.Mutex` | State |
| `tokens` | `float64` |  |
| `lastRefill` | `time.Time` |  |
| `initialized` | `bool` |  |
#### Methods
##### AllowN

AllowN reports whether n tokens are available at time now.
It returns true if the tokens were consumed, false otherwise.


```go
func (tb *TokenBucket) AllowN(now time.Time, n int) bool
```
##### Burst

Burst returns the bucket capacity.


```go
func (tb *TokenBucket) Burst() int
```
##### Rate

Rate returns the current token refill rate.


```go
func (tb *TokenBucket) Rate() Rate
```
##### Tokens

Tokens returns the current number of available tokens.


```go
func (tb *TokenBucket) Tokens() float64
```
##### WaitN

WaitN blocks until n tokens are available or the context is canceled.


```go
func (tb *TokenBucket) WaitN(ctx context.Context, n int) error
```
##### refillLocked

refillLocked adds tokens to the bucket based on elapsed time.
Must be called with tb.mu held.


```go
func (tb *TokenBucket) refillLocked(now time.Time)
```
##### waitSlow

waitSlow handles the blocking wait for tokens.


```go
func (tb *TokenBucket) waitSlow(ctx context.Context, n int, now time.Time) error
```
### config



```go
type config struct {
	name   string
	clock  Clock
	jitter float64
	obs    *shared.Observability
}
```
#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | `string` |  |
| `clock` | `Clock` |  |
| `jitter` | `float64` |  |
| `obs` | `*shared.Observability` |  |
### realClock

realClock implements Clock using the real time functions.


```go
type realClock struct{}
```
#### Methods
##### AfterFunc



```go
func (realClock) AfterFunc(d time.Duration, f func()) Timer
```
##### Now



```go
func (realClock) Now() time.Time
```
##### Sleep



```go
func (realClock) Sleep(d time.Duration)
```
### realTimer

realTimer wraps time.Timer to implement our Timer interface.


```go
type realTimer struct{ *time.Timer }
```
#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `` | `*time.Timer` |  |
#### Methods
##### Stop



```go
func (t *realTimer) Stop() bool
```
### testClock

testClock is a controllable clock implementation for testing.


```go
type testClock struct {
	mu     sync.Mutex
	now    time.Time
	timers []*testTimer
}
```
#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `mu` | `sync.Mutex` |  |
| `now` | `time.Time` |  |
| `timers` | `[]*testTimer` |  |
#### Methods
##### Advance

Advance advances the clock by the given duration and fires any timers.


```go
func (c *testClock) Advance(d time.Duration)
```
##### AfterFunc



```go
func (c *testClock) AfterFunc(d time.Duration, f func()) Timer
```
##### Now



```go
func (c *testClock) Now() time.Time
```
##### Set

Set sets the clock to a specific time.


```go
func (c *testClock) Set(t time.Time)
```
##### Sleep



```go
func (c *testClock) Sleep(d time.Duration)
```
### testTimer



```go
type testTimer struct {
	clock    *testClock
	deadline time.Time
	fn       func()
	stopped  bool
	mu       sync.Mutex
}
```
#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `clock` | `*testClock` |  |
| `deadline` | `time.Time` |  |
| `fn` | `func()` |  |
| `stopped` | `bool` |  |
| `mu` | `sync.Mutex` |  |
#### Methods
##### Stop



```go
func (t *testTimer) Stop() bool
```
