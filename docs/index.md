# Traffic2OpenAPI

Generate OpenAPI 3.0/3.1/3.2 specifications from HTTP traffic logs.

Traffic2OpenAPI captures HTTP request/response traffic from multiple sources, normalizes it to a shared Intermediate Representation (IR), and generates OpenAPI specs through intelligent inference.

## Features

- **Multi-source input**: IR files from HAR files, browser automation, proxy captures, or manual capture
- **Intelligent inference**: Automatically detects path parameters, query params, schemas
- **OpenAPI 3.0/3.1/3.2**: Generate specs for any version
- **Format detection**: Email, UUID, date-time, URI, IPv4/IPv6
- **Type inference**: String, integer, number, boolean, array, object
- **Required/optional**: Tracks field presence across requests
- **Provider pattern**: Symmetric read/write access for IR records
- **Storage integration**: Works with omnistorage for cloud storage backends

## Quick Example

```bash
# Install
go install github.com/grokify/traffic2openapi/cmd/traffic2openapi@latest

# Convert HAR file to IR format
traffic2openapi convert har -i recording.har -o traffic.ndjson

# Generate OpenAPI spec from IR
traffic2openapi generate -i traffic.ndjson -o openapi.yaml
```

## Go Package Usage

```go
import (
    "github.com/grokify/traffic2openapi/pkg/ir"
    "github.com/grokify/traffic2openapi/pkg/inference"
    "github.com/grokify/traffic2openapi/pkg/openapi"
)

// Read IR records
records, err := ir.ReadFile("traffic.ndjson")

// Infer API structure
engine := inference.NewEngine(inference.DefaultEngineOptions())
engine.ProcessRecords(records)
result := engine.Finalize()

// Generate OpenAPI spec
spec := openapi.GenerateFromInference(result, openapi.DefaultGeneratorOptions())
openapi.WriteFile("openapi.yaml", spec)
```

## Architecture Overview

```
Traffic Sources → Adapters → IR (Intermediate Representation) → Inference Engine → OpenAPI Spec
```

See the [Architecture](concepts/architecture.md) page for details.

## Documentation

- [Getting Started](getting-started/installation.md) - Installation and quick start guide
- [Concepts](concepts/architecture.md) - Architecture and IR format
- [Go Packages](go-packages/overview.md) - Provider pattern, LoggingTransport, inference
- [Adapters](adapters/overview.md) - HAR, browser automation, proxy captures
- [CLI Reference](cli/commands.md) - Command-line interface
