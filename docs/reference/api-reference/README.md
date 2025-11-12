# API Reference

Complete API documentation for ion.

## Overview

This section contains detailed API documentation for all packages. For package overviews and getting started guides, see the [Packages](../packages/README.md) section.

## Package APIs

### [circuit](circuit.md)

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


**[→ Full API Documentation](circuit.md)**

Key APIs:

- Types and interfaces
- Functions and methods
- Constants and variables
- Detailed usage examples

### [main](main.md)

Package main demonstrates circuit breaker usage in real-world scenarios.


**[→ Full API Documentation](main.md)**

Key APIs:

- Types and interfaces
- Functions and methods
- Constants and variables
- Detailed usage examples

### [main](main.md)

Package main demonstrates basic usage of the ion ratelimit package.

Package main demonstrates advanced multi-tier rate limiting with the ion ratelimit package.


**[→ Full API Documentation](main.md)**

Key APIs:

- Types and interfaces
- Functions and methods
- Constants and variables
- Detailed usage examples

### [main](main.md)

Package main demonstrates basic usage of the ion semaphore.


**[→ Full API Documentation](main.md)**

Key APIs:

- Types and interfaces
- Functions and methods
- Constants and variables
- Detailed usage examples

### [main](main.md)

Package main demonstrates basic usage of the ion workerpool.


**[→ Full API Documentation](main.md)**

Key APIs:

- Types and interfaces
- Functions and methods
- Constants and variables
- Detailed usage examples

### [observe](observe.md)

Package observe provides observability interfaces and implementations
for logging, metrics, and tracing across all Ion components.


**[→ Full API Documentation](observe.md)**

Key APIs:

- Types and interfaces
- Functions and methods
- Constants and variables
- Detailed usage examples

### [ratelimit](ratelimit.md)

Package ratelimit provides local process rate limiters for controlling function and I/O throughput.
It includes token bucket and leaky bucket implementations with configurable options.


**[→ Full API Documentation](ratelimit.md)**

Key APIs:

- Types and interfaces
- Functions and methods
- Constants and variables
- Detailed usage examples

### [semaphore](semaphore.md)

Package semaphore provides a weighted semaphore with configurable fairness modes.


**[→ Full API Documentation](semaphore.md)**

Key APIs:

- Types and interfaces
- Functions and methods
- Constants and variables
- Detailed usage examples

### [workerpool](workerpool.md)

Package workerpool provides a bounded worker pool with context-aware submission,
graceful shutdown, and observability hooks.


**[→ Full API Documentation](workerpool.md)**

Key APIs:

- Types and interfaces
- Functions and methods
- Constants and variables
- Detailed usage examples

## Navigation

- **[Packages](../packages/README.md)** - Package overviews and installation
- **[Examples](../examples/README.md)** - Working code examples
- **[Guides](../guides/README.md)** - Best practices and patterns

## External References

- [pkg.go.dev Documentation](https://pkg.go.dev/github.com/kolosys/ion) - Go module documentation
- [GitHub Repository](https://github.com/kolosys/ion) - Source code and issues
