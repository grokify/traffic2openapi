# Intermediate Representation (IR)

The IR is the shared contract between all traffic sources and the Go processing engine.

## Formats

IR supports two serialization formats:

| Format | Use Case | File Extension |
|--------|----------|----------------|
| **NDJSON** | Streaming, large datasets, cloud storage | `.ndjson`, `.ndjson.gz` |
| **Batch JSON** | File-based processing, small datasets | `.json` |

## Schema

The JSON Schema is available at: [`schemas/ir.v1.schema.json`](https://github.com/grokify/traffic2openapi/blob/main/schemas/ir.v1.schema.json)

## NDJSON Format (Recommended)

Each line is a complete JSON record:

```json
{"id":"req-001","request":{"method":"GET","host":"api.example.com","path":"/users"},"response":{"status":200,"body":{"users":[]}}}
{"id":"req-002","request":{"method":"POST","host":"api.example.com","path":"/users","body":{"name":"Bob"}},"response":{"status":201,"body":{"id":"456"}}}
```

Benefits:

- Streaming-friendly: process line by line
- Append-only: easy to add records
- Compression: gzip works well with NDJSON

## Batch JSON Format

Array of records with metadata wrapper:

```json
{
  "version": "ir.v1",
  "metadata": {
    "generatedAt": "2024-12-30T10:00:00Z",
    "source": "manual",
    "recordCount": 2
  },
  "records": [
    {
      "id": "req-001",
      "timestamp": "2024-12-30T09:00:00Z",
      "request": {
        "method": "GET",
        "host": "api.example.com",
        "path": "/users",
        "query": { "limit": "10" }
      },
      "response": {
        "status": 200,
        "contentType": "application/json",
        "body": { "users": [], "total": 0 }
      }
    }
  ]
}
```

## IR Record Fields

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `request.method` | string | HTTP method (GET, POST, PUT, PATCH, DELETE, etc.) |
| `request.path` | string | Raw request path without query string |
| `response.status` | integer | HTTP status code (100-599) |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique record identifier |
| `timestamp` | string | ISO 8601 timestamp |
| `source` | string | Adapter type: `har`, `playwright`, `logging-transport`, `proxy`, `manual` |
| `request.scheme` | string | `http` or `https` |
| `request.host` | string | Request host header |
| `request.pathTemplate` | string | Normalized path with parameters (e.g., `/users/{id}`) |
| `request.pathParams` | object | Extracted path parameter values |
| `request.query` | object | Query parameters |
| `request.headers` | object | Request headers (lowercase keys) |
| `request.contentType` | string | Content-Type header |
| `request.body` | any | Parsed request body (object, array, string, or null) |
| `response.headers` | object | Response headers |
| `response.contentType` | string | Response Content-Type |
| `response.body` | any | Parsed response body |
| `durationMs` | number | Round-trip time in milliseconds |

## Go Types

The `pkg/ir` package provides Go types for working with IR data:

```go
// IRRecord represents a single HTTP request/response capture
type IRRecord struct {
    ID         *string   `json:"id,omitempty"`
    Timestamp  *string   `json:"timestamp,omitempty"`
    Source     *string   `json:"source,omitempty"`
    Request    Request   `json:"request"`
    Response   Response  `json:"response"`
    DurationMs *float64  `json:"durationMs,omitempty"`
}

// Request contains HTTP request details
type Request struct {
    Method       RequestMethod          `json:"method"`
    Scheme       *string                `json:"scheme,omitempty"`
    Host         *string                `json:"host,omitempty"`
    Path         string                 `json:"path"`
    PathTemplate *string                `json:"pathTemplate,omitempty"`
    PathParams   map[string]string      `json:"pathParams,omitempty"`
    Query        map[string]string      `json:"query,omitempty"`
    Headers      map[string]string      `json:"headers,omitempty"`
    ContentType  *string                `json:"contentType,omitempty"`
    Body         interface{}            `json:"body,omitempty"`
}

// Response contains HTTP response details
type Response struct {
    Status      int                    `json:"status"`
    Headers     map[string]string      `json:"headers,omitempty"`
    ContentType *string                `json:"contentType,omitempty"`
    Body        interface{}            `json:"body,omitempty"`
}
```

## Creating Records

Use the builder pattern for creating records:

```go
record := ir.NewRecord(ir.RequestMethodPOST, "/api/users", 201).
    SetID("req-001").
    SetHost("api.example.com").
    SetRequestBody(map[string]string{"name": "Alice"}).
    SetResponseBody(map[string]interface{}{"id": "123", "name": "Alice"}).
    SetDuration(45.2)
```

## Reading and Writing

See [Provider Pattern](../go-packages/providers.md) for the recommended way to read and write IR records.

Quick example:

```go
// Read from file
records, err := ir.ReadFile("traffic.ndjson")

// Write to file
err := ir.WriteFile("output.ndjson", records)
```
