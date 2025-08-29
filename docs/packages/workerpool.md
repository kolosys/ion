# workerpool

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
}
```

## API Reference
### Types
- [Option](../api-reference/workerpool.md#option) - Option configures pool behavior

- [Pool](../api-reference/workerpool.md#pool) - Pool represents a bounded worker pool that executes tasks with controlled
concurrency and queue m...
- [PoolMetrics](../api-reference/workerpool.md#poolmetrics) - PoolMetrics holds runtime metrics for the pool

- [Task](../api-reference/workerpool.md#task) - Task represents a unit of work to be executed by the worker pool.
Tasks receive a context that wi...
- [config](../api-reference/workerpool.md#config) - 
- [taskSubmission](../api-reference/workerpool.md#tasksubmission) - taskSubmission wraps a task with its submission context


## Examples

See [examples](../examples/workerpool/README.md) for detailed usage examples.

## Resources

- [API Reference](../api-reference/workerpool.md) - Complete API documentation
- [Examples](../examples/workerpool/README.md) - Working examples
- [Best Practices](../guides/workerpool/best-practices.md) - Recommended patterns
