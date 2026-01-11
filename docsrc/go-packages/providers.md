# Provider Pattern

Providers offer symmetric read/write access to IR records through a unified interface.

## Overview

The Provider pattern decouples IR record I/O from the underlying storage mechanism. All providers implement the same interface, making it easy to switch between file storage, cloud storage, or in-memory channels.

## Interfaces

### Provider

The base interface for path-based I/O:

```go
type Provider interface {
    // NewWriter creates a writer for the given path.
    NewWriter(ctx context.Context, path string) (IRWriter, error)

    // NewReader creates a reader for the given path.
    NewReader(ctx context.Context, path string) (IRReader, error)
}
```

### StreamProvider

Extended interface for io.Reader/io.Writer based I/O:

```go
type StreamProvider interface {
    // NewStreamWriter creates a writer that writes to the given io.Writer.
    NewStreamWriter(w io.Writer) IRWriter

    // NewStreamReader creates a reader that reads from the given io.Reader.
    NewStreamReader(r io.Reader) (IRReader, error)
}
```

### IRWriter and IRReader

Common interfaces for all readers and writers:

```go
type IRWriter interface {
    Write(record *IRRecord) error
    Flush() error
    Close() error
}

type IRReader interface {
    Read() (*IRRecord, error)  // Returns io.EOF when done
    Close() error
}
```

## Built-in Providers

### NDJSONProvider

Plain NDJSON file I/O:

```go
provider := ir.NDJSON()

// Write to file
writer, err := provider.NewWriter(ctx, "/path/to/records.ndjson")
if err != nil {
    return err
}
defer writer.Close()

for _, record := range records {
    if err := writer.Write(record); err != nil {
        return err
    }
}

// Read from file
reader, err := provider.NewReader(ctx, "/path/to/records.ndjson")
if err != nil {
    return err
}
defer reader.Close()

for {
    record, err := reader.Read()
    if err == io.EOF {
        break
    }
    if err != nil {
        return err
    }
    // process record
}
```

### GzipNDJSONProvider

Gzip-compressed NDJSON:

```go
// Default compression
provider := ir.GzipNDJSON()

// Custom compression level (1-9, or gzip constants)
provider := ir.GzipNDJSON(ir.WithGzipCompressionLevel(gzip.BestCompression))

// Write compressed file
writer, _ := provider.NewWriter(ctx, "/path/to/records.ndjson.gz")

// Read compressed file
reader, _ := provider.NewReader(ctx, "/path/to/records.ndjson.gz")
```

### StorageProvider

Cloud storage via omnistorage:

```go
import (
    "github.com/grokify/omnistorage/backend/file"
    "github.com/grokify/omnistorage/backend/s3"
    "github.com/grokify/traffic2openapi/pkg/ir"
)

// Local file backend
backend := file.New(file.Config{Root: "/data"})
defer backend.Close()

// Or S3 backend
backend, _ := s3.New(ctx, s3.Config{
    Bucket: "my-bucket",
    Region: "us-east-1",
})
defer backend.Close()

// Create provider
provider := ir.Storage(backend)

// Auto-detects format from extension:
// - .ndjson → plain NDJSON
// - .ndjson.gz → gzip-compressed NDJSON
writer, _ := provider.NewWriter(ctx, "traffic/records.ndjson.gz")
reader, _ := provider.NewReader(ctx, "traffic/records.ndjson.gz")
```

### ChannelProvider

In-memory Go channels for pipelines:

```go
// Create provider with buffered channel
provider := ir.Channel(ir.WithChannelProviderBufferSize(100))

// Or use existing channel
ch := make(chan *ir.IRRecord, 100)
provider := ir.Channel(ir.WithExistingChannel(ch))

// Writer sends to channel
writer, _ := provider.NewWriter(ctx, "")  // path ignored

// Reader receives from channel
reader, _ := provider.NewReader(ctx, "")  // path ignored

// Concurrent pipeline example
go func() {
    writer, _ := provider.NewWriter(ctx, "")
    for _, record := range records {
        writer.Write(record)
    }
    writer.Close()  // Closes channel, signals EOF to reader
}()

reader, _ := provider.NewReader(ctx, "")
for {
    record, err := reader.Read()
    if err == io.EOF {
        break
    }
    // process record
}
```

## Stream-based I/O

Providers that implement `StreamProvider` support io.Reader/io.Writer directly:

```go
provider := ir.NDJSON()

// Write to any io.Writer
var buf bytes.Buffer
writer := provider.NewStreamWriter(&buf)
writer.Write(record)
writer.Close()

// Read from any io.Reader
reader, _ := provider.NewStreamReader(&buf)
```

This is useful for:

- HTTP request/response bodies
- Network connections
- Pipes
- Testing

## Common Patterns

### Processing Pipeline

```go
// Source provider (e.g., S3)
srcBackend, _ := s3.New(ctx, srcConfig)
srcProvider := ir.Storage(srcBackend)

// Destination provider (e.g., local file)
dstProvider := ir.GzipNDJSON()

// Copy records
reader, _ := srcProvider.NewReader(ctx, "input.ndjson")
writer, _ := dstProvider.NewWriter(ctx, "/local/output.ndjson.gz")

for {
    record, err := reader.Read()
    if err == io.EOF {
        break
    }
    if err != nil {
        return err
    }
    if err := writer.Write(record); err != nil {
        return err
    }
}

reader.Close()
writer.Close()
```

### Multi-writer (Tee)

Write to multiple destinations:

```go
// Create writers
fileWriter, _ := fileProvider.NewWriter(ctx, "records.ndjson")
channelWriter, _ := channelProvider.NewWriter(ctx, "")

// Use MultiWriter
multiWriter := ir.NewMultiWriter(fileWriter, channelWriter)
defer multiWriter.Close()

for _, record := range records {
    multiWriter.Write(record)  // Writes to both
}
```

### With LoggingTransport

Capture HTTP traffic directly to a provider:

```go
provider := ir.GzipNDJSON()
writer, _ := provider.NewWriter(ctx, "traffic.ndjson.gz")
defer writer.Close()

transport := ir.NewLoggingTransport(http.DefaultTransport, writer)
client := &http.Client{Transport: transport}

// All requests through this client are captured
resp, _ := client.Get("https://api.example.com/users")
```

## Error Handling

All providers follow Go error handling conventions:

```go
writer, err := provider.NewWriter(ctx, path)
if err != nil {
    return fmt.Errorf("creating writer: %w", err)
}
defer writer.Close()

for _, record := range records {
    if err := writer.Write(record); err != nil {
        return fmt.Errorf("writing record: %w", err)
    }
}

if err := writer.Close(); err != nil {
    return fmt.Errorf("closing writer: %w", err)
}
```

Context cancellation is supported:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

writer, err := provider.NewWriter(ctx, path)
// Returns error if context is cancelled
```
