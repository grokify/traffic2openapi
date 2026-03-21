# Postman Adapter

Convert Postman Collection v2.1 files to IR format.

## Overview

The Postman adapter provides lossless conversion from Postman Collection v2.1 format to IR. It preserves:

- **Request/response data**: Method, URL, headers, body, query params
- **Documentation**: Collection name, request descriptions, folder structure
- **Variables**: Collection and environment variable resolution
- **Authentication**: Bearer, Basic, API Key, OAuth2
- **Saved responses**: Each saved response becomes a separate IR record

## CLI Usage

```bash
# Convert a Postman collection
traffic2openapi convert postman -i collection.json -o traffic.ndjson

# With variable substitution
traffic2openapi convert postman -i collection.json -o traffic.ndjson \
    --var baseUrl=https://api.example.com \
    --var apiKey=sk-xxx

# With base URL override
traffic2openapi convert postman -i collection.json -o traffic.ndjson \
    --base-url https://api.example.com

# Filter by host or method
traffic2openapi convert postman -i collection.json -o traffic.ndjson \
    --host api.example.com \
    --method POST

# Output as batch JSON
traffic2openapi convert postman -i collection.json -o traffic.json --format json

# Exclude certain headers
traffic2openapi convert postman -i collection.json -o traffic.ndjson \
    --filter-headers "X-Internal-*,X-Debug-*"

# Skip authentication headers
traffic2openapi convert postman -i collection.json -o traffic.ndjson --auth=false
```

## Exporting from Postman

### Postman Desktop App

1. Select your collection in the sidebar
2. Click the three-dot menu (...)
3. Select **Export**
4. Choose **Collection v2.1** format
5. Save the JSON file

### Postman Web

1. Open your collection
2. Click **Export** in the toolbar
3. Select **Collection v2.1** format
4. Download the JSON file

### Via API

```bash
# Export collection via Postman API
curl --location --request GET 'https://api.getpostman.com/collections/{collection_id}' \
    --header 'X-Api-Key: {{your-postman-api-key}}' \
    --output collection.json
```

## Go Package Usage

```go
import "github.com/grokify/traffic2openapi/pkg/postman"

// Basic conversion
result, err := postman.ConvertFile("collection.json")
records := result.Records

// With options
conv := postman.NewConverter(
    postman.WithBaseURL("https://api.example.com"),
    postman.WithVariable("apiKey", "sk-xxx"),
    postman.WithVariable("userId", "12345"),
    postman.WithHeaderFilter([]string{"X-Debug-*"}),
)

result, err := conv.ConvertFile("collection.json")

// Access metadata
fmt.Printf("API: %s v%s\n", result.Metadata.Title, result.Metadata.APIVersion)
fmt.Printf("Tags: %v\n", result.TagDefinitions)

// Convert to batch format
batch, err := postman.ConvertFileToBatch("collection.json")
```

## Variable Resolution

The adapter resolves Postman variables in the format `{{variableName}}`:

### Collection Variables

Variables defined in the collection are automatically resolved:

```json
{
  "variable": [
    { "key": "baseUrl", "value": "https://api.example.com" },
    { "key": "apiVersion", "value": "v1" }
  ]
}
```

### CLI Variables

Override or add variables via CLI:

```bash
traffic2openapi convert postman -i collection.json -o traffic.ndjson \
    --var baseUrl=https://staging.example.com \
    --var apiKey=test-key
```

### Programmatic Variables

```go
conv := postman.NewConverter(
    postman.WithVariables(map[string]string{
        "baseUrl": "https://api.example.com",
        "apiKey":  "sk-xxx",
    }),
)
```

## Folder to Tags Mapping

Postman folder structure is automatically converted to OpenAPI tags:

```
Collection
├── Users/              → tag: "Users"
│   ├── List Users      → tags: ["Users"]
│   ├── Get User        → tags: ["Users"]
│   └── Admin/          → tag: "Users > Admin"
│       └── Delete User → tags: ["Users > Admin"]
└── Posts/              → tag: "Posts"
    └── Create Post     → tags: ["Posts"]
```

Tag definitions include folder descriptions when available.

## Authentication Handling

The adapter extracts authentication and adds appropriate headers:

| Postman Auth Type | IR Output |
|-------------------|-----------|
| Bearer Token | `Authorization: Bearer <token>` |
| Basic Auth | `Authorization: Basic <base64>` |
| API Key (header) | Custom header with key/value |
| API Key (query) | Query parameter |
| OAuth2 | `Authorization: Bearer <access_token>` |

Skip auth processing with `--auth=false` or `WithoutAuth()` option.

## Saved Responses

Each saved response in Postman creates a separate IR record:

```go
// Postman request with 3 saved responses:
// - 200 OK (success)
// - 404 Not Found
// - 500 Server Error
//
// Produces 3 IR records, one for each response
```

This enables generating OpenAPI specs with multiple response codes per endpoint.

## Configuration

### Header Filtering

Exclude headers matching patterns:

```go
conv := postman.NewConverter(
    postman.WithHeaderFilter([]string{
        "X-Debug-*",
        "X-Internal-*",
        "Postman-*",
    }),
)
```

Or exclude all headers:

```go
conv := postman.NewConverter(postman.WithoutHeaders())
```

### Including Disabled Items

By default, disabled requests are skipped. To include them:

```go
conv := postman.NewConverter(postman.WithDisabledItems())
```

### Operation IDs

By default, operation IDs are generated from request names. Disable with:

```go
conv := postman.NewConverter(postman.WithoutIDs())
```

## Full Workflow

```bash
# 1. Export collection from Postman

# 2. Convert to IR with variables
traffic2openapi convert postman \
    -i my-api-collection.json \
    -o traffic.ndjson \
    --var baseUrl=https://api.example.com \
    --var apiKey="${API_KEY}"

# 3. Generate OpenAPI spec
traffic2openapi generate \
    -i traffic.ndjson \
    -o openapi.yaml \
    --title "My API" \
    --server https://api.example.com

# 4. Validate the spec
traffic2openapi validate-spec openapi.yaml
```

## Comparison with Other Adapters

| Feature | Postman | HAR | Playwright |
|---------|:-------:|:---:|:----------:|
| Documentation (descriptions) | Yes | No | No |
| Folder structure → Tags | Yes | No | No |
| Variable resolution | Yes | No | No |
| Multiple responses per request | Yes | No | Yes |
| Real traffic timing | No | Yes | Yes |
| Request ordering | Yes | Yes | Yes |

## Limitations

- **Variables**: Unresolved variables remain as `{{varName}}` in the output
- **Pre-request scripts**: Scripts are not executed
- **Tests**: Test scripts are not evaluated
- **Dynamic values**: Postman dynamic variables (e.g., `{{$randomInt}}`) are not expanded
- **OAuth2 flows**: Only access token is extracted, not full flow configuration

## See Also

- [Adapters Overview](overview.md)
- [HAR Adapter](har.md)
- [IR Format Documentation](../concepts/ir-format.md)
