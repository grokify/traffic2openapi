package inference

import (
	"sort"
	"strings"
)

// ProcessBody extracts schema information from a JSON body into a SchemaStore.
func ProcessBody(store *SchemaStore, body any) {
	if body == nil {
		return
	}
	store.AddObservation()
	processValue(store, "", body)
}

// processValue recursively processes a value and records it in the store.
func processValue(store *SchemaStore, path string, value any) {
	if value == nil {
		store.AddValue(path, nil)
		return
	}

	switch v := value.(type) {
	case map[string]any:
		processObject(store, path, v)
	case []any:
		processArray(store, path, v)
	default:
		store.AddValue(path, value)
	}
}

// processObject processes a JSON object.
func processObject(store *SchemaStore, basePath string, obj map[string]any) {
	for key, val := range obj {
		newPath := joinPath(basePath, key)
		processValue(store, newPath, val)
	}
}

// processArray processes a JSON array.
func processArray(store *SchemaStore, basePath string, arr []any) {
	arrayPath := basePath + "[]"

	if len(arr) == 0 {
		// Record empty array
		store.AddValue(arrayPath, nil)
		return
	}

	// Check if array contains objects
	if isObjectArray(arr) {
		// Process each object's fields
		for _, item := range arr {
			if obj, ok := item.(map[string]any); ok {
				processObject(store, arrayPath, obj)
			}
		}
	} else {
		// Primitive array - record sample values
		for _, item := range arr {
			store.AddValue(arrayPath, item)
		}
	}
}

// isObjectArray checks if an array contains objects.
func isObjectArray(arr []any) bool {
	if len(arr) == 0 {
		return false
	}
	_, ok := arr[0].(map[string]any)
	return ok
}

// SchemaNode represents a node in the inferred schema tree.
type SchemaNode struct {
	Type       string                 // string, integer, number, boolean, array, object
	Format     string                 // uuid, email, date-time, etc.
	Properties map[string]*SchemaNode // for objects
	Items      *SchemaNode            // for arrays
	Required   []string               // required properties
	Nullable   bool                   // can be null
	Examples   []any                  // example values
	Enum       []string               // enum values for strings with few unique values
}

// BuildSchemaTree converts a SchemaStore into a hierarchical SchemaNode tree.
func BuildSchemaTree(store *SchemaStore) *SchemaNode {
	if store == nil || len(store.Examples) == 0 && len(store.Nullable) == 0 {
		return &SchemaNode{Type: TypeObject}
	}

	// Build a tree structure from dot-notation paths
	root := &treeNode{children: make(map[string]*treeNode)}

	// First pass: build tree structure
	allPaths := store.GetPaths()
	for _, path := range allPaths {
		if path == "" {
			continue
		}
		parts := parsePathSegments(path)
		insertPath(root, parts, path)
	}

	// Second pass: convert to SchemaNode
	return convertToSchemaNode(root, store, true)
}

// treeNode is an internal tree structure for building schemas.
type treeNode struct {
	children map[string]*treeNode
	fullPath string // the original path (for leaf nodes)
	isLeaf   bool
}

// insertPath inserts a path into the tree.
func insertPath(root *treeNode, parts []string, fullPath string) {
	current := root
	for i, part := range parts {
		if _, exists := current.children[part]; !exists {
			current.children[part] = &treeNode{children: make(map[string]*treeNode)}
		}
		current = current.children[part]
		if i == len(parts)-1 {
			current.isLeaf = true
			current.fullPath = fullPath
		}
	}
}

// convertToSchemaNode converts a tree node to a SchemaNode.
func convertToSchemaNode(node *treeNode, store *SchemaStore, isRoot bool) *SchemaNode {
	store.mu.RLock()
	defer store.mu.RUnlock()

	// Leaf node - create schema from examples
	if node.isLeaf && len(node.children) == 0 {
		return createLeafSchema(node.fullPath, store)
	}

	// Check if all children are array elements
	allArrays := true
	for key := range node.children {
		if !isArrayPath(key) {
			allArrays = false
			break
		}
	}

	// If root and only has array children with same key, it's a root array
	if isRoot && allArrays && len(node.children) == 1 {
		for _, child := range node.children {
			itemSchema := convertToSchemaNode(child, store, false)
			return &SchemaNode{
				Type:  TypeArray,
				Items: itemSchema,
			}
		}
	}

	// Build object schema
	schema := &SchemaNode{
		Type:       TypeObject,
		Properties: make(map[string]*SchemaNode),
		Required:   make([]string, 0),
	}

	for key, child := range node.children {
		propName := key
		var propSchema *SchemaNode

		if isArrayPath(key) {
			// Array property
			propName = stripArraySuffix(key)
			itemSchema := convertToSchemaNode(child, store, false)
			propSchema = &SchemaNode{
				Type:  TypeArray,
				Items: itemSchema,
			}
		} else if child.isLeaf && len(child.children) == 0 {
			// Leaf property
			propSchema = createLeafSchema(child.fullPath, store)
		} else {
			// Nested object
			propSchema = convertToSchemaNode(child, store, false)
		}

		schema.Properties[propName] = propSchema

		// Check if required
		if child.fullPath != "" && !store.Optional[child.fullPath] {
			schema.Required = append(schema.Required, propName)
		}
	}

	// Sort required for consistent output
	sort.Strings(schema.Required)

	return schema
}

// createLeafSchema creates a schema node for a leaf value.
func createLeafSchema(path string, store *SchemaStore) *SchemaNode {
	schema := &SchemaNode{
		Type: store.Types[path],
	}

	if schema.Type == "" {
		schema.Type = TypeString
	}

	// Set format
	if format, ok := store.Formats[path]; ok {
		schema.Format = format
	}

	// Set nullable
	if store.Nullable[path] {
		schema.Nullable = true
	}

	// Set examples
	if examples, ok := store.Examples[path]; ok && len(examples) > 0 {
		schema.Examples = examples
		// Note: We intentionally don't infer enums from observed values.
		// Just because we saw values like ["Alice", "Bob"] doesn't mean
		// those are the only allowed values - they're just examples.
		// Enum constraints should only be added through explicit configuration.
	}

	return schema
}

// MergeSchemas merges two schema nodes, combining their properties.
func MergeSchemas(a, b *SchemaNode) *SchemaNode {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}

	result := &SchemaNode{
		Type:       mergeTypes(a.Type, b.Type),
		Nullable:   a.Nullable || b.Nullable,
		Properties: make(map[string]*SchemaNode),
		Required:   make([]string, 0),
	}

	// Merge formats (prefer non-empty)
	if a.Format != "" {
		result.Format = a.Format
	} else if b.Format != "" {
		result.Format = b.Format
	}

	// Merge examples
	result.Examples = mergeExamples(a.Examples, b.Examples, 5)

	// Merge array items
	if a.Type == TypeArray || b.Type == TypeArray {
		result.Type = TypeArray
		result.Items = MergeSchemas(a.Items, b.Items)
	}

	// Merge object properties
	if a.Type == TypeObject || b.Type == TypeObject {
		result.Type = TypeObject

		// Collect all property names
		propNames := make(map[string]bool)
		for name := range a.Properties {
			propNames[name] = true
		}
		for name := range b.Properties {
			propNames[name] = true
		}

		// Merge each property
		for name := range propNames {
			propA := a.Properties[name]
			propB := b.Properties[name]
			result.Properties[name] = MergeSchemas(propA, propB)
		}

		// Required = intersection of both required sets
		requiredA := make(map[string]bool)
		for _, r := range a.Required {
			requiredA[r] = true
		}
		for _, r := range b.Required {
			if requiredA[r] {
				result.Required = append(result.Required, r)
			}
		}
		sort.Strings(result.Required)
	}

	// Merge enums (intersection)
	if len(a.Enum) > 0 && len(b.Enum) > 0 {
		enumA := make(map[string]bool)
		for _, e := range a.Enum {
			enumA[e] = true
		}
		for _, e := range b.Enum {
			if enumA[e] {
				result.Enum = append(result.Enum, e)
			}
		}
		sort.Strings(result.Enum)
	}

	return result
}

// mergeExamples combines examples from two sources.
func mergeExamples(a, b []any, max int) []any {
	seen := make(map[string]bool)
	result := make([]any, 0, max)

	addExample := func(ex any) {
		if len(result) >= max {
			return
		}
		// Simple dedup using string representation
		key := formatExample(ex)
		if !seen[key] {
			seen[key] = true
			result = append(result, ex)
		}
	}

	for _, ex := range a {
		addExample(ex)
	}
	for _, ex := range b {
		addExample(ex)
	}

	return result
}

// formatExample creates a string key for deduplication.
func formatExample(v any) string {
	switch val := v.(type) {
	case string:
		return "s:" + val
	case float64:
		return "n:" + strings.TrimRight(strings.TrimRight(
			strings.Replace(string(rune(int(val*1000000))), ".", "", 1), "0"), ".")
	case bool:
		if val {
			return "b:true"
		}
		return "b:false"
	case nil:
		return "null"
	default:
		return "?"
	}
}
