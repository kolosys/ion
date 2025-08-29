# semaphore

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
}
```

## API Reference
### Types
- [Fairness](../api-reference/semaphore.md#fairness) - Fairness defines the ordering behavior for semaphore waiters

- [Option](../api-reference/semaphore.md#option) - Option configures semaphore behavior

- [Semaphore](../api-reference/semaphore.md#semaphore) - Semaphore represents a weighted semaphore that controls access to a resource
with a fixed capacit...
- [config](../api-reference/semaphore.md#config) - 
- [waiter](../api-reference/semaphore.md#waiter) - waiter represents a goroutine waiting to acquire permits

- [waiterQueue](../api-reference/semaphore.md#waiterqueue) - waiterQueue manages the queue of waiting goroutines based on fairness mode

- [weightedSemaphore](../api-reference/semaphore.md#weightedsemaphore) - weightedSemaphore implements the Semaphore interface with weighted permits and fairness


## Examples

See [examples](../examples/semaphore/README.md) for detailed usage examples.

## Resources

- [API Reference](../api-reference/semaphore.md) - Complete API documentation
- [Examples](../examples/semaphore/README.md) - Working examples
- [Best Practices](../guides/semaphore-best-practices.md) - Recommended patterns
