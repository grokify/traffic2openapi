# Architecture

Traffic2OpenAPI uses a layered architecture to convert HTTP traffic from various sources into OpenAPI specifications.

## Overview

```
┌──────────────────────────────────────────────────────────────────────────────────────────┐
│                                    TRAFFIC SOURCES                                       │
├──────────────────┬──────────────────┬──────────────────┬──────────────────┬──────────────┤
│ Browser          │ Test Automation  │ Go Applications  │ Proxy Captures   │ Manual       │
│ HAR Files        │ Playwright, etc  │ LoggingTransport │ mitmproxy, etc   │ Capture      │
└────────┬─────────┴────────┬─────────┴────────┬─────────┴────────┬─────────┴───────┬──────┘
         │                  │                  │                  │                 │
         ▼                  ▼                  ▼                  ▼                 ▼
┌──────────────────────────────────────────────────────────────────────────────────────────┐
│                           ADAPTER LAYER (source-specific)                                │
│                    HAR Adapter | Playwright Adapter | LoggingTransport                   │
└─────────────────────────────────────────────┬────────────────────────────────────────────┘
                                              │
                                              ▼
┌──────────────────────────────────────────────────────────────────────────────────────────┐
│                        IR (Intermediate Representation)                                  │
│                           JSON Schema v1 contract                                        │
│                                                                                          │
│         Providers: NDJSON | GzipNDJSON | Storage | Channel                               │
└─────────────────────────────────────────────┬────────────────────────────────────────────┘
                                              │
                                              ▼
┌──────────────────────────────────────────────────────────────────────────────────────────┐
│                               GO CORE ENGINE                                             │
│        IR Reader → Endpoint Clustering → Schema Inference → OpenAPI Generator            │
└─────────────────────────────────────────────┬────────────────────────────────────────────┘
                                              │
                                              ▼
                                   OpenAPI 3.0/3.1/3.2 Spec
```

## Components

### Traffic Sources

HTTP traffic can be captured from various sources:

| Source | Method | Full Bodies | Setup Complexity |
|--------|--------|-------------|------------------|
| Browser DevTools | HAR export | Yes | Low |
| Playwright/Cypress | HAR or events | Yes | Low |
| LoggingTransport | Go http.Client | Yes | Low |
| Proxy captures | mitmproxy, Charles | Yes | Low-Medium |

### Adapter Layer

Adapters convert source-specific formats to the common IR format:

- **HAR Adapter**: Parses HTTP Archive files from browsers and proxies
- **Playwright Adapter**: Captures traffic during test automation
- **LoggingTransport**: Captures traffic from Go http.Client in real-time

### IR (Intermediate Representation)

The IR is the universal format that all adapters output and the inference engine consumes:

- **NDJSON format**: One JSON record per line, ideal for streaming
- **Batch JSON format**: Array of records for file-based processing
- **Provider pattern**: Symmetric read/write access via Providers

See [IR Format](ir-format.md) for the complete schema.

### Providers

Providers offer symmetric read/write access to IR records:

| Provider | Use Case | Streaming |
|----------|----------|-----------|
| `NDJSONProvider` | Plain NDJSON files | Yes |
| `GzipNDJSONProvider` | Compressed NDJSON | Yes |
| `StorageProvider` | Cloud storage via omnistorage | Yes |
| `ChannelProvider` | In-memory Go channels | Yes |

See [Provider Pattern](../go-packages/providers.md) for usage details.

### Inference Engine

The inference engine analyzes IR records to discover API structure:

1. **Endpoint Clustering**: Groups requests by HTTP method + path pattern
2. **Path Parameter Detection**: Identifies dynamic segments (UUIDs, IDs, etc.)
3. **Schema Inference**: Builds JSON Schema from request/response bodies
4. **Format Detection**: Recognizes email, date-time, URI, IP addresses

### OpenAPI Generator

Converts inference results to OpenAPI 3.0/3.1/3.2 specifications:

- Generates paths with parameters
- Creates component schemas
- Adds request/response examples
- Supports YAML and JSON output

## Data Flow

```
1. Capture    → Traffic captured at source (browser, proxy, code)
2. Convert    → Adapter converts to IR format
3. Store      → Provider writes IR records (file, cloud, memory)
4. Read       → Provider reads IR records
5. Analyze    → Inference engine processes records
6. Generate   → OpenAPI spec created from inference results
```

## Integration Points

### With omnistorage

The StorageProvider integrates with [omnistorage](https://github.com/grokify/omnistorage) for cloud storage:

```go
import (
    "github.com/grokify/omnistorage/backend/s3"
    "github.com/grokify/traffic2openapi/pkg/ir"
)

backend, _ := s3.New(ctx, s3.Config{...})
provider := ir.Storage(backend)
writer, _ := provider.NewWriter(ctx, "records.ndjson.gz")
```

### With http.Client

The LoggingTransport wraps `http.RoundTripper` to capture traffic:

```go
transport := ir.NewLoggingTransport(http.DefaultTransport, writer)
client := &http.Client{Transport: transport}
// All requests through this client are logged to the writer
```
