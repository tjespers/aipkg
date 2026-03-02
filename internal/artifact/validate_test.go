package artifact

import (
	"strings"
	"testing"

	"github.com/tjespers/aipkg/internal/manifest"
)

func TestValidateAll(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, root string)
		artifacts  []manifest.Artifact
		wantErr    bool
		wantErrStr string
		errCount   int // expected number of errors (0 = don't check)
	}{
		{
			name: "valid multi-type package",
			setup: func(t *testing.T, root string) {
				mkSkill(t, root, "writer")
				writeFile(t, root+"/prompts/review.md", "review content")
				writeFile(t, root+"/mcp-servers/github.json", `{"command":"npx"}`)
			},
			artifacts: []manifest.Artifact{
				{Name: "writer", Type: "skill", Path: "skills/writer/"},
				{Name: "review", Type: "prompt", Path: "prompts/review.md"},
				{Name: "github", Type: "mcp-server", Path: "mcp-servers/github.json"},
			},
		},
		{
			name: "invalid artifact name",
			setup: func(t *testing.T, root string) {
				writeFile(t, root+"/prompts/review.md", "content")
			},
			artifacts: []manifest.Artifact{
				{Name: "My_Bad_Name", Type: "prompt", Path: "prompts/review.md"},
			},
			wantErr:    true,
			wantErrStr: "invalid artifact name",
		},
		{
			name: "empty file for file-based type",
			setup: func(t *testing.T, root string) {
				writeFile(t, root+"/prompts/empty.md", "")
			},
			artifacts: []manifest.Artifact{
				{Name: "empty", Type: "prompt", Path: "prompts/empty.md"},
			},
			wantErr:    true,
			wantErrStr: "file must not be empty",
		},
		{
			name: "invalid JSON for mcp-server",
			setup: func(t *testing.T, root string) {
				writeFile(t, root+"/mcp-servers/bad.json", "{not json}")
			},
			artifacts: []manifest.Artifact{
				{Name: "bad", Type: "mcp-server", Path: "mcp-servers/bad.json"},
			},
			wantErr:    true,
			wantErrStr: "invalid JSON",
		},
		{
			name: "multiple errors collected",
			setup: func(t *testing.T, root string) {
				writeFile(t, root+"/prompts/empty.md", "")
				writeFile(t, root+"/mcp-servers/bad.json", "not json")
			},
			artifacts: []manifest.Artifact{
				{Name: "empty", Type: "prompt", Path: "prompts/empty.md"},
				{Name: "bad", Type: "mcp-server", Path: "mcp-servers/bad.json"},
			},
			wantErr:  true,
			errCount: 2,
		},
		{
			name: "valid non-md extensions",
			setup: func(t *testing.T, root string) {
				writeFile(t, root+"/prompts/review.txt", "content")
				writeFile(t, root+"/commands/deploy.prompt.md", "content")
			},
			artifacts: []manifest.Artifact{
				{Name: "review", Type: "prompt", Path: "prompts/review.txt"},
				{Name: "deploy", Type: "command", Path: "commands/deploy.prompt.md"},
			},
		},
		{
			name: "all file-based types validated as non-empty",
			setup: func(t *testing.T, root string) {
				writeFile(t, root+"/prompts/p.md", "content")
				writeFile(t, root+"/commands/c.md", "content")
				writeFile(t, root+"/agents/a.md", "content")
				writeFile(t, root+"/agent-instructions/ai.md", "content")
			},
			artifacts: []manifest.Artifact{
				{Name: "p", Type: "prompt", Path: "prompts/p.md"},
				{Name: "c", Type: "command", Path: "commands/c.md"},
				{Name: "a", Type: "agent", Path: "agents/a.md"},
				{Name: "ai", Type: "agent-instructions", Path: "agent-instructions/ai.md"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			tt.setup(t, root)

			err := ValidateAll(root, tt.artifacts)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrStr != "" && !strings.Contains(err.Error(), tt.wantErrStr) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.wantErrStr)
				}
				if tt.errCount > 0 {
					lines := strings.Split(err.Error(), "\n")
					if len(lines) != tt.errCount {
						t.Errorf("got %d errors, want %d:\n%s", len(lines), tt.errCount, err.Error())
					}
				}
				return
			}
			if err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestValidArtifactName(t *testing.T) {
	valid := []string{"a", "ab", "code-review", "my-prompt-2", "a1b2c3"}
	invalid := []string{"", "A", "code_review", "-leading", "trailing-", "con--secutive", strings.Repeat("a", 65)}

	for _, name := range valid {
		if !validArtifactName(name) {
			t.Errorf("validArtifactName(%q) should be true", name)
		}
	}
	for _, name := range invalid {
		if validArtifactName(name) {
			t.Errorf("validArtifactName(%q) should be false", name)
		}
	}
}
