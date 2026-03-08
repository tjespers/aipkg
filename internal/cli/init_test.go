package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func executeInit(dir string) (string, error) {
	root := newRootCmd()
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"init"})

	// Run from the specified directory.
	origDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if err := os.Chdir(dir); err != nil {
		return "", err
	}
	defer os.Chdir(origDir) //nolint:errcheck // best-effort restore in tests

	err = root.Execute()
	return stdout.String(), err
}

func TestInitHappyPath(t *testing.T) {
	dir := t.TempDir()

	out, err := executeInit(dir)
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	if out == "" {
		t.Error("expected success message, got empty output")
	}

	// Verify exactly one file created.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, e := range entries {
		if e.Name() == "aipkg-project.json" {
			found = true
		}
	}
	if !found {
		t.Error("aipkg-project.json not created")
	}

	// Verify file content.
	data, err := os.ReadFile(filepath.Join(dir, "aipkg-project.json"))
	if err != nil {
		t.Fatal(err)
	}

	var project map[string]any
	if err := json.Unmarshal(data, &project); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	specVersion, ok := project["specVersion"]
	if !ok {
		t.Fatal("missing specVersion")
	}
	if specVersion != float64(1) {
		t.Errorf("specVersion = %v, want 1", specVersion)
	}

	require, ok := project["require"]
	if !ok {
		t.Fatal("missing require")
	}
	requireMap, ok := require.(map[string]any)
	if !ok {
		t.Fatal("require is not an object")
	}
	if len(requireMap) != 0 {
		t.Errorf("require has %d entries, want 0", len(requireMap))
	}

	// Verify no .aipkg/ directory created (FR-018).
	if _, err := os.Stat(filepath.Join(dir, ".aipkg")); err == nil {
		t.Error(".aipkg/ directory should not be created by init")
	}
}

func TestInitWithExistingFiles(t *testing.T) {
	dir := t.TempDir()

	// Create some pre-existing files.
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := executeInit(dir)
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Verify pre-existing files are unchanged.
	data, err := os.ReadFile(filepath.Join(dir, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "package main\n" {
		t.Error("existing file was modified")
	}

	// Verify project file created.
	if _, err := os.Stat(filepath.Join(dir, "aipkg-project.json")); err != nil {
		t.Error("aipkg-project.json not created")
	}
}

func TestInitReInitGuard(t *testing.T) {
	dir := t.TempDir()

	// Create an existing project file with known content.
	existing := `{"specVersion": 1, "require": {"@alice/blog-tools": "1.0.0"}}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, "aipkg-project.json"), []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := executeInit(dir)
	if err == nil {
		t.Fatal("expected error for re-initialization, got nil")
	}

	// Verify error message indicates project already initialized.
	if got := err.Error(); !strings.Contains(got, "already initialized") {
		t.Errorf("error = %q, want message containing 'already initialized'", got)
	}

	// Verify existing file is unchanged (SC-006).
	data, err := os.ReadFile(filepath.Join(dir, "aipkg-project.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != existing {
		t.Error("existing aipkg-project.json was modified")
	}
}

func TestInitMutualExclusivityGuard(t *testing.T) {
	dir := t.TempDir()

	// Create an existing package manifest.
	if err := os.WriteFile(filepath.Join(dir, "aipkg.json"), []byte(`{"specVersion": 1}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := executeInit(dir)
	if err == nil {
		t.Fatal("expected error for mutual exclusivity, got nil")
	}

	// Verify error message mentions package manifest.
	got := err.Error()
	if !strings.Contains(got, "package manifest") && !strings.Contains(got, "aipkg.json") {
		t.Errorf("error = %q, want message mentioning package manifest", got)
	}

	// Verify error suggests alternative commands.
	if !strings.Contains(got, "aipkg require") || !strings.Contains(got, "aipkg install") {
		t.Errorf("error = %q, want suggestions for aipkg require and aipkg install", got)
	}

	// Verify no project file created.
	if _, err := os.Stat(filepath.Join(dir, "aipkg-project.json")); err == nil {
		t.Error("aipkg-project.json should not be created when aipkg.json exists")
	}
}

func TestInitReadOnlyDir(t *testing.T) {
	dir := t.TempDir()

	// Make directory read-only so Create() fails.
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755) }) //nolint:errcheck // best-effort restore

	_, err := executeInit(dir)
	if err == nil {
		t.Fatal("expected error for read-only directory, got nil")
	}

	if !strings.Contains(err.Error(), "cannot initialize project") {
		t.Errorf("error = %q, want message containing 'cannot initialize project'", err.Error())
	}
}
