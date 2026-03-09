package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateAndLoad(t *testing.T) {
	dir := t.TempDir()

	if err := Create(dir); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	p, err := LoadFile(dir)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	if p.SpecVersion != 1 {
		t.Errorf("SpecVersion = %d, want 1", p.SpecVersion)
	}
	if p.Require == nil {
		t.Fatal("Require is nil, want empty map")
	}
	if len(p.Require) != 0 {
		t.Errorf("Require has %d entries, want 0", len(p.Require))
	}
}

func TestCreateJSONStructure(t *testing.T) {
	dir := t.TempDir()

	if err := Create(dir); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "aipkg-project.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	// Verify trailing newline.
	if len(data) == 0 || data[len(data)-1] != '\n' {
		t.Error("file does not end with trailing newline")
	}

	// Verify valid JSON with expected structure.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if _, ok := raw["specVersion"]; !ok {
		t.Error("missing specVersion field")
	}
	if _, ok := raw["require"]; !ok {
		t.Error("missing require field")
	}
	if len(raw) != 2 {
		t.Errorf("got %d top-level fields, want 2", len(raw))
	}
}

func TestLoadFileNotFound(t *testing.T) {
	dir := t.TempDir()

	_, err := LoadFile(dir)
	if err == nil {
		t.Fatal("LoadFile() expected error for missing file")
	}
}

func TestLoadFileInvalidJSON(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "aipkg-project.json"), []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFile(dir)
	if err == nil {
		t.Fatal("LoadFile() expected error for invalid JSON")
	}
}

func TestCreateReadOnlyDir(t *testing.T) {
	dir := t.TempDir()

	// Make directory read-only.
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755) }) //nolint:errcheck // best-effort restore

	err := Create(dir)
	if err == nil {
		t.Fatal("Create() expected error for read-only directory")
	}
}
