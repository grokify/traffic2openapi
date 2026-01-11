package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/grokify/traffic2openapi/pkg/ir"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate [file or directory]",
	Short: "Validate IR files",
	Long: `Validate Intermediate Representation (IR) files for correctness.

This command reads IR files and checks that they conform to the IR schema,
reporting any parsing errors or invalid records.

Examples:
  # Validate a single file
  traffic2openapi validate traffic.ndjson

  # Validate all IR files in a directory
  traffic2openapi validate ./logs/

  # Validate with verbose output
  traffic2openapi validate ./logs/ --verbose`,
	Args: cobra.ExactArgs(1),
	RunE: runValidate,
}

var verboseValidate bool

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().BoolVarP(&verboseValidate, "verbose", "V", false, "Show detailed validation results")
}

func runValidate(cmd *cobra.Command, args []string) error {
	inputPath := args[0]

	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("input path error: %w", err)
	}

	var files []string
	if info.IsDir() {
		entries, err := os.ReadDir(inputPath)
		if err != nil {
			return fmt.Errorf("reading directory: %w", err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if ext == ".json" || ext == ".ndjson" {
				files = append(files, filepath.Join(inputPath, entry.Name()))
			}
		}
	} else {
		files = []string{inputPath}
	}

	if len(files) == 0 {
		return fmt.Errorf("no IR files found")
	}

	totalRecords := 0
	totalErrors := 0
	validFiles := 0

	for _, file := range files {
		records, err := ir.ReadFile(file)
		if err != nil {
			cmd.Printf("FAIL %s: %v\n", filepath.Base(file), err)
			totalErrors++
			continue
		}

		validFiles++
		totalRecords += len(records)

		if verboseValidate {
			cmd.Printf("OK   %s (%d records)\n", filepath.Base(file), len(records))

			// Show sample of endpoints
			methods := make(map[string]int)
			for _, r := range records {
				key := string(r.Request.Method)
				methods[key]++
			}
			for method, count := range methods {
				cmd.Printf("     %s: %d requests\n", method, count)
			}
		}
	}

	// Summary
	cmd.Printf("\nValidation Summary:\n")
	cmd.Printf("  Files:   %d valid, %d invalid, %d total\n", validFiles, totalErrors, len(files))
	cmd.Printf("  Records: %d total\n", totalRecords)

	if totalErrors > 0 {
		return fmt.Errorf("%d file(s) failed validation", totalErrors)
	}

	cmd.Printf("\nAll files valid.\n")
	return nil
}
