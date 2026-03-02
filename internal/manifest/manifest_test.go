package manifest

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestMarshalIndent(t *testing.T) {
	tests := []struct {
		name     string
		manifest PackageManifest
		golden   string
	}{
		{
			name: "minimal",
			manifest: PackageManifest{
				SpecVersion: 1,
				Name:        "@alice/blog-writer",
				Version:     "0.1.0",
			},
			golden: "minimal.json",
		},
		{
			name: "full",
			manifest: PackageManifest{
				SpecVersion: 1,
				Name:        "@alice/blog-writer",
				Version:     "1.0.0",
				Description: "AI blog writing assistant",
				License:     "MIT",
			},
			golden: "full.json",
		},
		{
			name: "with artifacts",
			manifest: PackageManifest{
				SpecVersion: 1,
				Name:        "@alice/blog-writer",
				Version:     "1.0.0",
				Artifacts: []Artifact{
					{Name: "test-writer", Type: "skill", Path: "skills/test-writer/"},
					{Name: "code-review", Type: "prompt", Path: "prompts/code-review.md"},
				},
			},
			golden: "with-artifacts.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.manifest.MarshalIndent()
			if err != nil {
				t.Fatalf("MarshalIndent() error = %v", err)
			}

			golden := filepath.Join("testdata", tt.golden)
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("reading golden file %s: %v", golden, err)
			}

			if !bytes.Equal(got, want) {
				t.Errorf("MarshalIndent() mismatch.\ngot:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}

func TestWriteFile(t *testing.T) {
	m := &PackageManifest{
		SpecVersion: 1,
		Name:        "@test/pkg",
		Version:     "0.1.0",
	}

	dir := t.TempDir()
	if err := m.WriteFile(dir); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "aipkg.json"))
	if err != nil {
		t.Fatalf("reading aipkg.json: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("WriteFile() produced empty file")
	}
}

func TestOmitEmptyFields(t *testing.T) {
	m := &PackageManifest{
		SpecVersion: 1,
		Name:        "@test/pkg",
		Version:     "0.1.0",
	}

	data, err := m.MarshalIndent()
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}

	s := string(data)
	if bytes.Contains([]byte(s), []byte("description")) {
		t.Error("empty description should be omitted")
	}
	if bytes.Contains([]byte(s), []byte("license")) {
		t.Error("empty license should be omitted")
	}
	if bytes.Contains([]byte(s), []byte("artifacts")) {
		t.Error("nil artifacts should be omitted")
	}
}

func TestLoadFile(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
		check   func(t *testing.T, m *PackageManifest)
	}{
		{
			name: "valid minimal manifest",
			json: `{"specVersion": 1, "name": "@test/pkg", "version": "0.1.0"}`,
			check: func(t *testing.T, m *PackageManifest) {
				if m.SpecVersion != 1 {
					t.Errorf("SpecVersion = %d, want 1", m.SpecVersion)
				}
				if m.Name != "@test/pkg" {
					t.Errorf("Name = %q, want %q", m.Name, "@test/pkg")
				}
				if m.Version != "0.1.0" {
					t.Errorf("Version = %q, want %q", m.Version, "0.1.0")
				}
				if len(m.Artifacts) != 0 {
					t.Errorf("Artifacts = %v, want empty", m.Artifacts)
				}
			},
		},
		{
			name: "manifest with artifacts",
			json: `{
				"specVersion": 1,
				"name": "@test/pkg",
				"version": "1.0.0",
				"artifacts": [
					{"name": "writer", "type": "skill", "path": "skills/writer/"}
				]
			}`,
			check: func(t *testing.T, m *PackageManifest) {
				if len(m.Artifacts) != 1 {
					t.Fatalf("Artifacts len = %d, want 1", len(m.Artifacts))
				}
				a := m.Artifacts[0]
				if a.Name != "writer" || a.Type != "skill" || a.Path != "skills/writer/" {
					t.Errorf("Artifact = %+v, want {writer skill skills/writer/}", a)
				}
			},
		},
		{
			name:    "invalid JSON",
			json:    `{not valid`,
			wantErr: true,
		},
		{
			name:    "missing file",
			json:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			if tt.name != "missing file" {
				err := os.WriteFile(filepath.Join(dir, "aipkg.json"), []byte(tt.json), 0o644)
				if err != nil {
					t.Fatalf("writing test file: %v", err)
				}
			}

			m, err := LoadFile(dir)
			if tt.wantErr {
				if err == nil {
					t.Fatal("LoadFile() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadFile() error = %v", err)
			}
			tt.check(t, m)
		})
	}
}

func TestRoundTrip(t *testing.T) {
	original := &PackageManifest{
		SpecVersion: 1,
		Name:        "@test/roundtrip",
		Version:     "2.0.0",
		Description: "round-trip test",
		License:     "MIT",
		Artifacts: []Artifact{
			{Name: "writer", Type: "skill", Path: "skills/writer/"},
			{Name: "review", Type: "prompt", Path: "prompts/review.md"},
		},
	}

	dir := t.TempDir()
	if err := original.WriteFile(dir); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	loaded, err := LoadFile(dir)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	if loaded.Name != original.Name {
		t.Errorf("Name = %q, want %q", loaded.Name, original.Name)
	}
	if loaded.Version != original.Version {
		t.Errorf("Version = %q, want %q", loaded.Version, original.Version)
	}
	if len(loaded.Artifacts) != len(original.Artifacts) {
		t.Fatalf("Artifacts len = %d, want %d", len(loaded.Artifacts), len(original.Artifacts))
	}
	for i, a := range loaded.Artifacts {
		want := original.Artifacts[i]
		if a.Name != want.Name || a.Type != want.Type || a.Path != want.Path {
			t.Errorf("Artifact[%d] = %+v, want %+v", i, a, want)
		}
	}
}
