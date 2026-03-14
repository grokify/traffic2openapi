package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/grokify/traffic2openapi/pkg/openapi/validate"
	"github.com/spf13/cobra"
)

var validateSpecCmd = &cobra.Command{
	Use:   "validate-spec [file or directory]",
	Short: "Validate OpenAPI specification files",
	Long: `Validate OpenAPI specification files using libopenapi.

This command reads OpenAPI specification files (YAML or JSON) and validates
them against the OpenAPI specification, reporting any errors or warnings.

Supports OpenAPI 3.0.x, 3.1.x, and 3.2.x specifications.

Examples:
  # Validate a single file
  traffic2openapi validate-spec openapi.yaml

  # Validate all OpenAPI files in a directory
  traffic2openapi validate-spec ./specs/

  # Validate with verbose output showing warnings
  traffic2openapi validate-spec openapi.yaml --verbose`,
	Args: cobra.ExactArgs(1),
	RunE: runValidateSpec,
}

var (
	verboseSpec  bool
	strictSpec   bool
	showWarnings bool
)

func init() {
	rootCmd.AddCommand(validateSpecCmd)

	validateSpecCmd.Flags().BoolVarP(&verboseSpec, "verbose", "V", false, "Show detailed validation results")
	validateSpecCmd.Flags().BoolVar(&strictSpec, "strict", false, "Treat warnings as errors")
	validateSpecCmd.Flags().BoolVarP(&showWarnings, "warnings", "w", true, "Show warnings (default: true)")
}

func runValidateSpec(cmd *cobra.Command, args []string) error {
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
			if ext == ".yaml" || ext == ".yml" || ext == ".json" {
				files = append(files, filepath.Join(inputPath, entry.Name()))
			}
		}
	} else {
		files = []string{inputPath}
	}

	if len(files) == 0 {
		return fmt.Errorf("no OpenAPI specification files found")
	}

	totalErrors := 0
	totalWarnings := 0
	validFiles := 0
	invalidFiles := 0

	for _, file := range files {
		result, err := validate.ValidateFile(file)
		if err != nil {
			cmd.Printf("ERROR %s: %v\n", filepath.Base(file), err)
			invalidFiles++
			totalErrors++
			continue
		}

		fileErrors := len(result.Errors)
		fileWarnings := len(result.Warnings)

		if strictSpec {
			fileErrors += fileWarnings
		}

		if result.Valid && fileErrors == 0 {
			validFiles++
			if verboseSpec {
				cmd.Printf("OK   %s (OpenAPI %s)\n", filepath.Base(file), result.Version)
			}
		} else {
			invalidFiles++
			cmd.Printf("FAIL %s (OpenAPI %s)\n", filepath.Base(file), result.Version)
		}

		// Show errors
		for _, e := range result.Errors {
			totalErrors++
			if verboseSpec || !result.Valid {
				cmd.Printf("     ERROR: %s\n", e.Message)
			}
		}

		// Show warnings
		if showWarnings {
			for _, w := range result.Warnings {
				totalWarnings++
				if verboseSpec {
					cmd.Printf("     WARN:  %s\n", w.Message)
				}
			}
		}
	}

	// Summary
	cmd.Printf("\nValidation Summary:\n")
	cmd.Printf("  Files:    %d valid, %d invalid, %d total\n", validFiles, invalidFiles, len(files))
	cmd.Printf("  Errors:   %d\n", totalErrors)
	if showWarnings {
		cmd.Printf("  Warnings: %d\n", totalWarnings)
	}

	if invalidFiles > 0 {
		return fmt.Errorf("%d file(s) failed validation", invalidFiles)
	}

	if strictSpec && totalWarnings > 0 {
		return fmt.Errorf("%d warning(s) found (strict mode)", totalWarnings)
	}

	cmd.Printf("\nAll files valid.\n")
	return nil
}
