package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/grokify/traffic2openapi/pkg/openapi"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff <old-spec> <new-spec>",
	Short: "Compare two OpenAPI specifications",
	Long: `Compare two OpenAPI specifications and show the differences.

Detects added, removed, and modified endpoints, parameters, and schemas.
Can identify breaking changes for API versioning.

Examples:
  # Compare two specs
  traffic2openapi diff old.yaml new.yaml

  # Output as JSON for CI/CD
  traffic2openapi diff old.yaml new.yaml --format json

  # Only show breaking changes
  traffic2openapi diff old.yaml new.yaml --breaking-only

  # Exit with non-zero code if breaking changes found (for CI)
  traffic2openapi diff old.yaml new.yaml --breaking-only --exit-code`,
	Args: cobra.ExactArgs(2),
	RunE: runDiff,
}

var (
	diffFormat       string
	diffBreakingOnly bool
	diffExitCode     bool
)

func init() {
	rootCmd.AddCommand(diffCmd)

	diffCmd.Flags().StringVarP(&diffFormat, "format", "f", "text", "Output format: text or json")
	diffCmd.Flags().BoolVar(&diffBreakingOnly, "breaking-only", false, "Only show breaking changes")
	diffCmd.Flags().BoolVar(&diffExitCode, "exit-code", false, "Exit with non-zero code if differences found")
}

// DiffResult holds the comparison results.
type DiffResult struct {
	AddedPaths      []string         `json:"addedPaths,omitempty"`
	RemovedPaths    []string         `json:"removedPaths,omitempty"`
	AddedOperations []string         `json:"addedOperations,omitempty"`
	RemovedOps      []string         `json:"removedOperations,omitempty"`
	ModifiedOps     []OpDiff         `json:"modifiedOperations,omitempty"`
	BreakingChanges []BreakingChange `json:"breakingChanges,omitempty"`
}

// OpDiff describes changes to an operation.
type OpDiff struct {
	Path             string   `json:"path"`
	Method           string   `json:"method"`
	AddedParams      []string `json:"addedParams,omitempty"`
	RemovedParams    []string `json:"removedParams,omitempty"`
	AddedResponses   []string `json:"addedResponses,omitempty"`
	RemovedResponses []string `json:"removedResponses,omitempty"`
}

// BreakingChange describes a breaking API change.
type BreakingChange struct {
	Type        string `json:"type"`
	Path        string `json:"path"`
	Method      string `json:"method,omitempty"`
	Description string `json:"description"`
}

func runDiff(cmd *cobra.Command, args []string) error {
	oldPath := args[0]
	newPath := args[1]

	// Read specs
	oldSpec, err := openapi.ReadFile(oldPath)
	if err != nil {
		return fmt.Errorf("reading old spec: %w", err)
	}

	newSpec, err := openapi.ReadFile(newPath)
	if err != nil {
		return fmt.Errorf("reading new spec: %w", err)
	}

	// Compare specs
	result := compareSpecs(oldSpec, newSpec)

	// Filter to breaking changes only if requested
	if diffBreakingOnly {
		result = filterBreakingOnly(result)
	}

	// Output results
	if diffFormat == "json" {
		return outputDiffJSON(result)
	}
	outputDiffText(cmd, result)

	// Exit code handling
	if diffExitCode {
		if hasChanges(result) {
			os.Exit(1)
		}
	}

	return nil
}

func compareSpecs(oldSpec, newSpec *openapi.Spec) *DiffResult {
	result := &DiffResult{}

	oldPaths := make(map[string]bool)
	newPaths := make(map[string]bool)

	for path := range oldSpec.Paths {
		oldPaths[path] = true
	}
	for path := range newSpec.Paths {
		newPaths[path] = true
	}

	// Find added and removed paths
	for path := range newPaths {
		if !oldPaths[path] {
			result.AddedPaths = append(result.AddedPaths, path)
		}
	}
	for path := range oldPaths {
		if !newPaths[path] {
			result.RemovedPaths = append(result.RemovedPaths, path)
			result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
				Type:        "path_removed",
				Path:        path,
				Description: fmt.Sprintf("Path %s was removed", path),
			})
		}
	}

	// Sort for consistent output
	sort.Strings(result.AddedPaths)
	sort.Strings(result.RemovedPaths)

	// Compare operations on shared paths
	for path := range oldPaths {
		if !newPaths[path] {
			continue
		}

		oldItem := oldSpec.Paths[path]
		newItem := newSpec.Paths[path]

		comparePaths(result, path, oldItem, newItem)
	}

	return result
}

func comparePaths(result *DiffResult, path string, oldItem, newItem *openapi.PathItem) {
	methods := []struct {
		name  string
		oldOp *openapi.Operation
		newOp *openapi.Operation
	}{
		{"GET", oldItem.Get, newItem.Get},
		{"POST", oldItem.Post, newItem.Post},
		{"PUT", oldItem.Put, newItem.Put},
		{"DELETE", oldItem.Delete, newItem.Delete},
		{"PATCH", oldItem.Patch, newItem.Patch},
		{"HEAD", oldItem.Head, newItem.Head},
		{"OPTIONS", oldItem.Options, newItem.Options},
	}

	for _, m := range methods {
		opKey := fmt.Sprintf("%s %s", m.name, path)

		if m.oldOp == nil && m.newOp != nil {
			result.AddedOperations = append(result.AddedOperations, opKey)
		} else if m.oldOp != nil && m.newOp == nil {
			result.RemovedOps = append(result.RemovedOps, opKey)
			result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
				Type:        "operation_removed",
				Path:        path,
				Method:      m.name,
				Description: fmt.Sprintf("Operation %s %s was removed", m.name, path),
			})
		} else if m.oldOp != nil && m.newOp != nil {
			diff := compareOperations(path, m.name, m.oldOp, m.newOp)
			if diff != nil {
				result.ModifiedOps = append(result.ModifiedOps, *diff)

				// Check for breaking parameter changes
				for _, param := range diff.RemovedParams {
					result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
						Type:        "parameter_removed",
						Path:        path,
						Method:      m.name,
						Description: fmt.Sprintf("Parameter '%s' was removed from %s %s", param, m.name, path),
					})
				}
			}
		}
	}
}

func compareOperations(path, method string, oldOp, newOp *openapi.Operation) *OpDiff {
	diff := &OpDiff{
		Path:   path,
		Method: method,
	}

	hasChanges := false

	// Compare parameters
	oldParams := make(map[string]bool)
	newParams := make(map[string]bool)

	for _, p := range oldOp.Parameters {
		oldParams[fmt.Sprintf("%s:%s", p.In, p.Name)] = true
	}
	for _, p := range newOp.Parameters {
		newParams[fmt.Sprintf("%s:%s", p.In, p.Name)] = true
	}

	for param := range newParams {
		if !oldParams[param] {
			diff.AddedParams = append(diff.AddedParams, param)
			hasChanges = true
		}
	}
	for param := range oldParams {
		if !newParams[param] {
			diff.RemovedParams = append(diff.RemovedParams, param)
			hasChanges = true
		}
	}

	// Compare responses
	for status := range newOp.Responses {
		if _, ok := oldOp.Responses[status]; !ok {
			diff.AddedResponses = append(diff.AddedResponses, status)
			hasChanges = true
		}
	}
	for status := range oldOp.Responses {
		if _, ok := newOp.Responses[status]; !ok {
			diff.RemovedResponses = append(diff.RemovedResponses, status)
			hasChanges = true
		}
	}

	if hasChanges {
		sort.Strings(diff.AddedParams)
		sort.Strings(diff.RemovedParams)
		sort.Strings(diff.AddedResponses)
		sort.Strings(diff.RemovedResponses)
		return diff
	}
	return nil
}

func filterBreakingOnly(result *DiffResult) *DiffResult {
	return &DiffResult{
		RemovedPaths:    result.RemovedPaths,
		RemovedOps:      result.RemovedOps,
		BreakingChanges: result.BreakingChanges,
	}
}

func hasChanges(result *DiffResult) bool {
	return len(result.AddedPaths) > 0 ||
		len(result.RemovedPaths) > 0 ||
		len(result.AddedOperations) > 0 ||
		len(result.RemovedOps) > 0 ||
		len(result.ModifiedOps) > 0 ||
		len(result.BreakingChanges) > 0
}

func outputDiffJSON(result *DiffResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func outputDiffText(cmd *cobra.Command, result *DiffResult) {
	if !hasChanges(result) {
		cmd.Println("No differences found.")
		return
	}

	if len(result.AddedPaths) > 0 {
		cmd.Println("\n+ Added Paths:")
		for _, path := range result.AddedPaths {
			cmd.Printf("  + %s\n", path)
		}
	}

	if len(result.RemovedPaths) > 0 {
		cmd.Println("\n- Removed Paths:")
		for _, path := range result.RemovedPaths {
			cmd.Printf("  - %s\n", path)
		}
	}

	if len(result.AddedOperations) > 0 {
		cmd.Println("\n+ Added Operations:")
		for _, op := range result.AddedOperations {
			cmd.Printf("  + %s\n", op)
		}
	}

	if len(result.RemovedOps) > 0 {
		cmd.Println("\n- Removed Operations:")
		for _, op := range result.RemovedOps {
			cmd.Printf("  - %s\n", op)
		}
	}

	if len(result.ModifiedOps) > 0 {
		cmd.Println("\n~ Modified Operations:")
		for _, op := range result.ModifiedOps {
			cmd.Printf("  ~ %s %s\n", op.Method, op.Path)
			for _, p := range op.AddedParams {
				cmd.Printf("    + param: %s\n", p)
			}
			for _, p := range op.RemovedParams {
				cmd.Printf("    - param: %s\n", p)
			}
			for _, r := range op.AddedResponses {
				cmd.Printf("    + response: %s\n", r)
			}
			for _, r := range op.RemovedResponses {
				cmd.Printf("    - response: %s\n", r)
			}
		}
	}

	if len(result.BreakingChanges) > 0 {
		cmd.Println("\n⚠️  Breaking Changes:")
		for _, bc := range result.BreakingChanges {
			cmd.Printf("  ⚠️  [%s] %s\n", strings.ToUpper(bc.Type), bc.Description)
		}
	}

	// Summary
	cmd.Printf("\nSummary: %d added, %d removed, %d modified, %d breaking\n",
		len(result.AddedPaths)+len(result.AddedOperations),
		len(result.RemovedPaths)+len(result.RemovedOps),
		len(result.ModifiedOps),
		len(result.BreakingChanges))
}
