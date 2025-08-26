package semaphore

import (
	"context"
	"time"

	"github.com/kolosys/ion/shared"
)

// Acquire blocks until n permits are available or the context is canceled.
// Returns an error if n is invalid, exceeds capacity, or if the context is canceled.
func (s *weightedSemaphore) Acquire(ctx context.Context, n int64) error {
	if n <= 0 {
		return shared.ErrInvalidWeight
	}

	if n > s.capacity {
		return shared.NewWeightExceedsCapacityError(s.name, n, s.capacity)
	}

	// Fast path: try to acquire without blocking
	if s.tryAcquireFast(n) {
		s.obs.Metrics.Inc("ion_semaphore_acquisitions_total",
			"semaphore_name", s.name, "result", "success")
		return nil
	}

	// Slow path: need to wait
	return s.acquireSlow(ctx, n)
}

// TryAcquire attempts to acquire n permits without blocking.
// Returns true if successful, false otherwise.
func (s *weightedSemaphore) TryAcquire(n int64) bool {
	if n <= 0 {
		return false
	}

	if n > s.capacity {
		return false
	}

	success := s.tryAcquireFast(n)

	result := "denied"
	if success {
		result = "success"
	}

	s.obs.Metrics.Inc("ion_semaphore_acquisitions_total",
		"semaphore_name", s.name, "result", result)

	return success
}

// tryAcquireFast attempts to acquire permits without blocking
func (s *weightedSemaphore) tryAcquireFast(n int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return false
	}

	if s.current >= n {
		s.current -= n
		s.obs.Metrics.Gauge("ion_semaphore_current_permits", float64(s.current), "semaphore_name", s.name)
		return true
	}

	return false
}

// acquireSlow handles the blocking acquisition path
func (s *weightedSemaphore) acquireSlow(ctx context.Context, n int64) error {
	// Apply timeout if configured
	if s.acquireTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.acquireTimeout)
		defer cancel()
	}

	// Create waiter
	w := &waiter{
		weight: n,
		ready:  make(chan struct{}),
		ctx:    ctx,
	}

	// Add to queue
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return shared.NewAcquireTimeoutError(s.name)
	}

	s.waiters.push(w)
	waitingCount := s.waiters.len()
	s.mu.Unlock()

	s.obs.Metrics.Gauge("ion_semaphore_waiting_goroutines", float64(waitingCount), "semaphore_name", s.name)
	s.obs.Logger.Debug("semaphore acquire waiting",
		"semaphore_name", s.name,
		"weight", n,
		"waiting_count", waitingCount,
	)

	start := time.Now()

	// Wait for either ready signal or context cancellation
	select {
	case <-w.ready:
		if w.acquired {
			duration := time.Since(start)
			s.obs.Metrics.Histogram("ion_semaphore_acquire_duration_seconds", duration.Seconds(), "semaphore_name", s.name)
			s.obs.Metrics.Inc("ion_semaphore_acquisitions_total",
				"semaphore_name", s.name, "result", "success")
			return nil
		}
		// waiter was notified but couldn't acquire (shouldn't happen with current impl)
		return shared.NewAcquireTimeoutError(s.name)

	case <-ctx.Done():
		// Remove waiter from queue on cancellation
		s.mu.Lock()
		removed := s.waiters.removeWaiter(w)
		waitingCount := s.waiters.len()
		s.mu.Unlock()

		if removed {
			s.obs.Metrics.Gauge("ion_semaphore_waiting_goroutines", float64(waitingCount), "semaphore_name", s.name)
			s.obs.Logger.Debug("semaphore acquire canceled",
				"semaphore_name", s.name,
				"weight", n,
			)
		}

		// Determine the appropriate error based on context
		if ctx.Err() == context.DeadlineExceeded {
			s.obs.Metrics.Inc("ion_semaphore_acquisitions_total",
				"semaphore_name", s.name, "result", "timeout")
			return shared.NewAcquireTimeoutError(s.name)
		}

		s.obs.Metrics.Inc("ion_semaphore_acquisitions_total",
			"semaphore_name", s.name, "result", "canceled")
		return ctx.Err()
	}
}

// notifyWaiters attempts to satisfy waiting acquire requests
// Must be called with s.mu held
func (s *weightedSemaphore) notifyWaiters() {
	for s.current > 0 && s.waiters.len() > 0 {
		w := s.waiters.popReady(s.current)
		if w == nil {
			// No waiters can be satisfied with current permits
			break
		}

		// Check if waiter's context is still valid
		select {
		case <-w.ctx.Done():
			// Waiter was canceled, continue to next
			continue
		default:
		}

		// Acquire permits for this waiter
		if s.current >= w.weight {
			s.current -= w.weight
			w.acquired = true

			// Signal the waiter (non-blocking)
			select {
			case w.ready <- struct{}{}:
			default:
			}
		}
	}

	// Update metrics
	s.obs.Metrics.Gauge("ion_semaphore_current_permits", float64(s.current), "semaphore_name", s.name)
	s.obs.Metrics.Gauge("ion_semaphore_waiting_goroutines", float64(s.waiters.len()), "semaphore_name", s.name)
}
