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

## Next Steps

- [API Reference](api-reference/README.md) - Complete API documentation
- [Examples](examples/README.md) - Working examples and tutorials
- [Guides](guides/README.md) - In-depth guides and best practices
- [GitHub Repository](https://github.com/kolosys/ion)
