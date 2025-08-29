# Getting Started with semaphore

Package semaphore provides a weighted semaphore with configurable fairness modes.


## Installation

```bash
go get github.com/kolosys/ion/semaphore
```

## Quick Start

```go
package main

import "github.com/kolosys/ion/semaphore"

func main() {
    // Your code here
    fmt.Println("Hello from semaphore!")
}
```

## Basic Usage
### Types
- **Fairness** - Fairness defines the ordering behavior for semaphore waiters
- **Option** - Option configures semaphore behavior
- **Semaphore** - Semaphore represents a weighted semaphore that controls access to a resource
- **config** - 
- **waiter** - waiter represents a goroutine waiting to acquire permits
- **waiterQueue** - waiterQueue manages the queue of waiting goroutines based on fairness mode
- **weightedSemaphore** - weightedSemaphore implements the Semaphore interface with weighted permits and fairness

## Next Steps

- [Package Overview](../packages/semaphore.md) - Complete package information
- [API Reference](../api-reference/semaphore.md) - Detailed API documentation
- [Examples](../examples/semaphore/README.md) - Working examples and tutorials  
- [Best Practices](../guides/semaphore/best-practices.md) - Recommended usage patterns
- [Common Patterns](../guides/semaphore/patterns.md) - Common implementation patterns
