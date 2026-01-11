# Quick Start

This guide walks you through generating an OpenAPI spec from HTTP traffic.

## Step 1: Capture Traffic

The easiest way to capture HTTP traffic is using a HAR file from your browser's DevTools:

1. Open Chrome/Firefox DevTools (F12)
2. Go to the Network tab
3. Perform your API operations
4. Right-click → Save all as HAR

Alternatively, use the LoggingTransport in your Go code to capture traffic programmatically.

## Step 2: Convert to IR Format

Convert the HAR file to the Intermediate Representation (IR) format:

```bash
traffic2openapi convert har -i recording.har -o traffic.ndjson
```

## Step 3: Generate OpenAPI Spec

Generate an OpenAPI specification from the IR file:

```bash
traffic2openapi generate -i traffic.ndjson -o openapi.yaml
```

## Step 4: Review and Customize

The generated spec will have:

- Endpoints discovered from traffic patterns
- Path parameters inferred from URL patterns
- Request/response schemas inferred from bodies
- Query parameters extracted from URLs

You can customize the output:

```bash
traffic2openapi generate -i traffic.ndjson -o openapi.yaml \
  --title "My API" \
  --api-version "2.0.0" \
  --server https://api.example.com
```

## Go Package Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/grokify/traffic2openapi/pkg/ir"
    "github.com/grokify/traffic2openapi/pkg/inference"
    "github.com/grokify/traffic2openapi/pkg/openapi"
)

func main() {
    ctx := context.Background()

    // Option 1: Read from file
    records, err := ir.ReadFile("traffic.ndjson")
    if err != nil {
        log.Fatal(err)
    }

    // Option 2: Create records programmatically
    record := ir.NewRecord(ir.RequestMethodGET, "/users", 200).
        SetID("req-001").
        SetHost("api.example.com").
        SetResponseBody(map[string]interface{}{
            "users": []interface{}{},
            "total": 0,
        })

    // Process records through inference engine
    engine := inference.NewEngine(inference.DefaultEngineOptions())
    engine.ProcessRecords(records)
    result := engine.Finalize()

    // Generate OpenAPI spec
    options := openapi.DefaultGeneratorOptions()
    options.Title = "My API"
    options.Version = openapi.Version31

    spec := openapi.GenerateFromInference(result, options)

    // Write to file
    if err := openapi.WriteFile("openapi.yaml", spec); err != nil {
        log.Fatal(err)
    }
}
```

## Next Steps

- Learn about the [Provider Pattern](../go-packages/providers.md) for symmetric IR read/write
- Explore [LoggingTransport](../go-packages/logging-transport.md) for capturing live HTTP traffic
- See [Adapters](../adapters/overview.md) for different traffic sources
