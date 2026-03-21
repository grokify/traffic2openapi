# Tasks

Current development tasks and priorities for traffic2openapi.

## In Progress

(none)

## Completed (Recent)

- [x] Postman Collection v2.1 to IR converter
  - Full lossless conversion with variable resolution
  - Folder structure → tags mapping
  - Auth conversion (Bearer, Basic, API Key, OAuth2)
  - CLI command: `convert postman`
- [x] Extended IR schema with documentation fields
  - operationId, summary, description, tags, deprecated, externalDocs
  - API-level metadata (title, description, version, contact, license)
- [x] OpenAPI generator uses IR documentation fields
  - Uses IR values with fallback to generated values
  - Supports tags, external docs, deprecation
- [x] Integration and E2E tests for Postman converter
  - PetStore sample collection test fixture
  - Integration tests: collection loading, endpoints, variable resolution
  - End-to-end tests: Postman → IR → OpenAPI pipeline
  - Round-trip tests: IR write/read, NDJSON streaming

## Backlog

### High Priority

- [x] Integration tests for Postman converter
- [ ] Insomnia collection import
- [ ] More comprehensive test fixtures for edge cases

### Medium Priority

- [ ] Swagger 2.0 / OpenAPI 3.0 import to IR
- [ ] Charles Proxy session import
- [ ] HAR converter improvements (auth extraction)

### Low Priority

- [ ] Bruno collection import
- [ ] HTTPie session import

## Documentation

- [x] docs/adapters/postman.md
- [x] docs/adapters/overview.md - Added Postman
- [x] docs/cli/commands.md - Added convert postman command
- [x] README.md - Added Postman adapter info
- [x] ROADMAP.md - Marked Postman complete
- [ ] Migration guide from Postman to OpenAPI
- [ ] Video tutorial for Postman import workflow

## Testing

- [x] End-to-end test: Postman → IR → OpenAPI round-trip
- [x] Test with PetStore sample collection
- [ ] Test with more real-world Postman collections
- [ ] Fuzz testing for variable resolution

## See Also

- [ROADMAP.md](ROADMAP.md) - Feature roadmap and version planning
- [CHANGELOG.md](CHANGELOG.md) - Release history
