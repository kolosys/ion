# Getting Started

This guide will help you get up and running quickly with ion.

## Installation

### Requirements

- Go 1.22 or later
- No external dependencies required

### Install via go get

```bash
go get github.com/kolosys/ion@latest
```

### Install specific version

```bash
go get github.com/kolosys/ion@v0.1.0
```

### Verify installation

Create a simple test file:

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

Run it:

```bash
go run main.go
```

### Module integration

Add to your `go.mod`:

```bash
go mod init your-project
go get github.com/kolosys/ion@latest
```

## Quick Start

Here's a simple example to get you started:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/kolosys/ion"
)

func main() {
    // Your code here
    fmt.Println("Hello from ion!")
}
```

## Available Packages

ion provides the following packages:
### [workerpool](packages/workerpool.md)

Bounded worker pool with context-aware submission and graceful shutdown

**Quick Links:**
- [Package Overview](packages/workerpool.md) - Installation and getting started
- [API Reference](api-reference/workerpool.md) - Complete API documentation  
- [Examples](examples/workerpool/README.md) - Working examples
- [Best Practices](guides/workerpool-best-practices.md) - Recommended patterns
### [ratelimit](packages/ratelimit.md)

Token bucket and leaky bucket rate limiters with configurable options

**Quick Links:**
- [Package Overview](packages/ratelimit.md) - Installation and getting started
- [API Reference](api-reference/ratelimit.md) - Complete API documentation  
- [Examples](examples/ratelimit/README.md) - Working examples
- [Best Practices](guides/ratelimit-best-practices.md) - Recommended patterns
### [semaphore](packages/semaphore.md)

Weighted semaphore with configurable fairness modes

**Quick Links:**
- [Package Overview](packages/semaphore.md) - Installation and getting started
- [API Reference](api-reference/semaphore.md) - Complete API documentation  
- [Examples](examples/semaphore/README.md) - Working examples
- [Best Practices](guides/semaphore-best-practices.md) - Recommended patterns

## Next Steps

- [API Reference](api-reference/README.md) - Complete API documentation
- [Examples](examples/README.md) - Working examples and tutorials  
- [Guides](guides/README.md) - In-depth guides and best practices
- [GitHub Repository](https://github.com/kolosys/ion)
