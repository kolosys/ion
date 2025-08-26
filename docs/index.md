---
layout: home
title: Ion
permalink: /
---

Build robust, high-performance Go applications with Ion's comprehensive suite of concurrency primitives. From rate limiting to worker pools, Ion provides production-ready tools for managing concurrent operations at scale.

## ⚡ Quick Start

Install Ion with a single command:

```bash
go get github.com/kolosys/ion
```

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kolosys/ion/workerpool"
    "github.com/kolosys/ion/ratelimit"
    "github.com/kolosys/ion/semaphore"
)

func main() {
    // 🚀 Worker Pool: Process tasks concurrently
    pool := workerpool.New(4, 10) // 4 workers, queue size 10
    defer pool.Close(context.Background())

    task := func(ctx context.Context) error {
        fmt.Println("Processing task...")
        return nil
    }

    pool.Submit(context.Background(), task)

    // 🛡️ Rate Limiting: Control request rates
    limiter := ratelimit.NewTokenBucket(10, time.Second) // 10 req/sec

    if limiter.Allow() {
        fmt.Println("Request allowed")
    }

    // 🔒 Semaphore: Manage resource access
    sem := semaphore.New(3) // Allow 3 concurrent operations
    defer sem.Close()

    sem.Acquire(context.Background(), 1)
    defer sem.Release(1)

    fmt.Println("Critical section")
}
```

## Why Choose Ion?

### 🏆 Battle-Tested Performance

Ion has been proven in production environments handling **millions of concurrent operations** with exceptional reliability and performance. Our benchmarks show 10M+ operations/sec with sub-50ns latency.

### 🔧 Zero Dependencies, Maximum Compatibility

Built exclusively with Go's standard library, Ion ensures maximum compatibility across Go versions while maintaining a minimal security footprint.

### ⚡ Context-First Design

Every Ion component natively supports Go's context patterns for proper cancellation, timeouts, and graceful shutdowns - essential for cloud-native applications.

### 🎯 Developer Experience Focused

Clean, intuitive APIs with comprehensive documentation and real-world examples. Get up and running in minutes, not hours.

## Core Components

### 🚀 Rate Limiting

Choose between **Token Bucket** and **Leaky Bucket** algorithms for precise request control. Perfect for API rate limiting, request throttling, and burst management in high-traffic applications.

### 🔒 Semaphores

**Weighted Semaphores** provide fine-grained resource management. Essential for connection pooling, resource allocation, and controlling concurrent operations across your application.

### ⚡ Worker Pools

High-performance **Worker Pools** for efficient task processing. Automatically manages worker lifecycle, optimizes CPU usage, and provides graceful shutdown capabilities.

---

## Performance Benchmarks

Ion delivers exceptional performance for production workloads:

| Component        | Operations/Second   | Latency | Use Case                      |
| ---------------- | ------------------- | ------- | ----------------------------- |
| **Rate Limiter** | 10M+ ops/sec        | <50ns   | API throttling, burst control |
| **Semaphore**    | 5M+ acquire/release | <100ns  | Resource management           |
| **Worker Pool**  | 100K+ tasks/sec     | <1μs    | Background processing         |

[View Detailed Benchmarks →]({{ site.baseurl }}/benchmarks/)
