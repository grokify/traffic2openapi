# LoggingTransport

LoggingTransport is an `http.RoundTripper` that captures HTTP traffic from Go's `http.Client`.

## Overview

LoggingTransport wraps any `http.RoundTripper` (typically `http.DefaultTransport`) and records all HTTP request/response pairs to an `IRWriter`.

```go
transport := ir.NewLoggingTransport(http.DefaultTransport, writer)
client := &http.Client{Transport: transport}

// All requests through this client are automatically logged
resp, err := client.Get("https://api.example.com/users")
```

## Basic Usage

```go
package main

import (
    "context"
    "net/http"

    "github.com/grokify/traffic2openapi/pkg/ir"
)

func main() {
    ctx := context.Background()

    // Create a provider and writer
    provider := ir.GzipNDJSON()
    writer, _ := provider.NewWriter(ctx, "traffic.ndjson.gz")
    defer writer.Close()

    // Wrap the default transport
    transport := ir.NewLoggingTransport(http.DefaultTransport, writer)

    // Create client with logging transport
    client := &http.Client{Transport: transport}

    // Make requests - they're automatically captured
    resp, _ := client.Get("https://api.example.com/users")
    defer resp.Body.Close()

    // POST with body
    client.Post("https://api.example.com/users",
        "application/json",
        strings.NewReader(`{"name": "Alice"}`))
}
```

## Configuration Options

### Header Filtering

Exclude sensitive headers from captured records:

```go
transport := ir.NewLoggingTransport(http.DefaultTransport, writer,
    ir.WithFilterHeaders("Authorization", "Cookie", "X-API-Key"),
)
```

Default filtered headers:

- `Authorization`
- `Cookie`
- `Set-Cookie`
- `X-API-Key`
- `X-Auth-Token`

### Path Filtering

Skip logging for specific paths:

```go
transport := ir.NewLoggingTransport(http.DefaultTransport, writer,
    ir.WithSkipPaths("/health", "/metrics", "/ping"),
)
```

### Method Filtering

Only log specific HTTP methods:

```go
transport := ir.NewLoggingTransport(http.DefaultTransport, writer,
    ir.WithAllowMethods("GET", "POST", "PUT", "DELETE"),
)
```

### Status Code Filtering

Skip logging for specific status codes:

```go
transport := ir.NewLoggingTransport(http.DefaultTransport, writer,
    ir.WithSkipStatusCodes(404, 500, 502, 503),
)
```

### Request ID Headers

Extract request IDs from headers:

```go
transport := ir.NewLoggingTransport(http.DefaultTransport, writer,
    ir.WithRequestIDHeaders("X-Request-ID", "X-Correlation-ID"),
)
```

If no header is found, a UUID is generated.

### Error Handler

Custom error handling:

```go
transport := ir.NewLoggingTransport(http.DefaultTransport, writer,
    ir.WithErrorHandler(func(err error) {
        log.Printf("logging error: %v", err)
    }),
)
```

## Full Example with Options

```go
transport := ir.NewLoggingTransport(http.DefaultTransport, writer,
    // Security
    ir.WithFilterHeaders("Authorization", "Cookie", "X-API-Key"),

    // Skip non-API paths
    ir.WithSkipPaths("/health", "/metrics", "/_next"),

    // Only log mutations
    ir.WithAllowMethods("POST", "PUT", "PATCH", "DELETE"),

    // Skip error responses
    ir.WithSkipStatusCodes(500, 502, 503),

    // Correlation
    ir.WithRequestIDHeaders("X-Request-ID", "X-Trace-ID"),

    // Error handling
    ir.WithErrorHandler(func(err error) {
        slog.Error("traffic logging failed", "error", err)
    }),
)

client := &http.Client{Transport: transport}
```

## Captured Data

Each request/response pair is captured as an `IRRecord`:

| Field | Captured |
|-------|----------|
| Request method | Yes |
| Request path | Yes |
| Request query params | Yes |
| Request headers | Yes (filtered) |
| Request body | Yes (JSON parsed) |
| Response status | Yes |
| Response headers | Yes |
| Response body | Yes (JSON parsed) |
| Duration | Yes |
| Timestamp | Yes |
| Request ID | Yes (from header or generated) |

## Integration Patterns

### With Channel Provider for Real-time Processing

```go
provider := ir.Channel(ir.WithChannelProviderBufferSize(1000))

// Writer for transport
writer, _ := provider.NewWriter(ctx, "")

// Reader for processing
reader, _ := provider.NewReader(ctx, "")

// Process records in background
go func() {
    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        }
        // Real-time processing: metrics, alerts, etc.
        processRecord(record)
    }
}()

// HTTP client with logging
transport := ir.NewLoggingTransport(http.DefaultTransport, writer)
client := &http.Client{Transport: transport}
```

### With MultiWriter for Multiple Destinations

```go
// Write to file and channel simultaneously
fileWriter, _ := fileProvider.NewWriter(ctx, "traffic.ndjson.gz")
channelWriter, _ := channelProvider.NewWriter(ctx, "")

multiWriter := ir.NewMultiWriter(fileWriter, channelWriter)

transport := ir.NewLoggingTransport(http.DefaultTransport, multiWriter)
client := &http.Client{Transport: transport}
```

### With Async Writer for Non-blocking Logging

```go
// Async writer doesn't block HTTP requests
asyncWriter := ir.NewAsyncNDJSONWriter(baseWriter, 1000)
defer asyncWriter.Close()

transport := ir.NewLoggingTransport(http.DefaultTransport, asyncWriter)
client := &http.Client{Transport: transport}
```

## Testing

LoggingTransport works well with `httptest`:

```go
func TestAPI(t *testing.T) {
    // Setup test server
    server := httptest.NewServer(http.HandlerFunc(handler))
    defer server.Close()

    // Capture traffic
    var records []*ir.IRRecord
    sliceWriter := ir.NewSliceWriter(&records)

    transport := ir.NewLoggingTransport(http.DefaultTransport, sliceWriter)
    client := &http.Client{Transport: transport}

    // Make requests
    client.Get(server.URL + "/users")

    // Verify captured traffic
    if len(records) != 1 {
        t.Errorf("expected 1 record, got %d", len(records))
    }
}
```

## Thread Safety

LoggingTransport is safe for concurrent use. Multiple goroutines can share the same client:

```go
client := &http.Client{Transport: transport}

var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        client.Get("https://api.example.com/users")
    }()
}
wg.Wait()
```

The underlying `IRWriter` must also be thread-safe. All built-in writers (NDJSONWriter, GzipNDJSONWriter, ChannelWriter) are thread-safe.
