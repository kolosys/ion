# semaphore API

Complete API documentation for the semaphore package.

**Import Path:** `github.com/kolosys/ion/semaphore`

## Package Documentation

Package semaphore provides a weighted semaphore with configurable fairness modes.


## Variables

### ErrInvalidWeight

Common sentinel errors for semaphore operations


```go
&{0xc00045a048 [ErrInvalidWeight] <nil> [0xc00049a6c0] <nil>}
```

## Types

### Fairness
Fairness defines the ordering behavior for semaphore waiters

#### Example Usage

```go
// Example usage of Fairness
var value Fairness
// Initialize with appropriate value
```

#### Type Definition

```go
type Fairness int
```

## Methods

### String

String returns the string representation of the fairness mode

```go
func (Fairness) String() string
```

**Parameters:**
  None

**Returns:**
- string

### Option
Option configures semaphore behavior

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

### WithAcquireTimeout

WithAcquireTimeout sets the default timeout for Acquire operations

```go
func WithAcquireTimeout(timeout time.Duration) Option
```

**Parameters:**
- `timeout` (time.Duration)

**Returns:**
- Option

### WithFairness

WithFairness sets the fairness mode for waiter ordering

```go
func WithFairness(fairness Fairness) Option
```

**Parameters:**
- `fairness` (Fairness)

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

WithName sets the semaphore name for observability and error reporting

```go
func WithName(name string) Option
```

**Parameters:**
- `name` (string)

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

### Semaphore
Semaphore represents a weighted semaphore that controls access to a resource with a fixed capacity. It supports configurable fairness modes and observability.

#### Example Usage

```go
// Example implementation of Semaphore
type MySemaphore struct {
    // Add your fields here
}

func (m MySemaphore) Acquire(param1 context.Context, param2 int64) error {
    // Implement your logic here
    return
}

func (m MySemaphore) TryAcquire(param1 int64) bool {
    // Implement your logic here
    return
}

func (m MySemaphore) Release(param1 int64)  {
    // Implement your logic here
    return
}

func (m MySemaphore) Current() int64 {
    // Implement your logic here
    return
}


```

#### Type Definition

```go
type Semaphore interface {
    Acquire(ctx context.Context, n int64) error
    TryAcquire(n int64) bool
    Release(n int64)
    Current() int64
}
```

## Methods

| Method | Description |
| ------ | ----------- |

### Constructor Functions

### NewWeighted

NewWeighted creates a new weighted semaphore with the specified capacity. The semaphore starts with all permits available.

```go
func NewWeighted(capacity int64, opts ...Option) Semaphore
```

**Parameters:**
- `capacity` (int64)
- `opts` (...Option)

**Returns:**
- Semaphore

### SemaphoreError
SemaphoreError represents semaphore-specific errors with context

#### Example Usage

```go
// Create a new SemaphoreError
semaphoreerror := SemaphoreError{
    Op: "example",
    Name: "example",
    Err: error{},
}
```

#### Type Definition

```go
type SemaphoreError struct {
    Op string
    Name string
    Err error
}
```

### Fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| Op | `string` | operation that failed |
| Name | `string` | name of the semaphore |
| Err | `error` | underlying error |

## Methods

### Error



```go
func (*SemaphoreError) Error() string
```

**Parameters:**
  None

**Returns:**
- string

### Unwrap



```go
func (*SemaphoreError) Unwrap() error
```

**Parameters:**
  None

**Returns:**
- error

## Functions

### NewAcquireTimeoutError
NewAcquireTimeoutError creates an error indicating an acquire operation timed out

```go
func NewAcquireTimeoutError(semaphoreName string) error
```

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `semaphoreName` | `string` | |

**Returns:**
| Type | Description |
|------|-------------|
| `error` | |

**Example:**

```go
// Example usage of NewAcquireTimeoutError
result := NewAcquireTimeoutError(/* parameters */)
```

### NewWeightExceedsCapacityError
NewWeightExceedsCapacityError creates an error indicating the requested weight exceeds capacity

```go
func NewWeightExceedsCapacityError(semaphoreName string, weight, capacity int64) error
```

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `semaphoreName` | `string` | |
| `weight` | `int64` | |
| `capacity` | `int64` | |

**Returns:**
| Type | Description |
|------|-------------|
| `error` | |

**Example:**

```go
// Example usage of NewWeightExceedsCapacityError
result := NewWeightExceedsCapacityError(/* parameters */)
```

## External Links

- [Package Overview](../packages/semaphore.md)
- [pkg.go.dev Documentation](https://pkg.go.dev/github.com/kolosys/ion/semaphore)
- [Source Code](https://github.com/kolosys/ion/tree/main/semaphore)
