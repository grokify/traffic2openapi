// Package openapibuilder provides a fluent builder API for constructing
// OpenAPI 3.x specifications.
//
// The builder uses a chainable pattern with parent references, allowing
// natural navigation through the spec structure using Done() methods.
//
// Basic usage:
//
//	spec, err := openapibuilder.NewSpec(openapibuilder.Version310).
//	    Title("Pet Store API").
//	    Version("1.0.0").
//	    Server("https://api.example.com").
//	    Path("/pets/{petId}").
//	        Get().
//	            Summary("Get pet by ID").
//	            OperationID("getPet").
//	            Response(200).Description("Success").JSON(RefSchema("Pet")).Done().
//	        Done().
//	    Done().
//	    Build()
//
// The package automatically handles OpenAPI 3.0 vs 3.1 differences such as
// nullable type representation and examples format.
package openapibuilder
