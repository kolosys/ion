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
- [Option](api-reference.md#option) - Option configures pool behavior

- [Pool](api-reference.md#pool) - Pool represents a bounded worker pool that executes tasks with controlled
concurrency and queue m...
- [PoolMetrics](api-reference.md#poolmetrics) - PoolMetrics holds runtime metrics for the pool

- [Task](api-reference.md#task) - Task represents a unit of work to be executed by the worker pool.
Tasks receive a context that wi...
- [config](api-reference.md#config) - 
- [taskSubmission](api-reference.md#tasksubmission) - taskSubmission wraps a task with its submission context


## Examples

See [examples](examples.md) for detailed usage examples.
