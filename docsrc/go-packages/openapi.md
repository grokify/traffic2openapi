# OpenAPI Generator

The OpenAPI generator creates OpenAPI 3.0/3.1/3.2 specifications from inference results.

## Overview

```go
import "github.com/grokify/traffic2openapi/pkg/openapi"

// Generate from inference results
spec := openapi.GenerateFromInference(result, openapi.DefaultGeneratorOptions())

// Write to file
openapi.WriteFile("openapi.yaml", spec)
```

## Supported Versions

| Version | Constant | Description |
|---------|----------|-------------|
| 3.0.3 | `openapi.Version30` | Wide compatibility |
| 3.1.0 | `openapi.Version31` | JSON Schema 2020-12 (default) |
| 3.2.0 | `openapi.Version32` | Latest features |

## Generator Options

```go
options := openapi.GeneratorOptions{
    // OpenAPI version
    Version: openapi.Version31,

    // API metadata
    Title:       "My API",
    Description: "API generated from traffic",
    APIVersion:  "1.0.0",

    // Server URLs
    Servers: []string{
        "https://api.example.com",
        "https://staging.example.com",
    },

    // Include 4xx/5xx responses
    IncludeErrors: true,

    // Contact information
    ContactName:  "API Support",
    ContactEmail: "support@example.com",
    ContactURL:   "https://example.com/support",

    // License
    LicenseName: "MIT",
    LicenseURL:  "https://opensource.org/licenses/MIT",
}

spec := openapi.GenerateFromInference(result, options)
```

## Output Formats

### YAML (recommended)

```go
// Write to YAML file
openapi.WriteFile("openapi.yaml", spec)

// Convert to YAML string
yaml, err := openapi.ToString(spec, openapi.FormatYAML)
```

### JSON

```go
// Write to JSON file
openapi.WriteFile("openapi.json", spec)

// Convert to JSON string
json, err := openapi.ToString(spec, openapi.FormatJSON)
```

## Generated Structure

The generator produces a complete OpenAPI specification:

```yaml
openapi: 3.1.0
info:
  title: My API
  version: 1.0.0
servers:
  - url: https://api.example.com
paths:
  /users:
    get:
      summary: Get users
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetUsersResponse'
  /users/{userId}:
    get:
      parameters:
        - name: userId
          in: path
          required: true
          schema:
            type: string
components:
  schemas:
    GetUsersResponse:
      type: object
      properties:
        users:
          type: array
          items:
            $ref: '#/components/schemas/User'
```

## Schema References

The generator automatically creates reusable schemas in `components/schemas`:

- Request body schemas: `{OperationId}Request`
- Response schemas: `{OperationId}Response`
- Nested objects: Extracted and referenced

## Customization

### Post-processing

Modify the spec after generation:

```go
spec := openapi.GenerateFromInference(result, options)

// Add security schemes
spec.Components.SecuritySchemes = map[string]interface{}{
    "bearerAuth": map[string]string{
        "type":   "http",
        "scheme": "bearer",
    },
}

// Add tags
spec.Tags = []map[string]string{
    {"name": "users", "description": "User operations"},
    {"name": "orders", "description": "Order operations"},
}

openapi.WriteFile("openapi.yaml", spec)
```

### Multiple Servers

```go
options.Servers = []string{
    "https://api.example.com",
    "https://staging.example.com",
    "http://localhost:8080",
}
```

## Full Example

```go
package main

import (
    "context"
    "log"

    "github.com/grokify/traffic2openapi/pkg/inference"
    "github.com/grokify/traffic2openapi/pkg/ir"
    "github.com/grokify/traffic2openapi/pkg/openapi"
)

func main() {
    ctx := context.Background()

    // Read traffic
    provider := ir.NDJSON()
    reader, err := provider.NewReader(ctx, "traffic.ndjson")
    if err != nil {
        log.Fatal(err)
    }

    var records []*ir.IRRecord
    for {
        record, err := reader.Read()
        if err != nil {
            break
        }
        records = append(records, record)
    }
    reader.Close()

    // Infer API structure
    engine := inference.NewEngine(inference.DefaultEngineOptions())
    engine.ProcessRecords(records)
    result := engine.Finalize()

    // Generate OpenAPI spec
    options := openapi.GeneratorOptions{
        Version:     openapi.Version31,
        Title:       "User Service API",
        Description: "API for managing users",
        APIVersion:  "2.0.0",
        Servers:     []string{"https://api.example.com/v2"},
    }

    spec := openapi.GenerateFromInference(result, options)

    // Write YAML
    if err := openapi.WriteFile("openapi.yaml", spec); err != nil {
        log.Fatal(err)
    }

    // Also write JSON
    if err := openapi.WriteFile("openapi.json", spec); err != nil {
        log.Fatal(err)
    }
}
```

## CLI Usage

The CLI provides the same functionality:

```bash
# Basic generation
traffic2openapi generate -i traffic.ndjson -o openapi.yaml

# With options
traffic2openapi generate -i traffic.ndjson -o openapi.yaml \
    --title "My API" \
    --api-version "2.0.0" \
    --version 3.1 \
    --server https://api.example.com \
    --server https://staging.example.com
```
