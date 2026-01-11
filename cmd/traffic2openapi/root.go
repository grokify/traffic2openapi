package main

import (
	"github.com/spf13/cobra"
)

var version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "traffic2openapi",
	Short: "Generate OpenAPI specs from HTTP traffic",
	Long: `traffic2openapi generates OpenAPI 3.0/3.1/3.2 specifications from HTTP traffic logs.

It reads Intermediate Representation (IR) files containing captured HTTP
request/response data and infers API structure through intelligent analysis.

Examples:
  # Generate OpenAPI 3.1 spec from IR files
  traffic2openapi generate -i ./logs/ -o openapi.yaml

  # Generate OpenAPI 3.0 spec in JSON format
  traffic2openapi generate -i traffic.ndjson -o api.json --version 3.0

  # Validate IR files
  traffic2openapi validate ./logs/`,
	Version: version,
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
