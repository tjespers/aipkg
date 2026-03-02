package artifact

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/tjespers/aipkg/internal/manifest"
)

func TestDiscover(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, root string)
		want    []manifest.Artifact
		wantErr bool
	}{
		{
			name: "single skill",
			setup: func(t *testing.T, root string) {
				mkSkill(t, root, "writer")
			},
			want: []manifest.Artifact{
				{Name: "writer", Type: "skill", Path: "skills/writer/"},
			},
		},
		{
			name: "single prompt",
			setup: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "prompts", "review.md"), "review this")
			},
			want: []manifest.Artifact{
				{Name: "review", Type: "prompt", Path: "prompts/review.md"},
			},
		},
		{
			name: "multi-type package",
			setup: func(t *testing.T, root string) {
				mkSkill(t, root, "writer")
				writeFile(t, filepath.Join(root, "prompts", "review.md"), "review")
				writeFile(t, filepath.Join(root, "mcp-servers", "github.json"), "{}")
			},
			want: []manifest.Artifact{
				{Name: "writer", Type: "skill", Path: "skills/writer/"},
				{Name: "review", Type: "prompt", Path: "prompts/review.md"},
				{Name: "github", Type: "mcp-server", Path: "mcp-servers/github.json"},
			},
		},
		{
			name: "compound extension",
			setup: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "prompts", "code-review.prompt.md"), "review")
			},
			want: []manifest.Artifact{
				{Name: "code-review", Type: "prompt", Path: "prompts/code-review.prompt.md"},
			},
		},
		{
			name: "no extension file",
			setup: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "prompts", "review"), "review")
			},
			want: []manifest.Artifact{
				{Name: "review", Type: "prompt", Path: "prompts/review"},
			},
		},
		{
			name: "skip hidden files",
			setup: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "prompts", ".draft"), "hidden")
				writeFile(t, filepath.Join(root, "prompts", "review.md"), "visible")
			},
			want: []manifest.Artifact{
				{Name: "review", Type: "prompt", Path: "prompts/review.md"},
			},
		},
		{
			name: "skip nested subdirectories in file-based types",
			setup: func(t *testing.T, root string) {
				mkDir(t, filepath.Join(root, "prompts", "drafts"))
				writeFile(t, filepath.Join(root, "prompts", "drafts", "review.md"), "nested")
				writeFile(t, filepath.Join(root, "prompts", "review.md"), "top-level")
			},
			want: []manifest.Artifact{
				{Name: "review", Type: "prompt", Path: "prompts/review.md"},
			},
		},
		{
			name: "empty well-known dirs",
			setup: func(t *testing.T, root string) {
				mkDir(t, filepath.Join(root, "prompts"))
				mkDir(t, filepath.Join(root, "skills"))
			},
			want: nil,
		},
		{
			name: "no well-known dirs",
			setup: func(t *testing.T, root string) {
				// root exists but no well-known dirs
			},
			want: nil,
		},
		{
			name: "all file-based types",
			setup: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "prompts", "review.md"), "content")
				writeFile(t, filepath.Join(root, "commands", "deploy.md"), "content")
				writeFile(t, filepath.Join(root, "agents", "expert.md"), "content")
				writeFile(t, filepath.Join(root, "agent-instructions", "rules.md"), "content")
				writeFile(t, filepath.Join(root, "mcp-servers", "api.json"), "{}")
			},
			want: []manifest.Artifact{
				{Name: "expert", Type: "agent", Path: "agents/expert.md"},
				{Name: "rules", Type: "agent-instructions", Path: "agent-instructions/rules.md"},
				{Name: "deploy", Type: "command", Path: "commands/deploy.md"},
				{Name: "api", Type: "mcp-server", Path: "mcp-servers/api.json"},
				{Name: "review", Type: "prompt", Path: "prompts/review.md"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			tt.setup(t, root)

			got, err := Discover(root, nil)
			if tt.wantErr {
				if err == nil {
					t.Fatal("Discover() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Discover() error = %v", err)
			}

			// Sort both for stable comparison (map iteration order is random).
			sortArtifacts(got)
			sortArtifacts(tt.want)

			if len(got) != len(tt.want) {
				t.Fatalf("Discover() returned %d artifacts, want %d\ngot: %+v", len(got), len(tt.want), got)
			}
			for i, a := range got {
				w := tt.want[i]
				if a.Name != w.Name || a.Type != w.Type || a.Path != w.Path {
					t.Errorf("artifact[%d] = %+v, want %+v", i, a, w)
				}
			}
		})
	}
}

func TestDeriveName(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"review.md", "review"},
		{"code-review.prompt.md", "code-review"},
		{"my-prompt.txt", "my-prompt"},
		{"review", "review"},
		{"a.b.c.d", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			if got := deriveName(tt.filename); got != tt.want {
				t.Errorf("deriveName(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

// helpers

func mkDir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mkSkill(t *testing.T, root, name string) {
	t.Helper()
	dir := filepath.Join(root, "skills", name)
	mkDir(t, dir)
	writeFile(t, filepath.Join(dir, "SKILL.md"), "---\nname: "+name+"\ndescription: test\n---\n")
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	mkDir(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDiscover_ExcludeFunc(t *testing.T) {
	root := t.TempDir()
	mkSkill(t, root, "writer")
	writeFile(t, filepath.Join(root, "prompts", "review.md"), "content")
	writeFile(t, filepath.Join(root, "prompts", "draft.md"), "content")

	// Exclude the "review" prompt and the "writer" skill.
	exclude := func(path string) bool {
		return path == "prompts/review.md" || path == "skills/writer/"
	}

	got, err := Discover(root, exclude)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 artifact, got %d: %+v", len(got), got)
	}
	if got[0].Name != "draft" {
		t.Errorf("expected draft artifact, got %+v", got[0])
	}
}

func sortArtifacts(arts []manifest.Artifact) {
	sort.Slice(arts, func(i, j int) bool {
		if arts[i].Type != arts[j].Type {
			return arts[i].Type < arts[j].Type
		}
		return arts[i].Name < arts[j].Name
	})
}
