package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// executeCreate runs the create command with the given args and returns stdout and error.
func executeCreate(args ...string) (string, error) {
	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(append([]string{"create"}, args...))

	// Force non-TTY mode by setting stdin to nil (cobra won't detect TTY).
	root.SetIn(bytes.NewReader(nil))

	err := root.Execute()
	return buf.String(), err
}

// parsedManifest reads and parses the aipkg.json from a directory.
type parsedManifest struct {
	SpecVersion int    `json:"specVersion"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
	License     string `json:"license,omitempty"`
}

func readManifest(t *testing.T, dir string) parsedManifest {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "aipkg.json"))
	if err != nil {
		t.Fatalf("reading aipkg.json: %v", err)
	}
	var m parsedManifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parsing aipkg.json: %v", err)
	}
	return m
}

func wellKnownDirs() []string {
	return []string{"agents", "agent-instructions", "commands", "mcp-servers", "prompts", "skills"}
}

// ===========================================================================
// US1: Create a new package from scratch
// ===========================================================================

func TestCreateNonInteractive_BasicFlow(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "blog-writer")

	out, err := executeCreate(
		"--name", "@alice/blog-writer",
		"--version", "0.1.0",
		"--path", target,
	)
	if err != nil {
		t.Fatalf("create failed: %v\noutput: %s", err, out)
	}

	// Verify aipkg.json content.
	m := readManifest(t, target)
	if m.SpecVersion != 1 {
		t.Errorf("specVersion = %d, want 1", m.SpecVersion)
	}
	if m.Name != "@alice/blog-writer" {
		t.Errorf("name = %q, want %q", m.Name, "@alice/blog-writer")
	}
	if m.Version != "0.1.0" {
		t.Errorf("version = %q, want %q", m.Version, "0.1.0")
	}

	// Verify well-known directories.
	for _, d := range wellKnownDirs() {
		info, err := os.Stat(filepath.Join(target, d))
		if err != nil {
			t.Errorf("expected directory %s: %v", d, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s should be a directory", d)
		}
	}
}

func TestCreateNonInteractive_AllFields(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "blog-writer")

	_, err := executeCreate(
		"--name", "@alice/blog-writer",
		"--version", "1.0.0",
		"--description", "AI blog writing assistant",
		"--license", "MIT",
		"--path", target,
	)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	m := readManifest(t, target)
	if m.Description != "AI blog writing assistant" {
		t.Errorf("description = %q, want %q", m.Description, "AI blog writing assistant")
	}
	if m.License != "MIT" {
		t.Errorf("license = %q, want %q", m.License, "MIT")
	}
}

func TestCreateNonInteractive_PositionalArg(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "my-pkg")

	_, err := executeCreate(
		"@alice/my-pkg",
		"--version", "0.1.0",
		"--path", target,
	)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	m := readManifest(t, target)
	if m.Name != "@alice/my-pkg" {
		t.Errorf("name = %q, want %q", m.Name, "@alice/my-pkg")
	}
}

func TestCreateNonInteractive_PositionalOverridesFlag(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "from-arg")

	_, err := executeCreate(
		"@alice/from-arg",
		"--name", "@alice/from-flag",
		"--version", "0.1.0",
		"--path", target,
	)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	m := readManifest(t, target)
	if m.Name != "@alice/from-arg" {
		t.Errorf("positional arg should take precedence, got name = %q", m.Name)
	}
}

func TestCreateNonInteractive_NoArtifactsField(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "pkg")

	_, err := executeCreate(
		"--name", "@alice/pkg",
		"--version", "0.1.0",
		"--path", target,
	)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(target, "aipkg.json"))
	if bytes.Contains(data, []byte("artifacts")) {
		t.Error("aipkg.json should not contain artifacts field at creation time")
	}
}

func TestCreateNonInteractive_OmitsEmptyFields(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "pkg")

	_, err := executeCreate(
		"--name", "@alice/pkg",
		"--version", "0.1.0",
		"--path", target,
	)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(target, "aipkg.json"))
	if bytes.Contains(data, []byte("description")) {
		t.Error("empty description should be omitted")
	}
	if bytes.Contains(data, []byte("license")) {
		t.Error("empty license should be omitted")
	}
}

// ===========================================================================
// US2: Create in existing directory
// ===========================================================================

func TestCreate_PreservesExistingFiles(t *testing.T) {
	dir := t.TempDir()

	// Create pre-existing files.
	readme := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readme, []byte("# My Package"), 0o644); err != nil {
		t.Fatal(err)
	}
	skillsDir := filepath.Join(dir, "skills", "my-skill")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	skillFile := filepath.Join(skillsDir, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte("# Skill"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := executeCreate(
		"--name", "@alice/my-pkg",
		"--version", "0.1.0",
		"--path", dir,
	)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// README should be preserved.
	data, err := os.ReadFile(readme)
	if err != nil {
		t.Fatal("README.md was removed")
	}
	if string(data) != "# My Package" {
		t.Error("README.md content was modified")
	}

	// Existing skill file preserved.
	data, err = os.ReadFile(skillFile)
	if err != nil {
		t.Fatal("skills/my-skill/SKILL.md was removed")
	}
	if string(data) != "# Skill" {
		t.Error("SKILL.md content was modified")
	}

	// aipkg.json should exist.
	if _, err := os.Stat(filepath.Join(dir, "aipkg.json")); err != nil {
		t.Error("aipkg.json should have been created")
	}
}

func TestCreate_RejectsExistingManifest(t *testing.T) {
	dir := t.TempDir()

	// Create an existing aipkg.json.
	if err := os.WriteFile(filepath.Join(dir, "aipkg.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := executeCreate(
		"--name", "@alice/my-pkg",
		"--version", "0.1.0",
		"--path", dir,
	)
	if err == nil {
		t.Fatal("expected error when aipkg.json already exists")
	}
}

func TestCreate_CreatesNonExistentPath(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "new-dir", "nested")

	_, err := executeCreate(
		"--name", "@alice/my-pkg",
		"--version", "0.1.0",
		"--path", target,
	)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(target, "aipkg.json")); err != nil {
		t.Error("aipkg.json should exist in new directory")
	}
}

func TestCreate_DerivesDirectoryFromName(t *testing.T) {
	// Change to a temp dir so the derived directory is created there.
	orig, _ := os.Getwd()
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	_, err := executeCreate(
		"--name", "@alice/blog-writer",
		"--version", "0.1.0",
	)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	target := filepath.Join(dir, "blog-writer")
	if _, err := os.Stat(filepath.Join(target, "aipkg.json")); err != nil {
		t.Error("should create blog-writer/ directory derived from package name")
	}
}

// ===========================================================================
// US3: Validate package name during creation
// ===========================================================================

func TestCreate_RejectsUnscopedName(t *testing.T) {
	dir := t.TempDir()

	_, err := executeCreate(
		"--name", "blog-writer",
		"--version", "0.1.0",
		"--path", dir,
	)
	if err == nil {
		t.Fatal("expected error for unscoped name")
	}
}

func TestCreate_RejectsReservedScope(t *testing.T) {
	dir := t.TempDir()

	_, err := executeCreate(
		"--name", "@aipkg/my-tool",
		"--version", "0.1.0",
		"--path", dir,
	)
	if err == nil {
		t.Fatal("expected error for reserved scope")
	}
}

func TestCreate_RejectsInvalidVersion(t *testing.T) {
	dir := t.TempDir()

	_, err := executeCreate(
		"--name", "@alice/my-pkg",
		"--version", "bad",
		"--path", dir,
	)
	if err == nil {
		t.Fatal("expected error for invalid version")
	}
}

func TestCreate_RejectsInvalidNameFormat(t *testing.T) {
	dir := t.TempDir()

	_, err := executeCreate(
		"--name", "@Alice/Blog_Writer",
		"--version", "0.1.0",
		"--path", dir,
	)
	if err == nil {
		t.Fatal("expected error for invalid name characters")
	}
}

func TestCreate_RejectsConsecutiveHyphens(t *testing.T) {
	dir := t.TempDir()

	_, err := executeCreate(
		"--name", "@alice/blog--writer",
		"--version", "0.1.0",
		"--path", dir,
	)
	if err == nil {
		t.Fatal("expected error for consecutive hyphens")
	}
}

// ===========================================================================
// US4: Non-interactive mode
// ===========================================================================

func TestCreate_NonInteractiveMissingName(t *testing.T) {
	// executeCreate always runs without TTY.
	_, err := executeCreate("--version", "0.1.0")
	if err == nil {
		t.Fatal("expected error for missing --name without TTY")
	}
}

func TestCreate_NonInteractiveMissingVersion(t *testing.T) {
	_, err := executeCreate("--name", "@alice/my-pkg")
	if err == nil {
		t.Fatal("expected error for missing --version without TTY")
	}
}

func TestCreate_NonInteractiveInvalidFlagValue(t *testing.T) {
	dir := t.TempDir()

	_, err := executeCreate(
		"--name", "@alice/my-pkg",
		"--version", "not-semver",
		"--path", dir,
	)
	if err == nil {
		t.Fatal("expected error for invalid flag value")
	}

	// No files should be created.
	entries, _ := os.ReadDir(dir)
	if len(entries) > 0 {
		t.Error("no files should be created on validation error")
	}
}

// ===========================================================================
// US5: License detection
// ===========================================================================

func TestCreate_DetectsLicenseFile(t *testing.T) {
	dir := t.TempDir()

	// Copy project's LICENSE (Apache-2.0) into the target.
	licenseData, err := os.ReadFile(filepath.Join("..", "..", "LICENSE"))
	if err != nil {
		t.Skip("project LICENSE not available")
	}
	if err := os.WriteFile(filepath.Join(dir, "LICENSE"), licenseData, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err = executeCreate(
		"--name", "@alice/my-pkg",
		"--version", "0.1.0",
		"--path", dir,
	)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	m := readManifest(t, dir)
	if m.License != "Apache-2.0" {
		t.Errorf("license = %q, want %q (detected from LICENSE file)", m.License, "Apache-2.0")
	}
}

func TestCreate_NoLicenseWithoutFile(t *testing.T) {
	dir := t.TempDir()

	_, err := executeCreate(
		"--name", "@alice/my-pkg",
		"--version", "0.1.0",
		"--path", dir,
	)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	m := readManifest(t, dir)
	if m.License != "" {
		t.Errorf("license = %q, want empty when no LICENSE file", m.License)
	}
}

func TestCreate_FlagOverridesDetection(t *testing.T) {
	dir := t.TempDir()

	// Create a LICENSE file.
	licenseData, err := os.ReadFile(filepath.Join("..", "..", "LICENSE"))
	if err != nil {
		t.Skip("project LICENSE not available")
	}
	if err := os.WriteFile(filepath.Join(dir, "LICENSE"), licenseData, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err = executeCreate(
		"--name", "@alice/my-pkg",
		"--version", "0.1.0",
		"--license", "MIT",
		"--path", dir,
	)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	m := readManifest(t, dir)
	if m.License != "MIT" {
		t.Errorf("license = %q, want %q (flag should override detection)", m.License, "MIT")
	}
}

// ===========================================================================
// Edge cases
// ===========================================================================

func TestCreate_HelpFlag(t *testing.T) {
	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"create", "--help"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("help failed: %v", err)
	}

	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("Create a new aipkg package")) {
		t.Error("help output should contain command description")
	}
}
