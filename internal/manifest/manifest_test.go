package manifest

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMarshal(t *testing.T) {
	tests := []struct {
		name     string
		manifest Manifest
		want     string
	}{
		{
			name: "package with all fields",
			manifest: Manifest{
				Type:        "package",
				Name:        "@myorg/cool-skill",
				Version:     "1.0.0",
				Description: "A cool skill",
				License:     "Apache-2.0",
			},
			want: `{
  "type": "package",
  "name": "@myorg/cool-skill",
  "version": "1.0.0",
  "description": "A cool skill",
  "license": "Apache-2.0"
}
`,
		},
		{
			name: "project with only type",
			manifest: Manifest{
				Type: "project",
			},
			want: `{
  "type": "project"
}
`,
		},
		{
			name: "project with name and description",
			manifest: Manifest{
				Type:        "project",
				Name:        "@myteam/my-project",
				Description: "My AI project",
			},
			want: `{
  "type": "project",
  "name": "@myteam/my-project",
  "description": "My AI project"
}
`,
		},
		{
			name: "package required fields only",
			manifest: Manifest{
				Type:    "package",
				Name:    "@myorg/my-tool",
				Version: "0.1.0",
			},
			want: `{
  "type": "package",
  "name": "@myorg/my-tool",
  "version": "0.1.0"
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.manifest.Marshal()
			if err != nil {
				t.Fatalf("Marshal() error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("Marshal() =\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestMarshal_TwoSpaceIndent(t *testing.T) {
	m := Manifest{Type: "package", Name: "@a/b", Version: "1.0.0"}
	data, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		indent := len(line) - len(trimmed)
		if indent > 0 && indent%2 != 0 {
			t.Errorf("non-2-space indent on line: %q", line)
		}
	}
}

func TestMarshal_TrailingNewline(t *testing.T) {
	m := Manifest{Type: "project"}
	data, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(data), "\n") {
		t.Error("Marshal() output missing trailing newline")
	}
}

func TestMarshal_FieldOrder(t *testing.T) {
	m := Manifest{
		Type:        "package",
		Name:        "@a/b",
		Version:     "1.0.0",
		Description: "desc",
		License:     "MIT",
	}
	data, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	// Verify field order: type, name, version, description, license
	var raw json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	s := string(data)
	indices := []int{
		strings.Index(s, `"type"`),
		strings.Index(s, `"name"`),
		strings.Index(s, `"version"`),
		strings.Index(s, `"description"`),
		strings.Index(s, `"license"`),
	}
	for i := 1; i < len(indices); i++ {
		if indices[i] <= indices[i-1] {
			t.Errorf("field order wrong: field at position %d appears before field at position %d", i, i-1)
		}
	}
}

func TestMarshal_OmitEmpty(t *testing.T) {
	m := Manifest{Type: "project"}
	data, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, field := range []string{"name", "version", "description", "license"} {
		if strings.Contains(s, `"`+field+`"`) {
			t.Errorf("Marshal() should omit empty %q field", field)
		}
	}
}
