package openapibuilder

// Version represents an OpenAPI specification version.
type Version string

// Supported OpenAPI versions.
const (
	// OpenAPI 3.0.x versions
	Version300 Version = "3.0.0"
	Version301 Version = "3.0.1"
	Version302 Version = "3.0.2"
	Version303 Version = "3.0.3"

	// OpenAPI 3.1.x versions
	Version310 Version = "3.1.0"
	Version311 Version = "3.1.1"

	// OpenAPI 3.2.x versions (draft/upcoming)
	Version320 Version = "3.2.0"
)

// Is3x returns true if this is an OpenAPI 3.x version.
func (v Version) Is3x() bool {
	return len(v) >= 3 && v[:2] == "3."
}

// Is31x returns true if this is an OpenAPI 3.1.x version.
func (v Version) Is31x() bool {
	return len(v) >= 3 && v[:3] == "3.1"
}

// Is30x returns true if this is an OpenAPI 3.0.x version.
func (v Version) Is30x() bool {
	return len(v) >= 3 && v[:3] == "3.0"
}

// Is32x returns true if this is an OpenAPI 3.2.x version.
func (v Version) Is32x() bool {
	return len(v) >= 3 && v[:3] == "3.2"
}

// String returns the version string.
func (v Version) String() string {
	return string(v)
}
