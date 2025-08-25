package workerpool

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/kolosys/ion/shared"
)

// Submit submits a task to the pool for execution. It respects the provided context
// for cancellation and timeouts. If the context is canceled before the task can be
// queued, it returns the context error wrapped. If the pool is closed or draining,
// it returns an appropriate error.
func (p *Pool) Submit(ctx context.Context, task Task) error {
	if task == nil {
		return errors.New("ion: nil task")
	}

	// Check if pool is closed
	select {
	case <-p.closed:
		return shared.NewPoolClosedError(p.name)
	default:
	}

	// Check if pool is draining
	if p.draining.Load() {
		return shared.NewPoolClosedError(p.name)
	}

	submission := taskSubmission{
		task: task,
		ctx:  ctx,
	}

	p.obs.Metrics.Inc("ion_workerpool_tasks_submitted_total", "pool_name", p.name)

	// Try to submit the task, respecting context cancellation and pool closure
	select {
	case p.taskCh <- submission:
		atomic.AddInt64(&p.metrics.Queued, 1)
		p.obs.Metrics.Gauge("ion_workerpool_queue_size", float64(atomic.LoadInt64(&p.metrics.Queued)), "pool_name", p.name)
		return nil

	case <-ctx.Done():
		return ctx.Err()

	case <-p.closed:
		return shared.NewPoolClosedError(p.name)
	}
}

// TrySubmit attempts to submit a task to the pool without blocking.
// It returns true if the task was successfully queued, false if the queue is full
// or the pool is closed/draining. It does not respect context cancellation since
// it returns immediately.
func (p *Pool) TrySubmit(task Task) error {
	if task == nil {
		return errors.New("ion: nil task")
	}

	// Check if pool is closed
	select {
	case <-p.closed:
		return shared.NewPoolClosedError(p.name)
	default:
	}

	// Check if pool is draining
	if p.draining.Load() {
		return shared.NewPoolClosedError(p.name)
	}

	submission := taskSubmission{
		task: task,
		ctx:  context.Background(), // TrySubmit uses background context
	}

	// Try to submit without blocking
	select {
	case p.taskCh <- submission:
		atomic.AddInt64(&p.metrics.Queued, 1)
		p.obs.Metrics.Inc("ion_workerpool_tasks_submitted_total", "pool_name", p.name)
		p.obs.Metrics.Gauge("ion_workerpool_queue_size", float64(atomic.LoadInt64(&p.metrics.Queued)), "pool_name", p.name)
		return nil

	default:
		// Queue is full
		return shared.NewQueueFullError(p.name, p.queueSize)
	}
}
