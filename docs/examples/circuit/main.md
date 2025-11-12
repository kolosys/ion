# main

This example demonstrates basic usage of the library.

## Source Code

```go
// Package main demonstrates circuit breaker usage in real-world scenarios.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kolosys/ion/circuit"
	"github.com/kolosys/ion/observe"
)

// ExampleLogger implements shared.Logger for demonstration
type ExampleLogger struct{}

func (l ExampleLogger) Debug(msg string, kv ...any) { log.Printf("[DEBUG] %s %v", msg, kv) }
func (l ExampleLogger) Info(msg string, kv ...any)  { log.Printf("[INFO] %s %v", msg, kv) }
func (l ExampleLogger) Warn(msg string, kv ...any)  { log.Printf("[WARN] %s %v", msg, kv) }
func (l ExampleLogger) Error(msg string, err error, kv ...any) {
	log.Printf("[ERROR] %s: %v %v", msg, err, kv)
}

// ExampleMetrics implements shared.Metrics for demonstration
type ExampleMetrics struct{}

func (m ExampleMetrics) Inc(name string, kv ...any) { log.Printf("[METRIC] %s++ %v", name, kv) }
func (m ExampleMetrics) Add(name string, v float64, kv ...any) {
	log.Printf("[METRIC] %s+=%f %v", name, v, kv)
}
func (m ExampleMetrics) Gauge(name string, v float64, kv ...any) {
	log.Printf("[METRIC] %s=%f %v", name, v, kv)
}
func (m ExampleMetrics) Histogram(name string, v float64, kv ...any) {
	log.Printf("[METRIC] %s~%f %v", name, v, kv)
}

// MockPaymentService simulates an unreliable payment service
type MockPaymentService struct {
	failureRate float64 // 0.0 to 1.0
	latency     time.Duration
}

func (ps *MockPaymentService) ProcessPayment(ctx context.Context, amount float64) (*PaymentResult, error) {
	// Simulate latency
	select {
	case <-time.After(ps.latency):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Simulate failures
	if rand.Float64() < ps.failureRate {
		return nil, errors.New("payment service temporarily unavailable")
	}

	return &PaymentResult{
		ID:     fmt.Sprintf("pay_%d", rand.Int63()),
		Amount: amount,
		Status: "completed",
	}, nil
}

type PaymentResult struct {
	ID     string
	Amount float64
	Status string
}

// ProtectedPaymentService wraps the payment service with circuit breaker protection
type ProtectedPaymentService struct {
	service *MockPaymentService
	circuit circuit.CircuitBreaker
}

func NewProtectedPaymentService(failureRate float64, latency time.Duration) *ProtectedPaymentService {
	obs := observe.New().
		WithLogger(ExampleLogger{}).
		WithMetrics(ExampleMetrics{})

	cb := circuit.New("payment-service",
		circuit.WithFailureThreshold(5),
		circuit.WithRecoveryTimeout(30*time.Second),
		circuit.WithHalfOpenMaxRequests(3),
		circuit.WithHalfOpenSuccessThreshold(2),
		circuit.WithObservability(obs),
		circuit.WithStateChangeCallback(func(from, to circuit.State) {
			log.Printf("Payment service circuit: %s -> %s", from, to)
		}),
	)

	return &ProtectedPaymentService{
		service: &MockPaymentService{
			failureRate: failureRate,
			latency:     latency,
		},
		circuit: cb,
	}
}

func (pps *ProtectedPaymentService) ProcessPayment(ctx context.Context, amount float64) (*PaymentResult, error) {
	result, err := pps.circuit.Execute(ctx, func(ctx context.Context) (any, error) {
		return pps.service.ProcessPayment(ctx, amount)
	})

	if err != nil {
		return nil, err
	}

	return result.(*PaymentResult), nil
}

func (pps *ProtectedPaymentService) Metrics() circuit.CircuitMetrics {
	return pps.circuit.Metrics()
}

// HTTPClientWrapper demonstrates circuit breaker for HTTP clients
type HTTPClientWrapper struct {
	client  *http.Client
	circuit circuit.CircuitBreaker
}

func NewHTTPClientWrapper(baseURL string) *HTTPClientWrapper {
	cb := circuit.New("http-client",
		circuit.WithFailureThreshold(3),
		circuit.WithRecoveryTimeout(15*time.Second),
		circuit.WithFailurePredicate(func(err error) bool {
			// Only count 5xx errors and timeouts as failures
			// 4xx errors (client errors) should not trip the circuit
			if err == nil {
				return false
			}
			// In a real implementation, you'd check the HTTP status code
			return true
		}),
	)

	return &HTTPClientWrapper{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		circuit: cb,
	}
}

func (hw *HTTPClientWrapper) Get(ctx context.Context, url string) (*http.Response, error) {
	result, err := hw.circuit.Execute(ctx, func(ctx context.Context) (any, error) {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}
		return hw.client.Do(req)
	})

	if err != nil {
		return nil, err
	}

	return result.(*http.Response), nil
}

func main() {
	fmt.Println("=== ION Circuit Breaker Examples ===")

	// Example 1: Payment Service Protection
	fmt.Println("1. Payment Service with Circuit Breaker Protection")
	paymentExample()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// Example 2: Concurrent Load Testing
	fmt.Println("2. Concurrent Load Testing")
	concurrentExample()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// Example 3: Different Configuration Strategies
	fmt.Println("3. Configuration Strategy Comparison")
	configurationExample()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// Example 4: HTTP Client with Circuit Breaker
	fmt.Println("4. HTTP Client Circuit Breaker Protection")
	httpClientExample()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// Example 5: Recovery Scenarios
	fmt.Println("5. Circuit Recovery Scenarios")
	recoveryExample()
}

func paymentExample() {
	// Create a payment service with 30% failure rate
	paymentService := NewProtectedPaymentService(0.3, 100*time.Millisecond)

	ctx := context.Background()

	// Process 20 payments
	for i := 1; i <= 20; i++ {
		amount := float64(i * 10)

		result, err := paymentService.ProcessPayment(ctx, amount)
		if err != nil {
			fmt.Printf("Payment %d ($%.2f) FAILED: %v\n", i, amount, err)
		} else {
			fmt.Printf("Payment %d ($%.2f) SUCCESS: %s\n", i, amount, result.ID)
		}

		// Small delay between requests
		time.Sleep(50 * time.Millisecond)
	}

	// Show final metrics
	metrics := paymentService.Metrics()
	fmt.Printf("\nPayment Service Metrics:\n")
	fmt.Printf("  State: %s\n", metrics.State)
	fmt.Printf("  Total Requests: %d\n", metrics.TotalRequests)
	fmt.Printf("  Successes: %d\n", metrics.TotalSuccesses)
	fmt.Printf("  Failures: %d\n", metrics.TotalFailures)
	fmt.Printf("  Failure Rate: %.2f%%\n", metrics.FailureRate()*100)
	fmt.Printf("  State Changes: %d\n", metrics.StateChanges)
}

func concurrentExample() {
	// Create a service with moderate failure rate
	paymentService := NewProtectedPaymentService(0.2, 50*time.Millisecond)

	const numWorkers = 10
	const requestsPerWorker = 5

	var wg sync.WaitGroup
	results := make(chan string, numWorkers*requestsPerWorker)

	ctx := context.Background()

	start := time.Now()

	// Launch concurrent workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < requestsPerWorker; j++ {
				amount := float64((workerID*requestsPerWorker + j + 1) * 10)

				_, err := paymentService.ProcessPayment(ctx, amount)
				if err != nil {
					results <- fmt.Sprintf("Worker %d: FAILED $%.2f - %v", workerID, amount, err)
				} else {
					results <- fmt.Sprintf("Worker %d: SUCCESS $%.2f", workerID, amount)
				}
			}
		}(i)
	}

	wg.Wait()
	close(results)

	duration := time.Since(start)

	// Collect and display results
	var successes, failures int
	for result := range results {
		fmt.Println(result)
		if strings.Contains(result, "SUCCESS") {
			successes++
		} else {
			failures++
		}
	}

	fmt.Printf("\nConcurrent Test Results (completed in %v):\n", duration)
	fmt.Printf("  Total: %d, Successes: %d, Failures: %d\n", successes+failures, successes, failures)

	metrics := paymentService.Metrics()
	fmt.Printf("  Circuit State: %s\n", metrics.State)
	fmt.Printf("  Circuit Failure Rate: %.2f%%\n", metrics.FailureRate()*100)
}

func configurationExample() {
	fmt.Println("Testing different circuit breaker configurations:")

	configs := map[string][]circuit.Option{
		"Quick Failover": circuit.QuickFailover(),
		"Conservative":   circuit.Conservative(),
		"Aggressive":     circuit.Aggressive(),
	}

	ctx := context.Background()

	for name, options := range configs {
		fmt.Printf("\n%s Configuration:\n", name)

		cb := circuit.New(name, options...)

		// Simulate failures to trip the circuit
		for i := 0; i < 15; i++ {
			_, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
				// 50% failure rate
				if i%2 == 0 {
					return nil, errors.New("simulated failure")
				}
				return "success", nil
			})

			state := cb.State()
			if err != nil {
				fmt.Printf("  Request %d: FAILED (state: %s)\n", i+1, state)
			} else {
				fmt.Printf("  Request %d: SUCCESS (state: %s)\n", i+1, state)
			}

			// Break early if circuit opens to show different thresholds
			if state == circuit.Open {
				fmt.Printf("  Circuit opened after %d requests\n", i+1)
				break
			}
		}

		metrics := cb.Metrics()
		fmt.Printf("  Final metrics: %d total, %d failures, %.2f%% failure rate\n",
			metrics.TotalRequests, metrics.TotalFailures, metrics.FailureRate()*100)
	}
}

func httpClientExample() {
	// Simulate HTTP service with circuit breaker
	httpClient := NewHTTPClientWrapper("https://api.example.com")
	ctx := context.Background()

	// Test URLs with different failure patterns
	testURLs := []string{
		"https://httpbin.org/status/200", // Should succeed
		"https://httpbin.org/status/500", // Server error
		"https://httpbin.org/delay/10",   // Timeout
		"https://invalid-url-404.com",    // DNS/connection error
	}

	for i, url := range testURLs {
		fmt.Printf("HTTP Request %d to %s\n", i+1, url)

		// Use shorter timeout for demo
		timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		resp, err := httpClient.Get(timeoutCtx, url)
		if err != nil {
			fmt.Printf("  FAILED: %v\n", err)
		} else {
			fmt.Printf("  SUCCESS: %s\n", resp.Status)
			resp.Body.Close()
		}

		// Show circuit state after each request
		fmt.Printf("  Circuit State: %s\n", httpClient.circuit.State())

		// Small delay between requests
		time.Sleep(100 * time.Millisecond)
	}
}

func recoveryExample() {
	fmt.Println("Demonstrating circuit recovery patterns...")

	// Create circuit with quick recovery for demo
	cb := circuit.New("recovery-demo",
		circuit.WithFailureThreshold(2),
		circuit.WithRecoveryTimeout(1*time.Second),
		circuit.WithHalfOpenMaxRequests(2),
		circuit.WithHalfOpenSuccessThreshold(1),
		circuit.WithStateChangeCallback(func(from, to circuit.State) {
			fmt.Printf("  🔄 Circuit state: %s → %s\n", from, to)
		}),
	)

	ctx := context.Background()
	requestCount := 0

	// Helper function to make a request
	makeRequest := func(shouldFail bool) {
		requestCount++
		result, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
			if shouldFail {
				return nil, fmt.Errorf("simulated failure %d", requestCount)
			}
			return fmt.Sprintf("success %d", requestCount), nil
		})

		state := cb.State()
		if err != nil {
			fmt.Printf("  ❌ Request %d: FAILED (%s) - %v\n", requestCount, state, err)
		} else {
			fmt.Printf("  ✅ Request %d: SUCCESS (%s) - %v\n", requestCount, state, result)
		}
	}

	// Phase 1: Trip the circuit
	fmt.Println("\n📍 Phase 1: Tripping the circuit with failures")
	makeRequest(true) // Failure 1
	makeRequest(true) // Failure 2 - should trip circuit

	// Phase 2: Circuit is open
	fmt.Println("\n📍 Phase 2: Circuit is open - requests fail fast")
	makeRequest(false) // Should fail fast even though operation would succeed

	// Phase 3: Wait for recovery timeout
	fmt.Println("\n📍 Phase 3: Waiting for recovery timeout...")
	time.Sleep(1200 * time.Millisecond) // Wait slightly longer than recovery timeout

	// Phase 4: Half-open state
	fmt.Println("\n📍 Phase 4: Circuit moves to half-open, testing recovery")
	makeRequest(false) // Should succeed and close circuit

	// Phase 5: Circuit recovered
	fmt.Println("\n📍 Phase 5: Circuit recovered - normal operation")
	makeRequest(false) // Normal successful operation
	makeRequest(false) // Another successful operation

	// Final metrics
	metrics := cb.Metrics()
	fmt.Printf("\nFinal Recovery Demo Metrics:\n")
	fmt.Printf("  Total Requests: %d\n", metrics.TotalRequests)
	fmt.Printf("  Successes: %d\n", metrics.TotalSuccesses)
	fmt.Printf("  Failures: %d\n", metrics.TotalFailures)
	fmt.Printf("  State Changes: %d\n", metrics.StateChanges)
	fmt.Printf("  Final State: %s\n", metrics.State)
	fmt.Printf("  Is Healthy: %t\n", metrics.IsHealthy())
}

```

## Running the Example

To run this example:

```bash
cd circuit
go run main.go
```

## Expected Output

```
Hello from Proton examples!
```
