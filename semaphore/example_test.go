package semaphore_test

import (
	"context"
	"fmt"
	"log"

	"github.com/kolosys/ion/semaphore"
)

func ExampleNewWeighted() {
	// Create a semaphore with capacity of 3
	sem := semaphore.NewWeighted(3)

	fmt.Printf("Initial permits: %d\n", sem.Current())

	// Acquire 2 permits
	sem.Acquire(context.Background(), 2)
	fmt.Printf("After acquiring 2: %d\n", sem.Current())

	// Try to acquire 2 more (should fail since only 1 permit available)
	if sem.TryAcquire(2) {
		fmt.Println("Acquired 2 more")
	} else {
		fmt.Println("Could not acquire 2 more")
	}

	// Release the 2 permits
	sem.Release(2)
	fmt.Printf("After releasing 2: %d\n", sem.Current())

	// Output:
	// Initial permits: 3
	// After acquiring 2: 1
	// Could not acquire 2 more
	// After releasing 2: 3
}

func ExampleSemaphore_Acquire() {
	// Resource pool with 2 connections
	sem := semaphore.NewWeighted(2)

	// Acquire a connection
	if err := sem.Acquire(context.Background(), 1); err != nil {
		log.Printf("Failed to acquire: %v", err)
		return
	}

	fmt.Printf("Acquired 1 permit, %d remaining\n", sem.Current())

	// Do work...

	// Release the connection
	sem.Release(1)
	fmt.Printf("Released 1 permit, %d available\n", sem.Current())

	// Output:
	// Acquired 1 permit, 1 remaining
	// Released 1 permit, 2 available
}

func ExampleSemaphore_TryAcquire() {
	// Semaphore with 1 permit
	sem := semaphore.NewWeighted(1)

	// Acquire the permit
	sem.Acquire(context.Background(), 1)

	// Try to acquire another permit (should fail)
	if sem.TryAcquire(1) {
		fmt.Println("Got permit")
	} else {
		fmt.Println("No permits available")
	}

	// Release and try again
	sem.Release(1)
	if sem.TryAcquire(1) {
		fmt.Println("Got permit after release")
		sem.Release(1)
	}

	// Output:
	// No permits available
	// Got permit after release
}
