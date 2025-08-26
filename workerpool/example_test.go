package workerpool_test

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/kolosys/ion/workerpool"
)

func ExamplePool() {
	// Create a worker pool with 2 workers and a queue size of 5
	pool := workerpool.New(2, 5, workerpool.WithName("example-pool"))
	defer pool.Close(context.Background())

	// Submit a simple task
	task := func(ctx context.Context) error {
		fmt.Println("Task executed")
		return nil
	}

	if err := pool.Submit(context.Background(), task); err != nil {
		log.Printf("Submit failed: %v", err)
		return
	}

	// Wait briefly for task completion
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("Completed tasks: %d", pool.Metrics().Completed)
	// Output:
	// Task executed
	// Completed tasks: 1
}

func ExamplePool_TrySubmit() {
	// Create a small pool to demonstrate non-blocking submission
	pool := workerpool.New(1, 0) // No queue
	defer pool.Close(context.Background())

	// Fill the single worker
	longTask := func(ctx context.Context) error {
		time.Sleep(200 * time.Millisecond)
		return nil
	}
	_ = pool.TrySubmit(longTask)

	// This should fail immediately since worker is busy
	quickTask := func(ctx context.Context) error {
		fmt.Println("Quick task executed")
		return nil
	}

	if err := pool.TrySubmit(quickTask); err != nil {
		fmt.Println("TrySubmit failed: queue full")
	}
	// Output: TrySubmit failed: queue full
}

func ExamplePool_Drain() {
	pool := workerpool.New(1, 2, workerpool.WithName("drain-example"))

	// Submit tasks with synchronized output
	var wg sync.WaitGroup
	var mu sync.Mutex
	var outputs []string

	for i := 0; i < 2; i++ {
		wg.Add(1)
		task := func(ctx context.Context) error {
			defer wg.Done()
			mu.Lock()
			outputs = append(outputs, "Task completed")
			mu.Unlock()
			return nil
		}
		pool.Submit(context.Background(), task)
	}

	// Drain waits for all queued tasks to complete
	if err := pool.Drain(context.Background()); err != nil {
		log.Printf("Drain failed: %v", err)
	}

	// Wait for all tasks and then print in order
	wg.Wait()
	mu.Lock()
	for _, output := range outputs {
		fmt.Println(output)
	}
	mu.Unlock()
	fmt.Println("All tasks finished")
	// Output:
	// Task completed
	// Task completed
	// All tasks finished
}
