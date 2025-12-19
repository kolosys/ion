# Observability

**Import Path:** `github.com/kolosys/ion/observe`

The observability package provides pluggable interfaces for logging, metrics, and tracing across all Ion components. It enables integration with any observability stack while maintaining zero overhead when not configured.

## Overview

Observability is built into every Ion component. By default, all components use no-op implementations that have zero overhead. You can plug in your own logging, metrics, and tracing implementations to integrate with your existing observability infrastructure.

### When to Use Observability

- **Production Monitoring**: Integrate with Prometheus, Datadog, or other metrics systems
- **Structured Logging**: Use `slog`, `zap`, or other logging libraries
- **Distributed Tracing**: Integrate with OpenTelemetry, Jaeger, or other tracing systems
- **Debugging**: Add detailed logging during development and troubleshooting

## Architecture

The observability package provides three interfaces:

```
┌─────────────────────────────────────┐
│      Observability                  │
├─────────────────────────────────────┤
│  Logger  │  Metrics  │  Tracer      │
└─────────────────────────────────────┘
      │         │          │
      │         │          │
      [Your Implementation ]
```

### Components

1. **Logger**: Structured logging interface
2. **Metrics**: Metrics recording interface
3. **Tracer**: Distributed tracing interface

## Core Concepts

### Zero Overhead Defaults

By default, all Ion components use no-op implementations:

```go
obs := observe.New() // No-op logger, metrics, and tracer

// These calls do nothing and have zero overhead
obs.Logger.Info("message")
obs.Metrics.Inc("counter")
obs.Tracer.Start(ctx, "operation")
```

### Pluggable Implementations

Replace any component with your own implementation:

```go
obs := observe.New().
    WithLogger(myLogger).
    WithMetrics(myMetrics).
    WithTracer(myTracer)
```

### Immutable Updates

The `With*` methods return new instances, making the API safe for concurrent use:

```go
baseObs := observe.New()
customObs := baseObs.WithLogger(myLogger) // baseObs unchanged
```

## Real-World Scenarios

### Scenario 1: Integration with `slog`

Use Go's standard structured logging:

```go
package main

import (
    "context"
    "log/slog"
    "os"

    "github.com/kolosys/ion/circuit"
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
    l.logger.Error(msg, append([]any{"error", err}, kv...)...)
}

func main() {
    // Create slog logger
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    // Create observability with slog
    obs := observe.New().
        WithLogger(SlogLogger{logger: logger})

    // Use with Ion components
    cb := circuit.New("payment-service",
        circuit.WithObservability(obs),
    )

    // All circuit breaker logs will go through slog
    ctx := context.Background()
    cb.Execute(ctx, func(ctx context.Context) (any, error) {
        return "success", nil
    })
}
```

### Scenario 2: Prometheus Metrics

Integrate with Prometheus for metrics collection:

```go
package main

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/kolosys/ion/circuit"
    "github.com/kolosys/ion/observe"
)

type PrometheusMetrics struct {
    counters   map[string]*prometheus.CounterVec
    histograms map[string]*prometheus.HistogramVec
    gauges     map[string]*prometheus.GaugeVec
    mu         sync.RWMutex
}

func NewPrometheusMetrics() *PrometheusMetrics {
    return &PrometheusMetrics{
        counters:   make(map[string]*prometheus.CounterVec),
        histograms: make(map[string]*prometheus.HistogramVec),
        gauges:     make(map[string]*prometheus.GaugeVec),
    }
}

func (m *PrometheusMetrics) Inc(name string, kv ...any) {
    labels := extractLabels(kv...)
    counter := m.getCounter(name, labels)
    counter.Inc()
}

func (m *PrometheusMetrics) Add(name string, v float64, kv ...any) {
    labels := extractLabels(kv...)
    counter := m.getCounter(name, labels)
    counter.Add(v)
}

func (m *PrometheusMetrics) Gauge(name string, v float64, kv ...any) {
    labels := extractLabels(kv...)
    gauge := m.getGauge(name, labels)
    gauge.Set(v)
}

func (m *PrometheusMetrics) Histogram(name string, v float64, kv ...any) {
    labels := extractLabels(kv...)
    histogram := m.getHistogram(name, labels)
    histogram.Observe(v)
}

func extractLabels(kv ...any) prometheus.Labels {
    labels := make(prometheus.Labels)
    for i := 0; i < len(kv)-1; i += 2 {
        if key, ok := kv[i].(string); ok {
            labels[key] = fmt.Sprintf("%v", kv[i+1])
        }
    }
    return labels
}

func main() {
    metrics := NewPrometheusMetrics()
    obs := observe.New().WithMetrics(metrics)

    cb := circuit.New("payment-service",
        circuit.WithObservability(obs),
    )

    // Metrics are automatically collected and exposed via Prometheus
}
```

### Scenario 3: OpenTelemetry Tracing

Integrate with OpenTelemetry for distributed tracing:

```go
package main

import (
    "context"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
    "github.com/kolosys/ion/circuit"
    "github.com/kolosys/ion/observe"
)

type OpenTelemetryTracer struct {
    tracer trace.Tracer
}

func NewOpenTelemetryTracer() *OpenTelemetryTracer {
    return &OpenTelemetryTracer{
        tracer: otel.Tracer("ion"),
    }
}

func (t *OpenTelemetryTracer) Start(ctx context.Context, name string, kv ...any) (context.Context, func(error)) {
    attrs := extractAttributes(kv...)
    ctx, span := t.tracer.Start(ctx, name, trace.WithAttributes(attrs...))
    return ctx, func(err error) {
        if err != nil {
            span.RecordError(err)
        }
        span.End()
    }
}

func extractAttributes(kv ...any) []attribute.KeyValue {
    attrs := make([]attribute.KeyValue, 0, len(kv)/2)
    for i := 0; i < len(kv)-1; i += 2 {
        if key, ok := kv[i].(string); ok {
            attrs = append(attrs, attribute.String(key, fmt.Sprintf("%v", kv[i+1])))
        }
    }
    return attrs
}

func main() {
    tracer := NewOpenTelemetryTracer()
    obs := observe.New().WithTracer(tracer)

    cb := circuit.New("payment-service",
        circuit.WithObservability(obs),
    )

    // All circuit breaker operations are traced
}
```

### Scenario 4: Custom Logger with Log Levels

Create a custom logger that respects log levels:

```go
package main

import (
    "io"
    "log"
    "github.com/kolosys/ion/observe"
)

type LevelLogger struct {
    debug *log.Logger
    info  *log.Logger
    warn  *log.Logger
    error *log.Logger
}

func NewLevelLogger(w io.Writer, prefix string) *LevelLogger {
    flags := log.LstdFlags | log.Lshortfile
    return &LevelLogger{
        debug: log.New(w, "[DEBUG] "+prefix, flags),
        info:  log.New(w, "[INFO] "+prefix, flags),
        warn:  log.New(w, "[WARN] "+prefix, flags),
        error: log.New(w, "[ERROR] "+prefix, flags),
    }
}

func (l *LevelLogger) Debug(msg string, kv ...any) {
    l.debug.Printf("%s %v", msg, kv)
}

func (l *LevelLogger) Info(msg string, kv ...any) {
    l.info.Printf("%s %v", msg, kv)
}

func (l *LevelLogger) Warn(msg string, kv ...any) {
    l.warn.Printf("%s %v", msg, kv)
}

func (l *LevelLogger) Error(msg string, err error, kv ...any) {
    l.error.Printf("%s: %v %v", msg, err, kv)
}

func main() {
    logger := NewLevelLogger(os.Stdout, "ion: ")
    obs := observe.New().WithLogger(logger)

    // Use with any Ion component
    cb := circuit.New("service", circuit.WithObservability(obs))
}
```

### Scenario 5: Multi-Component Observability

Share observability across multiple components:

```go
package main

import (
    "github.com/kolosys/ion/circuit"
    "github.com/kolosys/ion/ratelimit"
    "github.com/kolosys/ion/semaphore"
    "github.com/kolosys/ion/workerpool"
    "github.com/kolosys/ion/observe"
)

func setupObservability() *observe.Observability {
    // Create shared observability configuration
    return observe.New().
        WithLogger(myLogger).
        WithMetrics(myMetrics).
        WithTracer(myTracer)
}

func main() {
    obs := setupObservability()

    // All components share the same observability configuration
    cb := circuit.New("payment-service",
        circuit.WithObservability(obs),
    )

    limiter := ratelimit.NewTokenBucket(ratelimit.PerSecond(10), 20,
        ratelimit.WithLogger(obs.Logger),
        ratelimit.WithMetrics(obs.Metrics),
    )

    sem := semaphore.NewWeighted(10,
        semaphore.WithLogger(obs.Logger),
        semaphore.WithMetrics(obs.Metrics),
    )

    pool := workerpool.New(4, 20,
        workerpool.WithLogger(obs.Logger),
        workerpool.WithMetrics(obs.Metrics),
    )
}
```

## Interface Details

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
    Add(name string, v float64, kv ...any)        // Add to counter
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

## Best Practices

1. **Use Structured Logging**: Pass key-value pairs for better log analysis
2. **Include Context**: Always include component names and operation context
3. **Respect Log Levels**: Use appropriate log levels (Debug, Info, Warn, Error)
4. **Metric Naming**: Use consistent metric naming conventions
5. **Trace Propagation**: Ensure traces propagate through context
6. **Zero Overhead**: Use no-op implementations when observability isn't needed

## Common Pitfalls

### Pitfall 1: Not Using Observability

**Problem**: Missing visibility into component behavior

**Solution**: Always configure observability in production

```go
// Bad
cb := circuit.New("service")

// Good
obs := observe.New().WithLogger(myLogger).WithMetrics(myMetrics)
cb := circuit.New("service", circuit.WithObservability(obs))
```

### Pitfall 2: Inconsistent Metric Names

**Problem**: Metrics scattered across different naming conventions

**Solution**: Use consistent naming patterns

```go
// Good: Consistent prefix and naming
metrics.Inc("ion_circuit_requests_total", "name", "payment-service")
metrics.Inc("ion_circuit_requests_failed", "name", "payment-service")
```

### Pitfall 3: Not Propagating Context

**Problem**: Traces don't connect across service boundaries

**Solution**: Always pass context through operations

```go
// Good: Context propagates through all operations
result, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
    return service.Call(ctx, req)
})
```

## Integration Examples

### Zap Logger

```go
import (
    "go.uber.org/zap"
    "github.com/kolosys/ion/observe"
)

type ZapLogger struct {
    logger *zap.Logger
}

func (l ZapLogger) Debug(msg string, kv ...any) {
    l.logger.Debug(msg, convertKV(kv...)...)
}

func (l ZapLogger) Info(msg string, kv ...any) {
    l.logger.Info(msg, convertKV(kv...)...)
}

func (l ZapLogger) Warn(msg string, kv ...any) {
    l.logger.Warn(msg, convertKV(kv...)...)
}

func (l ZapLogger) Error(msg string, err error, kv ...any) {
    fields := append(convertKV(kv...), zap.Error(err))
    l.logger.Error(msg, fields...)
}

func convertKV(kv ...any) []zap.Field {
    fields := make([]zap.Field, 0, len(kv)/2)
    for i := 0; i < len(kv)-1; i += 2 {
        if key, ok := kv[i].(string); ok {
            fields = append(fields, zap.Any(key, kv[i+1]))
        }
    }
    return fields
}
```

### Datadog Metrics

```go
import (
    "gopkg.in/DataDog/dd-trace-go.v1/statsd"
    "github.com/kolosys/ion/observe"
)

type DatadogMetrics struct {
    client *statsd.Client
}

func (m *DatadogMetrics) Inc(name string, kv ...any) {
    tags := extractTags(kv...)
    m.client.Incr(name, tags, 1)
}

func (m *DatadogMetrics) Histogram(name string, v float64, kv ...any) {
    tags := extractTags(kv...)
    m.client.Histogram(name, v, tags, 1)
}
```

## Further Reading

- [API Reference](../api-reference/observe.md) - Complete API documentation
- [Examples](../examples/) - Practical examples
- [Best Practices](../advanced/best-practices.md) - Recommended patterns
