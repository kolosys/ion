# ratelimit API

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

| Field         | Type         | Description                 |
| ------------- | ------------ | --------------------------- |
| `rate`        | `Rate`       | Configuration               |
| `capacity`    | `int`        |                             |
| `cfg`         | `*config`    |                             |
| `mu`          | `sync.Mutex` | State                       |
| `level`       | `float64`    | Current level in the bucket |
| `lastLeak`    | `time.Time`  |                             |
| `initialized` | `bool`       |                             |

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

### MultiTierConfig

MultiTierConfig holds configuration for multi-tier rate limiting.

```go
type MultiTierConfig struct {
	// Global rate limit configuration
	GlobalRate  Rate
	GlobalBurst int

	// Default rate limits for routes and resources
	DefaultRouteRate     Rate
	DefaultRouteBurst    int
	DefaultResourceRate  Rate
	DefaultResourceBurst int

	// Queue configuration for request management
	QueueSize        int
	EnablePreemptive bool

	// Bucket management
	EnableBucketMapping bool
	BucketTTL           time.Duration

	// Route pattern matching
	RoutePatterns map[string]RouteConfig
}
```

#### Fields

| Field                  | Type                     | Description                                  |
| ---------------------- | ------------------------ | -------------------------------------------- |
| `GlobalRate`           | `Rate`                   | Global rate limit configuration              |
| `GlobalBurst`          | `int`                    |                                              |
| `DefaultRouteRate`     | `Rate`                   | Default rate limits for routes and resources |
| `DefaultRouteBurst`    | `int`                    |                                              |
| `DefaultResourceRate`  | `Rate`                   |                                              |
| `DefaultResourceBurst` | `int`                    |                                              |
| `QueueSize`            | `int`                    | Queue configuration for request management   |
| `EnablePreemptive`     | `bool`                   |                                              |
| `EnableBucketMapping`  | `bool`                   | Bucket management                            |
| `BucketTTL`            | `time.Duration`          |                                              |
| `RoutePatterns`        | `map[string]RouteConfig` | Route pattern matching                       |

### MultiTierLimiter

MultiTierLimiter implements a sophisticated multi-tier rate limiting system.
It supports global, per-route, and per-resource rate limiting with intelligent
bucket management and flexible API compatibility.

```go
type MultiTierLimiter struct {
	mu sync.RWMutex

	// Global limiter shared across all requests
	global Limiter

	// Route limiters for specific API endpoints
	routes sync.Map // map[string]Limiter

	// Resource limiters for specific resources (organizations, projects, etc.)
	resources sync.Map // map[string]Limiter

	// Bucket mapping for API-style rate limit buckets
	bucketMap sync.Map // map[string]string

	// Configuration
	config *MultiTierConfig
	cfg    *config

	// Metrics and observability
	metrics *MultiTierMetrics
}
```

#### Fields

| Field       | Type                | Description                                                              |
| ----------- | ------------------- | ------------------------------------------------------------------------ |
| `mu`        | `sync.RWMutex`      |                                                                          |
| `global`    | `Limiter`           | Global limiter shared across all requests                                |
| `routes`    | `sync.Map`          | Route limiters for specific API endpoints                                |
| `resources` | `sync.Map`          | Resource limiters for specific resources (organizations, projects, etc.) |
| `bucketMap` | `sync.Map`          | Bucket mapping for API-style rate limit buckets                          |
| `config`    | `*MultiTierConfig`  | Configuration                                                            |
| `cfg`       | `*config`           |                                                                          |
| `metrics`   | `*MultiTierMetrics` | Metrics and observability                                                |

#### Methods

##### Allow

Allow checks if a request is allowed without blocking.

```go
func (mtl *MultiTierLimiter) Allow(req *Request) bool
```

##### AllowN

AllowN checks if n requests are allowed without blocking.

```go
func (mtl *MultiTierLimiter) AllowN(req *Request, n int) bool
```

##### GetMetrics

GetMetrics returns current rate limiting metrics.

```go
func (mtl *MultiTierLimiter) GetMetrics() *MultiTierMetrics
```

##### Reset

Reset resets all rate limit buckets (useful for testing).

```go
func (mtl *MultiTierLimiter) Reset()
```

##### UpdateRateLimitFromHeaders

UpdateRateLimitFromHeaders updates rate limit information from API response headers.
This is designed for APIs that provide rate limit information in response headers.

```go
func (mtl *MultiTierLimiter) UpdateRateLimitFromHeaders(req *Request, headers map[string]string) error
```

##### Wait

Wait blocks until the request is allowed or context is canceled.

```go
func (mtl *MultiTierLimiter) Wait(req *Request) error
```

##### WaitN

WaitN blocks until n requests are allowed or context is canceled.

```go
func (mtl *MultiTierLimiter) WaitN(req *Request, n int) error
```

##### findRouteConfig

findRouteConfig finds the configuration for a specific route.

```go
func (mtl *MultiTierLimiter) findRouteConfig(method, endpoint string) RouteConfig
```

##### generateRouteKey

generateRouteKey creates a unique key for route identification.

```go
func (mtl *MultiTierLimiter) generateRouteKey(req *Request) string
```

##### getOrCreateRouteLimiter

getOrCreateRouteLimiter gets or creates a route-specific limiter.

```go
func (mtl *MultiTierLimiter) getOrCreateRouteLimiter(req *Request) Limiter
```

##### getResourceLimiter

getResourceLimiter gets a resource-specific limiter if applicable.

```go
func (mtl *MultiTierLimiter) getResourceLimiter(req *Request) Limiter
```

##### matchesPattern

matchesPattern checks if an endpoint matches a route pattern.

```go
func (mtl *MultiTierLimiter) matchesPattern(endpoint, pattern string) bool
```

##### normalizeRoute

normalizeRoute normalizes an API route for pattern matching.

```go
func (mtl *MultiTierLimiter) normalizeRoute(method, endpoint string) string
```

##### parseFloatHeader

parseFloatHeader parses a float header value.

```go
func (mtl *MultiTierLimiter) parseFloatHeader(headers map[string]string, key string, defaultValue float64) float64
```

##### parseIntHeader

parseIntHeader parses an integer header value.

```go
func (mtl *MultiTierLimiter) parseIntHeader(headers map[string]string, key string, defaultValue int) int
```

##### updateMetrics

updateMetrics safely updates metrics using a function.

```go
func (mtl *MultiTierLimiter) updateMetrics(fn func(*MultiTierMetrics))
```

### MultiTierMetrics

MultiTierMetrics tracks metrics for multi-tier rate limiting.

```go
type MultiTierMetrics struct {
	mu sync.RWMutex

	TotalRequests     int64
	GlobalLimitHits   int64
	RouteLimitHits    int64
	ResourceLimitHits int64
	QueuedRequests    int64
	DroppedRequests   int64
	AvgWaitTime       time.Duration
	MaxWaitTime       time.Duration
	BucketsActive     int64
}
```

#### Fields

| Field               | Type            | Description |
| ------------------- | --------------- | ----------- |
| `mu`                | `sync.RWMutex`  |             |
| `TotalRequests`     | `int64`         |             |
| `GlobalLimitHits`   | `int64`         |             |
| `RouteLimitHits`    | `int64`         |             |
| `ResourceLimitHits` | `int64`         |             |
| `QueuedRequests`    | `int64`         |             |
| `DroppedRequests`   | `int64`         |             |
| `AvgWaitTime`       | `time.Duration` |             |
| `MaxWaitTime`       | `time.Duration` |             |
| `BucketsActive`     | `int64`         |             |

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

| Field          | Type      | Description |
| -------------- | --------- | ----------- |
| `TokensPerSec` | `float64` |             |

#### Methods

##### String

String returns a string representation of the rate.

```go
func (r Rate) String() string
```

### Request

Request represents a request for rate limiting evaluation.

```go
type Request struct {
	// Route information
	Method   string
	Endpoint string

	// Resource identifiers (generic - applications define their own)
	ResourceID    string // Primary resource identifier
	SubResourceID string // Secondary resource identifier
	UserID        string // User/actor identifier

	// Major parameters for bucket identification
	MajorParameters map[string]string

	// Request metadata
	Priority int
	Context  context.Context
}
```

#### Fields

| Field             | Type                | Description                                                    |
| ----------------- | ------------------- | -------------------------------------------------------------- |
| `Method`          | `string`            | Route information                                              |
| `Endpoint`        | `string`            |                                                                |
| `ResourceID`      | `string`            | Resource identifiers (generic - applications define their own) |
| `SubResourceID`   | `string`            | Secondary resource identifier                                  |
| `UserID`          | `string`            | User/actor identifier                                          |
| `MajorParameters` | `map[string]string` | Major parameters for bucket identification                     |
| `Priority`        | `int`               | Request metadata                                               |
| `Context`         | `context.Context`   |                                                                |

### RouteConfig

RouteConfig defines rate limiting for specific route patterns.

```go
type RouteConfig struct {
	Rate  Rate
	Burst int
	// Major parameters that affect rate limiting (e.g., org_id, project_id)
	MajorParameters []string
}
```

#### Fields

| Field             | Type       | Description                                                           |
| ----------------- | ---------- | --------------------------------------------------------------------- |
| `Rate`            | `Rate`     |                                                                       |
| `Burst`           | `int`      |                                                                       |
| `MajorParameters` | `[]string` | Major parameters that affect rate limiting (e.g., org_id, project_id) |

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

| Field         | Type         | Description   |
| ------------- | ------------ | ------------- |
| `rate`        | `Rate`       | Configuration |
| `burst`       | `int`        |               |
| `cfg`         | `*config`    |               |
| `mu`          | `sync.Mutex` | State         |
| `tokens`      | `float64`    |               |
| `lastRefill`  | `time.Time`  |               |
| `initialized` | `bool`       |               |

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

| Field    | Type                    | Description |
| -------- | ----------------------- | ----------- |
| `name`   | `string`                |             |
| `clock`  | `Clock`                 |             |
| `jitter` | `float64`               |             |
| `obs`    | `*shared.Observability` |             |

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

| Field | Type          | Description |
| ----- | ------------- | ----------- |
| ``    | `*time.Timer` |             |

#### Methods

##### Stop

```go
func (t *realTimer) Stop() bool
```
