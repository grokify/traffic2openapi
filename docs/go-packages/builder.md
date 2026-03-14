# OpenAPI Builder

The `pkg/openapibuilder` package provides a fluent API for programmatically constructing OpenAPI 3.x specifications.

## Overview

```go
import "github.com/grokify/traffic2openapi/pkg/openapibuilder"

spec, err := openapibuilder.NewSpec(openapibuilder.Version310).
    Title("My API").
    Version("1.0.0").
    Build()
```

## Supported Versions

| Version | Constant | Description |
|---------|----------|-------------|
| 3.0.3 | `Version303` | Wide compatibility |
| 3.1.0 | `Version310` | JSON Schema 2020-12 (default) |
| 3.2.0 | `Version320` | Latest features |

## SpecBuilder

The entry point for building OpenAPI specifications.

```go
spec, err := openapibuilder.NewSpec(openapibuilder.Version310).
    Title("Pet Store API").
    Description("A sample Pet Store API").
    Version("1.0.0").
    TermsOfService("https://example.com/terms").
    Contact("API Support", "support@example.com", "https://example.com/support").
    License("MIT", "https://opensource.org/licenses/MIT").
    Server("https://api.example.com").
    ServerWithDescription("https://staging.example.com", "Staging server").
    Build()
```

## PathBuilder

Define API paths and their operations.

```go
spec, err := openapibuilder.NewSpec(openapibuilder.Version310).
    Title("User API").
    Version("1.0.0").
    Path("/users").
        Get().
            Summary("List all users").
            OperationID("listUsers").
            Tag("users").
            Response(200).Description("Success").JSON(usersSchema).Done().
        Done().
        Post().
            Summary("Create a user").
            OperationID("createUser").
            Tag("users").
            RequestBody().Required().JSON(createUserSchema).Done().
            Response(201).Description("Created").JSON(userSchema).Done().
        Done().
    Done().
    Path("/users/{userId}").
        Get().
            Summary("Get user by ID").
            OperationID("getUser").
            PathParam("userId").Description("User ID").Schema(openapibuilder.StringSchema()).Done().
            Response(200).Description("Success").JSON(userSchema).Done().
            Response(404).Description("Not found").Done().
        Done().
    Done().
    Build()
```

## SchemaBuilder

Build JSON Schema definitions for request/response bodies.

### Primitive Types

```go
// String
openapibuilder.StringSchema()
openapibuilder.StringSchema().Format("email")
openapibuilder.StringSchema().Format("uuid")
openapibuilder.StringSchema().MinLength(1).MaxLength(100)
openapibuilder.StringSchema().Pattern(`^[a-z]+$`)
openapibuilder.StringSchema().Enum("active", "inactive", "pending")

// Integer
openapibuilder.IntegerSchema()
openapibuilder.IntegerSchema().Format("int64")
openapibuilder.IntegerSchema().Minimum(0).Maximum(100)

// Number
openapibuilder.NumberSchema()
openapibuilder.NumberSchema().Format("double")

// Boolean
openapibuilder.BooleanSchema()
```

### Complex Types

```go
// Object
openapibuilder.ObjectSchema().
    Property("id", openapibuilder.IntegerSchema().Format("int64")).
    Property("name", openapibuilder.StringSchema()).
    Property("email", openapibuilder.StringSchema().Format("email")).
    Required("id", "name")

// Array
openapibuilder.ArraySchema(openapibuilder.StringSchema())
openapibuilder.ArraySchema(openapibuilder.RefSchema("User"))

// Reference
openapibuilder.RefSchema("User")  // References #/components/schemas/User

// Nullable (version-aware)
openapibuilder.StringSchema().Nullable()  // OpenAPI 3.0: nullable: true
                                           // OpenAPI 3.1+: type: ["string", "null"]
```

### Composition

```go
// AllOf
openapibuilder.AllOfSchema(
    openapibuilder.RefSchema("BaseModel"),
    openapibuilder.ObjectSchema().Property("extra", openapibuilder.StringSchema()),
)

// OneOf
openapibuilder.OneOfSchema(
    openapibuilder.RefSchema("Cat"),
    openapibuilder.RefSchema("Dog"),
)

// AnyOf
openapibuilder.AnyOfSchema(
    openapibuilder.StringSchema(),
    openapibuilder.IntegerSchema(),
)
```

## Components

Define reusable schemas, parameters, and security schemes.

```go
spec, err := openapibuilder.NewSpec(openapibuilder.Version310).
    Title("API").
    Version("1.0.0").
    Components().
        // Schemas
        Schema("User", openapibuilder.ObjectSchema().
            Property("id", openapibuilder.IntegerSchema()).
            Property("name", openapibuilder.StringSchema()).
            Required("id", "name")).
        Schema("Error", openapibuilder.ObjectSchema().
            Property("code", openapibuilder.IntegerSchema()).
            Property("message", openapibuilder.StringSchema())).

        // Security Schemes
        SecurityScheme("bearerAuth").BearerAuth().BearerFormat("JWT").Done().
        SecurityScheme("apiKey").APIKeyAuth("X-API-Key", "header").Done().
        SecurityScheme("oauth2").OAuth2().
            AuthorizationCodeFlow(
                "https://auth.example.com/authorize",
                "https://auth.example.com/token",
            ).
            Scope("read:users", "Read user data").
            Scope("write:users", "Write user data").
            Done().
        Done().
    Done().
    Build()
```

## Security

Apply security requirements to operations.

```go
// Global security (applies to all operations)
spec, err := openapibuilder.NewSpec(openapibuilder.Version310).
    Title("API").
    Version("1.0.0").
    Components().
        SecurityScheme("bearerAuth").BearerAuth().Done().
    Done().
    Security("bearerAuth").  // Apply globally
    Path("/users").
        Get().
            // Inherits global security
        Done().
    Done().
    Build()

// Per-operation security
spec, err := openapibuilder.NewSpec(openapibuilder.Version310).
    Title("API").
    Version("1.0.0").
    Path("/public").
        Get().
            Security().None().  // No auth required
        Done().
    Done().
    Path("/private").
        Get().
            Security().Scheme("bearerAuth").  // Requires bearer token
        Done().
    Done().
    Build()
```

## Parameters

Define query, path, header, and cookie parameters.

```go
openapibuilder.NewSpec(openapibuilder.Version310).
    Path("/users").
        Get().
            // Query parameters
            QueryParam("limit").
                Description("Maximum results").
                Schema(openapibuilder.IntegerSchema().Minimum(1).Maximum(100)).
                Done().
            QueryParam("offset").
                Description("Pagination offset").
                Schema(openapibuilder.IntegerSchema().Minimum(0)).
                Done().
            // Header parameter
            HeaderParam("X-Request-ID").
                Description("Request tracking ID").
                Schema(openapibuilder.StringSchema().Format("uuid")).
                Done().
        Done().
    Done().
    Path("/users/{userId}").
        Get().
            // Path parameter
            PathParam("userId").
                Description("User identifier").
                Schema(openapibuilder.StringSchema()).
                Done().
        Done().
    Done()
```

## Request Bodies

Define request body schemas.

```go
openapibuilder.NewSpec(openapibuilder.Version310).
    Path("/users").
        Post().
            RequestBody().
                Required().
                Description("User to create").
                JSON(openapibuilder.ObjectSchema().
                    Property("name", openapibuilder.StringSchema()).
                    Property("email", openapibuilder.StringSchema().Format("email")).
                    Required("name", "email")).
                Done().
            Response(201).Description("Created").Done().
        Done().
    Done()
```

## Responses

Define response schemas for different status codes.

```go
openapibuilder.NewSpec(openapibuilder.Version310).
    Path("/users/{userId}").
        Get().
            Response(200).
                Description("User found").
                JSON(openapibuilder.RefSchema("User")).
                Header("X-Request-ID", openapibuilder.StringSchema()).
                Done().
            Response(404).
                Description("User not found").
                JSON(openapibuilder.RefSchema("Error")).
                Done().
            Response(500).
                Description("Internal server error").
                JSON(openapibuilder.RefSchema("Error")).
                Done().
        Done().
    Done()
```

## Error Handling

The builder validates specifications at build time.

```go
spec, err := openapibuilder.NewSpec(openapibuilder.Version310).
    // Missing required title
    Version("1.0.0").
    Build()

if err != nil {
    // err: "title is required"
}
```

Use `MustBuild()` for panic-on-error behavior (useful in tests):

```go
spec := openapibuilder.NewSpec(openapibuilder.Version310).
    Title("API").
    Version("1.0.0").
    MustBuild()  // Panics if validation fails
```

## Full Example

```go
package main

import (
    "log"

    "github.com/grokify/traffic2openapi/pkg/openapi"
    "github.com/grokify/traffic2openapi/pkg/openapibuilder"
)

func main() {
    spec, err := openapibuilder.NewSpec(openapibuilder.Version310).
        Title("Pet Store API").
        Description("A sample API for managing pets").
        Version("1.0.0").
        Server("https://api.petstore.io/v1").
        Components().
            Schema("Pet", openapibuilder.ObjectSchema().
                Property("id", openapibuilder.IntegerSchema().Format("int64")).
                Property("name", openapibuilder.StringSchema()).
                Property("status", openapibuilder.StringSchema().Enum("available", "pending", "sold")).
                Required("id", "name")).
            Schema("Error", openapibuilder.ObjectSchema().
                Property("code", openapibuilder.IntegerSchema()).
                Property("message", openapibuilder.StringSchema())).
            SecurityScheme("bearerAuth").BearerAuth().BearerFormat("JWT").Done().
        Done().
        Security("bearerAuth").
        Path("/pets").
            Get().
                Summary("List all pets").
                OperationID("listPets").
                Tag("pets").
                QueryParam("limit").Schema(openapibuilder.IntegerSchema().Maximum(100)).Done().
                QueryParam("status").Schema(openapibuilder.StringSchema().Enum("available", "pending", "sold")).Done().
                Response(200).Description("A list of pets").
                    JSON(openapibuilder.ArraySchema(openapibuilder.RefSchema("Pet"))).Done().
            Done().
            Post().
                Summary("Create a pet").
                OperationID("createPet").
                Tag("pets").
                RequestBody().Required().JSON(openapibuilder.RefSchema("Pet")).Done().
                Response(201).Description("Pet created").JSON(openapibuilder.RefSchema("Pet")).Done().
                Response(400).Description("Invalid input").JSON(openapibuilder.RefSchema("Error")).Done().
            Done().
        Done().
        Path("/pets/{petId}").
            Get().
                Summary("Get a pet by ID").
                OperationID("getPet").
                Tag("pets").
                PathParam("petId").Description("Pet ID").Schema(openapibuilder.IntegerSchema().Format("int64")).Done().
                Response(200).Description("Pet found").JSON(openapibuilder.RefSchema("Pet")).Done().
                Response(404).Description("Pet not found").JSON(openapibuilder.RefSchema("Error")).Done().
            Done().
        Done().
        Build()

    if err != nil {
        log.Fatal(err)
    }

    // Write to file
    if err := openapi.WriteFile("petstore.yaml", spec); err != nil {
        log.Fatal(err)
    }
}
```
