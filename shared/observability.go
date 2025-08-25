package shared

import "context"

// Logger provides a simple logging interface that components can use
// without depending on specific logging libraries.
type Logger interface {
	Debug(msg string, kv ...any)
	Info(msg string, kv ...any)
	Warn(msg string, kv ...any)
	Error(msg string, err error, kv ...any)
}

// Metrics provides a simple metrics interface for recording component behavior
// without depending on specific metrics libraries.
type Metrics interface {
	Inc(name string, kv ...any)
	Add(name string, v float64, kv ...any)
	Gauge(name string, v float64, kv ...any)
	Histogram(name string, v float64, kv ...any)
}

// Tracer provides a simple tracing interface for observing component operations
// without depending on specific tracing libraries.
type Tracer interface {
	Start(ctx context.Context, name string, kv ...any) (context.Context, func(err error))
}

// NopLogger is a no-operation logger that discards all log messages
type NopLogger struct{}

func (NopLogger) Debug(msg string, kv ...any) {}
func (NopLogger) Info(msg string, kv ...any)  {}
func (NopLogger) Warn(msg string, kv ...any)  {}
func (NopLogger) Error(msg string, err error, kv ...any) {}

// NopMetrics is a no-operation metrics recorder that discards all metrics
type NopMetrics struct{}

func (NopMetrics) Inc(name string, kv ...any)              {}
func (NopMetrics) Add(name string, v float64, kv ...any)   {}
func (NopMetrics) Gauge(name string, v float64, kv ...any) {}
func (NopMetrics) Histogram(name string, v float64, kv ...any) {}

// NopTracer is a no-operation tracer that creates no spans
type NopTracer struct{}

func (NopTracer) Start(ctx context.Context, name string, kv ...any) (context.Context, func(err error)) {
	return ctx, func(err error) {}
}

// Observability holds observability hooks for a component
type Observability struct {
	Logger  Logger
	Metrics Metrics
	Tracer  Tracer
}

// NewObservability creates observability hooks with no-op defaults
func NewObservability() *Observability {
	return &Observability{
		Logger:  NopLogger{},
		Metrics: NopMetrics{},
		Tracer:  NopTracer{},
	}
}

// WithLogger sets the logger, returning a new Observability instance
func (o *Observability) WithLogger(logger Logger) *Observability {
	return &Observability{
		Logger:  logger,
		Metrics: o.Metrics,
		Tracer:  o.Tracer,
	}
}

// WithMetrics sets the metrics recorder, returning a new Observability instance
func (o *Observability) WithMetrics(metrics Metrics) *Observability {
	return &Observability{
		Logger:  o.Logger,
		Metrics: metrics,
		Tracer:  o.Tracer,
	}
}

// WithTracer sets the tracer, returning a new Observability instance
func (o *Observability) WithTracer(tracer Tracer) *Observability {
	return &Observability{
		Logger:  o.Logger,
		Metrics: o.Metrics,
		Tracer:  tracer,
	}
}
