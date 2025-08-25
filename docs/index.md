---
layout: page
title: Ion
permalink: /
---

# Ion

Robust, context-aware concurrency and scheduling primitives for Go applications.

Ion provides a comprehensive suite of concurrency primitives designed to help you build robust, scalable Go applications with confidence.

## Features

- **Rate Limiting**: Token bucket and leaky bucket algorithms for controlling request rates
- **Semaphores**: Weighted semaphores for resource management and access control
- **Worker Pools**: Efficient worker pool implementation for concurrent task processing
- **Context-Aware**: All components respect Go's context cancellation patterns
- **Production Ready**: Battle-tested in production environments
- **Zero Dependencies**: No external dependencies beyond the Go standard library

## Quick Start

```bash
go get github.com/kolosys/ion
```

## Components

- [Rate Limiting]({{ site.baseurl }}/ratelimit/) - Control request rates with token bucket and leaky bucket algorithms
- [Semaphores]({{ site.baseurl }}/semaphore/) - Manage resource access with weighted semaphores
- [Worker Pools]({{ site.baseurl }}/workerpool/) - Process tasks concurrently with efficient worker pools

## Getting Started

Check out our [Getting Started Guide]({{ site.baseurl }}/getting-started/) to begin using Ion in your projects.

## Examples

See the [Examples]({{ site.baseurl }}/examples/) section for practical usage patterns and code samples.

## API Reference

Complete API documentation is available in the [API Reference]({{ site.baseurl }}/api-reference/).
