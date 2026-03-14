// Example: Using openapibuilder to construct an OpenAPI spec programmatically
//
// This example demonstrates how to use the fluent builder API to create
// OpenAPI specifications without analyzing traffic.
package main

import (
	"fmt"
	"log"

	"github.com/grokify/traffic2openapi/pkg/openapibuilder"
)

func main() {
	// Build a Pet Store API spec using the fluent builder
	spec, err := openapibuilder.NewSpec(openapibuilder.Version310).
		Title("Pet Store API").
		Version("1.0.0").
		Description("A sample Pet Store API demonstrating openapibuilder").
		Contact("API Support", "https://support.petstore.io", "support@petstore.io").
		License("Apache 2.0", "https://www.apache.org/licenses/LICENSE-2.0").
		Server("https://api.petstore.io/v1").
		ServerWithDescription("https://staging.petstore.io/v1", "Staging server").
		// Define reusable components
		Components().
		Schema("Pet", openapibuilder.ObjectSchema().
			Property("id", openapibuilder.IntegerSchema().Format("int64")).
			Property("name", openapibuilder.StringSchema().MinLength(1).MaxLength(100)).
			Property("tag", openapibuilder.StringSchema()).
			Property("status", openapibuilder.StringSchema().Enum("available", "pending", "sold")).
			Required("id", "name")).
		Schema("Error", openapibuilder.ObjectSchema().
			Property("code", openapibuilder.IntegerSchema().Format("int32")).
			Property("message", openapibuilder.StringSchema()).
			Required("code", "message")).
		Schema("NewPet", openapibuilder.ObjectSchema().
			Property("name", openapibuilder.StringSchema().MinLength(1)).
			Property("tag", openapibuilder.StringSchema()).
			Required("name")).
		SecurityScheme("bearerAuth").BearerAuth().BearerFormat("JWT").
		Description("JWT Bearer token authentication").Done().
		SecurityScheme("apiKey").APIKeyHeader("X-API-Key").
		Description("API key in header").Done().
		Done().
		// Define paths
		Path("/pets").
		Get().
		Summary("List all pets").
		Description("Returns all pets from the system").
		OperationID("listPets").
		Tags("pets").
		QueryParam("limit").Type("integer").Format("int32").
		Description("Maximum number of pets to return").DoneOp().
		QueryParam("status").Type("string").Enum("available", "pending", "sold").
		Description("Filter by status").DoneOp().
		Response(200).Description("A list of pets").
		JSON(openapibuilder.ArraySchema(openapibuilder.RefSchema("Pet"))).
		Header("X-Total-Count").Type("integer").Description("Total number of pets").Done().
		Done().
		ResponseDefault().Description("Unexpected error").
		JSON(openapibuilder.RefSchema("Error")).Done().
		Done().
		Post().
		Summary("Create a pet").
		Description("Creates a new pet in the store").
		OperationID("createPet").
		Tags("pets").
		Security("bearerAuth").
		RequestBody().Required().JSON(openapibuilder.RefSchema("NewPet")).Done().
		Response(201).Description("Pet created successfully").
		JSON(openapibuilder.RefSchema("Pet")).Done().
		Response(400).Description("Invalid input").
		JSON(openapibuilder.RefSchema("Error")).Done().
		ResponseDefault().Description("Unexpected error").
		JSON(openapibuilder.RefSchema("Error")).Done().
		Done().
		Done().
		Path("/pets/{petId}").
		Get().
		Summary("Get a pet by ID").
		Description("Returns a single pet").
		OperationID("getPetById").
		Tags("pets").
		PathParam("petId").Type("string").Format("uuid").
		Description("The unique identifier of the pet").DoneOp().
		Response(200).Description("Pet found").
		JSON(openapibuilder.RefSchema("Pet")).Done().
		Response(404).Description("Pet not found").
		JSON(openapibuilder.RefSchema("Error")).Done().
		Done().
		Delete().
		Summary("Delete a pet").
		OperationID("deletePet").
		Tags("pets").
		Security("bearerAuth").
		PathParam("petId").Type("string").Format("uuid").DoneOp().
		Response(204).Description("Pet deleted").Done().
		Response(404).Description("Pet not found").
		JSON(openapibuilder.RefSchema("Error")).Done().
		Done().
		Done().
		Build()

	if err != nil {
		log.Fatalf("Failed to build spec: %v", err)
	}

	// Output as YAML
	yaml, err := openapibuilder.ToYAML(spec)
	if err != nil {
		log.Fatalf("Failed to render YAML: %v", err)
	}

	fmt.Println(string(yaml))
}
