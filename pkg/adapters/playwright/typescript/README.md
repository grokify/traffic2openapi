# traffic2openapi-playwright (TypeScript)

Capture Playwright HTTP traffic for OpenAPI generation.

## Features

- **Streaming capture** - Write records as they happen, not at the end
- **Type-safe** - Full TypeScript support with strict types
- **Flexible filtering** - Filter by host, method, path patterns
- **Security-first** - Automatically excludes sensitive headers
- **Compression** - Optional gzip output for large captures
- **ESM native** - Modern ES modules with tree-shaking support
- **Playwright Test integration** - Works with `@playwright/test`

## Installation

```bash
npm install traffic2openapi-playwright
```

Or install from source:

```bash
cd pkg/adapters/playwright/typescript
npm install
npm run build
```

## Quick Start

### Basic Usage

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

### With Gzip Compression

```typescript
import { chromium } from 'playwright';
import { PlaywrightCapture } from 'traffic2openapi-playwright';

// Auto-detects gzip from .gz extension
const capture = new PlaywrightCapture('traffic.ndjson.gz');

const browser = await chromium.launch();
const context = await browser.newContext();
capture.attach(context);

const page = await context.newPage();
await page.goto('https://api.example.com/users');

await capture.close();
await browser.close();
```

## Configuration

### Basic Options

```typescript
import { PlaywrightCapture } from 'traffic2openapi-playwright';

const capture = new PlaywrightCapture({
  output: 'traffic.ndjson.gz',

  // Filter by host - only capture these domains
  filterHosts: ['api.example.com', 'api.staging.example.com'],

  // Filter by HTTP method
  filterMethods: ['GET', 'POST', 'PUT', 'DELETE'],

  // Compression (auto-detected from .gz extension)
  gzip: true,
  compressionLevel: 9,
});
```

### Excluding Unwanted Traffic

```typescript
import { PlaywrightCapture } from 'traffic2openapi-playwright';

const capture = new PlaywrightCapture({
  output: 'traffic.ndjson',

  // Exclude specific paths
  excludePaths: [
    '/health',
    '/healthz',
    '/metrics',
    '/favicon.ico',
  ],

  // Exclude paths matching regex patterns
  excludePathPatterns: [
    /^\/_next\//,                    // Next.js internals
    /\.(js|css|png|jpg|svg|woff2?)$/, // Static assets
    /^\/sockjs-node\//,              // Webpack HMR
    /^\/api\/v1\/internal\//,        // Internal APIs
  ],
});
```

### Header Handling

```typescript
import { PlaywrightCapture } from 'traffic2openapi-playwright';

const capture = new PlaywrightCapture({
  output: 'traffic.ndjson',

  // Include headers in output (default: true)
  includeHeaders: true,

  // Headers to exclude (case-insensitive)
  // These are excluded by default for security:
  // authorization, cookie, set-cookie, x-api-key, x-auth-token
  excludeHeaders: new Set([
    'authorization',
    'cookie',
    'set-cookie',
    'x-api-key',
    'x-csrf-token',
    'x-custom-secret',
  ]),
});
```

### Body Capture

```typescript
import { PlaywrightCapture } from 'traffic2openapi-playwright';

const capture = new PlaywrightCapture({
  output: 'traffic.ndjson',

  // Capture request/response bodies
  captureRequestBody: true,
  captureResponseBody: true,

  // Maximum body size (bytes) - larger bodies are skipped
  maxBodySize: 1024 * 1024, // 1MB

  // Only capture bodies for these content types
  captureContentTypes: [
    'application/json',
    'application/xml',
    'text/plain',
    'text/xml',
  ],
});
```

### Error Handling

```typescript
import { PlaywrightCapture } from 'traffic2openapi-playwright';

const capture = new PlaywrightCapture({
  output: 'traffic.ndjson',

  // Custom error handler
  onError: (error) => {
    console.error('Capture error:', error);
    // Send to error tracking service
    // errorTracker.captureException(error);
  },
});
```

## Playwright Test Integration

### Basic Test Setup

```typescript
// playwright.config.ts
import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  use: {
    baseURL: 'https://api.example.com',
  },
});
```

```typescript
// tests/fixtures.ts
import { test as base } from '@playwright/test';
import { PlaywrightCapture } from 'traffic2openapi-playwright';

// Extend base test with capture fixture
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

test('should create user', async ({ page }) => {
  await page.goto('/');
  const response = await page.evaluate(async () => {
    const res = await fetch('/users', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: 'Alice', email: 'alice@example.com' }),
    });
    return { status: res.status, body: await res.json() };
  });
  expect(response.status).toBe(201);
});
```

### Per-Test Capture Files

```typescript
// tests/fixtures.ts
import { test as base } from '@playwright/test';
import { PlaywrightCapture } from 'traffic2openapi-playwright';

export const test = base.extend<{ capture: PlaywrightCapture }>({
  capture: async ({ context }, use, testInfo) => {
    // Create unique file per test
    const filename = `traffic-${testInfo.workerIndex}-${testInfo.testId}.ndjson`;
    const capture = new PlaywrightCapture({
      output: `test-results/${filename}`,
      filterHosts: ['api.example.com'],
    });
    capture.attach(context);
    await use(capture);
    await capture.close();
  },
});
```

### Global Setup/Teardown

```typescript
// global-setup.ts
import { FullConfig } from '@playwright/test';
import { PlaywrightCapture } from 'traffic2openapi-playwright';

let globalCapture: PlaywrightCapture;

export default async function globalSetup(config: FullConfig) {
  globalCapture = new PlaywrightCapture('all-traffic.ndjson.gz');
  // Store for access in tests
  (globalThis as any).__capture = globalCapture;
}

export async function globalTeardown() {
  const capture = (globalThis as any).__capture as PlaywrightCapture;
  if (capture) {
    await capture.close();
    console.log(`Total captured: ${capture.count} requests`);
  }
}
```

## Complete Example: API Testing Workflow

```typescript
/**
 * Complete example: Capture traffic from E2E tests and generate OpenAPI spec.
 */
import { chromium } from 'playwright';
import { PlaywrightCapture } from 'traffic2openapi-playwright';

async function main() {
  // Configure capture with production-ready settings
  const capture = new PlaywrightCapture({
    output: 'api-traffic.ndjson.gz',

    // Only capture your API
    filterHosts: ['api.example.com'],

    // Skip non-API endpoints
    excludePaths: ['/health', '/metrics', '/version'],

    // Skip static assets and internal routes
    excludePathPatterns: [
      /\.(js|css|png|jpg|gif|svg|ico|woff2?)$/,
      /^\/_/,
    ],

    // Capture bodies for schema inference
    captureRequestBody: true,
    captureResponseBody: true,
    maxBodySize: 512 * 1024, // 512KB

    // Compress output
    gzip: true,
  });

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext();
  capture.attach(context);

  const page = await context.newPage();

  console.log('Testing API endpoints...');

  // GET /users
  await page.goto('https://api.example.com/users');

  // GET /users/:id
  await page.goto('https://api.example.com/users/123');

  // POST /users
  await page.evaluate(async () => {
    await fetch('/users', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: 'Alice', email: 'alice@example.com' }),
    });
  });

  // PUT /users/:id
  await page.evaluate(async () => {
    await fetch('/users/123', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: 'Alice Updated' }),
    });
  });

  // DELETE /users/:id
  await page.evaluate(async () => {
    await fetch('/users/123', { method: 'DELETE' });
  });

  console.log(`Captured ${capture.count} requests`);

  await capture.close();
  await browser.close();

  console.log('Traffic saved to api-traffic.ndjson.gz');
  console.log('Generate OpenAPI spec with:');
  console.log('  traffic2openapi generate -i api-traffic.ndjson.gz -o openapi.yaml');
}

main().catch(console.error);
```

## Using with Node.js Scripts

### CommonJS (require)

```javascript
// Use dynamic import in CommonJS
async function main() {
  const { PlaywrightCapture } = await import('traffic2openapi-playwright');
  const { chromium } = await import('playwright');

  const capture = new PlaywrightCapture('traffic.ndjson');
  // ...
}
```

### ES Modules (import)

```javascript
// package.json: "type": "module"
import { PlaywrightCapture } from 'traffic2openapi-playwright';
import { chromium } from 'playwright';

const capture = new PlaywrightCapture('traffic.ndjson');
// ...
```

## Generate OpenAPI

After capturing traffic:

```bash
# Generate OpenAPI spec
traffic2openapi generate -i traffic.ndjson -o openapi.yaml

# With gzip input
traffic2openapi generate -i traffic.ndjson.gz -o openapi.yaml

# Output as JSON
traffic2openapi generate -i traffic.ndjson -o openapi.json
```

## API Reference

### PlaywrightCapture

Main class for capturing traffic.

```typescript
class PlaywrightCapture {
  constructor(options: CaptureOptions | string);
  attach(context: BrowserContext): void;
  flush(): Promise<void>;
  close(): Promise<void>;
  readonly count: number;
}
```

**Methods:**

| Method | Returns | Description |
|--------|---------|-------------|
| `attach(context)` | `void` | Attach to a BrowserContext |
| `flush()` | `Promise<void>` | Flush buffered records to disk |
| `close()` | `Promise<void>` | Close the capture and writer |

**Properties:**

| Property | Type | Description |
|----------|------|-------------|
| `count` | `number` | Number of records captured |

### CaptureOptions

Configuration interface.

```typescript
interface CaptureOptions {
  /** Output file path (required) */
  output: string;

  /** Only capture requests to these hosts */
  filterHosts?: string[];

  /** Only capture these HTTP methods */
  filterMethods?: RequestMethod[];

  /** Skip requests matching these exact paths */
  excludePaths?: string[];

  /** Skip requests matching these path patterns */
  excludePathPatterns?: RegExp[];

  /** Headers to exclude from capture */
  excludeHeaders?: Set<string>;

  /** Include headers in output (default: true) */
  includeHeaders?: boolean;

  /** Capture request bodies (default: true) */
  captureRequestBody?: boolean;

  /** Capture response bodies (default: true) */
  captureResponseBody?: boolean;

  /** Max body size in bytes (default: 1MB) */
  maxBodySize?: number;

  /** Content types to capture bodies for */
  captureContentTypes?: string[];

  /** Gzip compress output (default: false, auto-detected from .gz) */
  gzip?: boolean;

  /** Gzip compression level 1-9 (default: 9) */
  compressionLevel?: number;

  /** Error callback function */
  onError?: (error: Error) => void;
}
```

### IRRecord

The intermediate representation record.

```typescript
interface IRRecord {
  request: Request;
  response: Response;
  id?: string;           // Auto-generated UUID
  timestamp?: string;    // Auto-generated ISO timestamp
  source?: string;       // Default: "playwright"
  durationMs?: number;
}
```

### Request / Response

```typescript
interface Request {
  method: RequestMethod;           // HTTP method
  path: string;                    // URL path
  scheme?: string;                 // http or https
  host?: string;                   // Hostname
  query?: Record<string, string>;  // Query parameters
  headers?: Record<string, string>; // Request headers
  contentType?: string;            // Content-Type header
  body?: unknown;                  // Request body
}

interface Response {
  status: number;                  // HTTP status code
  headers?: Record<string, string>; // Response headers
  contentType?: string;            // Content-Type header
  body?: unknown;                  // Response body
}

type RequestMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE' | 'HEAD' | 'OPTIONS';
```

### Helper Functions

```typescript
import {
  createIRRecord,
  toJSON,
  fromJSON,
} from 'traffic2openapi-playwright';

// Create a record with auto-generated id/timestamp
const record = createIRRecord(
  { method: 'GET', path: '/users' },
  { status: 200, body: [{ id: 1, name: 'Alice' }] },
  { durationMs: 45.5 }
);

// Serialize to JSON string (compact)
const json = toJSON(record);

// Parse from JSON string
const parsed = fromJSON(json);
```

### NDJSONWriter / GzipNDJSONWriter

Low-level writers for IR records.

```typescript
import { NDJSONWriter, GzipNDJSONWriter, createIRRecord } from 'traffic2openapi-playwright';

// Plain NDJSON
const writer = new NDJSONWriter('output.ndjson');
writer.write(createIRRecord(
  { method: 'GET', path: '/users' },
  { status: 200 }
));
await writer.close();

// Gzip compressed
const gzWriter = new GzipNDJSONWriter('output.ndjson.gz', 9);
gzWriter.write(createIRRecord(
  { method: 'POST', path: '/users' },
  { status: 201 }
));
await gzWriter.close();
```

## Troubleshooting

### No requests captured

1. **Check filterHosts** - Ensure it includes the domain you're testing
2. **Check excludePaths** - Make sure you're not excluding the paths you need
3. **Attach before navigation** - Call `capture.attach(context)` before `page.goto()`
4. **Check browser context** - Each context needs its own attach call

### Missing request/response bodies

1. **Check maxBodySize** - Large bodies are skipped by default
2. **Check captureContentTypes** - Only specified content types are captured
3. **Enable body capture** - Set `captureRequestBody: true` and `captureResponseBody: true`

### Headers missing

1. **Enable includeHeaders** - Set `includeHeaders: true`
2. **Check excludeHeaders** - Sensitive headers are excluded by default

### File not created

1. **Call close()** - Always `await capture.close()` before exiting
2. **Check file path** - Ensure the directory exists and is writable
3. **Check for errors** - Use `onError` callback to catch write errors

### Large output files

1. **Enable gzip** - Use `.ndjson.gz` extension or set `gzip: true`
2. **Filter hosts** - Only capture traffic from your API
3. **Reduce body size** - Lower `maxBodySize` if bodies aren't needed for inference

### ESM/CommonJS Issues

This package is ESM-only. If you're using CommonJS:

```javascript
// Use dynamic import
const { PlaywrightCapture } = await import('traffic2openapi-playwright');
```

Or update your `package.json`:

```json
{
  "type": "module"
}
```

## Examples

See the `examples/` directory:

- `basic.ts` - Basic usage example

## Requirements

- Node.js 18+
- Playwright 1.40+

## License

MIT
