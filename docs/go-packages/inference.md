# Inference Engine

The inference engine analyzes IR records to discover API structure, path parameters, and schemas.

## Overview

The inference engine processes IR records and produces:

- **Endpoint patterns**: Discovered API endpoints with path templates
- **Path parameters**: Dynamic URL segments (UUIDs, IDs, slugs)
- **Request schemas**: JSON Schema for request bodies
- **Response schemas**: JSON Schema for response bodies by status code
- **Query parameters**: Discovered query string parameters

## Basic Usage

```go
import "github.com/grokify/traffic2openapi/pkg/inference"

// Create engine with default options
engine := inference.NewEngine(inference.DefaultEngineOptions())

// Process records
engine.ProcessRecords(records)

// Get results
result := engine.Finalize()

// Result contains discovered endpoints and schemas
for path, endpoint := range result.Endpoints {
    fmt.Printf("Endpoint: %s\n", path)
    for method, operation := range endpoint.Operations {
        fmt.Printf("  %s: %d requests\n", method, operation.RequestCount)
    }
}
```

## Engine Options

```go
options := inference.EngineOptions{
    // Path parameter detection
    DetectPathParams: true,

    // Minimum occurrences to consider a pattern
    MinOccurrences: 2,

    // Include 4xx/5xx responses in schema inference
    IncludeErrorResponses: true,

    // Maximum depth for schema inference
    MaxSchemaDepth: 10,

    // Merge similar schemas
    MergeSchemas: true,
}

engine := inference.NewEngine(options)
```

## Path Parameter Detection

The engine automatically detects dynamic path segments:

| Pattern | Detected As | Example |
|---------|-------------|---------|
| UUID | `{id}` | `/users/550e8400-e29b-41d4-a716-446655440000` |
| Numeric ID | `{id}` | `/users/12345` |
| Short hash | `{hash}` | `/commits/a1b2c3d` |
| Date | `{date}` | `/reports/2024-01-15` |
| Slug | `{slug}` | `/posts/hello-world` |

Context-aware naming:

```
/users/123        → /users/{userId}
/posts/456        → /posts/{postId}
/orders/789/items → /orders/{orderId}/items
```

## Schema Inference

### Type Detection

| JSON Type | Inferred Type |
|-----------|---------------|
| `"hello"` | `string` |
| `123` | `integer` |
| `12.5` | `number` |
| `true` | `boolean` |
| `[]` | `array` |
| `{}` | `object` |
| `null` | nullable |

### Format Detection

| Pattern | Format |
|---------|--------|
| `user@example.com` | `email` |
| `550e8400-e29b-...` | `uuid` |
| `2024-01-15T10:30:00Z` | `date-time` |
| `2024-01-15` | `date` |
| `https://example.com` | `uri` |
| `192.168.1.1` | `ipv4` |
| `::1` | `ipv6` |

### Required vs Optional

Fields are tracked across multiple requests:

```go
// Request 1: {"name": "Alice", "email": "alice@example.com"}
// Request 2: {"name": "Bob"}
// Request 3: {"name": "Charlie", "email": "charlie@example.com"}

// Result:
// - "name" is required (present in all requests)
// - "email" is optional (present in 2/3 requests)
```

## Result Structure

```go
type InferenceResult struct {
    // Discovered endpoints keyed by path template
    Endpoints map[string]*Endpoint

    // Global schemas that can be reused
    Schemas map[string]*Schema
}

type Endpoint struct {
    // Path template (e.g., "/users/{userId}")
    PathTemplate string

    // Path parameters
    PathParams []PathParam

    // Operations keyed by HTTP method
    Operations map[string]*Operation
}

type Operation struct {
    // HTTP method
    Method string

    // Number of requests observed
    RequestCount int

    // Query parameters
    QueryParams []QueryParam

    // Request body schema
    RequestSchema *Schema

    // Response schemas keyed by status code
    ResponseSchemas map[int]*Schema
}
```

## Processing Modes

### Batch Processing

```go
// Process all records at once
engine := inference.NewEngine(options)
engine.ProcessRecords(records)
result := engine.Finalize()
```

### Streaming Processing

```go
// Process records one at a time
engine := inference.NewEngine(options)

reader, _ := provider.NewReader(ctx, "traffic.ndjson")
for {
    record, err := reader.Read()
    if err == io.EOF {
        break
    }
    engine.ProcessRecord(record)
}

result := engine.Finalize()
```

### Incremental Processing

```go
// Add more records to existing engine
engine.ProcessRecords(batch1)
// ... later ...
engine.ProcessRecords(batch2)
// Only finalize when done
result := engine.Finalize()
```

## Convenience Functions

```go
// Infer from directory of IR files
result, err := inference.InferFromDir("./traffic/")

// Infer from single file
result, err := inference.InferFromFile("traffic.ndjson")
```

## Integration with OpenAPI Generator

```go
import (
    "github.com/grokify/traffic2openapi/pkg/inference"
    "github.com/grokify/traffic2openapi/pkg/openapi"
)

// Infer API structure
engine := inference.NewEngine(inference.DefaultEngineOptions())
engine.ProcessRecords(records)
result := engine.Finalize()

// Generate OpenAPI spec
options := openapi.DefaultGeneratorOptions()
options.Title = "My API"
options.Version = openapi.Version31

spec := openapi.GenerateFromInference(result, options)
openapi.WriteFile("openapi.yaml", spec)
```

## Best Practices

### Sufficient Sample Size

More requests lead to better inference:

- **Path parameters**: Need multiple values to detect patterns
- **Required fields**: Need multiple requests to distinguish required/optional
- **Response schemas**: Need examples of each status code

### Representative Traffic

Capture diverse traffic for best results:

- All API endpoints
- Various query parameter combinations
- Different request body shapes
- Success and error responses

### Pre-filtering

Filter traffic before inference:

```go
// Only process successful responses
var filtered []*ir.IRRecord
for _, record := range records {
    if record.Response.Status >= 200 && record.Response.Status < 300 {
        filtered = append(filtered, record)
    }
}
engine.ProcessRecords(filtered)
```
