# Guides

In-depth guides and best practices for ion.

## Getting Started

- [Installation & Setup](../getting-started.md)
- [Quick Start Guide](quick-start.md)
- [Basic Concepts](concepts.md)

## Best Practices

- [Performance Optimization](performance.md)
- [Error Handling](error-handling.md)
- [Testing Strategies](testing.md)
- [Production Deployment](deployment.md)

## Advanced Topics

- [Architecture Overview](architecture.md)
- [Extending ion](extending.md)
- [Integration Patterns](integration.md)
- [Troubleshooting](troubleshooting.md)

## Package-Specific Guides

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


- [circuit Best Practices](circuit/best-practices.md) - Recommended patterns and usage

### main

Package main demonstrates circuit breaker usage in real-world scenarios.


- [main Best Practices](main/best-practices.md) - Recommended patterns and usage

### main

Package main demonstrates basic usage of the ion ratelimit package.

Package main demonstrates advanced multi-tier rate limiting with the ion ratelimit package.


- [main Best Practices](main/best-practices.md) - Recommended patterns and usage

### main

Package main demonstrates basic usage of the ion semaphore.


- [main Best Practices](main/best-practices.md) - Recommended patterns and usage

### main

Package main demonstrates basic usage of the ion workerpool.


- [main Best Practices](main/best-practices.md) - Recommended patterns and usage

### observe

Package observe provides observability interfaces and implementations
for logging, metrics, and tracing across all Ion components.


- [observe Best Practices](observe/best-practices.md) - Recommended patterns and usage

### ratelimit

Package ratelimit provides local process rate limiters for controlling function and I/O throughput.
It includes token bucket and leaky bucket implementations with configurable options.


- [ratelimit Best Practices](ratelimit/best-practices.md) - Recommended patterns and usage

### semaphore

Package semaphore provides a weighted semaphore with configurable fairness modes.


- [semaphore Best Practices](semaphore/best-practices.md) - Recommended patterns and usage

### workerpool

Package workerpool provides a bounded worker pool with context-aware submission,
graceful shutdown, and observability hooks.


- [workerpool Best Practices](workerpool/best-practices.md) - Recommended patterns and usage

## Community Resources

- [Contributing Guide](contributing.md)
- [Code of Conduct](code-of-conduct.md)
- [Security Policy](security.md)
- [FAQ](faq.md)

## External Resources

- [GitHub Repository](https://github.com/kolosys/ion)
- [Discussions](https://github.com/kolosys/ion/discussions)
- [Issues](https://github.com/kolosys/ion/issues)
