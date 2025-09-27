package ratelimit

import (
	"errors"
	"fmt"
	"time"
)

// RateLimitError represents rate limiting specific errors with context
type RateLimitError struct {
	Op          string        // operation that failed
	LimiterName string        // name of the rate limiter
	Err         error         // underlying error
	RetryAfter  time.Duration // suggested retry delay
	Global      bool          // whether this is a global rate limit
	Bucket      string        // rate limit bucket identifier
	Remaining   int           // remaining requests in bucket
	Limit       int           // total limit for bucket
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
