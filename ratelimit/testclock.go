package ratelimit

import (
	"sync"
	"time"
)

// testClock is a controllable clock implementation for testing.
type testClock struct {
	mu    sync.Mutex
	now   time.Time
	timers []*testTimer
}

// newTestClock creates a new test clock starting at the given time.
func newTestClock(start time.Time) *testClock {
	return &testClock{
		now: start,
	}
}

func (c *testClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *testClock) Sleep(d time.Duration) {
	c.Advance(d)
}

func (c *testClock) AfterFunc(d time.Duration, f func()) Timer {
	c.mu.Lock()
	defer c.mu.Unlock()

	timer := &testTimer{
		clock:    c,
		deadline: c.now.Add(d),
		fn:       f,
		stopped:  false,
	}
	
	c.timers = append(c.timers, timer)
	return timer
}

// Advance advances the clock by the given duration and fires any timers.
func (c *testClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.now = c.now.Add(d)
	
	// Fire any timers that should trigger
	var remaining []*testTimer
	for _, timer := range c.timers {
		if !timer.stopped && !timer.deadline.After(c.now) {
			go timer.fn() // Fire timer function
		} else if !timer.stopped {
			remaining = append(remaining, timer)
		}
	}
	c.timers = remaining
}

// Set sets the clock to a specific time.
func (c *testClock) Set(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = t
}

type testTimer struct {
	clock    *testClock
	deadline time.Time
	fn       func()
	stopped  bool
	mu       sync.Mutex
}

func (t *testTimer) Stop() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if t.stopped {
		return false
	}
	
	t.stopped = true
	return true
}
