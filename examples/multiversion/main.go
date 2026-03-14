// Example: Multi-version OpenAPI output
//
// This example demonstrates how to generate the same API specification
// in multiple OpenAPI versions (3.0.3, 3.1.0, 3.2.0) simultaneously.
package main

import (
	"fmt"
	"log"

	"github.com/grokify/traffic2openapi/pkg/openapi/convert"
	"github.com/grokify/traffic2openapi/pkg/openapibuilder"
)

func main() {
	// Build a spec using the builder (defaults to 3.1.0)
	spec, err := openapibuilder.NewSpec(openapibuilder.Version310).
		Title("Multi-Version API").
		Version("1.0.0").
		Description("An API that demonstrates multi-version output").
		Server("https://api.example.com").
		Components().
		Schema("User", openapibuilder.ObjectSchema().
			Property("id", openapibuilder.IntegerSchema().Format("int64")).
			Property("name", openapibuilder.StringSchema()).
			Property("email", openapibuilder.StringSchema().Format("email").Nullable()). // Will be converted appropriately
			Property("age", openapibuilder.IntegerSchema().Nullable()).
			Required("id", "name")).
		Schema("Error", openapibuilder.ObjectSchema().
			Property("code", openapibuilder.IntegerSchema()).
			Property("message", openapibuilder.StringSchema()).
			Required("code", "message")).
		Done().
		Path("/users").
		Get().
		Summary("List users").
		OperationID("listUsers").
		Response(200).Description("Success").
		JSON(openapibuilder.ArraySchema(openapibuilder.RefSchema("User"))).Done().
		Done().
		Done().
		Path("/users/{userId}").
		Get().
		Summary("Get user by ID").
		OperationID("getUser").
		PathParam("userId").Type("integer").Format("int64").DoneOp().
		Response(200).Description("Success").
		JSON(openapibuilder.RefSchema("User")).Done().
		Response(404).Description("User not found").
		JSON(openapibuilder.RefSchema("Error")).Done().
		Done().
		Done().
		Build()

	if err != nil {
		log.Fatalf("Failed to build spec: %v", err)
	}

	// Convert to all standard versions (3.0.3 and 3.1.0)
	output, err := convert.StandardVersions(spec)
	if err != nil {
		log.Fatalf("Failed to convert: %v", err)
	}

	fmt.Println("=== Generated specs for versions:", output.Versions())
	fmt.Println()

	// Get YAML for each version
	yamlSpecs, err := output.ToYAML()
	if err != nil {
		log.Fatalf("Failed to render YAML: %v", err)
	}

	// Print each version
	for _, version := range output.Versions() {
		fmt.Printf("=== OpenAPI %s ===\n", version)
		fmt.Println(string(yamlSpecs[version]))
		fmt.Println()
	}

	// Demonstrate all versions (including 3.2.0)
	fmt.Println("=== All Versions (3.0.3, 3.1.0, 3.2.0) ===")
	allOutput, err := convert.AllVersions(spec)
	if err != nil {
		log.Fatalf("Failed to convert to all versions: %v", err)
	}

	for _, version := range allOutput.Versions() {
		s := allOutput.Get(version)
		fmt.Printf("Version %s: openapi=%s, info.title=%s\n",
			version, s.OpenAPI, s.Info.Title)
	}

	// Show how nullable is handled differently
	fmt.Println()
	fmt.Println("=== Nullable Handling Comparison ===")

	spec30 := output.Get(convert.Version303)
	spec31 := output.Get(convert.Version310)

	if spec30.Components != nil && spec30.Components.Schemas != nil {
		userSchema30 := spec30.Components.Schemas["User"]
		emailSchema30 := userSchema30.Properties["email"]
		fmt.Printf("3.0.3 - email.type=%v, email.nullable=%v\n",
			emailSchema30.Type, emailSchema30.Nullable)
	}

	if spec31.Components != nil && spec31.Components.Schemas != nil {
		userSchema31 := spec31.Components.Schemas["User"]
		emailSchema31 := userSchema31.Properties["email"]
		fmt.Printf("3.1.0 - email.type=%v, email.nullable=%v\n",
			emailSchema31.Type, emailSchema31.Nullable)
	}

	// Example: Write files to disk (commented out)
	// output.WriteFilesToDir("./output", "api", openapi.FormatYAML)
	// This would create:
	//   ./output/api-3.0.3.yaml
	//   ./output/api-3.1.0.yaml
}
