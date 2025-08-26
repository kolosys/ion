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
- [Fairness](api-reference.md#fairness) - Fairness defines the ordering behavior for semaphore waiters

- [Option](api-reference.md#option) - Option configures semaphore behavior

- [Semaphore](api-reference.md#semaphore) - Semaphore represents a weighted semaphore that controls access to a resource
with a fixed capacit...
- [config](api-reference.md#config) - 
- [waiter](api-reference.md#waiter) - waiter represents a goroutine waiting to acquire permits

- [waiterQueue](api-reference.md#waiterqueue) - waiterQueue manages the queue of waiting goroutines based on fairness mode

- [weightedSemaphore](api-reference.md#weightedsemaphore) - weightedSemaphore implements the Semaphore interface with weighted permits and fairness


## Examples

See [examples](examples.md) for detailed usage examples.
