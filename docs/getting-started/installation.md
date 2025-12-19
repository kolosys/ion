# Installation

This guide will help you install and set up Ion in your Go project.

## Prerequisites

Before installing Ion, ensure you have:

- **Go 1.21** or later installed ([download](https://go.dev/dl/))
- A Go module initialized in your project (run `go mod init` if needed)
- Access to the GitHub repository (for private repositories)

## Installation Steps

### Step 1: Install the Package

Use `go get` to install Ion:

```bash
go get github.com/kolosys/ion@latest
```

This will download the package and add it to your `go.mod` file.

### Step 2: Import in Your Code

Ion is organized into separate packages. Import only the packages you need:

```go
import (
    "github.com/kolosys/ion/circuit"     // Circuit breakers
    "github.com/kolosys/ion/ratelimit"   // Rate limiting
    "github.com/kolosys/ion/semaphore"  // Semaphores
    "github.com/kolosys/ion/workerpool"  // Worker pools
    "github.com/kolosys/ion/observe"     // Observability
)
```

### Step 3: Verify Installation

Create a simple test file to verify the installation:

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kolosys/ion/circuit"
)

func main() {
    // Create a simple circuit breaker
    cb := circuit.New("test-service",
        circuit.WithFailureThreshold(5),
        circuit.WithRecoveryTimeout(30*time.Second),
    )

    // Execute a simple operation
    ctx := context.Background()
    result, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
        return "success", nil
    })

    if err != nil {
        fmt.Printf("Error: %v\n", err)
    } else {
        fmt.Printf("Result: %v\n", result)
        fmt.Println("Ion installed successfully!")
    }
}
```

Run the test:

```bash
go run main.go
```

You should see: `Result: success` and `Ion installed successfully!`

## Package-Specific Installation

### Circuit Breaker

```go
import "github.com/kolosys/ion/circuit"

cb := circuit.New("payment-service",
    circuit.WithFailureThreshold(5),
    circuit.WithRecoveryTimeout(30*time.Second),
)
```

### Rate Limiting

```go
import "github.com/kolosys/ion/ratelimit"

limiter := ratelimit.NewTokenBucket(
    ratelimit.PerSecond(10), // 10 requests per second
    20,                       // burst capacity of 20
)
```

### Semaphore

```go
import "github.com/kolosys/ion/semaphore"

sem := semaphore.NewWeighted(10, // capacity of 10
    semaphore.WithName("db-pool"),
    semaphore.WithFairness(semaphore.FIFO),
)
```

### Worker Pool

```go
import "github.com/kolosys/ion/workerpool"

pool := workerpool.New(4, 20, // 4 workers, queue size 20
    workerpool.WithName("image-processor"),
)
```

### Observability

```go
import "github.com/kolosys/ion/observe"

obs := observe.New().
    WithLogger(myLogger).
    WithMetrics(myMetrics).
    WithTracer(myTracer)
```

## Updating the Package

To update to the latest version:

```bash
go get -u github.com/kolosys/ion
```

To update to a specific version:

```bash
go get github.com/kolosys/ion@latest
```

Check available versions on the [GitHub releases page](https://github.com/kolosys/ion/releases).

## Installing a Specific Version

To install a specific version of the package:

```bash
go get github.com/kolosys/ion@latest
```

## Development Setup

If you want to contribute or modify the library:

### Clone the Repository

```bash
git clone https://github.com/kolosys/ion.git
cd ion
```

### Install Dependencies

Ion has zero external dependencies, so no additional packages are required:

```bash
go mod download
```

### Run Tests

```bash
go test ./...
```

Run tests with race detection:

```bash
go test -race ./...
```

### Run Benchmarks

```bash
go test -bench=. -benchmem ./...
```

## Troubleshooting

### Module Not Found

If you encounter a "module not found" error:

1. Ensure your `GOPATH` is set correctly
2. Check that you have network access to GitHub
3. Verify Go version: `go version` (requires 1.21+)
4. Try running `go clean -modcache` and reinstall:

```bash
go clean -modcache
go get github.com/kolosys/ion@latest
```

### Private Repository Access

For private repositories, configure Git to use SSH or a personal access token:

```bash
git config --global url."git@github.com:".insteadOf "https://github.com/"
```

Or set up GOPRIVATE:

```bash
export GOPRIVATE=github.com/kolosys/ion
```

### Version Conflicts

If you encounter version conflicts with other packages:

1. Check your `go.mod` file for version constraints
2. Use `go mod tidy` to resolve dependencies:

```bash
go mod tidy
```

3. If issues persist, check for incompatible versions:

```bash
go list -m -versions github.com/kolosys/ion
```

## IDE Integration

### VS Code

Ion works seamlessly with VS Code's Go extension. Ensure you have:

1. The [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.Go) installed
2. `gopls` installed and up to date:

```bash
go install golang.org/x/tools/gopls@latest
```

### GoLand

Ion is fully supported in GoLand. The IDE will automatically:

- Provide code completion
- Show inline documentation
- Highlight errors and warnings
- Support refactoring

### Vim/Neovim

For Vim/Neovim users, ensure you have:

- `gopls` installed for LSP support
- A compatible LSP client (e.g., `vim-lsp`, `coc.nvim`)

## Next Steps

Now that you have Ion installed, check out the [Quick Start Guide](quick-start.md) to learn how to use it in your projects.

## Additional Resources

- [Quick Start Guide](quick-start.md) - Get started with practical examples
- [API Reference](../api-reference/) - Complete API documentation
- [Examples](../examples/) - Working code examples
- [Core Concepts](../core-concepts/) - Deep dive into each package
