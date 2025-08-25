// Package shared provides common error types and observability interfaces
// used across all ion components.
package shared

import (
	"errors"
	"fmt"
)

// Common sentinel errors used across ion components
var (
	// ErrInvalidWeight is returned when a negative or zero weight is provided to semaphore operations
	ErrInvalidWeight = errors.New("ion: invalid weight, must be positive")

	// ErrContextCanceled is returned when an operation is canceled due to context cancellation
	ErrContextCanceled = errors.New("ion: operation canceled")
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

// SemaphoreError represents semaphore-specific errors with context
type SemaphoreError struct {
	Op            string // operation that failed
	SemaphoreName string // name of the semaphore
	Err           error  // underlying error
}

func (e *SemaphoreError) Error() string {
	if e.SemaphoreName != "" {
		return fmt.Sprintf("ion: semaphore %q %s: %v", e.SemaphoreName, e.Op, e.Err)
	}
	return fmt.Sprintf("ion: semaphore %s: %v", e.Op, e.Err)
}

func (e *SemaphoreError) Unwrap() error {
	return e.Err
}

// NewWeightExceedsCapacityError creates an error indicating the requested weight exceeds capacity
func NewWeightExceedsCapacityError(semaphoreName string, weight, capacity int64) error {
	return &SemaphoreError{
		Op:            "acquire",
		SemaphoreName: semaphoreName,
		Err:           fmt.Errorf("weight %d exceeds capacity %d", weight, capacity),
	}
}

// NewAcquireTimeoutError creates an error indicating an acquire operation timed out
func NewAcquireTimeoutError(semaphoreName string) error {
	return &SemaphoreError{
		Op:            "acquire",
		SemaphoreName: semaphoreName,
		Err:           errors.New("acquire timeout"),
	}
}
