package circuit

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func BenchmarkCircuitBreakerExecute_Closed(b *testing.B) {
	cb := New("benchmark-circuit")
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
				return "success", nil
			})
			if err != nil {
				b.Fatalf("unexpected error: %v", err)
			}
		}
	})
}

func BenchmarkCircuitBreakerExecute_Open(b *testing.B) {
	cb := New("benchmark-circuit", WithFailureThreshold(1))
	ctx := context.Background()

	// Trip the circuit
	cb.Execute(ctx, func(ctx context.Context) (any, error) {
		return nil, errors.New("failure")
	})

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
				b.Fatal("function should not be called when circuit is open")
				return nil, nil
			})
			if err == nil {
				b.Fatal("expected error when circuit is open")
			}
		}
	})
}

func BenchmarkCircuitBreakerState(b *testing.B) {
	cb := New("benchmark-circuit")

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = cb.State()
		}
	})
}

func BenchmarkCircuitBreakerMetrics(b *testing.B) {
	cb := New("benchmark-circuit")

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = cb.Metrics()
		}
	})
}

func BenchmarkCircuitBreakerCall_Closed(b *testing.B) {
	cb := New("benchmark-circuit")
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := cb.Call(ctx, func(ctx context.Context) error {
				return nil
			})
			if err != nil {
				b.Fatalf("unexpected error: %v", err)
			}
		}
	})
}

func BenchmarkCircuitBreakerConcurrent(b *testing.B) {
	cb := New("benchmark-circuit")
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	var wg sync.WaitGroup
	numGoroutines := 100

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < b.N/numGoroutines; j++ {
				cb.Execute(ctx, func(ctx context.Context) (any, error) {
					return "success", nil
				})
			}
		}()
	}
	wg.Wait()
}

func BenchmarkCircuitBreakerMixed(b *testing.B) {
	cb := New("benchmark-circuit", WithFailureThreshold(1000)) // High threshold to avoid state changes
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			_, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
				// 90% success rate
				if i%10 == 0 {
					return nil, errors.New("failure")
				}
				return "success", nil
			})
			// We expect some errors due to failures, so don't fail the benchmark
			_ = err
		}
	})
}

// Benchmark the state checking logic specifically
func BenchmarkCircuitBreakerAllowRequest_Closed(b *testing.B) {
	cb := New("benchmark-circuit").(*circuitBreaker)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = cb.allowRequest()
		}
	})
}

func BenchmarkCircuitBreakerAllowRequest_Open(b *testing.B) {
	cb := New("benchmark-circuit", WithFailureThreshold(1)).(*circuitBreaker)
	ctx := context.Background()

	// Trip the circuit
	cb.Execute(ctx, func(ctx context.Context) (any, error) {
		return nil, errors.New("failure")
	})

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = cb.allowRequest()
		}
	})
}

// Benchmark atomic operations
func BenchmarkCircuitBreakerAtomicOperations(b *testing.B) {
	cb := New("benchmark-circuit").(*circuitBreaker)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cb.totalRequests.Add(1)
			cb.totalSuccesses.Add(1)
			_ = cb.state.Load()
		}
	})
}

// Memory allocation test
func BenchmarkCircuitBreakerAllocations(b *testing.B) {
	cb := New("benchmark-circuit")
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
			return nil, nil // No allocation return
		})
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
