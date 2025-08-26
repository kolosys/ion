# Ion API Reference

Complete API documentation for all Ion packages.

## Overview

Ion provides the following packages:

- **[workerpool](../workerpool/)** - Bounded worker pool with context-aware submission and graceful shutdown
- **[ratelimit](../ratelimit/)** - Token bucket and leaky bucket rate limiters with configurable options  
- **[semaphore](../semaphore/)** - Weighted semaphore with configurable fairness modes

## Quick Navigation

### workerpool

- [Full workerpool API Reference](../workerpool/api-reference.md)
- [Examples](../workerpool/examples.md)
- [Overview](../workerpool/README.md)

### ratelimit

- [Full ratelimit API Reference](../ratelimit/api-reference.md)
- [Examples](../ratelimit/examples.md)
- [Overview](../ratelimit/README.md)

### semaphore

- [Full semaphore API Reference](../semaphore/api-reference.md)
- [Examples](../semaphore/examples.md)
- [Overview](../semaphore/README.md)


## Package Summaries

### workerpool

[Full Documentation](../workerpool/api-reference.md)

#### Key Functions

- **TestMetrics**
- **TestNew**
- **TestPoolLifecycle**
- **TestSubmit**
- **TestTaskPanicRecovery**

#### Key Types

- **Option**
- **Pool**

### ratelimit

[Full Documentation](../ratelimit/api-reference.md)

#### Key Functions

- **BenchmarkComparison_AllowN**
- **BenchmarkHighContention**
- **BenchmarkLeakyBucketAlloc**
- **BenchmarkLeakyBucketAllowN**
- **BenchmarkLeakyBucketAllowN_Uncontended**

### semaphore

[Full Documentation](../semaphore/api-reference.md)

#### Key Functions

- **TestAcquire**
- **TestConcurrency**
- **TestFairness**
- **TestNewWeighted**
- **TestRelease**

#### Key Types

- **Fairness**
- **Option**
- **Semaphore**
- **config**
- **waiter**

## External References

- [pkg.go.dev Documentation](https://pkg.go.dev/github.com/kolosys/ion)
- [GitHub Repository](https://github.com/kolosys/ion)
- [Examples Directory](https://github.com/kolosys/ion/tree/main/examples)

