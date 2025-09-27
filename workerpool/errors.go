package workerpool

import (
	"errors"
	"fmt"
)

// PoolError represents workerpool-specific errors with context
type PoolError struct {
	Op       string // operation that failed
	PoolName string // name of the pool
	Err      error  // underlying error
}

func (e *PoolError) Error() string {
	if e.PoolName != "" {
		return fmt.Sprintf("ion: pool %q %s: %v", e.PoolName, e.Op, e.Err)
	}
	return fmt.Sprintf("ion: pool %s: %v", e.Op, e.Err)
}

func (e *PoolError) Unwrap() error {
	return e.Err
}

// NewPoolClosedError creates an error indicating the pool is closed
func NewPoolClosedError(poolName string) error {
	return &PoolError{
		Op:       "submit",
		PoolName: poolName,
		Err:      errors.New("pool is closed"),
	}
}

// NewQueueFullError creates an error indicating the queue is full
func NewQueueFullError(poolName string, queueSize int) error {
	return &PoolError{
		Op:       "submit",
		PoolName: poolName,
		Err:      fmt.Errorf("queue is full (size: %d)", queueSize),
	}
}
