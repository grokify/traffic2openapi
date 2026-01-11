package main

import (
	"github.com/spf13/cobra"
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert traffic logs to IR format",
	Long: `Convert traffic logs from various sources to Intermediate Representation (IR) format.

Supported sources:
  - har: HAR (HTTP Archive) files from browser DevTools, Playwright, etc.

Examples:
  # Convert HAR files to IR
  traffic2openapi convert har -i recording.har -o traffic.ndjson

  # Convert multiple HAR files from a directory
  traffic2openapi convert har -i ./har-files/ -o traffic.ndjson`,
}

func init() {
	rootCmd.AddCommand(convertCmd)
}
