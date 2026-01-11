package inference

import (
	"regexp"
	"strings"
)

// Type constants
const (
	TypeString  = "string"
	TypeInteger = "integer"
	TypeNumber  = "number"
	TypeBoolean = "boolean"
	TypeArray   = "array"
	TypeObject  = "object"
)

// Format constants
const (
	FormatUUID     = "uuid"
	FormatEmail    = "email"
	FormatDateTime = "date-time"
	FormatDate     = "date"
	FormatTime     = "time"
	FormatURI      = "uri"
	FormatIPv4     = "ipv4"
	FormatIPv6     = "ipv6"
)

// Regex patterns for format detection
var (
	uuidPattern     = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	emailPattern    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	dateTimePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}`)
	datePattern     = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	timePattern     = regexp.MustCompile(`^\d{2}:\d{2}:\d{2}`)
	uriPattern      = regexp.MustCompile(`^https?://`)
	ipv4Pattern     = regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
	ipv6Pattern     = regexp.MustCompile(`^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$`)
)

// inferType returns the JSON Schema type for a Go value.
func inferType(value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case bool:
		return TypeBoolean
	case float64:
		// Check if it's actually an integer
		if v == float64(int64(v)) {
			return TypeInteger
		}
		return TypeNumber
	case float32:
		if v == float32(int32(v)) {
			return TypeInteger
		}
		return TypeNumber
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return TypeInteger
	case string:
		return TypeString
	case []any:
		return TypeArray
	case map[string]any:
		return TypeObject
	default:
		return TypeString
	}
}

// mergeTypes returns a type that encompasses both types.
func mergeTypes(t1, t2 string) string {
	if t1 == "" {
		return t2
	}
	if t2 == "" {
		return t1
	}
	if t1 == t2 {
		return t1
	}

	// integer + number = number
	if (t1 == TypeInteger && t2 == TypeNumber) || (t1 == TypeNumber && t2 == TypeInteger) {
		return TypeNumber
	}

	// Conflicting types default to string (most permissive)
	return TypeString
}

// detectFormat detects the format of a string value.
func detectFormat(s string) string {
	if s == "" {
		return ""
	}

	switch {
	case uuidPattern.MatchString(s):
		return FormatUUID
	case emailPattern.MatchString(s):
		return FormatEmail
	case dateTimePattern.MatchString(s):
		return FormatDateTime
	case datePattern.MatchString(s):
		return FormatDate
	case timePattern.MatchString(s):
		return FormatTime
	case uriPattern.MatchString(s):
		return FormatURI
	case ipv4Pattern.MatchString(s):
		return FormatIPv4
	case ipv6Pattern.MatchString(s):
		return FormatIPv6
	default:
		return ""
	}
}

// valuesEqual compares two values for equality.
func valuesEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch v1 := a.(type) {
	case map[string]any:
		v2, ok := b.(map[string]any)
		if !ok || len(v1) != len(v2) {
			return false
		}
		for k, val1 := range v1 {
			val2, exists := v2[k]
			if !exists || !valuesEqual(val1, val2) {
				return false
			}
		}
		return true

	case []any:
		v2, ok := b.([]any)
		if !ok || len(v1) != len(v2) {
			return false
		}
		for i := range v1 {
			if !valuesEqual(v1[i], v2[i]) {
				return false
			}
		}
		return true

	case float64:
		v2, ok := b.(float64)
		return ok && v1 == v2

	case int:
		v2, ok := b.(int)
		return ok && v1 == v2

	case string:
		v2, ok := b.(string)
		return ok && v1 == v2

	case bool:
		v2, ok := b.(bool)
		return ok && v1 == v2

	default:
		return a == b
	}
}

// joinPath joins path segments with dots, handling array markers.
func joinPath(basePath, key string) string {
	if basePath == "" {
		return key
	}
	return basePath + "." + key
}

// parsePathSegments splits a path into segments.
func parsePathSegments(path string) []string {
	return strings.Split(path, ".")
}

// isArrayPath checks if a path segment indicates an array.
func isArrayPath(segment string) bool {
	return strings.HasSuffix(segment, "[]")
}

// stripArraySuffix removes the [] suffix from a path segment.
func stripArraySuffix(segment string) string {
	return strings.TrimSuffix(segment, "[]")
}
