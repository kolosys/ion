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
