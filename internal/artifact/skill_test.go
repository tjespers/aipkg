package artifact

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tjespers/aipkg/internal/manifest"
)

func TestValidateSkill(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, root string)
		wantErrStr string // substring expected in error, empty for no error
	}{
		{
			name: "valid skill",
			setup: func(t *testing.T, root string) {
				writeSkillMD(t, root, "writer", "---\nname: writer\ndescription: A test skill\n---\n\n# Instructions\n")
			},
		},
		{
			name: "valid skill with optional fields",
			setup: func(t *testing.T, root string) {
				writeSkillMD(t, root, "writer", "---\nname: writer\ndescription: A test skill\nlicense: MIT\ncompatibility:\n  - claude\nallowed-tools:\n  - bash\n---\n")
			},
		},
		{
			name: "missing SKILL.md",
			setup: func(t *testing.T, root string) {
				mkDir(t, filepath.Join(root, "skills", "broken"))
			},
			wantErrStr: "missing required SKILL.md file",
		},
		{
			name: "missing name field",
			setup: func(t *testing.T, root string) {
				writeSkillMD(t, root, "writer", "---\ndescription: A test skill\n---\n")
			},
			wantErrStr: "missing required field 'name'",
		},
		{
			name: "missing description field",
			setup: func(t *testing.T, root string) {
				writeSkillMD(t, root, "writer", "---\nname: writer\n---\n")
			},
			wantErrStr: "missing required field 'description'",
		},
		{
			name: "name mismatch with directory",
			setup: func(t *testing.T, root string) {
				writeSkillMD(t, root, "writer", "---\nname: reader\ndescription: Wrong name\n---\n")
			},
			wantErrStr: "does not match directory name",
		},
		{
			name: "unknown frontmatter key",
			setup: func(t *testing.T, root string) {
				writeSkillMD(t, root, "writer", "---\nname: writer\ndescription: A test skill\nauthor: alice\n---\n")
			},
			wantErrStr: "author",
		},
		{
			name: "invalid YAML",
			setup: func(t *testing.T, root string) {
				writeSkillMD(t, root, "writer", "---\n: invalid yaml [[\n---\n")
			},
			wantErrStr: "invalid frontmatter YAML",
		},
		{
			name: "missing frontmatter delimiters",
			setup: func(t *testing.T, root string) {
				writeSkillMD(t, root, "writer", "no frontmatter here\n")
			},
			wantErrStr: "invalid frontmatter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			tt.setup(t, root)

			art := manifest.Artifact{
				Name: "writer",
				Type: "skill",
				Path: "skills/writer/",
			}

			var errs ValidationErrors
			validateSkill(root, art, &errs)

			err := errs.Err()
			if tt.wantErrStr == "" {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErrStr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErrStr)
			}
		})
	}
}

func writeSkillMD(t *testing.T, root, name, content string) {
	t.Helper()
	dir := filepath.Join(root, "skills", name)
	mkDir(t, dir)
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
