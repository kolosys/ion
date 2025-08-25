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
