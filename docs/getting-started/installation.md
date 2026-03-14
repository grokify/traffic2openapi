# Installation

## CLI Tool

Install the CLI tool using Go:

```bash
go install github.com/grokify/traffic2openapi/cmd/traffic2openapi@latest
```

Verify the installation:

```bash
traffic2openapi --help
```

## Go Package

Add the package to your Go project:

```bash
go get github.com/grokify/traffic2openapi
```

Import the packages you need:

```go
import (
    "github.com/grokify/traffic2openapi/pkg/ir"
    "github.com/grokify/traffic2openapi/pkg/inference"
    "github.com/grokify/traffic2openapi/pkg/openapi"
)
```

## Requirements

- Go 1.21 or later
- For CLI: no additional dependencies
- For Go package: dependencies are managed via go.mod

## Optional Dependencies

For cloud storage integration:

```bash
go get github.com/grokify/omnistorage
```

This enables storing and reading IR records from various cloud storage backends (S3, GCS, Azure Blob, etc.).
