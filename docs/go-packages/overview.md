# Go Packages Overview

The `traffic2openapi` module provides several packages for working with HTTP traffic and generating OpenAPI specifications.

## Package Structure

```
pkg/
├── ir/                  # IR types, providers, readers, writers
├── inference/           # Traffic analysis and schema inference
└── openapi/             # OpenAPI spec generation
```

## pkg/ir

The `ir` package provides types and utilities for working with Intermediate Representation data.

### Key Features

- **IR Types**: `IRRecord`, `Request`, `Response`, `Batch`
- **Provider Pattern**: Symmetric read/write via `Provider` interface
- **Built-in Providers**: NDJSON, GzipNDJSON, Storage, Channel
- **LoggingTransport**: Capture traffic from `http.Client`
- **File I/O**: Read/write NDJSON and batch JSON files

```go
import "github.com/grokify/traffic2openapi/pkg/ir"

// Create a record
record := ir.NewRecord(ir.RequestMethodGET, "/users", 200)

// Use a provider for I/O
provider := ir.NDJSON()
writer, _ := provider.NewWriter(ctx, "output.ndjson")
writer.Write(record)
writer.Close()
```

See:

- [Provider Pattern](providers.md) - Detailed provider documentation
- [LoggingTransport](logging-transport.md) - HTTP client traffic capture

## pkg/inference

The `inference` package analyzes IR records to discover API structure.

### Key Features

- **Path Template Inference**: `/users/123` → `/users/{userId}`
- **Schema Inference**: JSON Schema from request/response bodies
- **Format Detection**: email, uuid, date-time, uri, ipv4, ipv6
- **Endpoint Clustering**: Groups requests by method + path

```go
import "github.com/grokify/traffic2openapi/pkg/inference"

engine := inference.NewEngine(inference.DefaultEngineOptions())
engine.ProcessRecords(records)
result := engine.Finalize()
```

See [Inference Engine](inference.md) for details.

## pkg/openapi

The `openapi` package generates OpenAPI 3.0/3.1/3.2 specifications.

### Key Features

- **Version Support**: OpenAPI 3.0.3, 3.1.0, 3.2.0
- **Output Formats**: YAML and JSON
- **Customization**: Title, description, servers, version

```go
import "github.com/grokify/traffic2openapi/pkg/openapi"

options := openapi.DefaultGeneratorOptions()
options.Title = "My API"
options.Version = openapi.Version31

spec := openapi.GenerateFromInference(result, options)
openapi.WriteFile("openapi.yaml", spec)
```

See [OpenAPI Generator](openapi.md) for details.

## Common Patterns

### End-to-End Pipeline

```go
// 1. Capture traffic
provider := ir.NDJSON()
writer, _ := provider.NewWriter(ctx, "traffic.ndjson")
transport := ir.NewLoggingTransport(http.DefaultTransport, writer)
client := &http.Client{Transport: transport}

// ... make HTTP requests with client ...

writer.Close()

// 2. Read and analyze
reader, _ := provider.NewReader(ctx, "traffic.ndjson")
var records []*ir.IRRecord
for {
    record, err := reader.Read()
    if err == io.EOF {
        break
    }
    records = append(records, record)
}

// 3. Infer API structure
engine := inference.NewEngine(inference.DefaultEngineOptions())
engine.ProcessRecords(records)
result := engine.Finalize()

// 4. Generate OpenAPI
spec := openapi.GenerateFromInference(result, openapi.DefaultGeneratorOptions())
openapi.WriteFile("openapi.yaml", spec)
```

### With Cloud Storage

```go
import (
    "github.com/grokify/omnistorage/backend/s3"
    "github.com/grokify/traffic2openapi/pkg/ir"
)

// Setup S3 backend
backend, _ := s3.New(ctx, s3.Config{
    Bucket: "my-bucket",
    Region: "us-east-1",
})
defer backend.Close()

// Use storage provider
provider := ir.Storage(backend)

// Write compressed NDJSON to S3
writer, _ := provider.NewWriter(ctx, "traffic/2024/01/records.ndjson.gz")
```
