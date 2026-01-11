# CLI Commands

The `traffic2openapi` CLI provides commands for converting traffic and generating OpenAPI specs.

## Installation

```bash
go install github.com/grokify/traffic2openapi/cmd/traffic2openapi@latest
```

## Commands Overview

| Command | Description |
|---------|-------------|
| `generate` | Generate OpenAPI spec from IR files |
| `convert har` | Convert HAR files to IR format |
| `validate` | Validate IR files |
| `site` | Generate static HTML documentation site |

## generate

Generate OpenAPI specification from IR files.

### Usage

```bash
traffic2openapi generate -i <input> -o <output> [flags]
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--input` | `-i` | (required) | Input file or directory |
| `--output` | `-o` | stdout | Output file path |
| `--version` | `-v` | `3.1` | OpenAPI version: 3.0, 3.1, or 3.2 |
| `--format` | `-f` | auto | Output format: json or yaml |
| `--title` | | `Generated API` | API title |
| `--description` | | | API description |
| `--api-version` | | `1.0.0` | API version |
| `--server` | | | Server URL (repeatable) |
| `--include-errors` | | `true` | Include 4xx/5xx responses |

### Examples

```bash
# Basic generation
traffic2openapi generate -i traffic.ndjson -o openapi.yaml

# OpenAPI 3.0 in JSON
traffic2openapi generate -i traffic.ndjson -o api.json --version 3.0 --format json

# With metadata
traffic2openapi generate -i traffic.ndjson -o openapi.yaml \
    --title "User Service API" \
    --description "API for managing users" \
    --api-version "2.0.0" \
    --server https://api.example.com \
    --server https://staging.example.com

# From directory
traffic2openapi generate -i ./traffic-logs/ -o openapi.yaml
```

## convert har

Convert HAR (HTTP Archive) files to IR format.

### Usage

```bash
traffic2openapi convert har -i <input> -o <output> [flags]
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--input` | `-i` | (required) | HAR file or directory |
| `--output` | `-o` | (required) | Output IR file |
| `--host` | | | Filter by host |
| `--method` | | | Filter by HTTP method |
| `--headers` | | `true` | Include headers |

### Examples

```bash
# Single file
traffic2openapi convert har -i recording.har -o traffic.ndjson

# Directory of HAR files
traffic2openapi convert har -i ./har-files/ -o traffic.ndjson

# Filter by host
traffic2openapi convert har -i recording.har -o traffic.ndjson --host api.example.com

# Filter by method
traffic2openapi convert har -i recording.har -o traffic.ndjson --method POST

# Without headers
traffic2openapi convert har -i recording.har -o traffic.ndjson --headers=false
```

## validate

Validate IR files against the schema.

### Usage

```bash
traffic2openapi validate <path> [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--verbose` | `false` | Show detailed validation output |

### Examples

```bash
# Validate file
traffic2openapi validate traffic.ndjson

# Validate directory
traffic2openapi validate ./logs/

# Verbose output
traffic2openapi validate traffic.ndjson --verbose
```

## Common Workflows

### HAR to OpenAPI

```bash
# 1. Convert HAR
traffic2openapi convert har -i browser.har -o traffic.ndjson

# 2. Generate spec
traffic2openapi generate -i traffic.ndjson -o openapi.yaml
```

### Combining Multiple Sources

```bash
# Combine multiple HAR files
traffic2openapi convert har -i ./har-files/ -o combined.ndjson

# Generate from combined
traffic2openapi generate -i combined.ndjson -o openapi.yaml \
    --title "My API" \
    --server https://api.example.com
```

### Pipeline

```bash
# Convert and generate in one pipeline
traffic2openapi convert har -i recording.har -o - | \
    traffic2openapi generate -i - -o openapi.yaml
```

## site

Generate a static HTML documentation site from IR traffic logs.

### Usage

```bash
traffic2openapi site -i <input_file_or_dir> -o <output_dir> [flags]
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--input` | `-i` | (required) | Input IR file (`.ndjson`, `.json`) or directory |
| `--output` | `-o` | `./site/` | Output directory for generated HTML files |
| `--title` | | `API Traffic Documentation` | Site title |
| `--base-url` | | | Base URL for links (e.g., `/docs/api/`) |

### Examples

```bash
# Basic site generation
traffic2openapi site -i traffic.ndjson -o ./site/

# With custom title
traffic2openapi site -i traffic.ndjson -o ./docs/ --title "My API Documentation"

# From directory of IR files
traffic2openapi site -i ./logs/ -o ./site/

# With base URL for hosting under a subdirectory
traffic2openapi site -i traffic.ndjson -o ./site/ --base-url /api-docs/
```

### Features

The generated site includes:

- **Index page**: Lists all endpoints with method badges, request counts, and status codes
- **Endpoint pages**: Detailed view of each endpoint grouped by HTTP status code
- **Two views per status code**:
    - **Deduped view**: Collapsed view showing all seen parameter values (e.g., `userId: 123, 456`)
    - **Distinct view**: Individual requests with full request/response details
- **Path template detection**: Automatically detects path parameters like `/users/{userId}`
- **Light/dark mode**: Theme toggle with localStorage persistence
- **Copy buttons**: One-click copying of JSON bodies
- **Syntax highlighting**: Color-coded JSON for better readability
- **Responsive design**: Works on desktop and mobile

### Output Structure

```
./site/
├── index.html              # Endpoint listing
├── get-users.html          # GET /users endpoint page
├── get-users-userid.html   # GET /users/{userId} endpoint page
├── post-users.html         # POST /users endpoint page
└── assets/
    ├── style.css           # Light/dark theme styles
    └── script.js           # Theme toggle, copy buttons, highlighting
```
