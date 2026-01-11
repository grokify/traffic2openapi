# Playwright Adapter

Capture HTTP traffic from Playwright tests and output IR format for OpenAPI generation.

## Language Support

Playwright supports multiple languages. This adapter provides implementations for:

| Language | Directory | Status |
|----------|-----------|--------|
| [Python](./python/) | `python/` | Available |
| [TypeScript/JavaScript](./typescript/) | `typescript/` | Available |
| Java | `java/` | Planned |
| .NET | `dotnet/` | Planned |

## Quick Start

### Python

```bash
cd python
pip install -e .
```

```python
from playwright.sync_api import sync_playwright
from traffic2openapi_playwright import PlaywrightCapture

with sync_playwright() as p:
    browser = p.chromium.launch()
    context = browser.new_context()

    # Start capturing
    capture = PlaywrightCapture("traffic.ndjson")
    capture.attach(context)

    page = context.new_page()
    page.goto("https://api.example.com")

    # Stop and save
    capture.close()
    browser.close()
```

### TypeScript

```bash
cd typescript
npm install
```

```typescript
import { chromium } from 'playwright';
import { PlaywrightCapture } from 'traffic2openapi-playwright';

const browser = await chromium.launch();
const context = await browser.newContext();

// Start capturing
const capture = new PlaywrightCapture('traffic.ndjson');
capture.attach(context);

const page = await context.newPage();
await page.goto('https://api.example.com');

// Stop and save
await capture.close();
await browser.close();
```

## Output Format

Both implementations output IR records in NDJSON format:

```json
{"id":"req-001","timestamp":"2024-01-08T12:00:00Z","source":"playwright","request":{"method":"GET","host":"api.example.com","path":"/users","headers":{}},"response":{"status":200,"headers":{},"body":{"users":[]}}}
```

## Generating OpenAPI

After capturing traffic:

```bash
# Generate OpenAPI spec
traffic2openapi generate -i traffic.ndjson -o openapi.yaml
```

## Features

- Capture all HTTP requests and responses
- Parse JSON request/response bodies
- Filter by host, method, or path pattern
- Exclude sensitive headers
- Gzip compression support
- Async/streaming writes

## Configuration Options

Both implementations support the same configuration:

| Option | Default | Description |
|--------|---------|-------------|
| `output` | (required) | Output file path |
| `filter_hosts` | `[]` | Only capture requests to these hosts |
| `exclude_paths` | `[]` | Skip requests matching these path patterns |
| `exclude_headers` | `["authorization", "cookie"]` | Headers to exclude |
| `capture_bodies` | `true` | Capture request/response bodies |
| `max_body_size` | `1048576` | Max body size in bytes (1MB) |
| `gzip` | `false` | Gzip compress output |

## Contributing

To add support for another language:

1. Create a new directory (e.g., `java/`)
2. Implement the IR types matching `schemas/ir.v1.schema.json`
3. Implement traffic capture using Playwright's request/response events
4. Implement NDJSON writer
5. Add examples and tests
6. Update this README
