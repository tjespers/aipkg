package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreate(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "my-pkg")

	if err := Create(target); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	for _, d := range WellKnownDirs {
		path := filepath.Join(target, d)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("expected directory %s to exist: %v", d, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected %s to be a directory", d)
		}
	}
}

func TestCreatePreservesExistingFiles(t *testing.T) {
	dir := t.TempDir()

	// Create a pre-existing file inside a well-known dir.
	skillsDir := filepath.Join(dir, "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	existing := filepath.Join(skillsDir, "SKILL.md")
	if err := os.WriteFile(existing, []byte("# Skill"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Create(dir); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	data, err := os.ReadFile(existing)
	if err != nil {
		t.Fatalf("existing file was removed: %v", err)
	}
	if string(data) != "# Skill" {
		t.Error("existing file content was modified")
	}
}

func TestCreateIdempotent(t *testing.T) {
	dir := t.TempDir()

	if err := Create(dir); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}
	if err := Create(dir); err != nil {
		t.Fatalf("second Create() error = %v", err)
	}

	// Verify dirs still exist.
	for _, d := range WellKnownDirs {
		if _, err := os.Stat(filepath.Join(dir, d)); err != nil {
			t.Errorf("directory %s missing after second Create: %v", d, err)
		}
	}
}
