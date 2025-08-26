package ratelimit_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kolosys/ion/ratelimit"
)

func ExampleNewTokenBucket() {
	// Create a token bucket that allows 10 requests per second with a burst of 5
	limiter := ratelimit.NewTokenBucket(ratelimit.PerSecond(10), 5)

	// Check current token count
	fmt.Printf("Initial tokens: %.1f\n", limiter.Tokens())

	// Allow some requests
	fmt.Printf("Allow 3: %t\n", limiter.AllowN(time.Now(), 3))
	fmt.Printf("Tokens after: %.1f\n", limiter.Tokens())

	// Try to exceed burst
	fmt.Printf("Allow 5: %t\n", limiter.AllowN(time.Now(), 5))

	// Output:
	// Initial tokens: 5.0
	// Allow 3: true
	// Tokens after: 2.0
	// Allow 5: false
}

func ExampleTokenBucket_WaitN() {
	// Create a token bucket with slow refill rate for demonstration
	limiter := ratelimit.NewTokenBucket(ratelimit.PerSecond(2), 1)

	// Use the initial token
	limiter.AllowN(time.Now(), 1)

	fmt.Println("Waiting for token...")
	start := time.Now()

	// This will wait for approximately 0.5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := limiter.WaitN(ctx, 1); err != nil {
		log.Printf("Wait failed: %v", err)
		return
	}

	elapsed := time.Since(start)
	fmt.Printf("Got token after %v\n", elapsed.Round(100*time.Millisecond))

	// Output:
	// Waiting for token...
	// Got token after 500ms
}

func ExampleNewLeakyBucket() {
	// Create a leaky bucket that processes 5 requests per second with capacity 3
	limiter := ratelimit.NewLeakyBucket(ratelimit.PerSecond(5), 3)

	// Check available space
	fmt.Printf("Available space: %d\n", limiter.Available())

	// Add requests to bucket
	fmt.Printf("Allow 2: %t\n", limiter.AllowN(time.Now(), 2))
	fmt.Printf("Available after: %d\n", limiter.Available())

	// Try to overfill
	fmt.Printf("Allow 2 more: %t\n", limiter.AllowN(time.Now(), 2))

	// Output:
	// Available space: 3
	// Allow 2: true
	// Available after: 1
	// Allow 2 more: false
}

func ExampleRate() {
	// Different ways to create rates
	rate1 := ratelimit.PerSecond(100)        // 100 per second
	rate2 := ratelimit.PerMinute(60)         // 1 per second
	rate3 := ratelimit.Per(5, 2*time.Second) // 2.5 per second

	fmt.Printf("PerSecond(100): %s\n", rate1)
	fmt.Printf("PerMinute(60): %s\n", rate2)
	fmt.Printf("Per(5, 2s): %s\n", rate3)

	// Output:
	// PerSecond(100): 100.0/s
	// PerMinute(60): 1.0/s
	// Per(5, 2s): 2.5/s
}

func ExampleTokenBucket_apiClient() {
	// Simulate an API client with rate limiting
	limiter := ratelimit.NewTokenBucket(
		ratelimit.PerSecond(10), // 10 requests per second
		20,                      // burst of 20
		ratelimit.WithName("api-client"),
	)

	// Function to make API request
	makeRequest := func(id int) {
		if !limiter.AllowN(time.Now(), 1) {
			fmt.Printf("Request %d: rate limited\n", id)
			return
		}
		fmt.Printf("Request %d: sent\n", id)
	}

	// Make requests
	for i := 1; i <= 25; i++ {
		makeRequest(i)
	}

	fmt.Printf("Tokens remaining: %.1f\n", limiter.Tokens())
}

func ExampleLeakyBucket_queueManagement() {
	// Simulate a queue with controlled processing rate
	queue := ratelimit.NewLeakyBucket(
		ratelimit.PerSecond(3), // process 3 items per second
		10,                     // queue capacity of 10
		ratelimit.WithName("task-queue"),
	)

	// Function to add task to queue
	addTask := func(id int) {
		if !queue.AllowN(time.Now(), 1) {
			fmt.Printf("Task %d: queue full\n", id)
			return
		}
		fmt.Printf("Task %d: queued (level: %.1f)\n", id, queue.Level())
	}

	// Add tasks
	for i := 1; i <= 12; i++ {
		addTask(i)
	}

	fmt.Printf("Queue available: %d\n", queue.Available())
}
