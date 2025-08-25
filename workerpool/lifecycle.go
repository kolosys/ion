package workerpool

import (
	"context"
	"time"
)

// Close immediately stops accepting new tasks and signals all workers to stop.
// It waits for currently running tasks to complete unless the provided context
// is canceled or times out. If the context expires, workers are asked to stop
// via task context cancellation.
func (p *Pool) Close(ctx context.Context) error {
	var err error

	p.closeOnce.Do(func() {
		p.obs.Logger.Info("closing workerpool", "pool", p.name)

		// Mark pool as closed to reject new submissions
		close(p.closed)

		// Cancel the base context to signal workers to stop
		p.cancel()

		// Close the task channel to prevent new tasks from being queued
		close(p.taskCh)

		// Wait for workers to finish with timeout
		done := make(chan struct{})
		go func() {
			p.workerWg.Wait()
			close(done)
		}()

		select {
		case <-done:
			p.obs.Logger.Info("workerpool closed gracefully", "pool", p.name)

		case <-ctx.Done():
			p.obs.Logger.Warn("workerpool close timed out, some tasks may have been interrupted",
				"pool", p.name, "error", ctx.Err())
			err = ctx.Err()
		}
	})

	return err
}

// Drain prevents new task submissions and waits for the queue to empty and all
// currently running tasks to complete. Unlike Close, Drain allows queued tasks
// to continue being processed until the queue is empty.
func (p *Pool) Drain(ctx context.Context) error {
	var err error

	p.drainOnce.Do(func() {
		p.obs.Logger.Info("draining workerpool", "pool", p.name)

		// Mark as draining to reject new submissions
		p.draining.Store(true)

		// Wait for queue to empty and workers to finish processing
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				p.obs.Logger.Warn("workerpool drain timed out",
					"pool", p.name, "error", ctx.Err())
				err = ctx.Err()
				// Still need to close after timeout
				p.Close(context.Background())
				return

			case <-ticker.C:
				metrics := p.Metrics()
				if metrics.Queued == 0 && metrics.Running == 0 {
					// Queue is empty and no tasks running, safe to close
					closeCtx, cancel := context.WithTimeout(context.Background(), p.drainTimeout)
					defer cancel()
					
					err = p.Close(closeCtx)
					p.obs.Logger.Info("workerpool drained successfully", "pool", p.name)
					return
				}

				p.obs.Logger.Debug("waiting for drain to complete",
					"pool", p.name,
					"queued", metrics.Queued,
					"running", metrics.Running,
				)
			}
		}
	})

	return err
}

// IsClosed returns true if the pool has been closed or is in the process of closing
func (p *Pool) IsClosed() bool {
	select {
	case <-p.closed:
		return true
	default:
		return false
	}
}

// IsDraining returns true if the pool is in draining mode (not accepting new tasks
// but still processing queued tasks)
func (p *Pool) IsDraining() bool {
	return p.draining.Load()
}
