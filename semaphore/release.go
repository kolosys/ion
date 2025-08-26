package semaphore

import "fmt"

// Release returns n permits to the semaphore, potentially unblocking waiters.
// Panics if n is negative or if releasing would exceed the semaphore capacity.
func (s *weightedSemaphore) Release(n int64) {
	if n < 0 {
		panic(fmt.Sprintf("semaphore: cannot release negative permits: %d", n))
	}

	if n == 0 {
		return // No-op
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for capacity overflow
	if s.current+n > s.capacity {
		panic(fmt.Sprintf("semaphore: release would exceed capacity (current: %d, releasing: %d, capacity: %d)",
			s.current, n, s.capacity))
	}

	s.obs.Logger.Debug("semaphore releasing permits",
		"semaphore_name", s.name,
		"permits", n,
		"current_before", s.current,
	)

	// Return permits
	s.current += n

	s.obs.Logger.Debug("semaphore permits released",
		"semaphore_name", s.name,
		"permits", n,
		"current_after", s.current,
	)

	// Notify waiters that permits are available
	s.notifyWaiters()
}

// Current returns the number of permits currently available
func (s *weightedSemaphore) Current() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.current
}
