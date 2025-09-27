package circuit

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCircuitBreakerBasicFunctionality(t *testing.T) {
	cb := New("test-circuit", WithFailureThreshold(3))

	// Initially should be closed
	if cb.State() != Closed {
		t.Errorf("expected initial state to be Closed, got %v", cb.State())
	}

	// Successful requests should work
	ctx := context.Background()
	result, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
		return "success", nil
	})

	if err != nil {
		t.Errorf("expected successful request to work, got error: %v", err)
	}

	if result != "success" {
		t.Errorf("expected result 'success', got %v", result)
	}
}

func TestCircuitBreakerStateTransitions(t *testing.T) {
	cb := New("test-circuit",
		WithFailureThreshold(2),
		WithRecoveryTimeout(100*time.Millisecond),
		WithHalfOpenMaxRequests(1),
		WithHalfOpenSuccessThreshold(1),
	)

	ctx := context.Background()

	// Start in Closed state
	if cb.State() != Closed {
		t.Errorf("expected initial state to be Closed, got %v", cb.State())
	}

	// Cause failures to trip the circuit
	for i := 0; i < 2; i++ {
		_, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
			return nil, errors.New("failure")
		})
		if err == nil {
			t.Errorf("expected error on iteration %d", i)
		}
	}

	// Should now be Open
	if cb.State() != Open {
		t.Errorf("expected state to be Open after failures, got %v", cb.State())
	}

	// Requests should fail fast
	_, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
		t.Error("function should not be called when circuit is open")
		return nil, nil
	})

	if err == nil {
		t.Error("expected error when circuit is open")
	}

	// Wait for recovery timeout
	time.Sleep(150 * time.Millisecond)

	// Next request should transition to HalfOpen
	_, err = cb.Execute(ctx, func(ctx context.Context) (any, error) {
		return "success", nil
	})

	if err != nil {
		t.Errorf("expected successful request after recovery timeout, got error: %v", err)
	}

	// Should now be back to Closed (since we have success threshold of 1)
	if cb.State() != Closed {
		t.Errorf("expected state to be Closed after successful recovery, got %v", cb.State())
	}
}

func TestCircuitBreakerConcurrentAccess(t *testing.T) {
	cb := New("test-circuit", WithFailureThreshold(10))

	const numGoroutines = 100
	const requestsPerGoroutine = 10

	var wg sync.WaitGroup
	var successes atomic.Int64
	var failures atomic.Int64

	ctx := context.Background()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < requestsPerGoroutine; j++ {
				_, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
					// Some requests succeed, some fail
					if (id+j)%3 == 0 {
						return nil, errors.New("failure")
					}
					return "success", nil
				})

				if err != nil {
					failures.Add(1)
				} else {
					successes.Add(1)
				}
			}
		}(i)
	}

	wg.Wait()

	totalRequests := successes.Load() + failures.Load()
	expectedRequests := numGoroutines * requestsPerGoroutine

	if totalRequests != int64(expectedRequests) {
		t.Errorf("expected %d total requests, got %d", expectedRequests, totalRequests)
	}

	metrics := cb.Metrics()
	if metrics.TotalRequests < totalRequests {
		t.Errorf("metrics show fewer requests (%d) than actual (%d)", metrics.TotalRequests, totalRequests)
	}
}

func TestCircuitBreakerContextCancellation(t *testing.T) {
	cb := New("test-circuit")

	ctx, cancel := context.WithCancel(context.Background())

	// Start a request that will be canceled
	done := make(chan bool)
	go func() {
		_, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
			// Wait for context cancellation
			<-ctx.Done()
			return nil, ctx.Err()
		})

		if err == nil {
			t.Error("expected error due to context cancellation")
		}
		done <- true
	}()

	// Cancel the context
	cancel()

	// Wait for completion
	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Error("timeout waiting for context cancellation")
	}
}

func TestCircuitBreakerMetrics(t *testing.T) {
	cb := New("test-circuit", WithFailureThreshold(3))

	ctx := context.Background()

	// Make some requests
	for i := 0; i < 5; i++ {
		if i < 2 {
			// Successful requests
			cb.Execute(ctx, func(ctx context.Context) (any, error) {
				return "success", nil
			})
		} else {
			// Failed requests
			cb.Execute(ctx, func(ctx context.Context) (any, error) {
				return nil, errors.New("failure")
			})
		}
	}

	metrics := cb.Metrics()

	if metrics.Name != "test-circuit" {
		t.Errorf("expected name 'test-circuit', got %s", metrics.Name)
	}

	if metrics.TotalRequests != 5 {
		t.Errorf("expected 5 total requests, got %d", metrics.TotalRequests)
	}

	if metrics.TotalSuccesses != 2 {
		t.Errorf("expected 2 successful requests, got %d", metrics.TotalSuccesses)
	}

	if metrics.TotalFailures != 3 {
		t.Errorf("expected 3 failed requests, got %d", metrics.TotalFailures)
	}

	if metrics.State != Open {
		t.Errorf("expected state to be Open, got %v", metrics.State)
	}

	failureRate := metrics.FailureRate()
	expectedFailureRate := 3.0 / 5.0
	if failureRate != expectedFailureRate {
		t.Errorf("expected failure rate %f, got %f", expectedFailureRate, failureRate)
	}
}

func TestCircuitBreakerCustomFailurePredicate(t *testing.T) {
	// Only specific errors should count as failures
	cb := New("test-circuit",
		WithFailureThreshold(2),
		WithFailurePredicate(func(err error) bool {
			return err != nil && err.Error() == "serious-error"
		}),
	)

	ctx := context.Background()

	// Non-serious errors shouldn't trip the circuit
	for i := 0; i < 5; i++ {
		_, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
			return nil, errors.New("minor-error")
		})
		if err == nil {
			t.Error("expected error from function")
		}
	}

	// Should still be closed
	if cb.State() != Closed {
		t.Errorf("expected state to be Closed after minor errors, got %v", cb.State())
	}

	// Now cause serious errors
	for i := 0; i < 2; i++ {
		cb.Execute(ctx, func(ctx context.Context) (any, error) {
			return nil, errors.New("serious-error")
		})
	}

	// Should now be open
	if cb.State() != Open {
		t.Errorf("expected state to be Open after serious errors, got %v", cb.State())
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	cb := New("test-circuit", WithFailureThreshold(2))

	ctx := context.Background()

	// Trip the circuit
	for i := 0; i < 2; i++ {
		cb.Execute(ctx, func(ctx context.Context) (any, error) {
			return nil, errors.New("failure")
		})
	}

	if cb.State() != Open {
		t.Errorf("expected state to be Open, got %v", cb.State())
	}

	// Reset the circuit
	cb.Reset()

	if cb.State() != Closed {
		t.Errorf("expected state to be Closed after reset, got %v", cb.State())
	}

	// Should work normally now
	_, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
		return "success", nil
	})

	if err != nil {
		t.Errorf("expected successful request after reset, got error: %v", err)
	}
}

func TestCircuitBreakerCall(t *testing.T) {
	cb := New("test-circuit")

	ctx := context.Background()

	// Test successful call
	err := cb.Call(ctx, func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Errorf("expected successful call, got error: %v", err)
	}

	// Test failed call
	err = cb.Call(ctx, func(ctx context.Context) error {
		return errors.New("call failed")
	})

	if err == nil {
		t.Error("expected error from failed call")
	}
}

func TestCircuitBreakerStateChangeCallback(t *testing.T) {
	var stateChanges []string
	cb := New("test-circuit",
		WithFailureThreshold(2),
		WithStateChangeCallback(func(from, to State) {
			stateChanges = append(stateChanges, fmt.Sprintf("%s->%s", from, to))
		}),
	)

	ctx := context.Background()

	// Cause failures to trigger state changes
	for i := 0; i < 2; i++ {
		cb.Execute(ctx, func(ctx context.Context) (any, error) {
			return nil, errors.New("failure")
		})
	}

	// Should have transitioned from Closed to Open
	if len(stateChanges) == 0 {
		t.Error("expected state change callback to be called")
	}
}

func TestCircuitBreakerHalfOpenTransition(t *testing.T) {
	cb := New("test-circuit",
		WithFailureThreshold(1),
		WithRecoveryTimeout(50*time.Millisecond),
		WithHalfOpenMaxRequests(2),
		WithHalfOpenSuccessThreshold(1),
	)

	ctx := context.Background()

	// Trip the circuit
	cb.Execute(ctx, func(ctx context.Context) (any, error) {
		return nil, errors.New("failure")
	})

	if cb.State() != Open {
		t.Errorf("expected state to be Open, got %v", cb.State())
	}

	// Wait for recovery timeout
	time.Sleep(60 * time.Millisecond)

	// First request should transition to half-open
	_, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
		return "success", nil
	})

	if err != nil {
		t.Errorf("expected successful request after recovery timeout, got error: %v", err)
	}

	// Should be back to closed after one success (threshold = 1)
	if cb.State() != Closed {
		t.Errorf("expected state to be Closed after recovery, got %v", cb.State())
	}
}

func TestCircuitBreakerHalfOpenFailure(t *testing.T) {
	cb := New("test-circuit",
		WithFailureThreshold(1),
		WithRecoveryTimeout(50*time.Millisecond),
		WithHalfOpenMaxRequests(2),
		WithHalfOpenSuccessThreshold(2),
	)

	ctx := context.Background()

	// Trip the circuit
	cb.Execute(ctx, func(ctx context.Context) (any, error) {
		return nil, errors.New("failure")
	})

	// Wait for recovery timeout
	time.Sleep(60 * time.Millisecond)

	// First request in half-open should succeed
	cb.Execute(ctx, func(ctx context.Context) (any, error) {
		return "success", nil
	})

	// Second request fails - should trip back to open
	_, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
		return nil, errors.New("failure in half-open")
	})

	if err == nil {
		t.Error("expected error from failed request")
	}

	// Should be back to open
	if cb.State() != Open {
		t.Errorf("expected state to be Open after failure in half-open, got %v", cb.State())
	}
}

func TestCircuitBreakerHalfOpenMaxRequests(t *testing.T) {
	cb := New("test-circuit",
		WithFailureThreshold(1),
		WithRecoveryTimeout(50*time.Millisecond),
		WithHalfOpenMaxRequests(2),
		WithHalfOpenSuccessThreshold(3), // More than max requests
	)

	ctx := context.Background()

	// Trip the circuit
	cb.Execute(ctx, func(ctx context.Context) (any, error) {
		return nil, errors.New("failure")
	})

	// Wait for recovery timeout
	time.Sleep(60 * time.Millisecond)

	// Make max requests in half-open
	for i := 0; i < 2; i++ {
		_, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
			return "success", nil
		})
		if err != nil {
			t.Errorf("unexpected error in half-open request %d: %v", i+1, err)
		}
	}

	// Third request should be rejected (exceeds max)
	_, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
		t.Error("function should not be called when half-open max requests exceeded")
		return nil, nil
	})

	if err == nil {
		t.Error("expected error when exceeding half-open max requests")
	}
}

func TestCircuitBreakerConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				FailureThreshold:         5,
				RecoveryTimeout:          30 * time.Second,
				HalfOpenMaxRequests:      3,
				HalfOpenSuccessThreshold: 2,
			},
			wantErr: false,
		},
		{
			name: "zero failure threshold",
			config: &Config{
				FailureThreshold:         0,
				RecoveryTimeout:          30 * time.Second,
				HalfOpenMaxRequests:      3,
				HalfOpenSuccessThreshold: 2,
			},
			wantErr: true,
		},
		{
			name: "negative recovery timeout",
			config: &Config{
				FailureThreshold:         5,
				RecoveryTimeout:          -time.Second,
				HalfOpenMaxRequests:      3,
				HalfOpenSuccessThreshold: 2,
			},
			wantErr: true,
		},
		{
			name: "success threshold exceeds max requests",
			config: &Config{
				FailureThreshold:         5,
				RecoveryTimeout:          30 * time.Second,
				HalfOpenMaxRequests:      2,
				HalfOpenSuccessThreshold: 5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCircuitBreakerMetricsHelpers(t *testing.T) {
	cb := New("test-circuit", WithFailureThreshold(5))
	ctx := context.Background()

	// Test metrics with no requests
	metrics := cb.Metrics()
	if metrics.FailureRate() != 0.0 {
		t.Errorf("expected failure rate 0.0 for no requests, got %f", metrics.FailureRate())
	}
	if metrics.SuccessRate() != 1.0 {
		t.Errorf("expected success rate 1.0 for no requests, got %f", metrics.SuccessRate())
	}
	if !metrics.IsHealthy() {
		t.Error("expected circuit to be healthy initially")
	}

	// Make some requests with failures
	for i := 0; i < 3; i++ {
		cb.Execute(ctx, func(ctx context.Context) (any, error) {
			if i < 1 {
				return "success", nil
			}
			return nil, errors.New("failure")
		})
	}

	metrics = cb.Metrics()
	expectedFailureRate := 2.0 / 3.0
	if abs := func(x float64) float64 {
		if x < 0 {
			return -x
		}
		return x
	}; abs(metrics.FailureRate()-expectedFailureRate) > 0.001 {
		t.Errorf("expected failure rate %f, got %f", expectedFailureRate, metrics.FailureRate())
	}

	expectedSuccessRate := 1.0 / 3.0
	if abs := func(x float64) float64 {
		if x < 0 {
			return -x
		}
		return x
	}; abs(metrics.SuccessRate()-expectedSuccessRate) > 0.001 {
		t.Errorf("expected success rate %f, got %f", expectedSuccessRate, metrics.SuccessRate())
	}

	if metrics.IsHealthy() {
		t.Error("expected circuit to not be healthy with consecutive failures")
	}
}

func TestCircuitBreakerClose(t *testing.T) {
	cb := New("test-circuit")

	err := cb.Close()
	if err != nil {
		t.Errorf("expected Close() to succeed, got error: %v", err)
	}
}

func TestCircuitBreakerStateString(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{Closed, "Closed"},
		{Open, "Open"},
		{HalfOpen, "HalfOpen"},
		{State(999), "State(999)"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.expected {
			t.Errorf("State.String() = %v, want %v", got, tt.expected)
		}
	}
}

func TestCircuitBreakerPresetConfigurations(t *testing.T) {
	testConfigs := map[string][]Option{
		"QuickFailover": QuickFailover(),
		"Conservative":  Conservative(),
		"Aggressive":    Aggressive(),
	}

	for name, options := range testConfigs {
		t.Run(name, func(t *testing.T) {
			cb := New(fmt.Sprintf("test-%s", name), options...)
			if cb == nil {
				t.Errorf("failed to create circuit breaker with %s preset", name)
			}

			// Basic functionality test
			ctx := context.Background()
			_, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
				return "success", nil
			})

			if err != nil {
				t.Errorf("preset %s failed basic execution: %v", name, err)
			}
		})
	}
}
