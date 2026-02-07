# Release Notes: v0.1.0

**Release Date:** 2026-02-07

## Overview

Traffic2OpenAPI v0.1.0 is the initial release of a tool that generates OpenAPI 3.0/3.1/3.2 specifications from HTTP traffic logs. Capture real API traffic from production, tests, or development environments and automatically generate accurate API documentation.

## Key Features

### Multi-Source Traffic Capture

Capture HTTP traffic from multiple sources, each with different fidelity levels:

| Source | Request Body | Response Body | Best For |
|--------|:------------:|:-------------:|----------|
| HAR Files | Yes | Yes | Browser DevTools, Playwright |
| Playwright Adapter | Yes | Yes | Test automation |
| LoggingTransport | Yes | Yes | Go http.Client capture |
| Proxy Captures | Yes | Yes | mitmproxy, Charles |

### Intelligent Inference

The inference engine automatically detects:

- **Path parameters**: `/users/123` becomes `/users/{userId}` with 180+ resource patterns
- **Parameter types**: UUIDs, numeric IDs, hashes, dates
- **String formats**: email, UUID, date-time, URI, IPv4, IPv6
- **Required fields**: Tracks field presence across requests
- **Schema types**: string, integer, number, boolean, array, object
- **Security schemes**: Bearer (with JWT detection), Basic auth, API keys
- **Pagination patterns**: page/limit, cursor/after/before
- **Rate limiting**: X-RateLimit-* header detection

### OpenAPI Generation

Generate specifications for any OpenAPI version:

- OpenAPI 3.0.3 (maximum compatibility)
- OpenAPI 3.1.0 (full JSON Schema 2020-12)
- OpenAPI 3.2.0 (latest features)

Output in JSON or YAML format.

### CLI Enhancements

New commands for spec management:

- **merge**: Combine multiple IR files or OpenAPI specs with deduplication
- **diff**: Compare OpenAPI specs with breaking change detection (CI-friendly)
- **serve**: Interactive documentation with Swagger UI or Redoc
- **site**: Generate static HTML documentation from traffic logs
- **watch mode**: Auto-regenerate specs when input files change

## Installation

```bash
go install github.com/grokify/traffic2openapi/cmd/traffic2openapi@latest
```

## Quick Start

```bash
# Convert HAR file to IR format
traffic2openapi convert har -i recording.har -o traffic.ndjson

# Generate OpenAPI spec
traffic2openapi generate -i traffic.ndjson -o openapi.yaml

# Generate directly from IR files
traffic2openapi generate -i ./logs/ -o openapi.yaml
```

## Architecture

```
Traffic Sources → Adapters → IR (Intermediate Representation) → Inference → OpenAPI Spec
```

The IR format provides a common contract between all traffic sources and the Go processing engine. It supports both batch JSON and streaming NDJSON formats.

## Package Structure

| Package | Description |
|---------|-------------|
| `pkg/ir` | IR types, file I/O, streaming |
| `pkg/har` | HAR file parsing and conversion |
| `pkg/inference` | Path template and schema inference |
| `pkg/openapi` | OpenAPI spec generation |
| `pkg/sitegen` | Static HTML documentation site generation |

## What's Next

Planned for future releases:

- Additional adapters (mitmproxy, Envoy)
- Schema merging and conflict resolution
- Web UI for spec review and editing
- SDK generation integration
- CI/CD pipeline examples

## Links

- [GitHub Repository](https://github.com/grokify/traffic2openapi)
- [Documentation](https://github.com/grokify/traffic2openapi#readme)
- [Issue Tracker](https://github.com/grokify/traffic2openapi/issues)
