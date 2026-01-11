# Browser & Test Automation

Capture HTTP traffic from browsers and test automation frameworks.

## Overview

| Method | Tool | Streaming | Best For |
|--------|------|:---------:|----------|
| **Playwright Adapter** | Playwright (Python/TS) | Yes | Recommended |
| HAR Export | DevTools, Playwright | No | Post-hoc analysis |
| Event-based | Playwright, Puppeteer | Yes | Real-time capture |
| Intercept | Cypress | Yes | E2E test suites |

## Playwright Adapter (Recommended)

We provide official Playwright adapters for Python and TypeScript. These handle all the complexity of capturing traffic and outputting IR format.

**Location:** `pkg/adapters/playwright/`

### Why Playwright Adapter?

- **Streaming** - Records are written as they happen, not buffered in memory
- **Filtering** - Built-in host, method, and path filtering
- **Security** - Automatically excludes sensitive headers
- **Compression** - Optional gzip output for large captures
- **Type-safe** - Full TypeScript types and Python type hints
- **Test integration** - Works with pytest and @playwright/test

### Installation

=== "Python"

    ```bash
    pip install traffic2openapi-playwright
    ```

=== "TypeScript"

    ```bash
    npm install traffic2openapi-playwright
    ```

### Quick Start

=== "Python (Sync)"

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
        page.goto("https://api.example.com/users")

        # Stop and save
        capture.close()
        browser.close()

    print(f"Captured {capture.count} requests")
    ```

=== "Python (Async)"

    ```python
    import asyncio
    from playwright.async_api import async_playwright
    from traffic2openapi_playwright import PlaywrightCapture

    async def main():
        async with async_playwright() as p:
            browser = await p.chromium.launch()
            context = await browser.new_context()

            capture = PlaywrightCapture("traffic.ndjson")
            await capture.attach_async(context)

            page = await context.new_page()
            await page.goto("https://api.example.com/users")

            capture.close()
            await browser.close()

    asyncio.run(main())
    ```

=== "TypeScript"

    ```typescript
    import { chromium } from 'playwright';
    import { PlaywrightCapture } from 'traffic2openapi-playwright';

    const browser = await chromium.launch();
    const context = await browser.newContext();

    // Start capturing
    const capture = new PlaywrightCapture('traffic.ndjson');
    capture.attach(context);

    const page = await context.newPage();
    await page.goto('https://api.example.com/users');

    // Stop and save
    await capture.close();
    await browser.close();

    console.log(`Captured ${capture.count} requests`);
    ```

### Configuration

Both adapters support the same configuration options:

| Option | Default | Description |
|--------|---------|-------------|
| `output` | (required) | Output file path |
| `filter_hosts` / `filterHosts` | `[]` | Only capture requests to these hosts |
| `filter_methods` / `filterMethods` | `[]` | Only capture these HTTP methods |
| `exclude_paths` / `excludePaths` | `[]` | Skip requests matching these paths |
| `exclude_path_patterns` / `excludePathPatterns` | `[]` | Skip paths matching regex patterns |
| `exclude_headers` / `excludeHeaders` | See below | Headers to exclude |
| `capture_request_body` / `captureRequestBody` | `true` | Capture request bodies |
| `capture_response_body` / `captureResponseBody` | `true` | Capture response bodies |
| `max_body_size` / `maxBodySize` | `1048576` | Max body size in bytes (1MB) |
| `gzip` | `false` | Gzip compress output (auto-detected from `.gz`) |
| `compression_level` / `compressionLevel` | `9` | Gzip level 1-9 |
| `on_error` / `onError` | `None` | Error callback function |

**Default excluded headers:**

- `authorization`
- `cookie`
- `set-cookie`
- `x-api-key`
- `x-auth-token`

### Filtering Traffic

Filter to only capture your API traffic:

=== "Python"

    ```python
    import re
    from traffic2openapi_playwright import PlaywrightCapture, CaptureOptions

    options = CaptureOptions(
        output="traffic.ndjson",

        # Only capture your API
        filter_hosts=["api.example.com"],

        # Only capture these methods
        filter_methods=["GET", "POST", "PUT", "DELETE"],

        # Exclude health checks and internal routes
        exclude_paths=["/health", "/metrics", "/version"],

        # Exclude static assets
        exclude_path_patterns=[
            re.compile(r"\.(js|css|png|jpg|svg|woff2?)$"),
            re.compile(r"^/_next/"),
        ],
    )

    capture = PlaywrightCapture(options)
    ```

=== "TypeScript"

    ```typescript
    import { PlaywrightCapture } from 'traffic2openapi-playwright';

    const capture = new PlaywrightCapture({
      output: 'traffic.ndjson',

      // Only capture your API
      filterHosts: ['api.example.com'],

      // Only capture these methods
      filterMethods: ['GET', 'POST', 'PUT', 'DELETE'],

      // Exclude health checks and internal routes
      excludePaths: ['/health', '/metrics', '/version'],

      // Exclude static assets
      excludePathPatterns: [
        /\.(js|css|png|jpg|svg|woff2?)$/,
        /^\/_next\//,
      ],
    });
    ```

### Test Framework Integration

#### pytest (Python)

```python
# conftest.py
import pytest
from playwright.sync_api import sync_playwright
from traffic2openapi_playwright import PlaywrightCapture

@pytest.fixture(scope="session")
def playwright_instance():
    with sync_playwright() as p:
        yield p

@pytest.fixture(scope="session")
def browser(playwright_instance):
    browser = playwright_instance.chromium.launch()
    yield browser
    browser.close()

@pytest.fixture(scope="session")
def traffic_capture():
    capture = PlaywrightCapture("test-traffic.ndjson")
    yield capture
    capture.close()

@pytest.fixture
def context(browser, traffic_capture):
    context = browser.new_context()
    traffic_capture.attach(context)
    yield context
    context.close()

@pytest.fixture
def page(context):
    page = context.new_page()
    yield page
    page.close()
```

```python
# test_api.py
def test_get_users(page):
    response = page.goto("https://api.example.com/users")
    assert response.status == 200

def test_create_user(page):
    page.goto("https://api.example.com")
    # ... perform actions
```

#### @playwright/test (TypeScript)

```typescript
// tests/fixtures.ts
import { test as base } from '@playwright/test';
import { PlaywrightCapture } from 'traffic2openapi-playwright';

export const test = base.extend<{}, { capture: PlaywrightCapture }>({
  capture: [async ({}, use) => {
    const capture = new PlaywrightCapture({
      output: 'test-traffic.ndjson',
      filterHosts: ['api.example.com'],
    });
    await use(capture);
    await capture.close();
  }, { scope: 'worker' }],
});

export { expect } from '@playwright/test';
```

```typescript
// tests/api.spec.ts
import { test, expect } from './fixtures';

test.beforeEach(async ({ context, capture }) => {
  capture.attach(context);
});

test('should get users', async ({ page }) => {
  const response = await page.goto('/users');
  expect(response?.status()).toBe(200);
});
```

### Generating OpenAPI

After capturing traffic:

```bash
# Generate OpenAPI spec
traffic2openapi generate -i traffic.ndjson -o openapi.yaml

# With gzip input
traffic2openapi generate -i traffic.ndjson.gz -o openapi.yaml
```

### Full Documentation

For complete API reference and examples:

- [Python README](https://github.com/grokify/traffic2openapi/tree/main/pkg/adapters/playwright/python)
- [TypeScript README](https://github.com/grokify/traffic2openapi/tree/main/pkg/adapters/playwright/typescript)

---

## Alternative: HAR Recording

If you prefer using Playwright's built-in HAR recording:

```typescript
const context = await browser.newContext({
  recordHar: { path: 'traffic.har', content: 'embed' }
});

const page = await context.newPage();
await page.goto('https://api.example.com');
// ... perform actions ...

await context.close();  // HAR written here

// Convert to IR
// $ traffic2openapi convert har -i traffic.har -o traffic.ndjson
```

**Pros:**

- Built into Playwright
- No additional dependencies

**Cons:**

- Not streaming (all in memory until close)
- Requires conversion step
- Less filtering control

---

## Alternative: Manual Event Capture

For custom capture logic or frameworks without adapters:

```typescript
import { chromium } from 'playwright';
import * as fs from 'fs';

const browser = await chromium.launch();
const context = await browser.newContext();
const page = await context.newPage();

const records: any[] = [];

// Capture responses (includes request info)
page.on('response', async response => {
  const request = response.request();
  const url = new URL(request.url());

  // Filter to API calls only
  if (!url.hostname.includes('api.example.com')) return;

  const record = {
    id: crypto.randomUUID(),
    timestamp: new Date().toISOString(),
    source: 'playwright',
    request: {
      method: request.method(),
      path: url.pathname,
      host: url.hostname,
      query: Object.fromEntries(url.searchParams),
      headers: request.headers(),
    },
    response: {
      status: response.status(),
      headers: response.headers(),
      body: await response.json().catch(() => null),
    },
  };
  records.push(record);
});

await page.goto('https://api.example.com');
// ... perform actions ...

// Write NDJSON
fs.writeFileSync('traffic.ndjson',
  records.map(r => JSON.stringify(r)).join('\n') + '\n');

await browser.close();
```

---

## Cypress

### Network Intercept

```typescript
// cypress/e2e/capture.cy.ts
describe('API Capture', () => {
  const records: any[] = [];

  beforeEach(() => {
    cy.intercept('**/api/**', (req) => {
      req.continue((res) => {
        records.push({
          id: crypto.randomUUID(),
          timestamp: new Date().toISOString(),
          source: 'cypress',
          request: {
            method: req.method,
            path: new URL(req.url).pathname,
            headers: req.headers,
            body: req.body,
          },
          response: {
            status: res.statusCode,
            headers: res.headers,
            body: res.body,
          },
        });
      });
    });
  });

  after(() => {
    cy.writeFile('traffic.ndjson',
      records.map(r => JSON.stringify(r)).join('\n') + '\n');
  });

  it('captures API traffic', () => {
    cy.visit('/');
    // ... test actions ...
  });
});
```

---

## Puppeteer

### CDP-based Capture

```typescript
import puppeteer from 'puppeteer';
import * as fs from 'fs';

const browser = await puppeteer.launch();
const page = await browser.newPage();

const records: any[] = [];

await page.setRequestInterception(true);

page.on('request', request => {
  request.continue();
});

page.on('response', async response => {
  const request = response.request();
  const url = new URL(request.url());

  records.push({
    id: crypto.randomUUID(),
    timestamp: new Date().toISOString(),
    source: 'puppeteer',
    request: {
      method: request.method(),
      path: url.pathname,
      host: url.hostname,
      headers: request.headers(),
    },
    response: {
      status: response.status(),
      headers: response.headers(),
      body: await response.text().catch(() => null),
    },
  });
});

await page.goto('https://api.example.com');
// ... perform actions ...

fs.writeFileSync('traffic.ndjson',
  records.map(r => JSON.stringify(r)).join('\n') + '\n');

await browser.close();
```

---

## Chrome DevTools

### Manual Export

1. Open DevTools (F12)
2. Go to Network tab
3. Check "Preserve log" to capture across navigations
4. Perform API operations
5. Right-click → "Save all as HAR with content"
6. Convert: `traffic2openapi convert har -i traffic.har -o traffic.ndjson`

### DevTools Protocol

```typescript
// Use Chrome DevTools Protocol for programmatic capture
const CDP = require('chrome-remote-interface');

const client = await CDP();
const { Network } = client;

await Network.enable();

Network.requestWillBeSent((params) => {
  console.log('Request:', params.request.url);
});

Network.responseReceived((params) => {
  console.log('Response:', params.response.status);
});
```

---

## Best Practices

### 1. Filter API Traffic

Exclude static assets, third-party requests, and internal routes:

```python
exclude_path_patterns=[
    re.compile(r"\.(js|css|png|jpg|gif|svg|ico|woff2?)$"),
    re.compile(r"^/_next/"),
    re.compile(r"^/sockjs-node/"),
    re.compile(r"^/__webpack_hmr"),
]
```

### 2. Capture During Tests

Use E2E tests to generate comprehensive traffic that covers:

- All API endpoints
- Different HTTP methods
- Various request payloads
- Success and error responses

### 3. Include Error Cases

Capture 4xx/5xx responses for complete schemas:

```python
# Don't filter by status - capture all responses
# The adapter captures everything by default
```

### 4. Use Realistic Data

Representative request bodies improve schema inference:

```typescript
// Good - realistic data
await page.evaluate(async () => {
  await fetch('/users', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      name: 'Alice Smith',
      email: 'alice@example.com',
      age: 30,
      roles: ['admin', 'user'],
    }),
  });
});
```

### 5. Compress Large Captures

For extensive test suites:

```python
capture = PlaywrightCapture("traffic.ndjson.gz")  # Auto-gzip
```

### 6. Exclude Sensitive Headers

The adapters exclude auth headers by default, but verify for your use case:

```python
exclude_headers={
    "authorization",
    "cookie",
    "x-api-key",
    "x-internal-token",
}
```
