// Package workerpool provides a bounded worker pool with context-aware submission,
// graceful shutdown, and observability hooks.
package workerpool

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kolosys/ion/shared"
)

// Task represents a unit of work to be executed by the worker pool.
// Tasks receive a context that will be canceled if either the submission
// context or the pool's base context is canceled.
type Task func(ctx context.Context) error

// Pool represents a bounded worker pool that executes tasks with controlled
// concurrency and queue management.
type Pool struct {
	// Configuration
	name         string
	size         int
	queueSize    int
	drainTimeout time.Duration

	// Observability
	obs *shared.Observability

	// Lifecycle management
	baseCtx   context.Context
	cancel    context.CancelFunc
	closed    chan struct{}
	draining  atomic.Bool
	closeOnce sync.Once
	drainOnce sync.Once

	// Task management
	taskCh   chan taskSubmission
	workerWg sync.WaitGroup

	// Metrics
	metrics PoolMetrics

	// Panic recovery
	panicHandler func(any)
	taskWrapper  func(Task) Task
}

// taskSubmission wraps a task with its submission context
type taskSubmission struct {
	task Task
	ctx  context.Context
}

// PoolMetrics holds runtime metrics for the pool
type PoolMetrics struct {
	Size      int    // configured pool size
	Queued    int64  // current queue length
	Running   int64  // currently running tasks
	Completed uint64 // total completed tasks
	Failed    uint64 // total failed tasks
	Panicked  uint64 // total panicked tasks
}

// Option configures pool behavior
type Option func(*config)

type config struct {
	name         string
	baseCtx      context.Context
	drainTimeout time.Duration
	obs          *shared.Observability
	panicHandler func(any)
	taskWrapper  func(Task) Task
}

// WithName sets the pool name for observability and error reporting
func WithName(name string) Option {
	return func(c *config) {
		c.name = name
	}
}

// WithBaseContext sets the base context for the pool.
// All task contexts will be derived from this context.
func WithBaseContext(ctx context.Context) Option {
	return func(c *config) {
		c.baseCtx = ctx
	}
}

// WithDrainTimeout sets the default timeout for Drain operations
func WithDrainTimeout(timeout time.Duration) Option {
	return func(c *config) {
		c.drainTimeout = timeout
	}
}

// WithLogger sets the logger for observability
func WithLogger(logger shared.Logger) Option {
	return func(c *config) {
		c.obs = c.obs.WithLogger(logger)
	}
}

// WithMetrics sets the metrics recorder for observability
func WithMetrics(metrics shared.Metrics) Option {
	return func(c *config) {
		c.obs = c.obs.WithMetrics(metrics)
	}
}

// WithTracer sets the tracer for observability
func WithTracer(tracer shared.Tracer) Option {
	return func(c *config) {
		c.obs = c.obs.WithTracer(tracer)
	}
}

// WithPanicRecovery sets a custom panic handler for task execution.
// If not set, panics are recovered and counted in metrics.
func WithPanicRecovery(handler func(any)) Option {
	return func(c *config) {
		c.panicHandler = handler
	}
}

// WithTaskWrapper sets a function to wrap tasks for instrumentation.
// The wrapper is applied to every submitted task.
func WithTaskWrapper(wrapper func(Task) Task) Option {
	return func(c *config) {
		c.taskWrapper = wrapper
	}
}

// New creates a new worker pool with the specified size and queue capacity.
// size determines the number of worker goroutines.
// queueSize determines the maximum number of queued tasks.
func New(size, queueSize int, opts ...Option) *Pool {
	if size <= 0 {
		size = runtime.GOMAXPROCS(0)
	}
	if queueSize < 0 {
		queueSize = 0
	}

	cfg := &config{
		name:         "",
		baseCtx:      context.Background(),
		drainTimeout: 30 * time.Second,
		obs:          shared.NewObservability(),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	ctx, cancel := context.WithCancel(cfg.baseCtx)

	p := &Pool{
		name:         cfg.name,
		size:         size,
		queueSize:    queueSize,
		drainTimeout: cfg.drainTimeout,
		obs:          cfg.obs,
		baseCtx:      ctx,
		cancel:       cancel,
		closed:       make(chan struct{}),
		taskCh:       make(chan taskSubmission, queueSize),
		panicHandler: cfg.panicHandler,
		taskWrapper:  cfg.taskWrapper,
		metrics: PoolMetrics{
			Size: size,
		},
	}

	// Start workers
	p.workerWg.Add(size)
	for i := 0; i < size; i++ {
		go p.worker(i)
	}

	p.obs.Logger.Info("workerpool started",
		"name", p.name,
		"size", size,
		"queue_size", queueSize,
	)

	return p
}

// worker runs the main worker loop
func (p *Pool) worker(id int) {
	defer p.workerWg.Done()

	p.obs.Logger.Debug("worker started", "worker_id", id, "pool", p.name)

	for {
		select {
		case submission := <-p.taskCh:
			atomic.AddInt64(&p.metrics.Queued, -1)
			p.executeTask(submission, id)

		case <-p.baseCtx.Done():
			p.obs.Logger.Debug("worker stopping due to context cancellation",
				"worker_id", id, "pool", p.name)
			return
		}
	}
}

// executeTask executes a single task with proper error handling and metrics
func (p *Pool) executeTask(submission taskSubmission, workerID int) {
	atomic.AddInt64(&p.metrics.Running, 1)
	defer atomic.AddInt64(&p.metrics.Running, -1)

	// Create task context that cancels when either submission context or pool context is done
	// Handle case where submission context might be nil
	submissionCtx := submission.ctx
	if submissionCtx == nil {
		submissionCtx = context.Background()
	}
	taskCtx, taskCancel := context.WithCancel(submissionCtx)
	defer taskCancel()

	// Monitor for pool context cancellation
	go func() {
		select {
		case <-p.baseCtx.Done():
			taskCancel()
		case <-taskCtx.Done():
		}
	}()

	task := submission.task
	if p.taskWrapper != nil {
		task = p.taskWrapper(task)
	}

	// Record metrics
	p.obs.Metrics.Inc("ion_workerpool_tasks_started_total",
		"pool_name", p.name, "worker_id", workerID)

	// Execute with panic recovery
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				atomic.AddUint64(&p.metrics.Panicked, 1)
				p.obs.Metrics.Inc("ion_workerpool_tasks_completed_total",
					"pool_name", p.name, "status", "panic")

				if p.panicHandler != nil {
					p.panicHandler(r)
				} else {
					p.obs.Logger.Error("task panicked",
						fmt.Errorf("panic: %v", r),
						"pool", p.name, "worker_id", workerID)
				}
			}
		}()

		err = task(taskCtx)
	}()

	// Update completion metrics
	if err != nil {
		atomic.AddUint64(&p.metrics.Failed, 1)
		p.obs.Metrics.Inc("ion_workerpool_tasks_completed_total",
			"pool_name", p.name, "status", "error")
		p.obs.Logger.Error("task failed", err,
			"pool", p.name, "worker_id", workerID)
	} else {
		atomic.AddUint64(&p.metrics.Completed, 1)
		p.obs.Metrics.Inc("ion_workerpool_tasks_completed_total",
			"pool_name", p.name, "status", "success")
	}
}

// Metrics returns a snapshot of the current pool metrics
func (p *Pool) Metrics() PoolMetrics {
	return PoolMetrics{
		Size:      p.metrics.Size,
		Queued:    atomic.LoadInt64(&p.metrics.Queued),
		Running:   atomic.LoadInt64(&p.metrics.Running),
		Completed: atomic.LoadUint64(&p.metrics.Completed),
		Failed:    atomic.LoadUint64(&p.metrics.Failed),
		Panicked:  atomic.LoadUint64(&p.metrics.Panicked),
	}
}
