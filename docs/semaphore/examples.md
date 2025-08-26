# semaphore Examples

## Basic Usage

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

## Advanced Examples

See the [examples directory](https://github.com/kolosys/ion/tree/main/examples/semaphore) for more comprehensive examples.
