package semaphore

import (
	"errors"
	"fmt"
)

// Common sentinel errors for semaphore operations
var (
	// ErrInvalidWeight is returned when a negative or zero weight is provided to semaphore operations
	ErrInvalidWeight = errors.New("ion: invalid weight, must be positive")
)

// SemaphoreError represents semaphore-specific errors with context
type SemaphoreError struct {
	Op   string // operation that failed
	Name string // name of the semaphore
	Err  error  // underlying error
}

func (e *SemaphoreError) Error() string {
	if e.Name != "" {
		return fmt.Sprintf("ion: semaphore %q %s: %v", e.Name, e.Op, e.Err)
	}
	return fmt.Sprintf("ion: semaphore %s: %v", e.Op, e.Err)
}

func (e *SemaphoreError) Unwrap() error {
	return e.Err
}

// NewWeightExceedsCapacityError creates an error indicating the requested weight exceeds capacity
func NewWeightExceedsCapacityError(semaphoreName string, weight, capacity int64) error {
	return &SemaphoreError{
		Op:   "acquire",
		Name: semaphoreName,
		Err:  fmt.Errorf("weight %d exceeds capacity %d", weight, capacity),
	}
}

// NewAcquireTimeoutError creates an error indicating an acquire operation timed out
func NewAcquireTimeoutError(semaphoreName string) error {
	return &SemaphoreError{
		Op:   "acquire",
		Name: semaphoreName,
		Err:  errors.New("acquire timeout"),
	}
}
