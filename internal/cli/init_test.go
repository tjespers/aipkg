package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// execInit runs the init command with the given args in the given directory.
// Returns stdout, stderr, and any error.
func execInit(t *testing.T, dir string, args ...string) (outStr, errStr string, execErr error) {
	t.Helper()

	orig, execErr := os.Getwd()
	if execErr != nil {
		t.Fatal(execErr)
	}
	if execErr = os.Chdir(dir); execErr != nil {
		t.Fatal(execErr)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	cmd := NewRootCmd()
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)
	cmd.SetArgs(append([]string{"init"}, args...))

	execErr = cmd.Execute()
	return outBuf.String(), errBuf.String(), execErr
}

func TestInit_PackageAllFlags(t *testing.T) {
	dir := t.TempDir()
	stdout, _, err := execInit(t, dir,
		"--type", "package",
		"--name", "@test/hello",
		"--version", "1.0.0",
		"--description", "A test package",
		"--license", "MIT",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "Created aipkg.json (package)") {
		t.Errorf("stdout = %q, want success message", stdout)
	}

	data, err := os.ReadFile(filepath.Join(dir, "aipkg.json")) //nolint:gosec // test file path
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m["type"] != "package" {
		t.Errorf("type = %v, want package", m["type"])
	}
	if m["name"] != "@test/hello" {
		t.Errorf("name = %v, want @test/hello", m["name"])
	}
	if m["version"] != "1.0.0" {
		t.Errorf("version = %v, want 1.0.0", m["version"])
	}
	if m["description"] != "A test package" {
		t.Errorf("description = %v, want 'A test package'", m["description"])
	}
	if m["license"] != "MIT" {
		t.Errorf("license = %v, want MIT", m["license"])
	}
}

func TestInit_PackageRequiredOnly(t *testing.T) {
	dir := t.TempDir()
	_, _, err := execInit(t, dir,
		"--type", "package",
		"--name", "@test/minimal",
		"--version", "2.0.0",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "aipkg.json")) //nolint:gosec // test file path
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	// Optional fields should be absent.
	if _, ok := m["description"]; ok {
		t.Errorf("description should be omitted, got %v", m["description"])
	}
	if _, ok := m["license"]; ok {
		t.Errorf("license should be omitted, got %v", m["license"])
	}
}

func TestInit_PackageDefaultVersion(t *testing.T) {
	dir := t.TempDir()
	_, _, err := execInit(t, dir,
		"--type", "package",
		"--name", "@test/defaultver",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "aipkg.json")) //nolint:gosec // test file path
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m["version"] != "0.1.0" {
		t.Errorf("version = %v, want 0.1.0 (default)", m["version"])
	}
}

func TestInit_ProjectMinimal(t *testing.T) {
	dir := t.TempDir()
	stdout, _, err := execInit(t, dir, "--type", "project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "Created aipkg.json (project)") {
		t.Errorf("stdout = %q, want project success message", stdout)
	}

	data, err := os.ReadFile(filepath.Join(dir, "aipkg.json")) //nolint:gosec // test file path
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m["type"] != "project" {
		t.Errorf("type = %v, want project", m["type"])
	}
	// Only type should be present.
	if len(m) != 1 {
		t.Errorf("expected 1 field, got %d: %v", len(m), m)
	}
}

func TestInit_ProjectWithOptionalFields(t *testing.T) {
	dir := t.TempDir()
	_, _, err := execInit(t, dir,
		"--type", "project",
		"--name", "@org/myproject",
		"--description", "My project",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "aipkg.json")) //nolint:gosec // test file path
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m["name"] != "@org/myproject" {
		t.Errorf("name = %v, want @org/myproject", m["name"])
	}
	if m["description"] != "My project" {
		t.Errorf("description = %v, want 'My project'", m["description"])
	}
}

func TestInit_OverwriteGuard(t *testing.T) {
	dir := t.TempDir()
	// Create initial file.
	if err := os.WriteFile(filepath.Join(dir, "aipkg.json"), []byte(`{"type":"project"}`), 0o600); err != nil { //nolint:gosec // test fixture
		t.Fatal(err)
	}

	_, _, err := execInit(t, dir, "--type", "project")
	if err == nil {
		t.Fatal("expected error for existing file")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %v, want 'already exists'", err)
	}
}

func TestInit_InvalidName(t *testing.T) {
	dir := t.TempDir()
	_, _, err := execInit(t, dir,
		"--type", "package",
		"--name", "bad-name",
		"--version", "1.0.0",
	)
	if err == nil {
		t.Fatal("expected error for invalid name")
	}
	if !strings.Contains(err.Error(), "invalid name") {
		t.Errorf("error = %v, want 'invalid name'", err)
	}
}

func TestInit_InvalidVersion(t *testing.T) {
	dir := t.TempDir()
	_, _, err := execInit(t, dir,
		"--type", "package",
		"--name", "@test/a",
		"--version", "v1.0",
	)
	if err == nil {
		t.Fatal("expected error for invalid version")
	}
	if !strings.Contains(err.Error(), "invalid version") {
		t.Errorf("error = %v, want 'invalid version'", err)
	}
}

func TestInit_InvalidType(t *testing.T) {
	dir := t.TempDir()
	_, _, err := execInit(t, dir, "--type", "library")
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
	if !strings.Contains(err.Error(), "invalid type") {
		t.Errorf("error = %v, want 'invalid type'", err)
	}
}

func TestInit_NonInteractiveMissingName(t *testing.T) {
	dir := t.TempDir()
	_, _, err := execInit(t, dir, "--type", "package")
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !strings.Contains(err.Error(), "missing required") {
		t.Errorf("error = %v, want 'missing required'", err)
	}
}

func TestInit_IrrelevantFlagWarnings(t *testing.T) {
	dir := t.TempDir()
	_, stderr, err := execInit(t, dir,
		"--type", "project",
		"--version", "1.0.0",
		"--license", "MIT",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Warning: --version is ignored") {
		t.Errorf("stderr = %q, want version warning", stderr)
	}
	if !strings.Contains(stderr, "Warning: --license is ignored") {
		t.Errorf("stderr = %q, want license warning", stderr)
	}
}

func TestInit_DescriptionTooLong(t *testing.T) {
	dir := t.TempDir()
	longDesc := strings.Repeat("x", 256)
	_, _, err := execInit(t, dir,
		"--type", "package",
		"--name", "@test/a",
		"--version", "1.0.0",
		"--description", longDesc,
	)
	if err == nil {
		t.Fatal("expected error for long description")
	}
	if !strings.Contains(err.Error(), "description too long") {
		t.Errorf("error = %v, want 'description too long'", err)
	}
}

func TestInit_ConsecutiveHyphensInName(t *testing.T) {
	dir := t.TempDir()
	_, _, err := execInit(t, dir,
		"--type", "package",
		"--name", "@test/bad--name",
		"--version", "1.0.0",
	)
	if err == nil {
		t.Fatal("expected error for consecutive hyphens")
	}
	if !strings.Contains(err.Error(), "invalid name") {
		t.Errorf("error = %v, want 'invalid name'", err)
	}
}

func TestInit_JSONFormatting(t *testing.T) {
	dir := t.TempDir()
	_, _, err := execInit(t, dir,
		"--type", "package",
		"--name", "@test/fmt",
		"--version", "1.0.0",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "aipkg.json")) //nolint:gosec // test file path
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	// 2-space indent.
	if !strings.Contains(content, "  \"type\"") {
		t.Error("expected 2-space indent")
	}
	// Trailing newline.
	if !strings.HasSuffix(content, "}\n") {
		t.Error("expected trailing newline")
	}
}
