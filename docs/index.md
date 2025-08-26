---
layout: home
title: Ion
permalink: /
---

Ion provides a comprehensive suite of concurrency primitives designed to help you build robust, scalable Go applications with confidence.

## Quick Start

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
    // Worker Pool Example
    pool := workerpool.New(4, 10)
    defer pool.Close(context.Background())

    task := func(ctx context.Context) error {
        fmt.Println("Processing task...")
        return nil
    }

    pool.Submit(context.Background(), task)

    // Rate Limiting Example
    limiter := ratelimit.NewTokenBucket(10, time.Second)

    if limiter.Allow() {
        fmt.Println("Request allowed")
    }

    // Semaphore Example
    sem := semaphore.New(3) // Allow 3 concurrent operations
    defer sem.Close()

    sem.Acquire(context.Background(), 1)
    defer sem.Release(1)

    fmt.Println("Critical section")
}
```

## Why Ion?

**Production Ready** - Ion has been battle-tested in high-traffic production environments, handling millions of concurrent operations with reliability and performance.

**Zero Dependencies** - Built using only Go's standard library, Ion has no external dependencies, ensuring maximum compatibility and minimal attack surface.

**Context-Aware** - All Ion components respect Go's context patterns, enabling proper cancellation, timeouts, and graceful shutdowns.

**Developer Friendly** - Clear APIs, comprehensive documentation, and extensive examples make Ion easy to integrate into your projects.

## Components Overview

### Rate Limiting

Control request rates with **Token Bucket** and **Leaky Bucket** algorithms. Perfect for API rate limiting, request throttling, and burst control.

### Semaphores

Manage resource access with **Weighted Semaphores**. Ideal for connection pooling, resource allocation, and concurrent operation control.

### Worker Pools

Process tasks efficiently with **Worker Pools**. Optimize CPU usage, control concurrency, and handle graceful shutdowns.

## Benchmarks

Ion components are optimized for performance:

- **Rate Limiter**: 10M+ operations/sec with <50ns latency
- **Semaphore**: 5M+ acquire/release operations/sec
- **Worker Pool**: Handles 100K+ tasks/sec with minimal overhead

See detailed [benchmarks]({{ site.baseurl }}/benchmarks/) for performance comparisons.
