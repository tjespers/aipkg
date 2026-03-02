package ignore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRules_DefaultsOnly(t *testing.T) {
	dir := t.TempDir()

	rules, err := LoadRules(dir, "tjespers--test-1.0.0.aipkg")
	if err != nil {
		t.Fatal(err)
	}

	// Built-in defaults.
	if !rules.IsExcluded(".git") {
		t.Error(".git should be excluded by default")
	}
	if !rules.IsExcluded(".git/config") {
		t.Error(".git/config should be excluded by default")
	}
	if !rules.IsExcluded(".aipkgignore") {
		t.Error(".aipkgignore should be excluded by default")
	}

	// Archive self-exclusion.
	if !rules.IsExcluded("tjespers--test-1.0.0.aipkg") {
		t.Error("archive output should be excluded")
	}
	if !rules.IsExcluded("tjespers--test-1.0.0.aipkg.sha256") {
		t.Error("sidecar should be excluded")
	}

	// Normal files should not be excluded.
	if rules.IsExcluded("skills/test-writer/SKILL.md") {
		t.Error("skills should not be excluded by default")
	}
	if rules.IsExcluded("aipkg.json") {
		t.Error("manifest should not be excluded by default")
	}
}

func TestLoadRules_CustomPatterns(t *testing.T) {
	dir := t.TempDir()

	ignoreContent := "*.log\nbuild/\n"
	if err := os.WriteFile(filepath.Join(dir, ".aipkgignore"), []byte(ignoreContent), 0o644); err != nil {
		t.Fatal(err)
	}

	rules, err := LoadRules(dir, "out.aipkg")
	if err != nil {
		t.Fatal(err)
	}

	if !rules.IsExcluded("debug.log") {
		t.Error("*.log pattern should exclude .log files")
	}
	if !rules.IsExcluded("build/output") {
		t.Error("build/ pattern should exclude build directory contents")
	}

	// Built-in defaults still apply.
	if !rules.IsExcluded(".git") {
		t.Error("built-in defaults should still apply with custom patterns")
	}

	// Non-matching files pass through.
	if rules.IsExcluded("skills/writer/SKILL.md") {
		t.Error("non-matching files should not be excluded")
	}
}

func TestLoadRules_CommentsAndEmpty(t *testing.T) {
	dir := t.TempDir()

	ignoreContent := "# This is a comment\n\n*.tmp\n# Another comment\n"
	if err := os.WriteFile(filepath.Join(dir, ".aipkgignore"), []byte(ignoreContent), 0o644); err != nil {
		t.Fatal(err)
	}

	rules, err := LoadRules(dir, "")
	if err != nil {
		t.Fatal(err)
	}

	if !rules.IsExcluded("notes.tmp") {
		t.Error("*.tmp pattern should work with comments in file")
	}
}

func TestLoadRules_NoArchivePath(t *testing.T) {
	dir := t.TempDir()

	rules, err := LoadRules(dir, "")
	if err != nil {
		t.Fatal(err)
	}

	// Defaults still work.
	if !rules.IsExcluded(".git") {
		t.Error(".git should be excluded")
	}
	if rules.IsExcluded("prompts/review.md") {
		t.Error("normal files should not be excluded")
	}
}

func TestLoadRules_ExcludeWellKnownDir(t *testing.T) {
	dir := t.TempDir()

	// FR-028: .aipkgignore can exclude well-known directory contents.
	ignoreContent := "skills/\n"
	if err := os.WriteFile(filepath.Join(dir, ".aipkgignore"), []byte(ignoreContent), 0o644); err != nil {
		t.Fatal(err)
	}

	rules, err := LoadRules(dir, "")
	if err != nil {
		t.Fatal(err)
	}

	if !rules.IsExcluded("skills/writer/SKILL.md") {
		t.Error("skills/ pattern should exclude skill files")
	}
	if rules.IsExcluded("prompts/review.md") {
		t.Error("prompts should not be affected by skills/ exclusion")
	}
}

func TestLoadRules_MalformedPattern(t *testing.T) {
	dir := t.TempDir()

	// Unclosed bracket is a malformed glob pattern. go-gitignore handles it
	// gracefully (no panic, no error). Valid patterns alongside it still work.
	ignoreContent := "[unclosed\n*.log\n"
	if err := os.WriteFile(filepath.Join(dir, ".aipkgignore"), []byte(ignoreContent), 0o644); err != nil {
		t.Fatal(err)
	}

	rules, err := LoadRules(dir, "")
	if err != nil {
		t.Fatalf("LoadRules should not fail on malformed patterns: %v", err)
	}

	// Valid pattern still works.
	if !rules.IsExcluded("debug.log") {
		t.Error("*.log pattern should still work alongside malformed pattern")
	}

	// Built-in defaults still apply.
	if !rules.IsExcluded(".git") {
		t.Error("built-in defaults should still apply")
	}
}

func TestNilRules_IsExcluded(t *testing.T) {
	var rules *Rules
	if rules.IsExcluded("anything") {
		t.Error("nil rules should not exclude anything")
	}
}
