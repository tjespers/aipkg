package frontmatter

import (
	"testing"
)

func TestExtract(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantYAML string
		wantBody string
		wantErr  bool
	}{
		{
			name:     "valid frontmatter with body",
			content:  "---\nname: test\ndescription: hello\n---\n\n# Body\n",
			wantYAML: "name: test\ndescription: hello",
			wantBody: "\n# Body\n",
		},
		{
			name:     "frontmatter only no body",
			content:  "---\nname: test\n---\n",
			wantYAML: "name: test",
			wantBody: "",
		},
		{
			name:     "frontmatter only no trailing newline",
			content:  "---\nname: test\n---",
			wantYAML: "name: test",
			wantBody: "",
		},
		{
			name:     "empty frontmatter",
			content:  "---\n---\n\nbody here\n",
			wantYAML: "",
			wantBody: "\nbody here\n",
		},
		{
			name:    "missing opening delimiter",
			content: "name: test\n---\nbody\n",
			wantErr: true,
		},
		{
			name:    "missing closing delimiter",
			content: "---\nname: test\nbody without closing\n",
			wantErr: true,
		},
		{
			name:    "empty content",
			content: "",
			wantErr: true,
		},
		{
			name:     "windows line endings",
			content:  "---\r\nname: test\r\n---\r\nbody\r\n",
			wantYAML: "name: test",
			wantBody: "body\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yaml, body, err := Extract([]byte(tt.content))
			if tt.wantErr {
				if err == nil {
					t.Fatal("Extract() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Extract() error = %v", err)
			}
			if string(yaml) != tt.wantYAML {
				t.Errorf("yaml = %q, want %q", string(yaml), tt.wantYAML)
			}
			if string(body) != tt.wantBody {
				t.Errorf("body = %q, want %q", string(body), tt.wantBody)
			}
		})
	}
}
