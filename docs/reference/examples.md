# Examples

Complete working examples for each Ion component.

## ratelimit

### Code

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
```

### Running this example

```bash
cd examples/ratelimit
go run main.go
```

[View on GitHub](https://github.com/kolosys/ion/tree/main/examples/ratelimit)

## semaphore

### Code

```go
// Package main demonstrates basic usage of the ion semaphore.
package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/kolosys/ion/semaphore"
)

func main() {
	fmt.Println("Ion Semaphore Examples")
	fmt.Println("=====================")

	// Example 1: Basic usage
	basicExample()

	// Example 2: Resource pool simulation
	resourcePoolExample()

	// Example 3: Weighted permits
	weightedExample()

	// Example 4: Fairness demonstration
	fairnessExample()
}

func basicExample() {
	fmt.Println("\n1. Basic Semaphore Usage:")
	
	// Create semaphore with capacity of 2
	sem := semaphore.NewWeighted(2, semaphore.WithName("basic-sem"))
	
	fmt.Printf("  Initial permits: %d\n", sem.Current())
	
	// Acquire permits
	sem.Acquire(context.Background(), 1)
	fmt.Printf("  After acquiring 1: %d permits left\n", sem.Current())
	
	// Try to acquire more than available
	if sem.TryAcquire(2) {
		fmt.Println("  Acquired 2 more permits")
	} else {
		fmt.Println("  Could not acquire 2 permits (only 1 available)")
	}
	
	// Release and acquire again
	sem.Release(1)
	fmt.Printf("  After releasing 1: %d permits available\n", sem.Current())
	
	if sem.TryAcquire(2) {
		fmt.Println("  Successfully acquired 2 permits")
		sem.Release(2)
	}
	
	fmt.Printf("  Final permits: %d\n", sem.Current())
}

func resourcePoolExample() {
	fmt.Println("\n2. Resource Pool Example (Database Connections):")
	
	// Simulate a database connection pool with 3 connections
	const maxConnections = 3
	sem := semaphore.NewWeighted(maxConnections, semaphore.WithName("db-pool"))
	
	var wg sync.WaitGroup
	
	// Simulate 5 clients trying to use database connections
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			
			fmt.Printf("  Client %d: requesting connection...\n", clientID)
			
			// Try to get a connection with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			
			start := time.Now()
			if err := sem.Acquire(ctx, 1); err != nil {
				fmt.Printf("  Client %d: timeout waiting for connection (%v)\n", clientID, time.Since(start))
				return
			}
			
			fmt.Printf("  Client %d: got connection after %v\n", clientID, time.Since(start))
			
			// Simulate database work
			time.Sleep(500 * time.Millisecond)
			
			// Release connection
			sem.Release(1)
			fmt.Printf("  Client %d: released connection\n", clientID)
		}(i)
	}
	
	wg.Wait()
	fmt.Printf("  All clients finished. Available connections: %d/%d\n", sem.Current(), maxConnections)
}

func weightedExample() {
	fmt.Println("\n3. Weighted Permits Example:")
	
	// CPU-bound task scheduler: small tasks use 1 core, large tasks use 3 cores
	const totalCores = 4
	sem := semaphore.NewWeighted(totalCores, semaphore.WithName("cpu-scheduler"))
	
	var wg sync.WaitGroup
	
	// Small task
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("  Small task: requesting 1 core\n")
		
		if err := sem.Acquire(context.Background(), 1); err != nil {
			log.Printf("Small task failed: %v", err)
			return
		}
		
		fmt.Printf("  Small task: running on 1 core\n")
		time.Sleep(800 * time.Millisecond)
		
		sem.Release(1)
		fmt.Printf("  Small task: completed\n")
	}()
	
	// Large task
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("  Large task: requesting 3 cores\n")
		
		if err := sem.Acquire(context.Background(), 3); err != nil {
			log.Printf("Large task failed: %v", err)
			return
		}
		
		fmt.Printf("  Large task: running on 3 cores\n")
		time.Sleep(600 * time.Millisecond)
		
		sem.Release(3)
		fmt.Printf("  Large task: completed\n")
	}()
	
	// Another small task
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(200 * time.Millisecond) // Start slightly later
		
		fmt.Printf("  Small task 2: requesting 1 core\n")
		
		start := time.Now()
		if err := sem.Acquire(context.Background(), 1); err != nil {
			log.Printf("Small task 2 failed: %v", err)
			return
		}
		
		fmt.Printf("  Small task 2: got core after %v\n", time.Since(start))
		time.Sleep(300 * time.Millisecond)
		
		sem.Release(1)
		fmt.Printf("  Small task 2: completed\n")
	}()
	
	wg.Wait()
	fmt.Printf("  All tasks completed. Available cores: %d/%d\n", sem.Current(), totalCores)
}

func fairnessExample() {
	fmt.Println("\n4. Fairness Example:")
	
	// Demonstrate FIFO vs LIFO fairness
	testFairness := func(fairness semaphore.Fairness, name string) {
		fmt.Printf("  Testing %s fairness:\n", name)
		
		sem := semaphore.NewWeighted(1, 
			semaphore.WithName(fmt.Sprintf("%s-sem", name)),
			semaphore.WithFairness(fairness),
		)
		
		// Acquire the single permit
		sem.Acquire(context.Background(), 1)
		
		var results []int
		var mu sync.Mutex
		var wg sync.WaitGroup
		
		// Start 3 waiters
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				
				// Small delay to ensure ordering
				time.Sleep(time.Duration(id) * 50 * time.Millisecond)
				
				if err := sem.Acquire(context.Background(), 1); err != nil {
					return
				}
				
				mu.Lock()
				results = append(results, id)
				mu.Unlock()
				
				// Quick work simulation
				time.Sleep(100 * time.Millisecond)
				sem.Release(1)
			}(i)
		}
		
		// Let waiters queue up
		time.Sleep(200 * time.Millisecond)
		
		// Release the initial permit
		sem.Release(1)
		
		wg.Wait()
		
		fmt.Printf("    Execution order: %v\n", results)
	}
	
	testFairness(semaphore.FIFO, "FIFO")
	testFairness(semaphore.LIFO, "LIFO")
}
```

### Running this example

```bash
cd examples/semaphore
go run main.go
```

[View on GitHub](https://github.com/kolosys/ion/tree/main/examples/semaphore)

## workerpool

### Code

```go
// Package main demonstrates basic usage of the ion workerpool.
package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/kolosys/ion/workerpool"
)

func main() {
	fmt.Println("Ion WorkerPool Example")
	fmt.Println("=====================")

	// Example 1: Basic usage
	basicExample()

	// Example 2: Error handling and observability
	observabilityExample()

	// Example 3: Graceful shutdown
	shutdownExample()
}

func basicExample() {
	fmt.Println("\n1. Basic Usage:")
	
	// Create a pool with 3 workers and queue size of 5
	pool := workerpool.New(3, 5, workerpool.WithName("basic-pool"))
	defer pool.Close(context.Background())

	var wg sync.WaitGroup

	// Submit 8 tasks
	for i := 0; i < 8; i++ {
		taskID := i
		wg.Add(1)

		task := func(ctx context.Context) error {
			defer wg.Done()
			
			// Simulate work
			workDuration := time.Duration(100+taskID*50) * time.Millisecond
			fmt.Printf("  Task %d starting (will take %v)\n", taskID, workDuration)
			
			select {
			case <-time.After(workDuration):
				fmt.Printf("  Task %d completed\n", taskID)
				return nil
			case <-ctx.Done():
				fmt.Printf("  Task %d canceled: %v\n", taskID, ctx.Err())
				return ctx.Err()
			}
		}

		if err := pool.Submit(context.Background(), task); err != nil {
			log.Printf("Failed to submit task %d: %v", taskID, err)
			wg.Done()
		}
	}

	wg.Wait()
	metrics := pool.Metrics()
	fmt.Printf("  Completed: %d, Failed: %d, Queue size: %d\n", 
		metrics.Completed, metrics.Failed, metrics.Queued)
}

func observabilityExample() {
	fmt.Println("\n2. Error Handling & Observability:")

	// Custom logger
	logger := &customLogger{}
	
	pool := workerpool.New(2, 3,
		workerpool.WithName("observable-pool"),
		workerpool.WithLogger(logger),
		workerpool.WithPanicRecovery(func(r any) {
			fmt.Printf("  Recovered from panic: %v\n", r)
		}),
	)
	defer pool.Close(context.Background())

	var wg sync.WaitGroup

	// Task that succeeds
	wg.Add(1)
	pool.Submit(context.Background(), func(ctx context.Context) error {
		defer wg.Done()
		fmt.Printf("  Success task completed\n")
		return nil
	})

	// Task that fails
	wg.Add(1)
	pool.Submit(context.Background(), func(ctx context.Context) error {
		defer wg.Done()
		return fmt.Errorf("simulated error")
	})

	// Task that panics
	wg.Add(1)
	pool.Submit(context.Background(), func(ctx context.Context) error {
		defer wg.Done()
		panic("simulated panic")
	})

	wg.Wait()
	metrics := pool.Metrics()
	fmt.Printf("  Final metrics - Completed: %d, Failed: %d, Panicked: %d\n",
		metrics.Completed, metrics.Failed, metrics.Panicked)
}

func shutdownExample() {
	fmt.Println("\n3. Graceful Shutdown:")

	pool := workerpool.New(2, 10, workerpool.WithName("shutdown-pool"))

	// Submit several long-running tasks
	for i := 0; i < 5; i++ {
		taskID := i
		task := func(ctx context.Context) error {
			for j := 0; j < 5; j++ {
				select {
				case <-time.After(200 * time.Millisecond):
					fmt.Printf("  Task %d: step %d/5\n", taskID, j+1)
				case <-ctx.Done():
					fmt.Printf("  Task %d: canceled at step %d/5\n", taskID, j+1)
					return ctx.Err()
				}
			}
			fmt.Printf("  Task %d: completed all steps\n", taskID)
			return nil
		}

		pool.Submit(context.Background(), task)
	}

	// Let tasks run for a bit
	time.Sleep(500 * time.Millisecond)

	// Drain the pool (wait for all tasks to complete)
	fmt.Printf("  Starting drain...\n")
	start := time.Now()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := pool.Drain(ctx); err != nil {
		fmt.Printf("  Drain failed: %v\n", err)
	} else {
		fmt.Printf("  Drain completed in %v\n", time.Since(start))
	}

	metrics := pool.Metrics()
	fmt.Printf("  Final state - Completed: %d, Running: %d, Queued: %d\n",
		metrics.Completed, metrics.Running, metrics.Queued)
}

// customLogger implements the shared.Logger interface
type customLogger struct{}

func (l *customLogger) Debug(msg string, kv ...any) {
	fmt.Printf("  [DEBUG] %s %v\n", msg, kv)
}

func (l *customLogger) Info(msg string, kv ...any) {
	fmt.Printf("  [INFO] %s %v\n", msg, kv)
}

func (l *customLogger) Warn(msg string, kv ...any) {
	fmt.Printf("  [WARN] %s %v\n", msg, kv)
}

func (l *customLogger) Error(msg string, err error, kv ...any) {
	fmt.Printf("  [ERROR] %s: %v %v\n", msg, err, kv)
}
```

### Running this example

```bash
cd examples/workerpool
go run main.go
```

[View on GitHub](https://github.com/kolosys/ion/tree/main/examples/workerpool)

