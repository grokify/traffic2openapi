# traffic2openapi-playwright (Python)

Capture Playwright HTTP traffic for OpenAPI generation.

## Features

- **Streaming capture** - Write records as they happen, not at the end
- **Sync and async support** - Works with both Playwright APIs
- **Flexible filtering** - Filter by host, method, path patterns
- **Security-first** - Automatically excludes sensitive headers
- **Compression** - Optional gzip output for large captures
- **Thread-safe** - Safe for concurrent use
- **Context manager** - Automatic cleanup with `with` statement

## Installation

```bash
pip install traffic2openapi-playwright
```

Or install from source:

```bash
cd pkg/adapters/playwright/python
pip install -e .
```

For development with test dependencies:

```bash
pip install -e ".[dev]"
```

## Quick Start

### Sync API

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

### Async API

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

### Context Manager

```python
from playwright.sync_api import sync_playwright
from traffic2openapi_playwright import PlaywrightCapture

with sync_playwright() as p:
    browser = p.chromium.launch()
    context = browser.new_context()

    with PlaywrightCapture("traffic.ndjson") as capture:
        capture.attach(context)
        page = context.new_page()
        page.goto("https://api.example.com")
        # Automatically closed when exiting context

    browser.close()
```

## Configuration

### Basic Options

```python
from traffic2openapi_playwright import PlaywrightCapture, CaptureOptions

options = CaptureOptions(
    output="traffic.ndjson.gz",

    # Filter by host - only capture these domains
    filter_hosts=["api.example.com", "api.staging.example.com"],

    # Filter by HTTP method
    filter_methods=["GET", "POST", "PUT", "DELETE"],

    # Compression (auto-detected from .gz extension)
    gzip=True,
    compression_level=9,
)

capture = PlaywrightCapture(options)
```

### Excluding Unwanted Traffic

```python
import re
from traffic2openapi_playwright import PlaywrightCapture, CaptureOptions

options = CaptureOptions(
    output="traffic.ndjson",

    # Exclude specific paths
    exclude_paths=[
        "/health",
        "/healthz",
        "/metrics",
        "/favicon.ico",
    ],

    # Exclude paths matching regex patterns
    exclude_path_patterns=[
        re.compile(r"^/_next/"),           # Next.js internals
        re.compile(r"\.(js|css|png|jpg|svg|woff2?)$"),  # Static assets
        re.compile(r"^/sockjs-node/"),     # Webpack HMR
    ],
)

capture = PlaywrightCapture(options)
```

### Header Handling

```python
from traffic2openapi_playwright import PlaywrightCapture, CaptureOptions

options = CaptureOptions(
    output="traffic.ndjson",

    # Include headers in output (default: True)
    include_headers=True,

    # Headers to exclude (case-insensitive)
    # These are excluded by default for security:
    # authorization, cookie, set-cookie, x-api-key, x-auth-token
    exclude_headers={
        "authorization",
        "cookie",
        "set-cookie",
        "x-api-key",
        "x-csrf-token",
        "x-custom-secret",
    },
)

capture = PlaywrightCapture(options)
```

### Body Capture

```python
from traffic2openapi_playwright import PlaywrightCapture, CaptureOptions

options = CaptureOptions(
    output="traffic.ndjson",

    # Capture request/response bodies
    capture_request_body=True,
    capture_response_body=True,

    # Maximum body size (bytes) - larger bodies are skipped
    max_body_size=1024 * 1024,  # 1MB

    # Only capture bodies for these content types
    capture_content_types=[
        "application/json",
        "application/xml",
        "text/plain",
        "text/xml",
    ],
)

capture = PlaywrightCapture(options)
```

### Error Handling

```python
import logging
from traffic2openapi_playwright import PlaywrightCapture, CaptureOptions

logging.basicConfig(level=logging.WARNING)
logger = logging.getLogger(__name__)

options = CaptureOptions(
    output="traffic.ndjson",

    # Custom error handler
    on_error=lambda e: logger.warning(f"Capture error: {e}"),
)

capture = PlaywrightCapture(options)
```

## Integration with pytest

### Basic pytest Integration

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
    # ... perform actions that trigger API calls
```

### Async pytest Integration

```python
# conftest.py
import pytest
import pytest_asyncio
from playwright.async_api import async_playwright
from traffic2openapi_playwright import PlaywrightCapture

@pytest_asyncio.fixture(scope="session")
async def browser():
    async with async_playwright() as p:
        browser = await p.chromium.launch()
        yield browser
        await browser.close()

@pytest.fixture(scope="session")
def traffic_capture():
    capture = PlaywrightCapture("test-traffic.ndjson")
    yield capture
    capture.close()

@pytest_asyncio.fixture
async def context(browser, traffic_capture):
    context = await browser.new_context()
    await traffic_capture.attach_async(context)
    yield context
    await context.close()
```

## Complete Example: API Testing Workflow

```python
"""
Complete example: Capture traffic from E2E tests and generate OpenAPI spec.
"""
import re
from playwright.sync_api import sync_playwright
from traffic2openapi_playwright import PlaywrightCapture, CaptureOptions

def main():
    # Configure capture with production-ready settings
    options = CaptureOptions(
        output="api-traffic.ndjson.gz",

        # Only capture your API
        filter_hosts=["api.example.com"],

        # Skip non-API endpoints
        exclude_paths=["/health", "/metrics", "/version"],

        # Skip static assets and internal routes
        exclude_path_patterns=[
            re.compile(r"\.(js|css|png|jpg|gif|svg|ico|woff2?)$"),
            re.compile(r"^/_"),
        ],

        # Capture bodies for schema inference
        capture_request_body=True,
        capture_response_body=True,
        max_body_size=512 * 1024,  # 512KB

        # Compress output
        gzip=True,
    )

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context()

        with PlaywrightCapture(options) as capture:
            capture.attach(context)
            page = context.new_page()

            # Perform API operations
            print("Testing API endpoints...")

            # GET /users
            page.goto("https://api.example.com/users")

            # GET /users/:id
            page.goto("https://api.example.com/users/123")

            # POST /users (via form or fetch)
            page.evaluate("""
                fetch('/users', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ name: 'Alice', email: 'alice@example.com' })
                })
            """)

            # PUT /users/:id
            page.evaluate("""
                fetch('/users/123', {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ name: 'Alice Updated' })
                })
            """)

            # DELETE /users/:id
            page.evaluate("fetch('/users/123', { method: 'DELETE' })")

            print(f"Captured {capture.count} requests")

        browser.close()

    print("Traffic saved to api-traffic.ndjson.gz")
    print("Generate OpenAPI spec with:")
    print("  traffic2openapi generate -i api-traffic.ndjson.gz -o openapi.yaml")

if __name__ == "__main__":
    main()
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

```python
class PlaywrightCapture:
    def __init__(self, options: CaptureOptions | str): ...
    def attach(self, context: BrowserContext) -> None: ...
    async def attach_async(self, context: BrowserContext) -> None: ...
    def flush(self) -> None: ...
    def close(self) -> None: ...
    @property
    def count(self) -> int: ...
```

**Methods:**

| Method | Description |
|--------|-------------|
| `attach(context)` | Attach to a sync BrowserContext |
| `attach_async(context)` | Attach to an async BrowserContext |
| `flush()` | Flush buffered records to disk |
| `close()` | Close the capture and writer |

**Properties:**

| Property | Type | Description |
|----------|------|-------------|
| `count` | `int` | Number of records captured |

### CaptureOptions

Configuration dataclass.

```python
@dataclass
class CaptureOptions:
    output: str                           # Output file path (required)
    filter_hosts: list[str] | None        # Only capture these hosts
    filter_methods: list[str] | None      # Only capture these HTTP methods
    exclude_paths: list[str] | None       # Skip these exact paths
    exclude_path_patterns: list[re.Pattern] | None  # Skip matching paths
    exclude_headers: set[str] | None      # Headers to exclude
    include_headers: bool = True          # Include headers in output
    capture_request_body: bool = True     # Capture request bodies
    capture_response_body: bool = True    # Capture response bodies
    max_body_size: int = 1048576          # Max body size (1MB)
    capture_content_types: list[str] | None  # Content types to capture
    gzip: bool = False                    # Gzip compress output
    compression_level: int = 9            # Gzip level (1-9)
    on_error: Callable[[Exception], None] | None  # Error callback
```

### IRRecord

The intermediate representation record.

```python
@dataclass
class IRRecord:
    request: Request
    response: Response
    id: str | None = None                 # Auto-generated UUID
    timestamp: str | None = None          # Auto-generated ISO timestamp
    source: str = "playwright"
    duration_ms: float | None = None
```

### Request / Response

```python
@dataclass
class Request:
    method: str                           # HTTP method
    path: str                             # URL path
    scheme: str | None = None             # http or https
    host: str | None = None               # Hostname
    query: dict[str, str] | None = None   # Query parameters
    headers: dict[str, str] | None = None # Request headers
    content_type: str | None = None       # Content-Type header
    body: Any | None = None               # Request body

@dataclass
class Response:
    status: int                           # HTTP status code
    headers: dict[str, str] | None = None # Response headers
    content_type: str | None = None       # Content-Type header
    body: Any | None = None               # Response body
```

### NDJSONWriter / GzipNDJSONWriter

Low-level writers for IR records.

```python
from traffic2openapi_playwright import NDJSONWriter, IRRecord, Request, Response

with NDJSONWriter("output.ndjson") as writer:
    record = IRRecord(
        request=Request(method="GET", path="/users"),
        response=Response(status=200),
    )
    writer.write(record)

# With gzip compression
with GzipNDJSONWriter("output.ndjson.gz", level=9) as writer:
    writer.write(record)
```

## Troubleshooting

### No requests captured

1. **Check filter_hosts** - Ensure it includes the domain you're testing
2. **Check exclude_paths** - Make sure you're not excluding the paths you need
3. **Attach before navigation** - Call `capture.attach(context)` before `page.goto()`

### Missing request/response bodies

1. **Check max_body_size** - Large bodies are skipped by default
2. **Check capture_content_types** - Only specified content types are captured
3. **Enable body capture** - Set `capture_request_body=True` and `capture_response_body=True`

### Headers missing

1. **Enable include_headers** - Set `include_headers=True`
2. **Check exclude_headers** - Sensitive headers are excluded by default

### File not created

1. **Call close()** - Always call `capture.close()` or use context manager
2. **Check file path** - Ensure the directory exists and is writable

### Large output files

1. **Enable gzip** - Use `.ndjson.gz` extension or set `gzip=True`
2. **Filter hosts** - Only capture traffic from your API
3. **Reduce body size** - Lower `max_body_size` if bodies aren't needed

## Examples

See the `examples/` directory:

- `basic.py` - Sync API examples
- `async_example.py` - Async API example

## Requirements

- Python 3.9+
- Playwright 1.40+

## License

MIT
