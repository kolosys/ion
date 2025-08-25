package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func BenchmarkTokenBucketAllowN(b *testing.B) {
	tb := NewTokenBucket(PerSecond(1000), 100)
	now := time.Now()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tb.AllowN(now, 1)
		}
	})
}

func BenchmarkTokenBucketAllowN_Uncontended(b *testing.B) {
	tb := NewTokenBucket(PerSecond(1000), 1000) // Large burst to avoid contention
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb.AllowN(now, 1)
	}
}

func BenchmarkTokenBucketAllowN_WithRefill(b *testing.B) {
	tb := NewTokenBucket(PerSecond(1000), 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb.AllowN(time.Now(), 1) // Use current time to trigger refill
	}
}

func BenchmarkLeakyBucketAllowN(b *testing.B) {
	lb := NewLeakyBucket(PerSecond(1000), 100)
	now := time.Now()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lb.AllowN(now, 1)
		}
	})
}

func BenchmarkLeakyBucketAllowN_Uncontended(b *testing.B) {
	lb := NewLeakyBucket(PerSecond(1000), 1000) // Large capacity
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lb.AllowN(now, 1)
	}
}

func BenchmarkLeakyBucketAllowN_WithLeak(b *testing.B) {
	lb := NewLeakyBucket(PerSecond(1000), 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lb.AllowN(time.Now(), 1) // Use current time to trigger leak
	}
}

func BenchmarkTokenBucketWaitN(b *testing.B) {
	tb := NewTokenBucket(PerSecond(10000), 1000) // High rate to minimize waiting
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = tb.WaitN(ctx, 1)
		}
	})
}

func BenchmarkLeakyBucketWaitN(b *testing.B) {
	lb := NewLeakyBucket(PerSecond(10000), 1000) // High rate, large capacity
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = lb.WaitN(ctx, 1)
		}
	})
}

// Benchmark comparing with golang.org/x/time/rate for reference
func BenchmarkComparison_AllowN(b *testing.B) {
	b.Run("TokenBucket", func(b *testing.B) {
		tb := NewTokenBucket(PerSecond(1000), 100)
		now := time.Now()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tb.AllowN(now, 1)
		}
	})

	b.Run("LeakyBucket", func(b *testing.B) {
		lb := NewLeakyBucket(PerSecond(1000), 100)
		now := time.Now()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			lb.AllowN(now, 1)
		}
	})
}

// Memory allocation benchmarks
func BenchmarkTokenBucketAlloc(b *testing.B) {
	tb := NewTokenBucket(PerSecond(1000), 100)
	now := time.Now()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb.AllowN(now, 1)
	}
}

func BenchmarkLeakyBucketAlloc(b *testing.B) {
	lb := NewLeakyBucket(PerSecond(1000), 100)
	now := time.Now()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lb.AllowN(now, 1)
	}
}

// High contention benchmark
func BenchmarkHighContention(b *testing.B) {
	const numGoroutines = 100

	b.Run("TokenBucket", func(b *testing.B) {
		tb := NewTokenBucket(PerSecond(10000), 1000)
		now := time.Now()

		b.ResetTimer()
		
		var wg sync.WaitGroup
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < b.N/numGoroutines; j++ {
					tb.AllowN(now, 1)
				}
			}()
		}
		wg.Wait()
	})

	b.Run("LeakyBucket", func(b *testing.B) {
		lb := NewLeakyBucket(PerSecond(10000), 1000)
		now := time.Now()

		b.ResetTimer()
		
		var wg sync.WaitGroup
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < b.N/numGoroutines; j++ {
					lb.AllowN(now, 1)
				}
			}()
		}
		wg.Wait()
	})
}

// Benchmark different burst/capacity sizes
func BenchmarkScalability(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("TokenBucket_Burst_%d", size), func(b *testing.B) {
			tb := NewTokenBucket(PerSecond(1000), size)
			now := time.Now()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tb.AllowN(now, 1)
			}
		})

		b.Run(fmt.Sprintf("LeakyBucket_Capacity_%d", size), func(b *testing.B) {
			lb := NewLeakyBucket(PerSecond(1000), size)
			now := time.Now()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				lb.AllowN(now, 1)
			}
		})
	}
}
