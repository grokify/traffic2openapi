# Adapters Overview

Adapters convert traffic from various sources to the IR format.

## Available Adapters

| Adapter | Source | Request Body | Response Body | Setup |
|---------|--------|:------------:|:-------------:|-------|
| [HAR](har.md) | Browser DevTools, proxies | Yes | Yes | Low |
| [Browser](browser.md) | Playwright, Cypress | Yes | Yes | Low |
| LoggingTransport | Go http.Client | Yes | Yes | Low |
| Proxy Captures | mitmproxy, Charles | Yes | Yes | Low-Medium |

## Choosing an Adapter

### For Development/Testing

- **HAR files**: Export from browser DevTools
- **LoggingTransport**: Capture from Go http.Client
- **Playwright/Cypress**: Capture during E2E tests

### For Production Traffic

- **Proxy captures**: mitmproxy, Charles Proxy
- **LoggingTransport**: Wrap production http.Client

### For Quick Discovery

- **HAR files**: Quick export from browser DevTools
- **Playwright**: Automated test traffic capture

## Fidelity Comparison

| Feature | HAR | Playwright | LoggingTransport | Proxy |
|---------|:---:|:----------:|:----------------:|:-----:|
| Request Headers | Yes | Yes | Yes | Yes |
| Request Body | Yes | Yes | Yes | Yes |
| Response Headers | Yes | Yes | Yes | Yes |
| Response Body | Yes | Yes | Yes | Yes |
| Query Params | Yes | Yes | Yes | Yes |
| Timing | Yes | Yes | Yes | Yes |
| Request ID | Varies | Yes | Yes | Varies |

## Common Workflow

1. **Capture**: Use an adapter to capture HTTP traffic
2. **Convert**: Convert to IR format (NDJSON)
3. **Analyze**: Run inference engine
4. **Generate**: Create OpenAPI spec

```bash
# Example with HAR
traffic2openapi convert har -i recording.har -o traffic.ndjson
traffic2openapi generate -i traffic.ndjson -o openapi.yaml
```

## Custom Adapters

You can create custom adapters by implementing the IR record format:

```go
// Convert your traffic format to IR records
func ConvertToIR(yourData YourFormat) *ir.IRRecord {
    return ir.NewRecord(
        ir.RequestMethod(yourData.Method),
        yourData.Path,
        yourData.StatusCode,
    ).
        SetHost(yourData.Host).
        SetRequestBody(yourData.RequestBody).
        SetResponseBody(yourData.ResponseBody)
}

// Write to IR file
provider := ir.NDJSON()
writer, _ := provider.NewWriter(ctx, "output.ndjson")
for _, data := range yourTraffic {
    writer.Write(ConvertToIR(data))
}
writer.Close()
```
