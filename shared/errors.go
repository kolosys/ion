// Package shared provides common error types and observability interfaces
// used across all ion components.
package shared

import (
	"errors"
	"fmt"
	"time"
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

// RateLimitError represents rate limiting specific errors with context
type RateLimitError struct {
	Op           string        // operation that failed
	LimiterName  string        // name of the rate limiter
	Err          error         // underlying error
	RetryAfter   time.Duration // suggested retry delay
	Global       bool          // whether this is a global rate limit
	Bucket       string        // rate limit bucket identifier
	Remaining    int           // remaining requests in bucket
	Limit        int           // total limit for bucket
}

func (e *RateLimitError) Error() string {
	if e.LimiterName != "" {
		return fmt.Sprintf("ion: rate limiter %q %s: %v (retry after: %v)", 
			e.LimiterName, e.Op, e.Err, e.RetryAfter)
	}
	return fmt.Sprintf("ion: rate limiter %s: %v (retry after: %v)", e.Op, e.Err, e.RetryAfter)
}

func (e *RateLimitError) Unwrap() error {
	return e.Err
}

// IsRetryable returns true if the rate limit error suggests retrying.
func (e *RateLimitError) IsRetryable() bool {
	return e.RetryAfter > 0
}

// NewRateLimitExceededError creates an error indicating rate limit was exceeded
func NewRateLimitExceededError(limiterName string, retryAfter time.Duration) error {
	return &RateLimitError{
		Op:          "wait",
		LimiterName: limiterName,
		Err:         errors.New("rate limit exceeded"),
		RetryAfter:  retryAfter,
	}
}

// NewGlobalRateLimitError creates an error for global rate limit hits
func NewGlobalRateLimitError(limiterName string, retryAfter time.Duration) error {
	return &RateLimitError{
		Op:          "wait",
		LimiterName: limiterName,
		Err:         errors.New("global rate limit exceeded"),
		RetryAfter:  retryAfter,
		Global:      true,
	}
}

// NewBucketLimitError creates an error for bucket-specific rate limits
func NewBucketLimitError(limiterName, bucket string, remaining, limit int, retryAfter time.Duration) error {
	return &RateLimitError{
		Op:          "wait",
		LimiterName: limiterName,
		Err:         fmt.Errorf("bucket rate limit exceeded (%d/%d remaining)", remaining, limit),
		RetryAfter:  retryAfter,
		Bucket:      bucket,
		Remaining:   remaining,
		Limit:       limit,
	}
}
