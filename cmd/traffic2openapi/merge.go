package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/grokify/traffic2openapi/pkg/ir"
	"github.com/grokify/traffic2openapi/pkg/openapi"
	"github.com/spf13/cobra"
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge multiple traffic files or OpenAPI specs",
	Long: `Merge multiple IR traffic files or OpenAPI specifications into a single output.

For IR files (.ndjson, .json), records are combined with optional deduplication.
For OpenAPI specs (.yaml, .yml, .json), paths and components are merged.

Examples:
  # Merge multiple traffic files
  traffic2openapi merge -i traffic1.ndjson -i traffic2.ndjson -o combined.ndjson

  # Merge all traffic files in a directory
  traffic2openapi merge -i ./traffic/ -o combined.ndjson

  # Merge with deduplication by record ID
  traffic2openapi merge -i traffic1.ndjson -i traffic2.ndjson -o combined.ndjson --dedupe

  # Merge OpenAPI specs
  traffic2openapi merge -i api-v1.yaml -i api-v2.yaml -o merged.yaml`,
	RunE: runMerge,
}

var (
	mergeInputs []string
	mergeOutput string
	mergeDedupe bool
)

func init() {
	rootCmd.AddCommand(mergeCmd)

	mergeCmd.Flags().StringArrayVarP(&mergeInputs, "input", "i", nil, "Input files or directories (can be repeated)")
	mergeCmd.Flags().StringVarP(&mergeOutput, "output", "o", "", "Output file path (required)")
	mergeCmd.Flags().BoolVar(&mergeDedupe, "dedupe", false, "Deduplicate records by ID")

	if err := mergeCmd.MarkFlagRequired("input"); err != nil {
		panic(fmt.Sprintf("failed to mark input flag required: %v", err))
	}
	if err := mergeCmd.MarkFlagRequired("output"); err != nil {
		panic(fmt.Sprintf("failed to mark output flag required: %v", err))
	}
}

func runMerge(cmd *cobra.Command, args []string) error {
	// Determine merge type based on output extension
	outputExt := strings.ToLower(filepath.Ext(mergeOutput))

	switch outputExt {
	case ".ndjson", ".json":
		return mergeIRFiles(cmd)
	case ".yaml", ".yml":
		return mergeOpenAPISpecs(cmd)
	default:
		// Try to detect from input files
		if len(mergeInputs) > 0 {
			inputExt := strings.ToLower(filepath.Ext(mergeInputs[0]))
			if inputExt == ".yaml" || inputExt == ".yml" {
				return mergeOpenAPISpecs(cmd)
			}
		}
		return mergeIRFiles(cmd)
	}
}

func mergeIRFiles(cmd *cobra.Command) error {
	var allRecords []ir.IRRecord
	seenIDs := make(map[string]bool)

	for _, input := range mergeInputs {
		info, err := os.Stat(input)
		if err != nil {
			return fmt.Errorf("input path error for %s: %w", input, err)
		}

		var records []ir.IRRecord
		if info.IsDir() {
			records, err = ir.ReadDir(input)
		} else {
			records, err = ir.ReadFile(input)
		}
		if err != nil {
			return fmt.Errorf("reading %s: %w", input, err)
		}

		cmd.Printf("Read %d records from %s\n", len(records), input)

		// Add records with optional deduplication
		for _, rec := range records {
			if mergeDedupe && rec.Id != nil {
				if seenIDs[*rec.Id] {
					continue
				}
				seenIDs[*rec.Id] = true
			}
			allRecords = append(allRecords, rec)
		}
	}

	if len(allRecords) == 0 {
		return fmt.Errorf("no records found in inputs")
	}

	// Write merged records
	if err := ir.WriteFile(mergeOutput, allRecords); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	cmd.Printf("Wrote %d records to %s\n", len(allRecords), mergeOutput)
	if mergeDedupe {
		cmd.Printf("Deduplicated %d duplicate records\n", countDuplicates(mergeInputs)-len(allRecords))
	}

	return nil
}

func countDuplicates(inputs []string) int {
	total := 0
	for _, input := range inputs {
		info, err := os.Stat(input)
		if err != nil {
			continue
		}
		var records []ir.IRRecord
		if info.IsDir() {
			records, _ = ir.ReadDir(input)
		} else {
			records, _ = ir.ReadFile(input)
		}
		total += len(records)
	}
	return total
}

func mergeOpenAPISpecs(cmd *cobra.Command) error {
	var mergedSpec *openapi.Spec

	for _, input := range mergeInputs {
		spec, err := openapi.ReadFile(input)
		if err != nil {
			return fmt.Errorf("reading %s: %w", input, err)
		}

		cmd.Printf("Read spec from %s (%d paths)\n", input, len(spec.Paths))

		if mergedSpec == nil {
			mergedSpec = spec
			continue
		}

		// Merge paths
		for path, pathItem := range spec.Paths {
			if existing, ok := mergedSpec.Paths[path]; ok {
				// Merge operations
				mergePathItem(existing, pathItem)
			} else {
				mergedSpec.Paths[path] = pathItem
			}
		}

		// Merge servers (dedupe by URL)
		serverURLs := make(map[string]bool)
		for _, s := range mergedSpec.Servers {
			serverURLs[s.URL] = true
		}
		for _, s := range spec.Servers {
			if !serverURLs[s.URL] {
				mergedSpec.Servers = append(mergedSpec.Servers, s)
				serverURLs[s.URL] = true
			}
		}

		// Merge components
		if spec.Components != nil {
			if mergedSpec.Components == nil {
				mergedSpec.Components = &openapi.Components{}
			}
			mergeComponents(mergedSpec.Components, spec.Components)
		}
	}

	if mergedSpec == nil {
		return fmt.Errorf("no specs found in inputs")
	}

	// Write merged spec
	if err := openapi.WriteFile(mergeOutput, mergedSpec); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	cmd.Printf("Wrote merged spec to %s (%d paths)\n", mergeOutput, len(mergedSpec.Paths))

	return nil
}

func mergePathItem(target, source *openapi.PathItem) {
	if source.Get != nil && target.Get == nil {
		target.Get = source.Get
	}
	if source.Post != nil && target.Post == nil {
		target.Post = source.Post
	}
	if source.Put != nil && target.Put == nil {
		target.Put = source.Put
	}
	if source.Delete != nil && target.Delete == nil {
		target.Delete = source.Delete
	}
	if source.Patch != nil && target.Patch == nil {
		target.Patch = source.Patch
	}
	if source.Head != nil && target.Head == nil {
		target.Head = source.Head
	}
	if source.Options != nil && target.Options == nil {
		target.Options = source.Options
	}
	if source.Trace != nil && target.Trace == nil {
		target.Trace = source.Trace
	}
}

func mergeComponents(target, source *openapi.Components) {
	if source.Schemas != nil {
		if target.Schemas == nil {
			target.Schemas = make(map[string]*openapi.Schema)
		}
		for name, schema := range source.Schemas {
			if _, exists := target.Schemas[name]; !exists {
				target.Schemas[name] = schema
			}
		}
	}

	if source.SecuritySchemes != nil {
		if target.SecuritySchemes == nil {
			target.SecuritySchemes = make(map[string]*openapi.SecurityScheme)
		}
		for name, scheme := range source.SecuritySchemes {
			if _, exists := target.SecuritySchemes[name]; !exists {
				target.SecuritySchemes[name] = scheme
			}
		}
	}

	if source.Parameters != nil {
		if target.Parameters == nil {
			target.Parameters = make(map[string]*openapi.Parameter)
		}
		for name, param := range source.Parameters {
			if _, exists := target.Parameters[name]; !exists {
				target.Parameters[name] = param
			}
		}
	}
}
