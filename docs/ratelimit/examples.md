# ratelimit Examples

## Basic Usage

The examples below demonstrate the core rate limiting functionality including token bucket, leaky bucket, and multi-tier rate limiting.

```go
// Package main demonstrates basic usage of the ion ratelimit package.
package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kolosys/ion/ratelimit"
)

func main() {
	fmt.Println("Ion RateLimit Examples")
	fmt.Println("======================")

	// Example 1: Token Bucket Basic Usage
	tokenBucketExample()

	// Example 2: Leaky Bucket Basic Usage
	leakyBucketExample()

	// Example 3: API Client Rate Limiting
	apiClientExample()

	// Example 4: Queue Processing with Leaky Bucket
	queueProcessingExample()

	// Example 5: Burst vs Sustained Traffic
	burstVsSustainedExample()

	// Example 6: Multi-Tier Rate Limiting
	multiTierExample()
}

func tokenBucketExample() {
	fmt.Println("\n1. Token Bucket Basic Usage:")

	// Create a token bucket: 5 requests per second, burst of 10
	limiter := ratelimit.NewTokenBucket(ratelimit.PerSecond(5), 10)

	fmt.Printf("  Initial tokens: %.1f\n", limiter.Tokens())

	// Allow some requests immediately (using burst)
	for i := 1; i <= 12; i++ {
		if limiter.AllowN(time.Now(), 1) {
			fmt.Printf("  Request %d: allowed\n", i)
		} else {
			fmt.Printf("  Request %d: denied\n", i)
		}
	}

	fmt.Printf("  Tokens remaining: %.1f\n", limiter.Tokens())
}

func leakyBucketExample() {
	fmt.Println("\n2. Leaky Bucket Basic Usage:")

	// Create a leaky bucket: process 3 per second, capacity of 5
	limiter := ratelimit.NewLeakyBucket(ratelimit.PerSecond(3), 5)

	fmt.Printf("  Initial available space: %d\n", limiter.Available())

	// Add requests to the bucket
	for i := 1; i <= 7; i++ {
		if limiter.AllowN(time.Now(), 1) {
			fmt.Printf("  Request %d: queued (level: %.1f)\n", i, limiter.Level())
		} else {
			fmt.Printf("  Request %d: rejected (bucket full)\n", i)
		}
	}

	fmt.Printf("  Final bucket level: %.1f\n", limiter.Level())
}

func apiClientExample() {
	fmt.Println("\n3. API Client Rate Limiting:")

	// Simulate API with different rate limits for different endpoints
	authLimiter := ratelimit.NewTokenBucket(ratelimit.PerMinute(100), 10,
		ratelimit.WithName("auth-api"))
	dataLimiter := ratelimit.NewTokenBucket(ratelimit.PerSecond(10), 20,
		ratelimit.WithName("data-api"))

	// Function to make API request
	makeAPIRequest := func(endpoint string, limiter ratelimit.Limiter, id int) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		if err := limiter.WaitN(ctx, 1); err != nil {
			fmt.Printf("  %s request %d: timeout (%v)\n", endpoint, id, err)
			return
		}
		fmt.Printf("  %s request %d: sent\n", endpoint, id)
	}

	var wg sync.WaitGroup

	// Auth requests (slower rate)
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			makeAPIRequest("Auth", authLimiter, id)
		}(i)
	}

	// Data requests (faster rate)
	for i := 1; i <= 8; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			makeAPIRequest("Data", dataLimiter, id)
		}(i)
	}

	wg.Wait()
}

func queueProcessingExample() {
	fmt.Println("\n4. Queue Processing with Leaky Bucket:")

	// Simulate a message queue with controlled processing rate
	processor := ratelimit.NewLeakyBucket(ratelimit.PerSecond(2), 8,
		ratelimit.WithName("message-processor"))

	// Simulate incoming messages
	messages := []string{
		"Process user signup", "Send email notification", "Update database",
		"Generate report", "Backup data", "Clean cache", "Send SMS",
		"Process payment", "Update inventory", "Log analytics",
	}

	fmt.Printf("  Processing %d messages at 2/second (capacity: %d)\n",
		len(messages), processor.Capacity())

	accepted := 0
	rejected := 0

	for i, msg := range messages {
		if processor.AllowN(time.Now(), 1) {
			fmt.Printf("  Message %d: %s -> QUEUED\n", i+1, msg)
			accepted++
		} else {
			fmt.Printf("  Message %d: %s -> REJECTED (queue full)\n", i+1, msg)
			rejected++
		}
	}

	fmt.Printf("  Results: %d accepted, %d rejected\n", accepted, rejected)
	fmt.Printf("  Queue level: %.1f/%d\n", processor.Level(), processor.Capacity())
}

func burstVsSustainedExample() {
	fmt.Println("\n5. Burst vs Sustained Traffic:")

	// Token bucket allows bursts but sustains at rate limit
	tokenBucket := ratelimit.NewTokenBucket(ratelimit.PerSecond(3), 6)

	// Leaky bucket smooths out bursts
	leakyBucket := ratelimit.NewLeakyBucket(ratelimit.PerSecond(3), 6)

	fmt.Println("  Sending 10 requests in a burst:")

	fmt.Println("  Token Bucket (allows burst):")
	for i := 1; i <= 10; i++ {
		if tokenBucket.AllowN(time.Now(), 1) {
			fmt.Printf("    Request %d: ✓\n", i)
		} else {
			fmt.Printf("    Request %d: ✗\n", i)
		}
	}

	fmt.Println("  Leaky Bucket (smooths burst):")
	for i := 1; i <= 10; i++ {
		if leakyBucket.AllowN(time.Now(), 1) {
			fmt.Printf("    Request %d: ✓\n", i)
		} else {
			fmt.Printf("    Request %d: ✗\n", i)
		}
	}

	fmt.Printf("  Token bucket tokens remaining: %.1f\n", tokenBucket.Tokens())
	fmt.Printf("  Leaky bucket level: %.1f\n", leakyBucket.Level())
}

## Multi-Tier Rate Limiting

The multi-tier rate limiter provides sophisticated rate limiting capabilities for API gateways and complex applications. It supports:

- **Global limits**: Overall rate limits across all requests
- **Route limits**: Per-endpoint rate limiting with pattern matching
- **Resource limits**: Per-resource (e.g., per-organization) rate limiting
- **Concurrent safety**: Thread-safe operations
- **Metrics**: Comprehensive observability

### Key Features

1. **Route Pattern Matching**: Define specific rate limits for different API endpoints
2. **Resource Isolation**: Different resources (organizations, users, etc.) get separate rate limit buckets
3. **Header Integration**: Support for external API rate limit headers
4. **Metrics Collection**: Track limit hits, wait times, and active buckets

func multiTierExample() {
	fmt.Println("\n6. Multi-Tier Rate Limiting:")

	// Create a multi-tier rate limiter configuration
	config := ratelimit.DefaultMultiTierConfig()
	config.GlobalRate = ratelimit.PerSecond(20)      // Global: 20 req/sec
	config.GlobalBurst = 20
	config.DefaultRouteRate = ratelimit.PerSecond(10) // Default route: 10 req/sec
	config.DefaultRouteBurst = 10
	config.DefaultResourceRate = ratelimit.PerSecond(5) // Per-resource: 5 req/sec
	config.DefaultResourceBurst = 5

	// Add specific route patterns
	config.RoutePatterns = map[string]ratelimit.RouteConfig{
		"POST:/api/v1/users": {
			Rate:  ratelimit.PerSecond(2), // User creation: 2 req/sec
			Burst: 2,
		},
		"GET:/api/v1/users/{id}": {
			Rate:  ratelimit.PerSecond(15), // User lookup: 15 req/sec
			Burst: 15,
		},
	}

	limiter := ratelimit.NewMultiTierLimiter(config, ratelimit.WithName("api-gateway"))

	// Test different types of requests
	testRequests := []struct {
		name     string
		method   string
		endpoint string
		resource string
	}{
		{"User Creation", "POST", "/api/v1/users", ""},
		{"User Lookup", "GET", "/api/v1/users/123", ""},
		{"Data Query", "GET", "/api/v1/data", "org123"},
		{"Data Query (diff org)", "GET", "/api/v1/data", "org456"},
	}

	fmt.Println("  Testing different request types:")
	for _, test := range testRequests {
		req := &ratelimit.Request{
			Method:     test.method,
			Endpoint:   test.endpoint,
			ResourceID: test.resource,
			Context:    context.Background(),
		}

		allowed := 0
		for i := 0; i < 5; i++ {
			if limiter.Allow(req) {
				allowed++
			}
		}

		fmt.Printf("    %s: %d/5 allowed\n", test.name, allowed)
	}

	// Test concurrent access
	fmt.Println("\n  Testing concurrent access:")
	var wg sync.WaitGroup
	results := make(chan string, 20)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			req := &ratelimit.Request{
				Method:   "GET",
				Endpoint: "/api/v1/users/123",
				Context:  context.Background(),
			}
			if limiter.Allow(req) {
				results <- fmt.Sprintf("Request %d: allowed", id)
			} else {
				results <- fmt.Sprintf("Request %d: denied", id)
			}
		}(i)
	}

	wg.Wait()
	close(results)

	allowed := 0
	for result := range results {
		fmt.Printf("    %s\n", result)
		if result[len(result)-7:] == "allowed" {
			allowed++
		}
	}

	fmt.Printf("  Concurrent results: %d/10 allowed\n", allowed)

	// Show metrics
	metrics := limiter.GetMetrics()
	fmt.Printf("\n  Metrics:\n")
	fmt.Printf("    Total requests: %d\n", metrics.TotalRequests)
	fmt.Printf("    Global limit hits: %d\n", metrics.GlobalLimitHits)
	fmt.Printf("    Route limit hits: %d\n", metrics.RouteLimitHits)
	fmt.Printf("    Resource limit hits: %d\n", metrics.ResourceLimitHits)
	fmt.Printf("    Active buckets: %d\n", metrics.BucketsActive)
}

```

## Advanced Examples

See the [examples directory](https://github.com/kolosys/ion/tree/main/examples/ratelimit) for more comprehensive examples.
