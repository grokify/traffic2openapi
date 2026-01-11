# pkg/ir API Reference

The `ir` package provides types and utilities for working with Intermediate Representation data.

## Types

### IRRecord

Represents a single HTTP request/response capture.

```go
type IRRecord struct {
    ID         *string           `json:"id,omitempty"`
    Timestamp  *string           `json:"timestamp,omitempty"`
    Source     *IRRecordSource   `json:"source,omitempty"`
    Request    Request           `json:"request"`
    Response   Response          `json:"response"`
    DurationMs *float64          `json:"durationMs,omitempty"`
}
```

### Request

HTTP request details.

```go
type Request struct {
    Method       RequestMethod     `json:"method"`
    Scheme       *string           `json:"scheme,omitempty"`
    Host         *string           `json:"host,omitempty"`
    Path         string            `json:"path"`
    PathTemplate *string           `json:"pathTemplate,omitempty"`
    PathParams   map[string]string `json:"pathParams,omitempty"`
    Query        map[string]string `json:"query,omitempty"`
    Headers      map[string]string `json:"headers,omitempty"`
    ContentType  *string           `json:"contentType,omitempty"`
    Body         interface{}       `json:"body,omitempty"`
}
```

### Response

HTTP response details.

```go
type Response struct {
    Status      int               `json:"status"`
    Headers     map[string]string `json:"headers,omitempty"`
    ContentType *string           `json:"contentType,omitempty"`
    Body        interface{}       `json:"body,omitempty"`
}
```

### RequestMethod

HTTP method enum.

```go
type RequestMethod string

const (
    RequestMethodGET     RequestMethod = "GET"
    RequestMethodPOST    RequestMethod = "POST"
    RequestMethodPUT     RequestMethod = "PUT"
    RequestMethodPATCH   RequestMethod = "PATCH"
    RequestMethodDELETE  RequestMethod = "DELETE"
    RequestMethodHEAD    RequestMethod = "HEAD"
    RequestMethodOPTIONS RequestMethod = "OPTIONS"
)
```

## Interfaces

### Provider

Base interface for path-based I/O.

```go
type Provider interface {
    NewWriter(ctx context.Context, path string) (IRWriter, error)
    NewReader(ctx context.Context, path string) (IRReader, error)
}
```

### StreamProvider

Extended interface for stream-based I/O.

```go
type StreamProvider interface {
    NewStreamWriter(w io.Writer) IRWriter
    NewStreamReader(r io.Reader) (IRReader, error)
}
```

### IRWriter

Writer interface for IR records.

```go
type IRWriter interface {
    Write(record *IRRecord) error
    Flush() error
    Close() error
}
```

### IRReader

Reader interface for IR records.

```go
type IRReader interface {
    Read() (*IRRecord, error)
    Close() error
}
```

## Functions

### NewRecord

Create a new IR record with builder methods.

```go
func NewRecord(method RequestMethod, path string, status int) *IRRecord
```

Example:

```go
record := ir.NewRecord(ir.RequestMethodGET, "/users", 200).
    SetID("req-001").
    SetHost("api.example.com").
    SetRequestBody(requestBody).
    SetResponseBody(responseBody).
    SetDuration(45.2)
```

### NDJSON

Create an NDJSON provider.

```go
func NDJSON(opts ...ProviderOption) *NDJSONProvider
```

### GzipNDJSON

Create a gzip-compressed NDJSON provider.

```go
func GzipNDJSON(opts ...GzipNDJSONOption) *GzipNDJSONProvider

func WithGzipCompressionLevel(level int) GzipNDJSONOption
```

### Storage

Create a storage provider using an omnistorage backend.

```go
func Storage(backend omnistorage.Backend, opts ...StorageProviderOption) *StorageProvider
```

### Channel

Create a channel provider for in-memory I/O.

```go
func Channel(opts ...ChannelProviderOption) *ChannelProvider

func WithChannelProviderBufferSize(size int) ChannelProviderOption
func WithExistingChannel(ch chan *IRRecord) ChannelProviderOption
```

### NewLoggingTransport

Create an http.RoundTripper that logs traffic.

```go
func NewLoggingTransport(rt http.RoundTripper, writer IRWriter, opts ...LoggingOption) *LoggingTransport

func WithFilterHeaders(headers ...string) LoggingOption
func WithSkipPaths(paths ...string) LoggingOption
func WithAllowMethods(methods ...string) LoggingOption
func WithSkipStatusCodes(codes ...int) LoggingOption
func WithRequestIDHeaders(headers ...string) LoggingOption
func WithErrorHandler(handler func(error)) LoggingOption
```

### ReadFile

Read IR records from a file.

```go
func ReadFile(path string) ([]*IRRecord, error)
```

### WriteFile

Write IR records to a file.

```go
func WriteFile(path string, records []*IRRecord) error
```

### ReadDir

Read IR records from a directory.

```go
func ReadDir(path string) ([]*IRRecord, error)
```

## Example

```go
package main

import (
    "context"
    "fmt"
    "io"
    "net/http"

    "github.com/grokify/traffic2openapi/pkg/ir"
)

func main() {
    ctx := context.Background()

    // Create provider
    provider := ir.GzipNDJSON()

    // Write records
    writer, _ := provider.NewWriter(ctx, "traffic.ndjson.gz")

    record := ir.NewRecord(ir.RequestMethodGET, "/users", 200).
        SetID("req-001").
        SetHost("api.example.com")

    writer.Write(record)
    writer.Close()

    // Read records
    reader, _ := provider.NewReader(ctx, "traffic.ndjson.gz")
    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        }
        fmt.Printf("Request: %s %s\n", record.Request.Method, record.Request.Path)
    }
    reader.Close()

    // Use LoggingTransport
    writer2, _ := provider.NewWriter(ctx, "live-traffic.ndjson.gz")
    transport := ir.NewLoggingTransport(http.DefaultTransport, writer2,
        ir.WithFilterHeaders("Authorization"),
    )
    client := &http.Client{Transport: transport}

    resp, _ := client.Get("https://api.example.com/users")
    resp.Body.Close()
    writer2.Close()
}
```
