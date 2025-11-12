# Overview

[![Go Reference](https://pkg.go.dev/badge/github.com/kolosys/ion.svg)](https://pkg.go.dev/github.com/kolosys/ion)

## About ion

This documentation provides comprehensive guidance for using ion, a Go library designed to help you build better software.

## Project Information

- **Repository**: [https://github.com/kolosys/ion](https://github.com/kolosys/ion)
- **Import Path**: `github.com/kolosys/ion`
- **License**: MIT
- **Version**: latest

## What You'll Find Here

This documentation is organized into several sections to help you find what you need:

- **[Getting Started](../getting-started/)** - Installation instructions and quick start guides
- **[Core Concepts](../core-concepts/)** - Fundamental concepts and architecture details
- **[Advanced Topics](../advanced/)** - Performance tuning and advanced usage patterns
- **[Reference](../reference/)** - Complete API reference and examples

## Project Features

ion provides:
- **circuit** - Package circuit provides circuit breaker functionality for resilient microservices.
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

- **main** - Package main demonstrates circuit breaker usage in real-world scenarios.

- **main** - Package main demonstrates basic usage of the ion ratelimit package.

Package main demonstrates advanced multi-tier rate limiting with the ion ratelimit package.

- **main** - Package main demonstrates basic usage of the ion semaphore.

- **main** - Package main demonstrates basic usage of the ion workerpool.

- **observe** - Package observe provides observability interfaces and implementations
for logging, metrics, and tracing across all Ion components.

- **ratelimit** - Package ratelimit provides local process rate limiters for controlling function and I/O throughput.
It includes token bucket and leaky bucket implementations with configurable options.

- **semaphore** - Package semaphore provides a weighted semaphore with configurable fairness modes.

- **workerpool** - Package workerpool provides a bounded worker pool with context-aware submission,
graceful shutdown, and observability hooks.


## Quick Links

- [Installation Guide](installation.md)
- [Quick Start Guide](quick-start.md)
- [API Reference](../reference/api-reference/README.md)
- [Examples](../reference/examples/README.md)

## Community & Support

- **GitHub Issues**: [https://github.com/kolosys/ion/issues](https://github.com/kolosys/ion/issues)
- **Discussions**: [https://github.com/kolosys/ion/discussions](https://github.com/kolosys/ion/discussions)
- **Repository Owner**: [kolosys](https://github.com/kolosys)

## Getting Help

If you encounter any issues or have questions:

1. Check the [API Reference](../reference/api-reference/README.md) for detailed documentation
2. Browse the [Examples](../reference/examples/README.md) for common use cases
3. Search existing [GitHub Issues](https://github.com/kolosys/ion/issues)
4. Open a new issue if you've found a bug or have a feature request

## Next Steps

Ready to get started? Head over to the [Installation Guide](installation.md) to begin using ion.

