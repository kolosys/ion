# Installation

This guide will help you install and set up ion in your Go project.

## Prerequisites

Before installing ion, ensure you have:

- **Go ** or later installed
- A Go module initialized in your project (run `go mod init` if needed)
- Access to the GitHub repository (for private repositories)

## Installation Steps

### Step 1: Install the Package

Use `go get` to install ion:

```bash
go get github.com/kolosys/ion
```

This will download the package and add it to your `go.mod` file.

### Step 2: Import in Your Code

Import the package in your Go source files:

```go
import "github.com/kolosys/ion"
```

### Multiple Packages

ion includes several packages. Import the ones you need:

```go
// Package circuit provides circuit breaker functionality for resilient microservices.
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

import "github.com/kolosys/ion/circuit"
```

```go
// Package observe provides observability interfaces and implementations
for logging, metrics, and tracing across all Ion components.

import "github.com/kolosys/ion/observe"
```

```go
// Package ratelimit provides local process rate limiters for controlling function and I/O throughput.
It includes token bucket and leaky bucket implementations with configurable options.

import "github.com/kolosys/ion/ratelimit"
```

```go
// Package semaphore provides a weighted semaphore with configurable fairness modes.

import "github.com/kolosys/ion/semaphore"
```

```go
// Package workerpool provides a bounded worker pool with context-aware submission,
graceful shutdown, and observability hooks.

import "github.com/kolosys/ion/workerpool"
```

### Step 3: Verify Installation

Create a simple test file to verify the installation:

```go
package main

import (
    "fmt"
    "github.com/kolosys/ion"
)

func main() {
    fmt.Println("ion installed successfully!")
}
```

Run the test:

```bash
go run main.go
```

## Updating the Package

To update to the latest version:

```bash
go get -u github.com/kolosys/ion
```

To update to a specific version:

```bash
go get github.com/kolosys/ion@v1.2.3
```

## Installing a Specific Version

To install a specific version of the package:

```bash
go get github.com/kolosys/ion@v1.0.0
```

Check available versions on the [GitHub releases page](https://github.com/kolosys/ion/releases).

## Development Setup

If you want to contribute or modify the library:

### Clone the Repository

```bash
git clone https://github.com/kolosys/ion.git
cd ion
```

### Install Dependencies

```bash
go mod download
```

### Run Tests

```bash
go test ./...
```

## Troubleshooting

### Module Not Found

If you encounter a "module not found" error:

1. Ensure your `GOPATH` is set correctly
2. Check that you have network access to GitHub
3. Try running `go clean -modcache` and reinstall

### Private Repository Access

For private repositories, configure Git to use SSH or a personal access token:

```bash
git config --global url."git@github.com:".insteadOf "https://github.com/"
```

Or set up GOPRIVATE:

```bash
export GOPRIVATE=github.com/kolosys/ion
```

## Next Steps

Now that you have ion installed, check out the [Quick Start Guide](quick-start.md) to learn how to use it.

## Additional Resources

- [Quick Start Guide](quick-start.md)
- [API Reference](../reference/api-reference/README.md)
- [Examples](../reference/examples/README.md)

