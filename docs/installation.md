# Installation

## Requirements

- Go 1.22 or later
- No external dependencies required

## Install via go get

```bash
go get github.com/kolosys/ion@latest
```

## Install specific version

```bash
go get github.com/kolosys/ion@v0.1.0
```

## Verify installation

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

## Module integration

Add to your `go.mod`:

```bash
go mod init your-project
go get github.com/kolosys/ion@latest
```

## Import packages

```go
import "github.com/kolosys/ion"
```
