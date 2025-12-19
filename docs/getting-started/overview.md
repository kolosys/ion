# Overview

[![Go Reference](https://pkg.go.dev/badge/github.com/kolosys/ion.svg)](https://pkg.go.dev/github.com/kolosys/ion)

## About Ion

Ion is a production-grade concurrency and resilience toolkit for Go, designed to help you build reliable, high-performance applications. Whether you're building microservices, API gateways, data processing pipelines, or distributed systems, Ion provides the primitives you need with enterprise-grade reliability, observability, and performance.

**Zero dependencies. Context-first. Production-ready.**

## Project Information

- **Repository**: [https://github.com/kolosys/ion](https://github.com/kolosys/ion)
- **Import Path**: `github.com/kolosys/ion`
- **License**: MIT
- **Version**: latest

## Core Philosophy

Ion is built on three fundamental principles:

1. **Context-First**: All operations respect `context.Context` for cancellation, timeouts, and request-scoped values
2. **Zero-Panic**: Library code returns errors, never panics, ensuring predictable behavior
3. **Observable**: Built-in hooks for logging, metrics, and tracing with zero overhead when not configured

## What You'll Find Here

This documentation is organized into several sections to help you find what you need:

- **[Getting Started](./)** - Installation instructions and quick start guides
- **[Core Concepts](../core-concepts/)** - Fundamental concepts and architecture details
- **[Advanced Topics](../advanced/)** - Performance tuning and advanced usage patterns
- **[API Reference](../api-reference/)** - Complete API reference documentation
- **[Examples](../examples/)** - Working code examples and tutorials

## Package Overview

Ion provides five core packages, each designed for specific concurrency and resilience patterns:

### Circuit Breaker (`circuit`)

Protect your services from cascading failures with automatic failure detection and recovery.

**Use when:**

- Calling external services or APIs
- Accessing databases or caches
- Performing operations that can fail under load
- Building resilient microservices

**Real-world scenarios:**

- Payment processing services that need fast-fail behavior
- HTTP clients calling third-party APIs
- Database connections that may timeout
- Service mesh integration for distributed systems

### Rate Limiting (`ratelimit`)

Control the rate at which operations execute with token bucket, leaky bucket, and multi-tier limiters.

**Use when:**

- Protecting APIs from being overwhelmed
- Respecting external API rate limits
- Controlling resource consumption
- Building API gateways

**Real-world scenarios:**

- API gateway rate limiting per user, route, or resource
- Client libraries respecting external API limits
- Background job processing with controlled throughput
- Multi-tenant systems with per-tenant limits

### Semaphore (`semaphore`)

Control access to shared resources with weighted semaphores and configurable fairness.

**Use when:**

- Limiting concurrent database connections
- Controlling memory or CPU usage
- Managing access to external resources
- Implementing resource pools

**Real-world scenarios:**

- Database connection pool management
- Limiting concurrent file operations
- Controlling memory-intensive operations
- CPU-bound task scheduling

### Worker Pool (`workerpool`)

Execute tasks with bounded concurrency, graceful shutdown, and queue management.

**Use when:**

- Processing background jobs
- Handling request queues
- Managing concurrent operations
- Building task processors

**Real-world scenarios:**

- Image processing pipelines
- Email sending services
- Data transformation jobs
- Event processing systems

### Observability (`observe`)

Pluggable interfaces for logging, metrics, and tracing across all Ion components.

**Use when:**

- Integrating with existing observability stacks
- Adding structured logging
- Collecting performance metrics
- Implementing distributed tracing

**Real-world scenarios:**

- Integration with Prometheus, Grafana, or Datadog
- Structured logging with `slog` or `zap`
- OpenTelemetry tracing integration
- Custom metrics collection

## Quick Links

- [Installation Guide](installation.md) - Get started with Ion
- [Quick Start Guide](quick-start.md) - Your first Ion program
- [API Reference](../api-reference/) - Complete API documentation
- [Examples](../examples/) - Working code examples

## Design Principles

### Performance

Ion is optimized for performance with:

- **<200ns hot paths** for critical operations
- **Zero allocations** in steady state
- **Lock-free algorithms** where possible
- **Minimal overhead** when not configured

### Reliability

Ion ensures reliable operation through:

- **Deterministic behavior** under load
- **Graceful degradation** when resources are exhausted
- **Comprehensive error handling** with context
- **Thread-safe** public APIs

### Developer Experience

Ion prioritizes developer experience with:

- **Intuitive APIs** that are easy to use correctly
- **Functional options** for flexible configuration
- **Rich error messages** with context
- **Comprehensive documentation** and examples

## Use Cases

Ion powers production systems across various domains:

- **🌐 API Gateways**: Multi-tier rate limiting, circuit breakers, request routing
- **📊 Data Pipelines**: Bounded processing, backpressure handling, error recovery
- **⏰ Background Jobs**: Controlled concurrency, graceful shutdown, resource management
- **🔄 Microservices**: Service protection, cascading failure prevention, observability
- **🏦 Financial Systems**: High-frequency trading, payment processing, risk management
- **🎮 Gaming Platforms**: Matchmaking, leaderboards, real-time event processing

## Community & Support

- **GitHub Issues**: [https://github.com/kolosys/ion/issues](https://github.com/kolosys/ion/issues)
- **Discussions**: [https://github.com/kolosys/ion/discussions](https://github.com/kolosys/ion/discussions)
- **Repository Owner**: [kolosys](https://github.com/kolosys)

## Getting Help

If you encounter any issues or have questions:

1. Check the [API Reference](../api-reference/) for detailed documentation
2. Browse the [Examples](../examples/) for common use cases
3. Search existing [GitHub Issues](https://github.com/kolosys/ion/issues)
4. Open a new issue if you've found a bug or have a feature request

## Next Steps

Ready to get started? Head over to the [Installation Guide](installation.md) to begin using Ion in your projects.
