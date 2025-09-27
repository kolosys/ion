# Observe

[![Go Reference](https://pkg.go.dev/badge/github.com/kolosys/ion/observe.svg)](https://pkg.go.dev/github.com/kolosys/ion/observe)

Pluggable observability interfaces for logging, metrics, and tracing across all Ion components.

## Features

- **Pluggable Interfaces**: Simple interfaces that work with any observability stack
- **No-Op Defaults**: Zero-overhead defaults when observability is not configured
- **Zero Dependencies**: No external dependencies beyond the Go standard library
- **Type Safety**: Strongly typed interfaces for compile-time safety

## Interfaces

### Logger Interface

```go
type Logger interface {
    Debug(msg string, kv ...any)
    Info(msg string, kv ...any)
    Warn(msg string, kv ...any)
    Error(msg string, err error, kv ...any)
}
```

### Metrics Interface

```go
type Metrics interface {
    Inc(name string, kv ...any)                  // Increment counter
    Add(name string, v float64, kv ...any)       // Add to counter
    Gauge(name string, v float64, kv ...any)     // Set gauge value
    Histogram(name string, v float64, kv ...any) // Record histogram value
}
```

### Tracer Interface

```go
type Tracer interface {
    Start(ctx context.Context, name string, kv ...any) (context.Context, func(err error))
}
```

## Usage

### Basic Configuration

```go
import "github.com/kolosys/ion/observe"

// Create observability with defaults (no-op implementations)
obs := observe.New()

// Use with any Ion component
pool := workerpool.New(4, 20, workerpool.WithLogger(obs.Logger))
```

### Custom Implementations

#### Structured Logging (slog)

```go
import (
    "log/slog"
    "github.com/kolosys/ion/observe"
)

type SlogLogger struct {
    logger *slog.Logger
}

func (l SlogLogger) Debug(msg string, kv ...any) {
    l.logger.Debug(msg, kv...)
}

func (l SlogLogger) Info(msg string, kv ...any) {
    l.logger.Info(msg, kv...)
}

func (l SlogLogger) Warn(msg string, kv ...any) {
    l.logger.Warn(msg, kv...)
}

func (l SlogLogger) Error(msg string, err error, kv ...any) {
    args := append([]any{"error", err}, kv...)
    l.logger.Error(msg, args...)
}

// Usage
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
pool := workerpool.New(4, 20, workerpool.WithLogger(SlogLogger{logger}))
```

#### Prometheus Metrics

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/kolosys/ion/observe"
)

type PromMetrics struct {
    counters   map[string]*prometheus.CounterVec
    gauges     map[string]*prometheus.GaugeVec
    histograms map[string]*prometheus.HistogramVec
}

func NewPromMetrics(registry prometheus.Registerer) *PromMetrics {
    return &PromMetrics{
        counters:   make(map[string]*prometheus.CounterVec),
        gauges:     make(map[string]*prometheus.GaugeVec),
        histograms: make(map[string]*prometheus.HistogramVec),
    }
}

func (m *PromMetrics) Inc(name string, kv ...any) {
    counter, exists := m.counters[name]
    if !exists {
        counter = prometheus.NewCounterVec(
            prometheus.CounterOpts{Name: name},
            labelsFromKV(kv),
        )
        m.counters[name] = counter
    }

    counter.With(kvToPrometheusLabels(kv)).Inc()
}

// Similar implementations for Gauge and Histogram...

// Usage
metrics := NewPromMetrics(prometheus.DefaultRegisterer)
sem := semaphore.NewWeighted(10, semaphore.WithMetrics(metrics))
```

#### OpenTelemetry Tracing

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
    "github.com/kolosys/ion/observe"
)

type OTelTracer struct {
    tracer trace.Tracer
}

func NewOTelTracer(name string) *OTelTracer {
    return &OTelTracer{
        tracer: otel.Tracer(name),
    }
}

func (t *OTelTracer) Start(ctx context.Context, name string, kv ...any) (context.Context, func(err error)) {
    ctx, span := t.tracer.Start(ctx, name)

    // Add attributes
    for i := 0; i < len(kv); i += 2 {
        if i+1 < len(kv) {
            key := fmt.Sprint(kv[i])
            value := fmt.Sprint(kv[i+1])
            span.SetAttributes(attribute.String(key, value))
        }
    }

    return ctx, func(err error) {
        if err != nil {
            span.RecordError(err)
            span.SetStatus(codes.Error, err.Error())
        }
        span.End()
    }
}

// Usage
tracer := NewOTelTracer("ion-circuit")
cb := circuit.New("payment", circuit.WithTracer(tracer))
```

### Complete Configuration

```go
// Build complete observability configuration
obs := observe.New().
    WithLogger(myLogger).
    WithMetrics(myMetrics).
    WithTracer(myTracer)

// Use with Ion components
pool := workerpool.New(4, 20,
    workerpool.WithLogger(obs.Logger),
    workerpool.WithMetrics(obs.Metrics),
    workerpool.WithTracer(obs.Tracer),
)
```

## Default Implementations

All interfaces have no-op implementations that discard output:

- `observe.NopLogger{}` - Discards all log messages
- `observe.NopMetrics{}` - Discards all metrics
- `observe.NopTracer{}` - Creates no spans

These allow Ion components to work without requiring observability setup.

## Integration Examples

### Complete Observability Stack

```go
package main

import (
    "log/slog"
    "os"

    "github.com/prometheus/client_golang/prometheus"
    "go.opentelemetry.io/otel"

    "github.com/kolosys/ion/observe"
    "github.com/kolosys/ion/workerpool"
)

func main() {
    // Setup logging
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelDebug,
    }))

    // Setup metrics
    registry := prometheus.NewRegistry()
    metrics := NewPromMetrics(registry)

    // Setup tracing
    tracer := otel.Tracer("ion-example")

    // Create observability
    obs := observe.New().
        WithLogger(SlogLogger{logger}).
        WithMetrics(metrics).
        WithTracer(OTelTracer{tracer})

    // Use with Ion components
    pool := workerpool.New(4, 20,
        workerpool.WithName("main-pool"),
        workerpool.WithLogger(obs.Logger),
        workerpool.WithMetrics(obs.Metrics),
        workerpool.WithTracer(obs.Tracer),
    )
    defer pool.Close(context.Background())

    // Your application logic...
}
```

## Best Practices

1. **Use Structured Logging**: Pass key-value pairs for better observability
2. **Consistent Naming**: Use consistent metric and span names across components
3. **Error Context**: Always include relevant context in error messages
4. **Performance**: No-op implementations have zero overhead when not used

## Contributing

See the main [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

## License

Licensed under the [MIT License](../LICENSE).
