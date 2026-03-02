package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Artifact represents a single artifact entry in the package manifest.
type Artifact struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Path string `json:"path"`
}

// PackageManifest represents the contents of an aipkg.json file for a package.
type PackageManifest struct {
	SpecVersion int        `json:"specVersion"`
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	Description string     `json:"description,omitempty"`
	License     string     `json:"license,omitempty"`
	Artifacts   []Artifact `json:"artifacts,omitempty"`
}

// MarshalIndent returns the manifest as indented JSON with a trailing newline.
func (m *PackageManifest) MarshalIndent() ([]byte, error) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')
	return data, nil
}

// WriteFile writes the manifest as indented JSON to the given path.
func (m *PackageManifest) WriteFile(dir string) error {
	data, err := m.MarshalIndent()
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "aipkg.json"), data, 0o644) //nolint:gosec // 0o644 is intentional for non-secret config
}

// LoadFile reads and parses an aipkg.json file from the given directory.
func LoadFile(dir string) (*PackageManifest, error) {
	path := filepath.Join(dir, "aipkg.json")
	data, err := os.ReadFile(path) //nolint:gosec // path constructed from caller-provided directory
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}

	var m PackageManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}
	return &m, nil
}
