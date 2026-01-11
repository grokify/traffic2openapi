# Traffic2OpenAPI

[![Build Status][build-status-svg]][build-status-url]
[![Lint Status][lint-status-svg]][lint-status-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Documentation][docs-site-svg]][docs-site-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

Generate OpenAPI 3.0/3.1/3.2 specifications from HTTP traffic logs.

**[Documentation](https://grokify.github.io/traffic2openapi/)** | **[Getting Started](https://grokify.github.io/traffic2openapi/getting-started/quickstart/)**

Traffic2OpenAPI captures HTTP request/response traffic from multiple sources (HAR files, Playwright, proxy captures), normalizes it to a shared Intermediate Representation (IR), and generates OpenAPI specs through intelligent inference.

## Quick Start

```bash
# Install
go install github.com/grokify/traffic2openapi/cmd/traffic2openapi@latest

# Convert HAR file to IR format
traffic2openapi convert har -i recording.har -o traffic.ndjson

# Generate OpenAPI spec from IR
traffic2openapi generate -i traffic.ndjson -o openapi.yaml

# Generate static HTML documentation site
traffic2openapi site -i traffic.ndjson -o ./site/

# Validate IR files
traffic2openapi validate ./logs/
```

## Features

- **Multi-source input**: HAR files, Playwright captures, proxy logs, or manual capture
- **Intelligent inference**: Automatically detects path parameters, query params, schemas
- **OpenAPI 3.0/3.1/3.2**: Generate specs for any version
- **Format detection**: Email, UUID, date-time, URI, IPv4/IPv6
- **Type inference**: String, integer, number, boolean, array, object
- **Required/optional**: Tracks field presence across requests
- **Security detection**: Automatically detects Bearer, Basic, API Key authentication
- **Pagination detection**: Identifies page/limit/offset/cursor patterns
- **Rate limit detection**: Captures X-RateLimit-* headers from responses
- **Provider pattern**: Symmetric read/write for IR records (NDJSON, Gzip, Storage, Channel)
- **LoggingTransport**: Capture HTTP traffic from Go `http.Client`
- **Spec comparison**: Diff two OpenAPI specs to detect breaking changes
- **Documentation server**: Serve specs with Swagger UI or Redoc
- **Static site generator**: Generate HTML documentation from traffic logs
- **Watch mode**: Auto-regenerate specs when IR files change

## Architecture

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                              TRAFFIC SOURCES                                  │
├──────────────────┬──────────────────┬──────────────────┬─────────────────────┤
│ Browser DevTools │ Test Automation  │ Go http.Client   │ Proxy Captures      │
│ (HAR export)     │ Playwright, etc  │ LoggingTransport │ mitmproxy, Charles  │
└────────┬─────────┴────────┬─────────┴────────┬─────────┴───────────┬─────────┘
         │                  │                  │                     │
         ▼                  ▼                  ▼                     ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                       ADAPTER LAYER (source-specific)                         │
│                  HAR Converter | Playwright Adapter | IR Writer               │
└─────────────────────────────────────┬────────────────────────────────────────┘
                                      │
                                      ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                    IR (Intermediate Representation)                           │
│                       JSON Schema v1 contract                                 │
└─────────────────────────────────────┬────────────────────────────────────────┘
                                      │
                                      ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                           GO CORE ENGINE                                      │
│    IR Reader → Endpoint Clustering → Schema Inference → OpenAPI Generator     │
└─────────────────────────────────────┬────────────────────────────────────────┘
                                      │
                                      ▼
                           OpenAPI 3.0/3.1/3.2 Spec
```

## Adapter Fidelity Comparison

Different traffic sources capture different levels of detail:

| Adapter | Req Headers | Req Body | Res Headers | Res Body | Query | Timing | Setup |
|---------|:-----------:|:--------:|:-----------:|:--------:|:-----:|:------:|:-----:|
| **HAR Files** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Low |
| **Playwright** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Low |
| **LoggingTransport** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Low |
| **mitmproxy** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Medium |

**Legend:** ✅ Full support | ⚠️ Partial/Limited | ❌ Not available

## IR (Intermediate Representation)

The IR is the shared contract between all traffic sources and the Go processing engine. It supports two formats:

| Format | Use Case | File Extension |
|--------|----------|----------------|
| **Batch** | File-based processing, batch uploads | `.json` |
| **NDJSON** | Streaming, large datasets | `.ndjson` |

### Schema Location

- **JSON Schema**: [`schemas/ir.v1.schema.json`](schemas/ir.v1.schema.json)
- **Examples**: [`examples/`](examples/)

### Batch Format

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

### NDJSON Format (Streaming)

```
{"id":"req-001","request":{"method":"GET","host":"api.example.com","path":"/users"},"response":{"status":200,"body":{"users":[]}}}
{"id":"req-002","request":{"method":"POST","host":"api.example.com","path":"/users","body":{"name":"Bob"}},"response":{"status":201,"body":{"id":"456"}}}
```

### IR Record Fields

#### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `request.method` | string | HTTP method (GET, POST, PUT, PATCH, DELETE, etc.) |
| `request.path` | string | Raw request path without query string |
| `response.status` | integer | HTTP status code (100-599) |

#### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique record identifier |
| `timestamp` | string | ISO 8601 timestamp |
| `source` | string | Adapter type: `har`, `playwright`, `proxy`, `manual` |
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

## Go Package

The `pkg/ir` package provides Go types and utilities for working with IR data.

### Reading IR Files

```go
import "github.com/grokify/traffic2openapi/pkg/ir"

// Read from file (auto-detects format by extension)
records, err := ir.ReadFile("traffic.ndjson")

// Read from directory (all .json and .ndjson files)
records, err := ir.ReadDir("./logs/")

// Stream large NDJSON files
f, _ := os.Open("large-file.ndjson")
recordCh, errCh := ir.StreamNDJSON(f)
for record := range recordCh {
    // Process each record
}
```

### Writing IR Files

```go
// Write to batch JSON format
err := ir.WriteFile("output.json", records)

// Write to NDJSON format
err := ir.WriteFile("output.ndjson", records)

// Stream writes for large datasets
w, _ := ir.NewNDJSONFileWriter("output.ndjson")
defer w.Close()
for _, record := range records {
    w.Write(&record)
}
```

### Creating Records Programmatically

```go
// Using the builder pattern
record := ir.NewRecord(ir.RequestMethodPOST, "/api/users", 201).
    SetID("req-001").
    SetHost("api.example.com").
    SetRequestBody(map[string]string{"name": "Alice"}).
    SetResponseBody(map[string]interface{}{"id": "123", "name": "Alice"}).
    SetDuration(45.2)

// Create a batch
batch := ir.NewBatch([]ir.IRRecord{*record})
```

## Inference Engine

The `pkg/inference` package analyzes IR records and infers API structure.

### Features

- **Path Template Inference**: Converts `/users/123` to `/users/{userId}`
  - Detects UUIDs, numeric IDs, hashes, dates
  - Context-aware naming with 180+ resource patterns (e.g., `/users/123` → `{userId}`, `/orders/456` → `{orderId}`)
- **Schema Inference**: Builds JSON Schema from request/response bodies
  - Type detection (string, integer, number, boolean, array, object)
  - Format detection (email, uuid, date-time, uri, ipv4, ipv6)
  - Optional vs required field tracking
- **Security Detection**: Automatically detects authentication schemes
  - Bearer token (with JWT format detection)
  - Basic authentication
  - API key headers (X-API-Key, X-Auth-Token, etc.)
  - Outputs OpenAPI securitySchemes component
- **Pagination Detection**: Identifies pagination patterns from query parameters
  - Page/limit (offset-based)
  - Cursor/after/before (cursor-based)
  - Tracks min/max values and examples
- **Rate Limit Detection**: Captures rate limiting from response headers
  - X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset
  - Retry-After headers
- **Endpoint Clustering**: Groups requests by method + path template

### Usage

```go
import (
    "github.com/grokify/traffic2openapi/pkg/inference"
    "github.com/grokify/traffic2openapi/pkg/openapi"
)

// Infer from IR files
result, err := inference.InferFromDir("./logs/")

// Or process records directly
engine := inference.NewEngine(inference.DefaultEngineOptions())
engine.ProcessRecords(records)
result := engine.Finalize()

// Generate OpenAPI spec
options := openapi.DefaultGeneratorOptions()
options.Title = "My API"
options.Version = openapi.Version31

spec := openapi.GenerateFromInference(result, options)

// Write to file
openapi.WriteFile("openapi.yaml", spec)
```

## OpenAPI Generation

The `pkg/openapi` package generates OpenAPI 3.0/3.1 specifications.

### Supported Versions

| Version | Description |
|---------|-------------|
| `Version31` | OpenAPI 3.1.0 (default) - Full JSON Schema 2020-12 |
| `Version30` | OpenAPI 3.0.3 - For compatibility |

### Output Formats

```go
// Write to file (format detected by extension)
openapi.WriteFile("spec.yaml", spec)  // YAML
openapi.WriteFile("spec.json", spec)  // JSON

// Convert to string
yamlStr, _ := openapi.ToString(spec, openapi.FormatYAML)
jsonStr, _ := openapi.ToString(spec, openapi.FormatJSON)
```

## HAR Adapter

The `pkg/har` package converts HAR (HTTP Archive) files to IR format.

### Supported Sources

HAR is a standard format supported by:

- **Browser DevTools** - Chrome, Firefox, Safari (Network tab → Save as HAR)
- **Playwright** - Built-in `recordHar` option
- **Puppeteer** - Via Chrome DevTools Protocol
- **Charles Proxy** - File → Export Session as HAR
- **Fiddler** - File → Export → HTTPArchive
- **mitmproxy** - `mitmdump --save-stream-file`
- **Postman** - Collection export

### CLI Usage

```bash
# Convert a single HAR file
traffic2openapi convert har -i recording.har -o traffic.ndjson

# Convert multiple HAR files from a directory
traffic2openapi convert har -i ./har-files/ -o traffic.ndjson

# Filter by host
traffic2openapi convert har -i recording.har -o traffic.ndjson --host api.example.com

# Filter by method
traffic2openapi convert har -i recording.har -o traffic.ndjson --method POST

# Generate OpenAPI from HAR
traffic2openapi convert har -i recording.har -o traffic.ndjson
traffic2openapi generate -i traffic.ndjson -o openapi.yaml
```

## Browser & Test Automation Capture

Capture HTTP traffic from browsers and test automation frameworks like Playwright, Cypress, and Puppeteer.

### Supported Methods

| Method | Tool | Streaming | Best For |
|--------|------|-----------|----------|
| HAR Export | DevTools, Playwright | No | Post-hoc analysis |
| Playwright IR Plugin | Playwright | Yes | Real-time test capture |
| Cypress Intercept | Cypress | Yes | E2E test suites |
| Puppeteer CDP | Puppeteer | Yes | Headless automation |

### Quick Example (Playwright)

```typescript
// Record HAR during tests
const context = await browser.newContext({
  recordHar: { path: 'traffic.har', content: 'embed' }
});
await page.goto('https://api.example.com');
await context.close();  // HAR written here

// Convert to IR and generate spec
// $ traffic2openapi convert har -i traffic.har -o traffic.ndjson
// $ traffic2openapi generate -i traffic.ndjson -o openapi.yaml
```

For real-time streaming and Playwright event-based capture, see **[README_BROWSER.md](README_BROWSER.md)**.

## CLI Usage

### Convert Command

Convert traffic logs to IR format:

```bash
# Convert HAR file to IR
traffic2openapi convert har -i recording.har -o traffic.ndjson
```

### Generate Command

Generate OpenAPI specs from IR files:

```bash
# Generate OpenAPI 3.1 spec from IR files (YAML)
traffic2openapi generate -i ./logs/ -o openapi.yaml

# Generate OpenAPI 3.0 spec (JSON)
traffic2openapi generate -i ./logs/ -o api.json --version 3.0

# Generate with custom title and servers
traffic2openapi generate -i ./logs/ -o api.yaml \
  --title "My API" \
  --api-version "2.0.0" \
  --server https://api.example.com

# Watch mode - auto-regenerate on file changes
traffic2openapi generate -i ./logs/ -o api.yaml --watch
```

### Validate Command

Validate IR files:

```bash
# Validate IR files
traffic2openapi validate ./logs/

# Validate with verbose output
traffic2openapi validate ./logs/ --verbose
```

### Merge Command

Merge multiple IR files or OpenAPI specs:

```bash
# Merge IR files with deduplication
traffic2openapi merge -i file1.ndjson -i file2.ndjson -o merged.ndjson --dedupe

# Merge OpenAPI specs
traffic2openapi merge --openapi -i spec1.yaml -i spec2.yaml -o merged.yaml
```

### Diff Command

Compare two OpenAPI specifications:

```bash
# Compare two specs
traffic2openapi diff old.yaml new.yaml

# Output as JSON (for CI/CD)
traffic2openapi diff old.yaml new.yaml --format json

# Only show breaking changes
traffic2openapi diff old.yaml new.yaml --breaking-only

# Exit with non-zero code if breaking changes found
traffic2openapi diff old.yaml new.yaml --breaking-only --exit-code
```

### Serve Command

Serve OpenAPI spec with interactive documentation:

```bash
# Serve with Swagger UI (default)
traffic2openapi serve openapi.yaml

# Serve on a specific port
traffic2openapi serve openapi.yaml --port 8080

# Serve with Redoc
traffic2openapi serve openapi.yaml --ui redoc

# Auto-reload when spec changes
traffic2openapi serve openapi.yaml --watch
```

### Site Command

Generate a static HTML documentation site from IR traffic logs:

```bash
traffic2openapi site -i <input_file_or_dir> -o <output_dir> [flags]
```

```bash
# Generate site with default settings
traffic2openapi site -i traffic.ndjson -o ./site/

# With custom title
traffic2openapi site -i traffic.ndjson -o ./site/ --title "My API Docs"

# From directory of IR files
traffic2openapi site -i ./logs/ -o ./docs/
```

Features:

- **Index page**: Lists all endpoints with method badges and status codes
- **Endpoint pages**: Detailed view of each endpoint grouped by status code
- **Deduped view**: Collapsed view showing all seen parameter values (e.g., `userId: 123, 456`)
- **Distinct view**: Individual requests with full details
- **Path template detection**: Automatically detects parameters like `/users/{userId}`
- **Light/dark mode**: Toggle with localStorage persistence
- **Copy buttons**: One-click JSON copying
- **Syntax highlighting**: Color-coded JSON bodies

### Generate Command Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--input` | `-i` | (required) | Input file or directory |
| `--output` | `-o` | stdout | Output file path |
| `--version` | `-v` | `3.1` | OpenAPI version: 3.0, 3.1, or 3.2 |
| `--format` | `-f` | auto | Output format: json or yaml |
| `--title` | | `Generated API` | API title |
| `--description` | | | API description |
| `--api-version` | | `1.0.0` | API version |
| `--server` | | | Server URL (repeatable) |
| `--include-errors` | | `true` | Include 4xx/5xx responses |
| `--watch` | `-w` | `false` | Watch for file changes and regenerate |
| `--debounce` | | `500ms` | Debounce interval for watch mode |

## Project Structure

```
traffic2openapi/
├── cmd/
│   └── traffic2openapi/     # CLI application
│       ├── main.go          # Entry point
│       ├── root.go          # Root command
│       ├── generate.go      # Generate command (with watch mode)
│       ├── validate.go      # Validate command
│       ├── convert_har.go   # Convert command (HAR)
│       ├── merge.go         # Merge command (IR/OpenAPI)
│       ├── diff.go          # Diff command (OpenAPI comparison)
│       ├── serve.go         # Serve command (Swagger UI/Redoc)
│       └── site.go          # Site command (static HTML generator)
├── pkg/
│   ├── ir/                  # IR types and I/O
│   │   ├── ir_gen.go        # Generated types from JSON Schema
│   │   ├── ir.go            # Batch type, helpers
│   │   ├── reader.go        # File/dir reading, streaming
│   │   └── writer.go        # File writing, streaming
│   ├── har/                 # HAR file parsing
│   │   └── har.go           # HAR → IR conversion
│   ├── inference/           # Traffic analysis
│   │   ├── engine.go        # Main orchestrator
│   │   ├── endpoint.go      # Endpoint clustering
│   │   ├── path.go          # Path template inference
│   │   ├── schema.go        # JSON Schema inference
│   │   ├── detection.go     # Security, pagination, rate limit detection
│   │   ├── types.go         # Internal types
│   │   └── helpers.go       # Utility functions
│   ├── openapi/             # OpenAPI generation
│   │   ├── generator.go     # Spec builder
│   │   ├── types.go         # OpenAPI 3.x types
│   │   └── writer.go        # JSON/YAML output
│   └── sitegen/             # Static HTML site generator
│       ├── engine.go        # Site engine (wraps inference)
│       ├── generator.go     # HTML generation
│       ├── dedup.go         # Deduplication logic
│       ├── templates.go     # HTML templates
│       └── assets/          # CSS and JavaScript
├── pkg/adapters/playwright/ # Playwright adapters
│   ├── python/              # Python package
│   └── typescript/          # TypeScript package
├── schemas/
│   └── ir.v1.schema.json    # IR JSON Schema
├── examples/
│   ├── sample-batch.json    # Batch format example
│   ├── sample-stream.ndjson # NDJSON format example
│   └── har/                 # HAR file examples
│       └── sample.har
├── go.mod
└── README.md
```

## License

MIT

 [build-status-svg]: https://github.com/grokify/traffic2openapi/actions/workflows/ci.yaml/badge.svg?branch=main
 [build-status-url]: https://github.com/grokify/traffic2openapi/actions/workflows/ci.yaml
 [lint-status-svg]: https://github.com/grokify/traffic2openapi/actions/workflows/lint.yaml/badge.svg?branch=main
 [lint-status-url]: https://github.com/grokify/traffic2openapi/actions/workflows/lint.yaml
 [coverage-svg]: https://img.shields.io/badge/coverage-98.1%25-brightgreen
 [coverage-url]: https://github.com/grokify/traffic2openapi
 [goreport-svg]: https://goreportcard.com/badge/github.com/grokify/traffic2openapi
 [goreport-url]: https://goreportcard.com/report/github.com/grokify/traffic2openapi
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/grokify/traffic2openapi
 [docs-godoc-url]: https://pkg.go.dev/github.com/grokify/traffic2openapi
 [docs-site-svg]: https://img.shields.io/badge/docs-MkDocs-blue.svg
 [docs-site-url]: https://grokify.github.io/traffic2openapi/
 [viz-svg]: https://img.shields.io/badge/visualizaton-Go-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=grokify%2Ftraffic2openapi
 [loc-svg]: https://tokei.rs/b1/github/grokify/traffic2openapi
 [repo-url]: https://github.com/grokify/traffic2openapi
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/grokify/traffic2openapi/blob/master/LICENSE
