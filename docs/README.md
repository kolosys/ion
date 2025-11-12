# ion Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/kolosys/ion.svg)](https://pkg.go.dev/github.com/kolosys/ion)

## Quick Navigation

### 🚀 [Getting Started](getting-started/README.md)

Everything you need to get up and running with ion.

### 📚 [API Reference](api-reference/README.md)

Complete API documentation for all packages.

### 📖 [Examples](examples/README.md)

Working examples and tutorials.

### 📘 [Guides](guides/README.md)

In-depth guides and best practices.

## Package Overview

### circuit

Package circuit provides circuit breaker functionality for resilient microservices.
Circuit breakers prevent cascading failures by temporarily blocking requests to failing services,
allowing them time to recover while providing fast-fail behavior to callers.

The circuit breaker implements a three-state machine:
- Closed: Normal operation, requests pass through
- Open: Circuit is tripped, requests fail fast
- Half-Open: Testing recovery, limited requests allowed

Usage:

	cb := circuit.New("payment-service",
		circuit.WithFailureThreshold(5),
		circuit.WithRecoveryTimeout(30*time.Second),
	)

	result, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
		return paymentService.ProcessPayment(ctx, payment)
	})

The circuit breaker integrates with ION's observability system and supports
context cancellation, timeouts, and comprehensive metrics collection.


- [Getting Started](getting-started/circuit.md)
- [API Reference](api-reference/circuit.md)
- [Examples](examples/README.md)
- [Best Practices](guides/circuit/best-practices.md)

### main

Package main demonstrates circuit breaker usage in real-world scenarios.


- [Getting Started](getting-started/main.md)
- [API Reference](api-reference/main.md)
- [Examples](examples/README.md)
- [Best Practices](guides/main/best-practices.md)

### main

Package main demonstrates basic usage of the ion ratelimit package.

Package main demonstrates advanced multi-tier rate limiting with the ion ratelimit package.


- [Getting Started](getting-started/main.md)
- [API Reference](api-reference/main.md)
- [Examples](examples/README.md)
- [Best Practices](guides/main/best-practices.md)

### main

Package main demonstrates basic usage of the ion semaphore.


- [Getting Started](getting-started/main.md)
- [API Reference](api-reference/main.md)
- [Examples](examples/README.md)
- [Best Practices](guides/main/best-practices.md)

### main

Package main demonstrates basic usage of the ion workerpool.


- [Getting Started](getting-started/main.md)
- [API Reference](api-reference/main.md)
- [Examples](examples/README.md)
- [Best Practices](guides/main/best-practices.md)

### observe

Package observe provides observability interfaces and implementations
for logging, metrics, and tracing across all Ion components.


- [Getting Started](getting-started/observe.md)
- [API Reference](api-reference/observe.md)
- [Examples](examples/README.md)
- [Best Practices](guides/observe/best-practices.md)

### ratelimit

Package ratelimit provides local process rate limiters for controlling function and I/O throughput.
It includes token bucket and leaky bucket implementations with configurable options.


- [Getting Started](getting-started/ratelimit.md)
- [API Reference](api-reference/ratelimit.md)
- [Examples](examples/README.md)
- [Best Practices](guides/ratelimit/best-practices.md)

### semaphore

Package semaphore provides a weighted semaphore with configurable fairness modes.


- [Getting Started](getting-started/semaphore.md)
- [API Reference](api-reference/semaphore.md)
- [Examples](examples/README.md)
- [Best Practices](guides/semaphore/best-practices.md)

### workerpool

Package workerpool provides a bounded worker pool with context-aware submission,
graceful shutdown, and observability hooks.


- [Getting Started](getting-started/workerpool.md)
- [API Reference](api-reference/workerpool.md)
- [Examples](examples/README.md)
- [Best Practices](guides/workerpool/best-practices.md)

## External Resources

- [GitHub Repository](https://github.com/kolosys/ion)
- [pkg.go.dev Documentation](https://pkg.go.dev/github.com/kolosys/ion)
- [Issues & Support](https://github.com/kolosys/ion/issues)

## Contributing

See our [Contributing Guide](guides/contributing.md) to get started.
