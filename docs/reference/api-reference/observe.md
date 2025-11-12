# observe API

Complete API documentation for the observe package.

**Import Path:** `github.com/kolosys/ion/observe`

## Package Documentation

Package observe provides observability interfaces and implementations
for logging, metrics, and tracing across all Ion components.


## Types

### Logger
Logger provides a simple logging interface that components can use without depending on specific logging libraries.

#### Example Usage

```go
// Example implementation of Logger
type MyLogger struct {
    // Add your fields here
}

func (m MyLogger) Debug(param1 string, param2 ...any)  {
    // Implement your logic here
    return
}

func (m MyLogger) Info(param1 string, param2 ...any)  {
    // Implement your logic here
    return
}

func (m MyLogger) Warn(param1 string, param2 ...any)  {
    // Implement your logic here
    return
}

func (m MyLogger) Error(param1 string, param2 error, param3 ...any)  {
    // Implement your logic here
    return
}


```

#### Type Definition

```go
type Logger interface {
    Debug(msg string, kv ...any)
    Info(msg string, kv ...any)
    Warn(msg string, kv ...any)
    Error(msg string, err error, kv ...any)
}
```

## Methods

| Method | Description |
| ------ | ----------- |

### Metrics
Metrics provides a simple metrics interface for recording component behavior without depending on specific metrics libraries.

#### Example Usage

```go
// Example implementation of Metrics
type MyMetrics struct {
    // Add your fields here
}

func (m MyMetrics) Inc(param1 string, param2 ...any)  {
    // Implement your logic here
    return
}

func (m MyMetrics) Add(param1 string, param2 float64, param3 ...any)  {
    // Implement your logic here
    return
}

func (m MyMetrics) Gauge(param1 string, param2 float64, param3 ...any)  {
    // Implement your logic here
    return
}

func (m MyMetrics) Histogram(param1 string, param2 float64, param3 ...any)  {
    // Implement your logic here
    return
}


```

#### Type Definition

```go
type Metrics interface {
    Inc(name string, kv ...any)
    Add(name string, v float64, kv ...any)
    Gauge(name string, v float64, kv ...any)
    Histogram(name string, v float64, kv ...any)
}
```

## Methods

| Method | Description |
| ------ | ----------- |

### NopLogger
NopLogger is a no-operation logger that discards all log messages

#### Example Usage

```go
// Create a new NopLogger
noplogger := NopLogger{

}
```

#### Type Definition

```go
type NopLogger struct {
}
```

## Methods

### Debug



```go
func (NopLogger) Debug(msg string, kv ...any)
```

**Parameters:**
- `msg` (string)
- `kv` (...any)

**Returns:**
  None

### Error



```go
func (NopLogger) Error(msg string, err error, kv ...any)
```

**Parameters:**
- `msg` (string)
- `err` (error)
- `kv` (...any)

**Returns:**
  None

### Info



```go
func (NopLogger) Info(msg string, kv ...any)
```

**Parameters:**
- `msg` (string)
- `kv` (...any)

**Returns:**
  None

### Warn



```go
func (NopLogger) Warn(msg string, kv ...any)
```

**Parameters:**
- `msg` (string)
- `kv` (...any)

**Returns:**
  None

### NopMetrics
NopMetrics is a no-operation metrics recorder that discards all metrics

#### Example Usage

```go
// Create a new NopMetrics
nopmetrics := NopMetrics{

}
```

#### Type Definition

```go
type NopMetrics struct {
}
```

## Methods

### Add



```go
func (NopMetrics) Add(name string, v float64, kv ...any)
```

**Parameters:**
- `name` (string)
- `v` (float64)
- `kv` (...any)

**Returns:**
  None

### Gauge



```go
func (NopMetrics) Gauge(name string, v float64, kv ...any)
```

**Parameters:**
- `name` (string)
- `v` (float64)
- `kv` (...any)

**Returns:**
  None

### Histogram



```go
func (NopMetrics) Histogram(name string, v float64, kv ...any)
```

**Parameters:**
- `name` (string)
- `v` (float64)
- `kv` (...any)

**Returns:**
  None

### Inc



```go
func (NopMetrics) Inc(name string, kv ...any)
```

**Parameters:**
- `name` (string)
- `kv` (...any)

**Returns:**
  None

### NopTracer
NopTracer is a no-operation tracer that creates no spans

#### Example Usage

```go
// Create a new NopTracer
noptracer := NopTracer{

}
```

#### Type Definition

```go
type NopTracer struct {
}
```

## Methods

### Start



```go
func (NopTracer) Start(ctx context.Context, name string, kv ...any) (context.Context, func(err error))
```

**Parameters:**
- `ctx` (context.Context)
- `name` (string)
- `kv` (...any)

**Returns:**
- context.Context
- func(err error)

### Observability
Observability holds observability hooks for a component

#### Example Usage

```go
// Create a new Observability
observability := Observability{
    Logger: Logger{},
    Metrics: Metrics{},
    Tracer: Tracer{},
}
```

#### Type Definition

```go
type Observability struct {
    Logger Logger
    Metrics Metrics
    Tracer Tracer
}
```

### Fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| Logger | `Logger` |  |
| Metrics | `Metrics` |  |
| Tracer | `Tracer` |  |

### Constructor Functions

### New

New creates observability hooks with no-op defaults

```go
func New() *Observability
```

**Parameters:**
  None

**Returns:**
- *Observability

## Methods

### WithLogger

WithLogger sets the logger, returning a new Observability instance

```go
func (*Observability) WithLogger(logger Logger) *Observability
```

**Parameters:**
- `logger` (Logger)

**Returns:**
- *Observability

### WithMetrics

WithMetrics sets the metrics recorder, returning a new Observability instance

```go
func (*Observability) WithMetrics(metrics Metrics) *Observability
```

**Parameters:**
- `metrics` (Metrics)

**Returns:**
- *Observability

### WithTracer

WithTracer sets the tracer, returning a new Observability instance

```go
func (*Observability) WithTracer(tracer Tracer) *Observability
```

**Parameters:**
- `tracer` (Tracer)

**Returns:**
- *Observability

### Tracer
Tracer provides a simple tracing interface for observing component operations without depending on specific tracing libraries.

#### Example Usage

```go
// Example implementation of Tracer
type MyTracer struct {
    // Add your fields here
}

func (m MyTracer) Start(param1 context.Context, param2 string, param3 ...any) context.Context {
    // Implement your logic here
    return
}


```

#### Type Definition

```go
type Tracer interface {
    Start(ctx context.Context, name string, kv ...any) (context.Context, func(err error))
}
```

## Methods

| Method | Description |
| ------ | ----------- |

## External Links

- [Package Overview](../packages/observe.md)
- [pkg.go.dev Documentation](https://pkg.go.dev/github.com/kolosys/ion/observe)
- [Source Code](https://github.com/kolosys/ion/tree/main/observe)
