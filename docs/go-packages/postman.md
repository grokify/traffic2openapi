# Postman Converter

The `postman` package converts Postman Collection v2.1 files to IR format.

## Overview

Postman collections are a popular way to document and share API definitions. This package extracts HTTP traffic information from collections and converts it to the IR format for OpenAPI generation.

## Features

- Convert Postman Collection v2.1 format
- Extract request method, URL, headers, and body
- Parse query parameters and path parameters
- Extract authentication settings (bearer, basic, API key)
- Preserve collection metadata (name, description, version)
- Handle folder hierarchy

## Installation

```go
import "github.com/grokify/traffic2openapi/pkg/postman"
```

## Usage

### Basic Conversion

```go
package main

import (
    "context"
    "github.com/grokify/traffic2openapi/pkg/postman"
    "github.com/grokify/traffic2openapi/pkg/ir"
)

func main() {
    ctx := context.Background()

    // Convert Postman collection to IR records
    records, metadata, err := postman.ConvertFile(ctx, "collection.json", nil)
    if err != nil {
        panic(err)
    }

    // Create batch with metadata
    batch := ir.NewBatchWithMetadata(records, metadata)

    // Write to NDJSON
    provider := ir.NDJSON()
    writer, _ := provider.NewWriter(ctx, "traffic.ndjson")
    for _, record := range records {
        writer.Write(record)
    }
    writer.Close()
}
```

### With Options

```go
opts := &postman.ConvertOptions{
    IncludeDisabled: false,  // Skip disabled requests
    BaseURL:         "https://api.example.com",  // Override base URL
}

records, metadata, err := postman.ConvertFile(ctx, "collection.json", opts)
```

### From Reader

```go
file, _ := os.Open("collection.json")
defer file.Close()

records, metadata, err := postman.ConvertReader(ctx, file, nil)
```

## Supported Features

### Request Elements

| Element | Supported | Notes |
|---------|:---------:|-------|
| Method | Yes | GET, POST, PUT, PATCH, DELETE, etc. |
| URL | Yes | Includes path and host |
| Headers | Yes | All headers preserved |
| Query Params | Yes | Extracted from URL |
| Path Params | Yes | Variables like `:id` → `{id}` |
| Body (raw) | Yes | JSON, XML, text |
| Body (form-data) | Yes | Multipart forms |
| Body (urlencoded) | Yes | Form submissions |

### Authentication

| Auth Type | Supported |
|-----------|:---------:|
| Bearer Token | Yes |
| Basic Auth | Yes |
| API Key | Yes |
| OAuth2 | Partial |

### Metadata

| Field | Mapped To |
|-------|-----------|
| Collection name | APIMetadata.Name |
| Collection description | APIMetadata.Description |
| Collection version | APIMetadata.Version |
| Item name | Documentation.Summary |
| Item description | Documentation.Description |

## CLI Usage

```bash
# Convert Postman collection to IR format
traffic2openapi convert postman -i collection.json -o traffic.ndjson

# Then generate OpenAPI spec
traffic2openapi generate -i traffic.ndjson -o openapi.yaml
```

## Example

Given a Postman collection:

```json
{
  "info": {
    "name": "Pet Store API",
    "description": "API for managing pets"
  },
  "item": [
    {
      "name": "List Pets",
      "request": {
        "method": "GET",
        "url": "https://api.example.com/pets"
      }
    },
    {
      "name": "Get Pet",
      "request": {
        "method": "GET",
        "url": "https://api.example.com/pets/:petId"
      }
    }
  ]
}
```

The converter produces IR records with:
- Proper path parameter detection (`:petId` → `{petId}`)
- Documentation from item names and descriptions
- API metadata from collection info

## Limitations

- Response bodies are not included (Postman collections typically don't contain actual responses)
- Environment variables are not resolved (use `BaseURL` option to override)
- Pre-request scripts and tests are ignored
