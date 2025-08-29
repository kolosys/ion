# Getting Started with workerpool

Package workerpool provides a bounded worker pool with context-aware submission,
graceful shutdown, and observability hooks.


## Installation

```bash
go get github.com/kolosys/ion/workerpool
```

## Quick Start

```go
package main

import "github.com/kolosys/ion/workerpool"

func main() {
    // Your code here
    fmt.Println("Hello from workerpool!")
}
```

## Basic Usage
### Types
- **Option** - Option configures pool behavior
- **Pool** - Pool represents a bounded worker pool that executes tasks with controlled
- **PoolMetrics** - PoolMetrics holds runtime metrics for the pool
- **Task** - Task represents a unit of work to be executed by the worker pool.
- **config** - 
- **taskSubmission** - taskSubmission wraps a task with its submission context

## Next Steps

- [Package Overview](../packages/workerpool.md) - Complete package information
- [API Reference](../api-reference/workerpool.md) - Detailed API documentation
- [Examples](../examples/workerpool/README.md) - Working examples and tutorials  
- [Best Practices](../guides/workerpool/best-practices.md) - Recommended usage patterns
- [Common Patterns](../guides/workerpool/patterns.md) - Common implementation patterns
