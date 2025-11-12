# multitier_demo

This example demonstrates basic usage of the library.

## Source Code

```go
// Package main demonstrates advanced multi-tier rate limiting with the ion ratelimit package.
package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/kolosys/ion/ratelimit"
)

func basicMultiTierExample() {
	fmt.Println("\n1. Basic Multi-Tier Configuration:")

	// Create a simple multi-tier configuration
	config := ratelimit.DefaultMultiTierConfig()
	config.GlobalRate = ratelimit.PerSecond(10)      // Global limit
	config.GlobalBurst = 10
	config.DefaultRouteRate = ratelimit.PerSecond(5)  // Per-route limit
	config.DefaultRouteBurst = 5

	limiter := ratelimit.NewMultiTierLimiter(config, ratelimit.WithName("basic-multitier"))

	// Test global vs route limits
	fmt.Println("  Testing global vs route limits:")
	
	// This should hit the route limit first (5 requests)
	req := &ratelimit.Request{
		Method:   "GET",
		Endpoint: "/api/test",
		Context:  context.Background(),
	}

	allowed := 0
	for i := 0; i < 8; i++ {
		if limiter.Allow(req) {
			allowed++
			fmt.Printf("    Request %d: OK\n", i+1)
		} else {
			fmt.Printf("    Request %d: DENIED (limit exceeded)\n", i+1)
		}
	}

	fmt.Printf("  Results: %d/8 requests allowed\n", allowed)
	fmt.Printf("  (Route limit: 5/sec, Global limit: 10/sec)\n")
}

func apiGatewayExample() {
	fmt.Println("\n2. API Gateway with Route Patterns:")

	// Create an API gateway configuration
	config := ratelimit.DefaultMultiTierConfig()
	config.GlobalRate = ratelimit.PerSecond(100)     // High global limit
	config.GlobalBurst = 100
	config.DefaultRouteRate = ratelimit.PerSecond(20) // Default route limit
	config.DefaultRouteBurst = 20

	// Define specific route patterns
	config.RoutePatterns = map[string]ratelimit.RouteConfig{
		"POST:/api/v1/users": {
			Rate:  ratelimit.PerSecond(2), // User creation: very limited
			Burst: 2,
		},
		"GET:/api/v1/users/{id}": {
			Rate:  ratelimit.PerSecond(30), // User lookup: higher limit
			Burst: 30,
		},
		"DELETE:/api/v1/users/{id}": {
			Rate:  ratelimit.PerSecond(1), // User deletion: very limited
			Burst: 1,
		},
		"GET:/api/v1/health": {
			Rate:  ratelimit.PerSecond(100), // Health check: very high limit
			Burst: 100,
		},
	}

	limiter := ratelimit.NewMultiTierLimiter(config, ratelimit.WithName("api-gateway"))

	// Test different endpoints
	endpoints := []struct {
		name     string
		method   string
		endpoint string
		expected int
	}{
		{"User Creation", "POST", "/api/v1/users", 2},
		{"User Lookup", "GET", "/api/v1/users/123", 30},
		{"User Deletion", "DELETE", "/api/v1/users/123", 1},
		{"Health Check", "GET", "/api/v1/health", 100},
		{"Unknown Route", "GET", "/api/v1/unknown", 20}, // Uses default
	}

	fmt.Println("  Testing different endpoints:")
	for _, ep := range endpoints {
		req := &ratelimit.Request{
			Method:   ep.method,
			Endpoint: ep.endpoint,
			Context:  context.Background(),
		}

		allowed := 0
		for i := 0; i < ep.expected+2; i++ { // Try 2 more than expected
			if limiter.Allow(req) {
				allowed++
			}
		}

		fmt.Printf("    %s: %d/%d allowed (limit: %d/sec)\n", 
			ep.name, allowed, ep.expected+2, ep.expected)
	}
}

func resourceBasedExample() {
	fmt.Println("\n3. Resource-Based Rate Limiting:")

	// Create configuration with resource limits
	config := ratelimit.DefaultMultiTierConfig()
	config.GlobalRate = ratelimit.PerSecond(50)       // High global limit
	config.GlobalBurst = 50
	config.DefaultRouteRate = ratelimit.PerSecond(25)  // Moderate route limit
	config.DefaultRouteBurst = 25
	config.DefaultResourceRate = ratelimit.PerSecond(5) // Per-resource limit
	config.DefaultResourceBurst = 5

	limiter := ratelimit.NewMultiTierLimiter(config, ratelimit.WithName("resource-based"))

	// Test resource isolation
	fmt.Println("  Testing resource isolation:")
	
	resources := []string{"org123", "org456", "org789"}
	
	for _, resource := range resources {
		req := &ratelimit.Request{
			Method:     "GET",
			Endpoint:   "/api/v1/data",
			ResourceID: resource,
			Context:    context.Background(),
		}

		allowed := 0
		for i := 0; i < 8; i++ {
			if limiter.Allow(req) {
				allowed++
			}
		}

		fmt.Printf("    Resource %s: %d/8 allowed (limit: 5/sec per resource)\n", 
			resource, allowed)
	}

	// Test that different resources don't interfere
	fmt.Println("  Testing resource independence:")
	req1 := &ratelimit.Request{
		Method:     "GET",
		Endpoint:   "/api/v1/data",
		ResourceID: "org123",
		Context:    context.Background(),
	}
	req2 := &ratelimit.Request{
		Method:     "GET",
		Endpoint:   "/api/v1/data",
		ResourceID: "org456",
		Context:    context.Background(),
	}

	// Use up org123's limit
	for i := 0; i < 5; i++ {
		limiter.Allow(req1)
	}

	// org456 should still be able to make requests
	if limiter.Allow(req2) {
		fmt.Println("    OK: Different resources are isolated")
	} else {
		fmt.Println("    ERROR: Resources are not properly isolated")
	}
}

func headerUpdateExample() {
	fmt.Println("\n4. Header-Based Rate Limit Updates:")

	config := ratelimit.DefaultMultiTierConfig()
	config.GlobalRate = ratelimit.PerSecond(10)
	config.GlobalBurst = 10
	config.EnableBucketMapping = true

	limiter := ratelimit.NewMultiTierLimiter(config, ratelimit.WithName("header-updates"))

	req := &ratelimit.Request{
		Method:   "GET",
		Endpoint: "/api/v1/external",
		Context:  context.Background(),
	}

	// Simulate receiving rate limit headers from external API
	headers := map[string]string{
		"X-RateLimit-Bucket":        "external-api-bucket-123",
		"X-RateLimit-Limit":         "100",
		"X-RateLimit-Remaining":     "95",
		"X-RateLimit-Reset":         "1640995200",
		"X-RateLimit-Reset-After":   "60.5",
		"X-RateLimit-Global":        "false",
	}

	err := limiter.UpdateRateLimitFromHeaders(req, headers)
	if err != nil {
		fmt.Printf("  Error updating headers: %v\n", err)
	} else {
		fmt.Println("  OK: Successfully processed rate limit headers")
		fmt.Println("  OK: Bucket mapping enabled for external API integration")
	}
}

func concurrentAccessExample() {
	fmt.Println("\n5. Concurrent Multi-Tier Access:")

	config := ratelimit.DefaultMultiTierConfig()
	config.GlobalRate = ratelimit.PerSecond(50)
	config.GlobalBurst = 50
	config.DefaultRouteRate = ratelimit.PerSecond(20)
	config.DefaultRouteBurst = 20
	config.DefaultResourceRate = ratelimit.PerSecond(10)
	config.DefaultResourceBurst = 10

	limiter := ratelimit.NewMultiTierLimiter(config, ratelimit.WithName("concurrent"))

	// Test concurrent access to different routes and resources
	fmt.Println("  Testing concurrent access patterns:")

	var wg sync.WaitGroup
	results := make(chan string, 100)

	// Launch concurrent requests to different endpoints
	endpoints := []struct {
		method   string
		endpoint string
		resource string
	}{
		{"GET", "/api/v1/users", ""},
		{"POST", "/api/v1/users", ""},
		{"GET", "/api/v1/data", "org123"},
		{"GET", "/api/v1/data", "org456"},
		{"GET", "/api/v1/health", ""},
	}

	for _, ep := range endpoints {
		for i := 0; i < 5; i++ { // 5 requests per endpoint
			wg.Add(1)
			go func(method, endpoint, resource string, id int) {
				defer wg.Done()
				
				req := &ratelimit.Request{
					Method:     method,
					Endpoint:   endpoint,
					ResourceID: resource,
					Context:    context.Background(),
				}

				if limiter.Allow(req) {
					results <- fmt.Sprintf("OK %s %s (%s) - Request %d", method, endpoint, resource, id)
				} else {
					results <- fmt.Sprintf("DENIED %s %s (%s) - Request %d", method, endpoint, resource, id)
				}
			}(ep.method, ep.endpoint, ep.resource, i+1)
		}
	}

	wg.Wait()
	close(results)

	// Collect and display results
	allowed := 0
	total := 0
	for result := range results {
		fmt.Printf("    %s\n", result)
		total++
		if len(result) >= 2 && result[:2] == "OK" {
			allowed++
		}
	}

	fmt.Printf("  Concurrent results: %d/%d requests allowed\n", allowed, total)

	// Show final metrics
	metrics := limiter.GetMetrics()
	fmt.Printf("\n  Final Metrics:\n")
	fmt.Printf("    Total requests: %d\n", metrics.TotalRequests)
	fmt.Printf("    Global limit hits: %d\n", metrics.GlobalLimitHits)
	fmt.Printf("    Route limit hits: %d\n", metrics.RouteLimitHits)
	fmt.Printf("    Resource limit hits: %d\n", metrics.ResourceLimitHits)
	fmt.Printf("    Active buckets: %d\n", metrics.BucketsActive)
	fmt.Printf("    Average wait time: %v\n", metrics.AvgWaitTime)
	fmt.Printf("    Maximum wait time: %v\n", metrics.MaxWaitTime)
}

```

## Running the Example

To run this example:

```bash
cd ratelimit
go run multitier_demo.go
```

## Expected Output

```
Hello from Proton examples!
```
