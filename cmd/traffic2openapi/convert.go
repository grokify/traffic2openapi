package main

import (
	"github.com/spf13/cobra"
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert traffic logs to IR format",
	Long: `Convert traffic logs from various sources to Intermediate Representation (IR) format.

Supported sources:
  - har:     HAR (HTTP Archive) files from browser DevTools, Playwright, etc.
  - postman: Postman Collection v2.1 files

Examples:
  # Convert HAR files to IR
  traffic2openapi convert har -i recording.har -o traffic.ndjson

  # Convert multiple HAR files from a directory
  traffic2openapi convert har -i ./har-files/ -o traffic.ndjson

  # Convert Postman collection to IR
  traffic2openapi convert postman -i collection.json -o api.ndjson

  # Convert Postman collection with base URL
  traffic2openapi convert postman -i collection.json -o api.ndjson --base-url https://api.example.com`,
}

func init() {
	rootCmd.AddCommand(convertCmd)
}
