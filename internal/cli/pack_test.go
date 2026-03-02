package cli

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tjespers/aipkg/internal/manifest"
)

// executePack runs the pack command with the given args and returns stderr and error.
func executePack(args ...string) (string, error) {
	root := newRootCmd()
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs(append([]string{"pack"}, args...))

	err := root.Execute()
	return stderr.String(), err
}

// setupPackageDir creates a valid package directory for testing.
func setupPackageDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	m := &manifest.PackageManifest{
		SpecVersion: 1,
		Name:        "@tjespers/test-writer",
		Version:     "1.0.0",
	}
	if err := m.WriteFile(dir); err != nil {
		t.Fatal(err)
	}

	return dir
}

func addSkill(t *testing.T, dir, name string) {
	t.Helper()
	skillDir := filepath.Join(dir, "skills", name)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := fmt.Sprintf("---\nname: %s\ndescription: A test skill\n---\n\n# Instructions\n", name)
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func addPrompt(t *testing.T, dir, name string) {
	t.Helper()
	promptDir := filepath.Join(dir, "prompts")
	if err := os.MkdirAll(promptDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(promptDir, name+".md"), []byte("prompt content"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func addMCPServer(t *testing.T, dir, name string) {
	t.Helper()
	mcpDir := filepath.Join(dir, "mcp-servers")
	if err := os.MkdirAll(mcpDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mcpDir, name+".json"), []byte(`{"command":"test"}`), 0o644); err != nil {
		t.Fatal(err)
	}
}

// ===========================================================================
// US1: Pack a package into a distributable archive
// ===========================================================================

// AS-1: Basic pack produces archive and sidecar with correct names.
func TestPack_BasicFlow(t *testing.T) {
	dir := setupPackageDir(t)
	addSkill(t, dir, "test-writer")

	stderr, err := executePack(dir)
	if err != nil {
		t.Fatalf("pack failed: %v\nstderr: %s", err, stderr)
	}

	archivePath := "tjespers--test-writer-1.0.0.aipkg"
	sidecarPath := archivePath + ".sha256"

	if _, err := os.Stat(archivePath); err != nil {
		t.Fatalf("archive not created: %v", err)
	}
	defer os.Remove(archivePath) //nolint:errcheck // test cleanup
	defer os.Remove(sidecarPath) //nolint:errcheck // test cleanup

	if _, err := os.Stat(sidecarPath); err != nil {
		t.Fatalf("sidecar not created: %v", err)
	}

	// Verify summary output.
	if !strings.Contains(stderr, "1 artifact") {
		t.Errorf("stderr should mention artifact count, got: %s", stderr)
	}
}

// AS-2: Archive contains correct structure and enriched manifest.
func TestPack_ArchiveContents(t *testing.T) {
	dir := setupPackageDir(t)
	addSkill(t, dir, "test-writer")
	addPrompt(t, dir, "code-review")

	stderr, err := executePack(dir)
	if err != nil {
		t.Fatalf("pack failed: %v\nstderr: %s", err, stderr)
	}

	archivePath := "tjespers--test-writer-1.0.0.aipkg"
	defer os.Remove(archivePath)             //nolint:errcheck // test cleanup
	defer os.Remove(archivePath + ".sha256") //nolint:errcheck // test cleanup

	// Read the archive.
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatalf("opening archive: %v", err)
	}
	defer reader.Close() //nolint:errcheck // test cleanup

	entries := make(map[string]bool)
	var manifestData []byte
	for _, f := range reader.File {
		entries[f.Name] = true
		if f.Name == "test-writer/aipkg.json" {
			rc, err := f.Open()
			if err != nil {
				t.Fatal(err)
			}
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(rc)
			_ = rc.Close()
			manifestData = buf.Bytes()
		}
	}

	// Verify top-level directory.
	if !entries["test-writer/aipkg.json"] {
		t.Error("archive missing test-writer/aipkg.json")
	}
	if !entries["test-writer/skills/test-writer/SKILL.md"] {
		t.Error("archive missing test-writer/skills/test-writer/SKILL.md")
	}
	if !entries["test-writer/prompts/code-review.md"] {
		t.Error("archive missing test-writer/prompts/code-review.md")
	}

	// Verify enriched manifest has artifacts array.
	if manifestData == nil {
		t.Fatal("manifest not found in archive")
	}
	var m manifest.PackageManifest
	if err := json.Unmarshal(manifestData, &m); err != nil {
		t.Fatalf("parsing archived manifest: %v", err)
	}
	if len(m.Artifacts) != 2 {
		t.Errorf("archived manifest has %d artifacts, want 2", len(m.Artifacts))
	}
}

// AS-2 continued: Original aipkg.json is unchanged.
func TestPack_OriginalManifestUnchanged(t *testing.T) {
	dir := setupPackageDir(t)
	addSkill(t, dir, "test-writer")

	original, _ := os.ReadFile(filepath.Join(dir, "aipkg.json"))

	stderr, err := executePack(dir)
	if err != nil {
		t.Fatalf("pack failed: %v\nstderr: %s", err, stderr)
	}
	defer os.Remove("tjespers--test-writer-1.0.0.aipkg")        //nolint:errcheck // test cleanup
	defer os.Remove("tjespers--test-writer-1.0.0.aipkg.sha256") //nolint:errcheck // test cleanup

	after, _ := os.ReadFile(filepath.Join(dir, "aipkg.json"))
	if !bytes.Equal(original, after) {
		t.Error("original aipkg.json was modified by pack")
	}
}

// AS-3: Missing SKILL.md fails validation.
func TestPack_MissingSKILLmd(t *testing.T) {
	dir := setupPackageDir(t)

	// Create a skill directory without SKILL.md.
	skillDir := filepath.Join(dir, "skills", "broken")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}

	stderr, err := executePack(dir)
	if err == nil {
		t.Fatal("expected error for missing SKILL.md")
	}
	if !strings.Contains(stderr, "validation") {
		t.Errorf("stderr should mention validation, got: %s", stderr)
	}

	// No archive should be created.
	if _, err := os.Stat("tjespers--test-writer-1.0.0.aipkg"); err == nil {
		_ = os.Remove("tjespers--test-writer-1.0.0.aipkg")
		t.Error("archive should not be created on validation failure")
	}
}

// AS-4: No artifacts fails.
func TestPack_NoArtifacts(t *testing.T) {
	dir := setupPackageDir(t)

	_, err := executePack(dir)
	if err == nil {
		t.Fatal("expected error when no artifacts discovered")
	}
	if !strings.Contains(err.Error(), "no artifacts") {
		t.Errorf("error should mention no artifacts, got: %v", err)
	}
}

// AS-5: Invalid JSON in mcp-server fails.
func TestPack_InvalidJSON(t *testing.T) {
	dir := setupPackageDir(t)

	mcpDir := filepath.Join(dir, "mcp-servers")
	if err := os.MkdirAll(mcpDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mcpDir, "github.json"), []byte("{bad json}"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := executePack(dir)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("error should mention invalid JSON, got: %v", err)
	}
}

// AS-6: Multiple artifact types all included.
func TestPack_MultipleTypes(t *testing.T) {
	dir := setupPackageDir(t)
	addSkill(t, dir, "test-writer")
	addPrompt(t, dir, "review")
	addMCPServer(t, dir, "github")

	stderr, err := executePack(dir)
	if err != nil {
		t.Fatalf("pack failed: %v\nstderr: %s", err, stderr)
	}

	archivePath := "tjespers--test-writer-1.0.0.aipkg"
	defer os.Remove(archivePath)             //nolint:errcheck // test cleanup
	defer os.Remove(archivePath + ".sha256") //nolint:errcheck // test cleanup

	if !strings.Contains(stderr, "3 artifacts") {
		t.Errorf("stderr should mention 3 artifacts, got: %s", stderr)
	}
}

// Sidecar verifies correctly with sha256sum format.
func TestPack_SidecarVerifies(t *testing.T) {
	dir := setupPackageDir(t)
	addSkill(t, dir, "test-writer")

	_, err := executePack(dir)
	if err != nil {
		t.Fatalf("pack failed: %v", err)
	}

	archivePath := "tjespers--test-writer-1.0.0.aipkg"
	sidecarPath := archivePath + ".sha256"
	defer os.Remove(archivePath) //nolint:errcheck // test cleanup
	defer os.Remove(sidecarPath) //nolint:errcheck // test cleanup

	// Read sidecar.
	sidecarData, _ := os.ReadFile(sidecarPath)
	parts := strings.SplitN(strings.TrimSpace(string(sidecarData)), "  ", 2)
	if len(parts) != 2 {
		t.Fatalf("invalid sidecar format: %q", string(sidecarData))
	}

	// Compute actual hash.
	archiveData, _ := os.ReadFile(archivePath)
	actualHash := fmt.Sprintf("%x", sha256.Sum256(archiveData))

	if parts[0] != actualHash {
		t.Errorf("sidecar hash mismatch: got %q, computed %q", parts[0], actualHash)
	}
	if parts[1] != filepath.Base(archivePath) {
		t.Errorf("sidecar filename = %q, want %q", parts[1], filepath.Base(archivePath))
	}
}

// Edge case: Missing aipkg.json.
func TestPack_MissingManifest(t *testing.T) {
	dir := t.TempDir()

	_, err := executePack(dir)
	if err == nil {
		t.Fatal("expected error for missing aipkg.json")
	}
}

// Edge case: Invalid manifest (schema validation fails before discovery).
func TestPack_InvalidManifest(t *testing.T) {
	dir := t.TempDir()
	// Write a manifest missing required fields.
	if err := os.WriteFile(filepath.Join(dir, "aipkg.json"), []byte(`{"specVersion": 1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	addSkill(t, dir, "writer")

	_, err := executePack(dir)
	if err == nil {
		t.Fatal("expected error for invalid manifest")
	}
	if !strings.Contains(err.Error(), "manifest validation") {
		t.Errorf("error should mention manifest validation, got: %v", err)
	}
}

// ===========================================================================
// US2: Exclude files from the archive
// ===========================================================================

// US2 AS-1: .aipkgignore patterns exclude matching files.
func TestPack_IgnorePattern(t *testing.T) {
	dir := setupPackageDir(t)
	addSkill(t, dir, "test-writer")
	addPrompt(t, dir, "review")
	addPrompt(t, dir, "draft")

	// Exclude the draft prompt via .aipkgignore.
	if err := os.WriteFile(filepath.Join(dir, ".aipkgignore"), []byte("prompts/draft.md\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stderr, err := executePack(dir)
	if err != nil {
		t.Fatalf("pack failed: %v\nstderr: %s", err, stderr)
	}

	archivePath := "tjespers--test-writer-1.0.0.aipkg"
	defer os.Remove(archivePath)             //nolint:errcheck // test cleanup
	defer os.Remove(archivePath + ".sha256") //nolint:errcheck // test cleanup

	// Should have 2 artifacts (skill + review prompt), not 3.
	if !strings.Contains(stderr, "2 artifacts") {
		t.Errorf("expected 2 artifacts, got stderr: %s", stderr)
	}
}

// US2 AS-2: Built-in defaults exclude .git and .aipkgignore without user config.
func TestPack_DefaultExclusions(t *testing.T) {
	dir := setupPackageDir(t)
	addSkill(t, dir, "test-writer")

	// Create .git directory and .aipkgignore (should not appear in archive).
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".git", "config"), []byte("gitconfig"), 0o644); err != nil {
		t.Fatal(err)
	}

	stderr, err := executePack(dir)
	if err != nil {
		t.Fatalf("pack failed: %v\nstderr: %s", err, stderr)
	}

	archivePath := "tjespers--test-writer-1.0.0.aipkg"
	defer os.Remove(archivePath)             //nolint:errcheck // test cleanup
	defer os.Remove(archivePath + ".sha256") //nolint:errcheck // test cleanup

	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatalf("opening archive: %v", err)
	}
	defer reader.Close() //nolint:errcheck // test cleanup

	for _, f := range reader.File {
		if strings.Contains(f.Name, ".git") {
			t.Errorf("archive should not contain .git entries, found: %s", f.Name)
		}
		if strings.Contains(f.Name, ".aipkgignore") {
			t.Errorf("archive should not contain .aipkgignore, found: %s", f.Name)
		}
	}
}

// US2 AS-3: .aipkgignore can exclude well-known directory contents.
func TestPack_IgnoreOverridesConvention(t *testing.T) {
	dir := setupPackageDir(t)
	addSkill(t, dir, "test-writer")
	addPrompt(t, dir, "review")

	// Exclude all prompts.
	if err := os.WriteFile(filepath.Join(dir, ".aipkgignore"), []byte("prompts/\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stderr, err := executePack(dir)
	if err != nil {
		t.Fatalf("pack failed: %v\nstderr: %s", err, stderr)
	}

	archivePath := "tjespers--test-writer-1.0.0.aipkg"
	defer os.Remove(archivePath)             //nolint:errcheck // test cleanup
	defer os.Remove(archivePath + ".sha256") //nolint:errcheck // test cleanup

	// Only the skill should remain.
	if !strings.Contains(stderr, "1 artifact") {
		t.Errorf("expected 1 artifact after ignore, got stderr: %s", stderr)
	}
}

// US2 edge case: .aipkgignore excludes all artifacts -> error.
func TestPack_IgnoreExcludesAll(t *testing.T) {
	dir := setupPackageDir(t)
	addPrompt(t, dir, "review")

	if err := os.WriteFile(filepath.Join(dir, ".aipkgignore"), []byte("prompts/\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := executePack(dir)
	if err == nil {
		t.Fatal("expected error when all artifacts excluded")
	}
	if !strings.Contains(err.Error(), "no artifacts") {
		t.Errorf("error should mention no artifacts, got: %v", err)
	}
}

// ===========================================================================
// US3: Control output location
// ===========================================================================

// US3 AS-1: --output with a custom file path.
func TestPack_OutputFilePath(t *testing.T) {
	dir := setupPackageDir(t)
	addSkill(t, dir, "test-writer")

	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "custom.aipkg")

	stderr, err := executePack(dir, "--output", outPath)
	if err != nil {
		t.Fatalf("pack failed: %v\nstderr: %s", err, stderr)
	}

	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("archive not at custom path: %v", err)
	}
	if _, err := os.Stat(outPath + ".sha256"); err != nil {
		t.Fatalf("sidecar not at custom path: %v", err)
	}

	// Default location should not have an archive.
	if _, err := os.Stat("tjespers--test-writer-1.0.0.aipkg"); err == nil {
		_ = os.Remove("tjespers--test-writer-1.0.0.aipkg")
		t.Error("archive should not be at default location when --output is used")
	}
}

// US3 AS-2: --output with a directory path uses conventional filename.
func TestPack_OutputDirectory(t *testing.T) {
	dir := setupPackageDir(t)
	addSkill(t, dir, "test-writer")

	outDir := t.TempDir()

	stderr, err := executePack(dir, "--output", outDir+"/")
	if err != nil {
		t.Fatalf("pack failed: %v\nstderr: %s", err, stderr)
	}

	expected := filepath.Join(outDir, "tjespers--test-writer-1.0.0.aipkg")
	if _, err := os.Stat(expected); err != nil {
		t.Fatalf("archive not in output directory: %v", err)
	}
	if _, err := os.Stat(expected + ".sha256"); err != nil {
		t.Fatalf("sidecar not in output directory: %v", err)
	}
}

// US3 AS-3: --output overwrites existing file silently.
func TestPack_OutputOverwrite(t *testing.T) {
	dir := setupPackageDir(t)
	addSkill(t, dir, "test-writer")

	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "existing.aipkg")

	// Create a pre-existing file.
	if err := os.WriteFile(outPath, []byte("old content"), 0o644); err != nil {
		t.Fatal(err)
	}

	stderr, err := executePack(dir, "--output", outPath)
	if err != nil {
		t.Fatalf("pack failed: %v\nstderr: %s", err, stderr)
	}

	// Verify it was overwritten (should be a valid zip, not "old content").
	data, _ := os.ReadFile(outPath)
	if string(data) == "old content" {
		t.Error("existing file was not overwritten")
	}
	if len(data) < 10 {
		t.Error("overwritten file is suspiciously small")
	}
}

// US3 edge case: --output parent directory does not exist.
func TestPack_OutputMissingParent(t *testing.T) {
	dir := setupPackageDir(t)
	addSkill(t, dir, "test-writer")

	_, err := executePack(dir, "--output", "/nonexistent/dir/archive.aipkg")
	if err == nil {
		t.Fatal("expected error for missing parent directory")
	}
	if !strings.Contains(err.Error(), "output directory does not exist") {
		t.Errorf("error should mention missing directory, got: %v", err)
	}
}

// ===========================================================================
// Edge cases
// ===========================================================================

// Edge case: Same artifact name across two well-known directories.
func TestPack_DuplicateNameAcrossDirs(t *testing.T) {
	dir := setupPackageDir(t)
	addPrompt(t, dir, "review")
	addMCPServer(t, dir, "review") // same name, different type

	stderr, err := executePack(dir)
	if err != nil {
		t.Fatalf("pack failed: %v\nstderr: %s", err, stderr)
	}

	archivePath := "tjespers--test-writer-1.0.0.aipkg"
	defer os.Remove(archivePath)             //nolint:errcheck // test cleanup
	defer os.Remove(archivePath + ".sha256") //nolint:errcheck // test cleanup

	if !strings.Contains(stderr, "2 artifacts") {
		t.Errorf("expected 2 artifacts (both types), got stderr: %s", stderr)
	}

	// Verify both files present in archive.
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatalf("opening archive: %v", err)
	}
	defer reader.Close() //nolint:errcheck // test cleanup

	entries := make(map[string]bool)
	for _, f := range reader.File {
		entries[f.Name] = true
	}
	if !entries["test-writer/prompts/review.md"] {
		t.Error("archive missing prompts/review.md")
	}
	if !entries["test-writer/mcp-servers/review.json"] {
		t.Error("archive missing mcp-servers/review.json")
	}
}

// Edge case: Write permission failure on output path.
func TestPack_OutputWritePermission(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping permission test as root")
	}

	dir := setupPackageDir(t)
	addSkill(t, dir, "test-writer")

	// Create a read-only directory.
	readonlyDir := t.TempDir()
	if err := os.Chmod(readonlyDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(readonlyDir, 0o755) }) //nolint:errcheck // test cleanup

	outPath := filepath.Join(readonlyDir, "output.aipkg")

	_, err := executePack(dir, "--output", outPath)
	if err == nil {
		t.Fatal("expected error for write permission failure")
	}
}

// Edge case: Empty file in file-based type.
func TestPack_EmptyFile(t *testing.T) {
	dir := setupPackageDir(t)
	promptDir := filepath.Join(dir, "prompts")
	if err := os.MkdirAll(promptDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(promptDir, "empty.md"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := executePack(dir)
	if err == nil {
		t.Fatal("expected error for empty file")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("error should mention empty, got: %v", err)
	}
}
