# HAR Adapter

Convert HAR (HTTP Archive) files to IR format.

## Overview

HAR is a standard format supported by:

- **Browser DevTools**: Chrome, Firefox, Safari (Network tab → Save as HAR)
- **Playwright**: Built-in `recordHar` option
- **Charles Proxy**: File → Export Session as HAR
- **Fiddler**: File → Export → HTTPArchive
- **mitmproxy**: `mitmdump --save-stream-file`
- **Postman**: Collection export

## CLI Usage

```bash
# Convert a single HAR file
traffic2openapi convert har -i recording.har -o traffic.ndjson

# Convert multiple HAR files from a directory
traffic2openapi convert har -i ./har-files/ -o traffic.ndjson

# Filter by host
traffic2openapi convert har -i recording.har -o traffic.ndjson --host api.example.com

# Filter by method
traffic2openapi convert har -i recording.har -o traffic.ndjson --method POST

# Exclude headers from output
traffic2openapi convert har -i recording.har -o traffic.ndjson --headers=false
```

## Capturing HAR Files

### Browser DevTools

1. Open DevTools (F12)
2. Go to Network tab
3. Perform your API operations
4. Right-click in the network panel → "Save all as HAR with content"

### Playwright

```typescript
const context = await browser.newContext({
  recordHar: { path: 'traffic.har', content: 'embed' }
});

await page.goto('https://api.example.com');
// ... perform actions ...

await context.close();  // HAR file written here
```

### Charles Proxy

1. File → Start Recording
2. Perform API operations
3. File → Export Session → HTTPArchive (.har)

## Go Package Usage

```go
import "github.com/grokify/traffic2openapi/pkg/adapters/har"

// Read HAR file
reader := har.NewReader()
records, err := reader.ReadFile("recording.har")

// Read directory of HAR files
records, err := reader.ReadDir("./har-files/")

// Parse HAR and filter
h, err := har.ParseFile("recording.har")
apiEntries := har.FilterByHost(h, "api.example.com")
postEntries := har.FilterByMethod(h, "POST")
```

## Configuration

### Header Filtering

By default, sensitive headers are excluded:

- `authorization`
- `cookie` / `set-cookie`
- `x-api-key`
- `x-auth-token`
- `x-csrf-token`
- `proxy-authorization`

Customize:

```go
reader := har.NewReader()
reader.Converter.IncludeHeaders = true
reader.Converter.FilterHeaders = []string{"authorization", "x-api-key"}
```

### Cookie Handling

```go
reader.Converter.IncludeCookies = false  // Exclude cookies (default)
```

## Full Workflow

```bash
# 1. Export HAR from browser
# 2. Convert to IR
traffic2openapi convert har -i browser-recording.har -o traffic.ndjson

# 3. Generate OpenAPI
traffic2openapi generate -i traffic.ndjson -o openapi.yaml \
    --title "My API" \
    --server https://api.example.com
```
