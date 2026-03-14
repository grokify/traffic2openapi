package convert

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/grokify/traffic2openapi/pkg/openapi"
)

// MultiVersionOutput contains specs converted to multiple versions.
type MultiVersionOutput struct {
	Specs map[TargetVersion]*openapi.Spec
}

// NewMultiVersionOutput creates a MultiVersionOutput from a source spec.
func NewMultiVersionOutput(spec *openapi.Spec, targets ...TargetVersion) (*MultiVersionOutput, error) {
	output := &MultiVersionOutput{
		Specs: make(map[TargetVersion]*openapi.Spec, len(targets)),
	}

	for _, target := range targets {
		converted, err := ToVersion(spec, target)
		if err != nil {
			return nil, fmt.Errorf("converting to %s: %w", target, err)
		}
		output.Specs[target] = converted
	}

	return output, nil
}

// AllVersions returns a MultiVersionOutput with all supported versions.
func AllVersions(spec *openapi.Spec) (*MultiVersionOutput, error) {
	return NewMultiVersionOutput(spec, Version303, Version310, Version320)
}

// StandardVersions returns a MultiVersionOutput with commonly used versions (3.0.3, 3.1.0).
func StandardVersions(spec *openapi.Spec) (*MultiVersionOutput, error) {
	return NewMultiVersionOutput(spec, Version303, Version310)
}

// ToYAML returns all specs as YAML bytes keyed by version.
func (m *MultiVersionOutput) ToYAML() (map[TargetVersion][]byte, error) {
	result := make(map[TargetVersion][]byte, len(m.Specs))
	for version, spec := range m.Specs {
		data, err := openapi.ToYAML(spec)
		if err != nil {
			return nil, fmt.Errorf("converting %s to YAML: %w", version, err)
		}
		result[version] = data
	}
	return result, nil
}

// ToJSON returns all specs as JSON bytes keyed by version.
func (m *MultiVersionOutput) ToJSON() (map[TargetVersion][]byte, error) {
	result := make(map[TargetVersion][]byte, len(m.Specs))
	for version, spec := range m.Specs {
		data, err := openapi.ToJSON(spec)
		if err != nil {
			return nil, fmt.Errorf("converting %s to JSON: %w", version, err)
		}
		result[version] = data
	}
	return result, nil
}

// WriteFiles writes all specs to files with version-specific names.
// The basePath should include the base filename without extension.
// Example: basePath="./output/openapi" creates:
//   - ./output/openapi-3.0.3.yaml
//   - ./output/openapi-3.1.0.yaml
//   - ./output/openapi-3.2.0.yaml
func (m *MultiVersionOutput) WriteFiles(basePath string, format openapi.Format) error {
	ext := ".yaml"
	if format == openapi.FormatJSON {
		ext = ".json"
	}

	for version, spec := range m.Specs {
		filename := fmt.Sprintf("%s-%s%s", basePath, version, ext)
		if err := openapi.WriteFile(filename, spec); err != nil {
			return fmt.Errorf("writing %s: %w", filename, err)
		}
	}
	return nil
}

// WriteFilesToDir writes all specs to a directory with version-specific names.
// Example: dir="./output", basename="openapi" creates:
//   - ./output/openapi-3.0.3.yaml
//   - ./output/openapi-3.1.0.yaml
func (m *MultiVersionOutput) WriteFilesToDir(dir, basename string, format openapi.Format) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	basePath := filepath.Join(dir, basename)
	return m.WriteFiles(basePath, format)
}

// Get returns the spec for a specific version.
func (m *MultiVersionOutput) Get(version TargetVersion) *openapi.Spec {
	return m.Specs[version]
}

// Versions returns a sorted list of versions in this output.
func (m *MultiVersionOutput) Versions() []TargetVersion {
	versions := make([]TargetVersion, 0, len(m.Specs))
	for v := range m.Specs {
		versions = append(versions, v)
	}
	// Sort by version string
	for i := 0; i < len(versions)-1; i++ {
		for j := i + 1; j < len(versions); j++ {
			if string(versions[i]) > string(versions[j]) {
				versions[i], versions[j] = versions[j], versions[i]
			}
		}
	}
	return versions
}

// VersionedFilename generates a version-specific filename.
// Example: VersionedFilename("openapi.yaml", Version310) -> "openapi-3.1.0.yaml"
func VersionedFilename(filename string, version TargetVersion) string {
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	return fmt.Sprintf("%s-%s%s", base, version, ext)
}
