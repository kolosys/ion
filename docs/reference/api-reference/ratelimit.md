# ratelimit API

Complete API documentation for the ratelimit package.

**Import Path:** `github.com/kolosys/ion/ratelimit`

## Package Documentation

Package ratelimit provides local process rate limiters for controlling function and I/O throughput.
It includes token bucket and leaky bucket implementations with configurable options.


## Types

### Clock
Clock abstracts time operations for testability.

#### Example Usage

```go
// Example implementation of Clock
type MyClock struct {
    // Add your fields here
}

func (m MyClock) Now() time.Time {
    // Implement your logic here
    return
}

func (m MyClock) Sleep(param1 time.Duration)  {
    // Implement your logic here
    return
}

func (m MyClock) AfterFunc(param1 time.Duration, param2 func()) Timer {
    // Implement your logic here
    return
}


```

#### Type Definition

```go
type Clock interface {
    Now() time.Time
    Sleep(time.Duration)
    AfterFunc(time.Duration, func()) Timer
}
```

## Methods

| Method | Description |
| ------ | ----------- |

### LeakyBucket
LeakyBucket implements a leaky bucket rate limiter. Requests are added to the bucket, and the bucket leaks at a constant rate. If the bucket is full, requests are denied or must wait.

#### Example Usage

```go
// Create a new LeakyBucket
leakybucket := LeakyBucket{
    rate: Rate{},
    capacity: 42,
    cfg: &config{}{},
    mu: /* value */,
    level: 3.14,
    lastLeak: /* value */,
    initialized: true,
}
```

#### Type Definition

```go
type LeakyBucket struct {
    rate Rate
    capacity int
    cfg *config
    mu sync.Mutex
    level float64
    lastLeak time.Time
    initialized bool
}
```

### Fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| rate | `Rate` | Configuration |
| capacity | `int` |  |
| cfg | `*config` |  |
| mu | `sync.Mutex` | State |
| level | `float64` | Current level in the bucket |
| lastLeak | `time.Time` |  |
| initialized | `bool` |  |

### Constructor Functions

### NewLeakyBucket

NewLeakyBucket creates a new leaky bucket rate limiter. rate determines how fast the bucket leaks (processes requests). capacity is the maximum number of requests the bucket can hold.

```go
func NewLeakyBucket(rate Rate, capacity int, opts ...Option) *LeakyBucket
```

**Parameters:**
- `rate` (Rate)
- `capacity` (int)
- `opts` (...Option)

**Returns:**
- *LeakyBucket

## Methods

### AllowN

AllowN reports whether n requests can be added to the bucket at time now. It returns true if the requests were accepted, false otherwise.

```go
func (*LeakyBucket) AllowN(now time.Time, n int) bool
```

**Parameters:**
- `now` (time.Time)
- `n` (int)

**Returns:**
- bool

### Available

Available returns the number of requests that can be immediately accepted.

```go
func (*LeakyBucket) Available() int
```

**Parameters:**
  None

**Returns:**
- int

### Capacity

Capacity returns the bucket capacity.

```go
func (*LeakyBucket) Capacity() int
```

**Parameters:**
  None

**Returns:**
- int

### Level

Level returns the current level of the bucket.

```go
func (*LeakyBucket) Level() float64
```

**Parameters:**
  None

**Returns:**
- float64

### Rate

Rate returns the current leak rate.

```go
func (*LeakyBucket) Rate() Rate
```

**Parameters:**
  None

**Returns:**
- Rate

### WaitN

WaitN blocks until n requests can be added to the bucket or the context is canceled.

```go
func (*LeakyBucket) WaitN(ctx context.Context, n int) error
```

**Parameters:**
- `ctx` (context.Context)
- `n` (int)

**Returns:**
- error

### leakLocked

leakLocked removes requests from the bucket based on elapsed time. Must be called with lb.mu held.

```go
func (*LeakyBucket) leakLocked(now time.Time)
```

**Parameters:**
- `now` (time.Time)

**Returns:**
  None

### waitSlow

waitSlow handles the blocking wait for bucket space.

```go
func (*LeakyBucket) waitSlow(ctx context.Context, n int, now time.Time) error
```

**Parameters:**
- `ctx` (context.Context)
- `n` (int)
- `now` (time.Time)

**Returns:**
- error

### Limiter
Limiter represents a rate limiter that controls the rate at which events are allowed to occur.

#### Example Usage

```go
// Example implementation of Limiter
type MyLimiter struct {
    // Add your fields here
}

func (m MyLimiter) AllowN(param1 time.Time, param2 int) bool {
    // Implement your logic here
    return
}

func (m MyLimiter) WaitN(param1 context.Context, param2 int) error {
    // Implement your logic here
    return
}


```

#### Type Definition

```go
type Limiter interface {
    AllowN(now time.Time, n int) bool
    WaitN(ctx context.Context, n int) error
}
```

## Methods

| Method | Description |
| ------ | ----------- |

### MultiTierConfig
MultiTierConfig holds configuration for multi-tier rate limiting.

#### Example Usage

```go
// Create a new MultiTierConfig
multitierconfig := MultiTierConfig{
    GlobalRate: Rate{},
    GlobalBurst: 42,
    DefaultRouteRate: Rate{},
    DefaultRouteBurst: 42,
    DefaultResourceRate: Rate{},
    DefaultResourceBurst: 42,
    QueueSize: 42,
    EnablePreemptive: true,
    EnableBucketMapping: true,
    BucketTTL: /* value */,
    RoutePatterns: map[],
}
```

#### Type Definition

```go
type MultiTierConfig struct {
    GlobalRate Rate
    GlobalBurst int
    DefaultRouteRate Rate
    DefaultRouteBurst int
    DefaultResourceRate Rate
    DefaultResourceBurst int
    QueueSize int
    EnablePreemptive bool
    EnableBucketMapping bool
    BucketTTL time.Duration
    RoutePatterns map[string]RouteConfig
}
```

### Fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| GlobalRate | `Rate` | Global rate limit configuration |
| GlobalBurst | `int` |  |
| DefaultRouteRate | `Rate` | Default rate limits for routes and resources |
| DefaultRouteBurst | `int` |  |
| DefaultResourceRate | `Rate` |  |
| DefaultResourceBurst | `int` |  |
| QueueSize | `int` | Queue configuration for request management |
| EnablePreemptive | `bool` |  |
| EnableBucketMapping | `bool` | Bucket management |
| BucketTTL | `time.Duration` |  |
| RoutePatterns | `map[string]RouteConfig` | Route pattern matching |

### Constructor Functions

### DefaultMultiTierConfig

DefaultMultiTierConfig returns a default configuration for multi-tier rate limiting. Applications should customize this configuration for their specific needs.

```go
func DefaultMultiTierConfig() *MultiTierConfig
```

**Parameters:**
  None

**Returns:**
- *MultiTierConfig

### MultiTierLimiter
MultiTierLimiter implements a sophisticated multi-tier rate limiting system. It supports global, per-route, and per-resource rate limiting with intelligent bucket management and flexible API compatibility.

#### Example Usage

```go
// Create a new MultiTierLimiter
multitierlimiter := MultiTierLimiter{
    mu: /* value */,
    global: Limiter{},
    routes: /* value */,
    resources: /* value */,
    bucketMap: /* value */,
    config: &MultiTierConfig{}{},
    cfg: &config{}{},
    metrics: &MultiTierMetrics{}{},
}
```

#### Type Definition

```go
type MultiTierLimiter struct {
    mu sync.RWMutex
    global Limiter
    routes sync.Map
    resources sync.Map
    bucketMap sync.Map
    config *MultiTierConfig
    cfg *config
    metrics *MultiTierMetrics
}
```

### Fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| mu | `sync.RWMutex` |  |
| global | `Limiter` | Global limiter shared across all requests |
| routes | `sync.Map` | Route limiters for specific API endpoints |
| resources | `sync.Map` | Resource limiters for specific resources (organizations, projects, etc.) |
| bucketMap | `sync.Map` | Bucket mapping for API-style rate limit buckets |
| config | `*MultiTierConfig` | Configuration |
| cfg | `*config` |  |
| metrics | `*MultiTierMetrics` | Metrics and observability |

### Constructor Functions

### NewMultiTierLimiter

NewMultiTierLimiter creates a new multi-tier rate limiter.

```go
func NewMultiTierLimiter(config *MultiTierConfig, opts ...Option) *MultiTierLimiter
```

**Parameters:**
- `config` (*MultiTierConfig)
- `opts` (...Option)

**Returns:**
- *MultiTierLimiter

## Methods

### Allow

Allow checks if a request is allowed without blocking.

```go
func (*MultiTierLimiter) Allow(req *Request) bool
```

**Parameters:**
- `req` (*Request)

**Returns:**
- bool

### AllowN

AllowN checks if n requests are allowed without blocking.

```go
func (*TokenBucket) AllowN(now time.Time, n int) bool
```

**Parameters:**
- `now` (time.Time)
- `n` (int)

**Returns:**
- bool

### GetMetrics

GetMetrics returns current rate limiting metrics.

```go
func (*MultiTierLimiter) GetMetrics() *MultiTierMetrics
```

**Parameters:**
  None

**Returns:**
- *MultiTierMetrics

### Reset

Reset resets all rate limit buckets (useful for testing).

```go
func (*MultiTierLimiter) Reset()
```

**Parameters:**
  None

**Returns:**
  None

### UpdateRateLimitFromHeaders

UpdateRateLimitFromHeaders updates rate limit information from API response headers. This is designed for APIs that provide rate limit information in response headers.

```go
func (*MultiTierLimiter) UpdateRateLimitFromHeaders(req *Request, headers map[string]string) error
```

**Parameters:**
- `req` (*Request)
- `headers` (map[string]string)

**Returns:**
- error

### Wait

Wait blocks until the request is allowed or context is canceled.

```go
func (*MultiTierLimiter) Wait(req *Request) error
```

**Parameters:**
- `req` (*Request)

**Returns:**
- error

### WaitN

WaitN blocks until n requests are allowed or context is canceled.

```go
func (*LeakyBucket) WaitN(ctx context.Context, n int) error
```

**Parameters:**
- `ctx` (context.Context)
- `n` (int)

**Returns:**
- error

### findRouteConfig

findRouteConfig finds the configuration for a specific route.

```go
func (*MultiTierLimiter) findRouteConfig(method, endpoint string) RouteConfig
```

**Parameters:**
- `method` (string)
- `endpoint` (string)

**Returns:**
- RouteConfig

### generateRouteKey

generateRouteKey creates a unique key for route identification.

```go
func (*MultiTierLimiter) generateRouteKey(req *Request) string
```

**Parameters:**
- `req` (*Request)

**Returns:**
- string

### getOrCreateRouteLimiter

getOrCreateRouteLimiter gets or creates a route-specific limiter.

```go
func (*MultiTierLimiter) getOrCreateRouteLimiter(req *Request) Limiter
```

**Parameters:**
- `req` (*Request)

**Returns:**
- Limiter

### getResourceLimiter

getResourceLimiter gets a resource-specific limiter if applicable.

```go
func (*MultiTierLimiter) getResourceLimiter(req *Request) Limiter
```

**Parameters:**
- `req` (*Request)

**Returns:**
- Limiter

### matchesPattern

matchesPattern checks if an endpoint matches a route pattern.

```go
func (*MultiTierLimiter) matchesPattern(endpoint, pattern string) bool
```

**Parameters:**
- `endpoint` (string)
- `pattern` (string)

**Returns:**
- bool

### normalizeRoute

normalizeRoute normalizes an API route for pattern matching.

```go
func (*MultiTierLimiter) normalizeRoute(method, endpoint string) string
```

**Parameters:**
- `method` (string)
- `endpoint` (string)

**Returns:**
- string

### parseFloatHeader

parseFloatHeader parses a float header value.

```go
func (*MultiTierLimiter) parseFloatHeader(headers map[string]string, key string, defaultValue float64) float64
```

**Parameters:**
- `headers` (map[string]string)
- `key` (string)
- `defaultValue` (float64)

**Returns:**
- float64

### parseIntHeader

parseIntHeader parses an integer header value.

```go
func (*MultiTierLimiter) parseIntHeader(headers map[string]string, key string, defaultValue int) int
```

**Parameters:**
- `headers` (map[string]string)
- `key` (string)
- `defaultValue` (int)

**Returns:**
- int

### updateMetrics

updateMetrics safely updates metrics using a function.

```go
func (*MultiTierLimiter) updateMetrics(fn func(*MultiTierMetrics))
```

**Parameters:**
- `fn` (func(*MultiTierMetrics))

**Returns:**
  None

### MultiTierMetrics
MultiTierMetrics tracks metrics for multi-tier rate limiting.

#### Example Usage

```go
// Create a new MultiTierMetrics
multitiermetrics := MultiTierMetrics{
    mu: /* value */,
    TotalRequests: 42,
    GlobalLimitHits: 42,
    RouteLimitHits: 42,
    ResourceLimitHits: 42,
    QueuedRequests: 42,
    DroppedRequests: 42,
    AvgWaitTime: /* value */,
    MaxWaitTime: /* value */,
    BucketsActive: 42,
}
```

#### Type Definition

```go
type MultiTierMetrics struct {
    mu sync.RWMutex
    TotalRequests int64
    GlobalLimitHits int64
    RouteLimitHits int64
    ResourceLimitHits int64
    QueuedRequests int64
    DroppedRequests int64
    AvgWaitTime time.Duration
    MaxWaitTime time.Duration
    BucketsActive int64
}
```

### Fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| mu | `sync.RWMutex` |  |
| TotalRequests | `int64` |  |
| GlobalLimitHits | `int64` |  |
| RouteLimitHits | `int64` |  |
| ResourceLimitHits | `int64` |  |
| QueuedRequests | `int64` |  |
| DroppedRequests | `int64` |  |
| AvgWaitTime | `time.Duration` |  |
| MaxWaitTime | `time.Duration` |  |
| BucketsActive | `int64` |  |

### Option
Option configures rate limiter behavior.

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

### WithClock

WithClock sets a custom clock implementation (useful for testing).

```go
func WithClock(clock Clock) Option
```

**Parameters:**
- `clock` (Clock)

**Returns:**
- Option

### WithJitter

WithJitter sets the jitter factor for WaitN operations (0.0 to 1.0). Jitter helps prevent thundering herd problems by randomizing wait times.

```go
func WithJitter(jitter float64) Option
```

**Parameters:**
- `jitter` (float64)

**Returns:**
- Option

### WithLogger

WithLogger sets the logger for observability.

```go
func WithLogger(logger observe.Logger) Option
```

**Parameters:**
- `logger` (observe.Logger)

**Returns:**
- Option

### WithMetrics

WithMetrics sets the metrics recorder for observability.

```go
func WithMetrics(metrics observe.Metrics) Option
```

**Parameters:**
- `metrics` (observe.Metrics)

**Returns:**
- Option

### WithName

WithName sets the rate limiter name for observability and error reporting.

```go
func WithName(name string) Option
```

**Parameters:**
- `name` (string)

**Returns:**
- Option

### WithTracer

WithTracer sets the tracer for observability.

```go
func WithTracer(tracer observe.Tracer) Option
```

**Parameters:**
- `tracer` (observe.Tracer)

**Returns:**
- Option

### Rate
Rate represents the rate at which tokens are added to the bucket.

#### Example Usage

```go
// Create a new Rate
rate := Rate{
    TokensPerSec: 3.14,
}
```

#### Type Definition

```go
type Rate struct {
    TokensPerSec float64
}
```

### Fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| TokensPerSec | `float64` |  |

### Constructor Functions

### NewRate

NewRate creates a new Rate from the given number of tokens per time duration.

```go
func NewRate(tokens int, duration time.Duration) Rate
```

**Parameters:**
- `tokens` (int)
- `duration` (time.Duration)

**Returns:**
- Rate

### Per

Per is a convenience function for creating rates. For example: Per(100, time.Second) creates a rate of 100 tokens per second.

```go
func Per(tokens int, duration time.Duration) Rate
```

**Parameters:**
- `tokens` (int)
- `duration` (time.Duration)

**Returns:**
- Rate

### PerHour

PerHour creates a rate of the given number of tokens per hour.

```go
func PerHour(tokens int) Rate
```

**Parameters:**
- `tokens` (int)

**Returns:**
- Rate

### PerMinute

PerMinute creates a rate of the given number of tokens per minute.

```go
func PerMinute(tokens int) Rate
```

**Parameters:**
- `tokens` (int)

**Returns:**
- Rate

### PerSecond

PerSecond creates a rate of the given number of tokens per second.

```go
func PerSecond(tokens int) Rate
```

**Parameters:**
- `tokens` (int)

**Returns:**
- Rate

## Methods

### String

String returns a string representation of the rate.

```go
func (Rate) String() string
```

**Parameters:**
  None

**Returns:**
- string

### RateLimitError
RateLimitError represents rate limiting specific errors with context

#### Example Usage

```go
// Create a new RateLimitError
ratelimiterror := RateLimitError{
    Op: "example",
    LimiterName: "example",
    Err: error{},
    RetryAfter: /* value */,
    Global: true,
    Bucket: "example",
    Remaining: 42,
    Limit: 42,
}
```

#### Type Definition

```go
type RateLimitError struct {
    Op string
    LimiterName string
    Err error
    RetryAfter time.Duration
    Global bool
    Bucket string
    Remaining int
    Limit int
}
```

### Fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| Op | `string` | operation that failed |
| LimiterName | `string` | name of the rate limiter |
| Err | `error` | underlying error |
| RetryAfter | `time.Duration` | suggested retry delay |
| Global | `bool` | whether this is a global rate limit |
| Bucket | `string` | rate limit bucket identifier |
| Remaining | `int` | remaining requests in bucket |
| Limit | `int` | total limit for bucket |

## Methods

### Error



```go
func (*RateLimitError) Error() string
```

**Parameters:**
  None

**Returns:**
- string

### IsRetryable

IsRetryable returns true if the rate limit error suggests retrying.

```go
func (*RateLimitError) IsRetryable() bool
```

**Parameters:**
  None

**Returns:**
- bool

### Unwrap



```go
func (*RateLimitError) Unwrap() error
```

**Parameters:**
  None

**Returns:**
- error

### Request
Request represents a request for rate limiting evaluation.

#### Example Usage

```go
// Create a new Request
request := Request{
    Method: "example",
    Endpoint: "example",
    ResourceID: "example",
    SubResourceID: "example",
    UserID: "example",
    MajorParameters: map[],
    Priority: 42,
    Context: /* value */,
}
```

#### Type Definition

```go
type Request struct {
    Method string
    Endpoint string
    ResourceID string
    SubResourceID string
    UserID string
    MajorParameters map[string]string
    Priority int
    Context context.Context
}
```

### Fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| Method | `string` | Route information |
| Endpoint | `string` |  |
| ResourceID | `string` | Resource identifiers (generic - applications define their own) |
| SubResourceID | `string` | Secondary resource identifier |
| UserID | `string` | User/actor identifier |
| MajorParameters | `map[string]string` | Major parameters for bucket identification |
| Priority | `int` | Request metadata |
| Context | `context.Context` |  |

### RouteConfig
RouteConfig defines rate limiting for specific route patterns.

#### Example Usage

```go
// Create a new RouteConfig
routeconfig := RouteConfig{
    Rate: Rate{},
    Burst: 42,
    MajorParameters: [],
}
```

#### Type Definition

```go
type RouteConfig struct {
    Rate Rate
    Burst int
    MajorParameters []string
}
```

### Fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| Rate | `Rate` |  |
| Burst | `int` |  |
| MajorParameters | `[]string` | Major parameters that affect rate limiting (e.g., org_id, project_id) |

### Timer
Timer represents a timer that can be stopped.

#### Example Usage

```go
// Example implementation of Timer
type MyTimer struct {
    // Add your fields here
}

func (m MyTimer) Stop() bool {
    // Implement your logic here
    return
}


```

#### Type Definition

```go
type Timer interface {
    Stop() bool
}
```

## Methods

| Method | Description |
| ------ | ----------- |

### TokenBucket
TokenBucket implements a token bucket rate limiter. Tokens are added to the bucket at a fixed rate, and requests consume tokens. If no tokens are available, requests must wait or are denied.

#### Example Usage

```go
// Create a new TokenBucket
tokenbucket := TokenBucket{
    rate: Rate{},
    burst: 42,
    cfg: &config{}{},
    mu: /* value */,
    tokens: 3.14,
    lastRefill: /* value */,
    initialized: true,
}
```

#### Type Definition

```go
type TokenBucket struct {
    rate Rate
    burst int
    cfg *config
    mu sync.Mutex
    tokens float64
    lastRefill time.Time
    initialized bool
}
```

### Fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| rate | `Rate` | Configuration |
| burst | `int` |  |
| cfg | `*config` |  |
| mu | `sync.Mutex` | State |
| tokens | `float64` |  |
| lastRefill | `time.Time` |  |
| initialized | `bool` |  |

### Constructor Functions

### NewTokenBucket

NewTokenBucket creates a new token bucket rate limiter. rate determines how fast tokens are added to the bucket. burst is the maximum number of tokens the bucket can hold.

```go
func NewTokenBucket(rate Rate, burst int, opts ...Option) *TokenBucket
```

**Parameters:**
- `rate` (Rate)
- `burst` (int)
- `opts` (...Option)

**Returns:**
- *TokenBucket

## Methods

### AllowN

AllowN reports whether n tokens are available at time now. It returns true if the tokens were consumed, false otherwise.

```go
func (*LeakyBucket) AllowN(now time.Time, n int) bool
```

**Parameters:**
- `now` (time.Time)
- `n` (int)

**Returns:**
- bool

### Burst

Burst returns the bucket capacity.

```go
func (*TokenBucket) Burst() int
```

**Parameters:**
  None

**Returns:**
- int

### Rate

Rate returns the current token refill rate.

```go
func (*LeakyBucket) Rate() Rate
```

**Parameters:**
  None

**Returns:**
- Rate

### Tokens

Tokens returns the current number of available tokens.

```go
func (*TokenBucket) Tokens() float64
```

**Parameters:**
  None

**Returns:**
- float64

### WaitN

WaitN blocks until n tokens are available or the context is canceled.

```go
func (*TokenBucket) WaitN(ctx context.Context, n int) error
```

**Parameters:**
- `ctx` (context.Context)
- `n` (int)

**Returns:**
- error

### refillLocked

refillLocked adds tokens to the bucket based on elapsed time. Must be called with tb.mu held.

```go
func (*TokenBucket) refillLocked(now time.Time)
```

**Parameters:**
- `now` (time.Time)

**Returns:**
  None

### waitSlow

waitSlow handles the blocking wait for tokens.

```go
func (*TokenBucket) waitSlow(ctx context.Context, n int, now time.Time) error
```

**Parameters:**
- `ctx` (context.Context)
- `n` (int)
- `now` (time.Time)

**Returns:**
- error

## Functions

### NewBucketLimitError
NewBucketLimitError creates an error for bucket-specific rate limits

```go
func NewBucketLimitError(limiterName, bucket string, remaining, limit int, retryAfter time.Duration) error
```

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `limiterName` | `string` | |
| `bucket` | `string` | |
| `remaining` | `int` | |
| `limit` | `int` | |
| `retryAfter` | `time.Duration` | |

**Returns:**
| Type | Description |
|------|-------------|
| `error` | |

**Example:**

```go
// Example usage of NewBucketLimitError
result := NewBucketLimitError(/* parameters */)
```

### NewGlobalRateLimitError
NewGlobalRateLimitError creates an error for global rate limit hits

```go
func NewGlobalRateLimitError(limiterName string, retryAfter time.Duration) error
```

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `limiterName` | `string` | |
| `retryAfter` | `time.Duration` | |

**Returns:**
| Type | Description |
|------|-------------|
| `error` | |

**Example:**

```go
// Example usage of NewGlobalRateLimitError
result := NewGlobalRateLimitError(/* parameters */)
```

### NewRateLimitExceededError
NewRateLimitExceededError creates an error indicating rate limit was exceeded

```go
func NewRateLimitExceededError(limiterName string, retryAfter time.Duration) error
```

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `limiterName` | `string` | |
| `retryAfter` | `time.Duration` | |

**Returns:**
| Type | Description |
|------|-------------|
| `error` | |

**Example:**

```go
// Example usage of NewRateLimitExceededError
result := NewRateLimitExceededError(/* parameters */)
```

## External Links

- [Package Overview](../packages/ratelimit.md)
- [pkg.go.dev Documentation](https://pkg.go.dev/github.com/kolosys/ion/ratelimit)
- [Source Code](https://github.com/kolosys/ion/tree/main/ratelimit)
