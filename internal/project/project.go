package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const fileName = "aipkg-project.json"

// File represents the contents of an aipkg-project.json project file.
type File struct {
	SpecVersion int               `json:"specVersion"`
	Require     map[string]string `json:"require"`
}

// Create writes a new aipkg-project.json with specVersion 1 and an empty
// require map to the given directory.
func Create(dir string) error {
	p := &File{
		SpecVersion: 1,
		Require:     map[string]string{},
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling project file: %w", err)
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(dir, fileName), data, 0o644) //nolint:gosec // 0o644 is intentional for non-secret config
}

// LoadFile reads and parses an aipkg-project.json file from the given directory.
func LoadFile(dir string) (*File, error) {
	path := filepath.Join(dir, fileName)
	data, err := os.ReadFile(path) //nolint:gosec // path constructed from caller-provided directory
	if err != nil {
		return nil, fmt.Errorf("reading project file: %w", err)
	}

	var p File
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parsing project file: %w", err)
	}
	return &p, nil
}
