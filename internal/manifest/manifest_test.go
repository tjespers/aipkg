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
	if contains(s, "description") {
		t.Error("empty description should be omitted")
	}
	if contains(s, "license") {
		t.Error("empty license should be omitted")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
