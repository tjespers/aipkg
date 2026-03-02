package archive

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestCreateArchive(t *testing.T) {
	root := t.TempDir()

	// Set up test files.
	writeTestFile(t, filepath.Join(root, "aipkg.json"), `{"specVersion":1}`)
	writeTestFile(t, filepath.Join(root, "prompts", "review.md"), "review content")
	writeTestFile(t, filepath.Join(root, "skills", "writer", "SKILL.md"), "skill content")
	writeTestFile(t, filepath.Join(root, "skills", "writer", "scripts", "run.sh"), "#!/bin/sh")

	enrichedManifest := []byte(`{"specVersion":1,"artifacts":[]}`)

	paths := []string{
		"prompts/review.md",
		"skills/writer/",
	}

	var buf bytes.Buffer
	err := CreateArchive(&buf, root, "test-pkg", paths, enrichedManifest)
	if err != nil {
		t.Fatalf("CreateArchive() error = %v", err)
	}

	// Read back and verify.
	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("reading zip: %v", err)
	}

	var names []string
	contents := make(map[string]string)
	for _, f := range reader.File {
		names = append(names, f.Name)
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("opening %s: %v", f.Name, err)
		}
		data, err := readAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("reading %s: %v", f.Name, err)
		}
		contents[f.Name] = string(data)
	}

	sort.Strings(names)

	wantNames := []string{
		"test-pkg/aipkg.json",
		"test-pkg/prompts/review.md",
		"test-pkg/skills/writer/SKILL.md",
		"test-pkg/skills/writer/scripts/run.sh",
	}
	sort.Strings(wantNames)

	if len(names) != len(wantNames) {
		t.Fatalf("archive has %d entries, want %d\ngot: %v", len(names), len(wantNames), names)
	}
	for i, name := range names {
		if name != wantNames[i] {
			t.Errorf("entry[%d] = %q, want %q", i, name, wantNames[i])
		}
	}

	// Verify enriched manifest is used (not the on-disk version).
	if got := contents["test-pkg/aipkg.json"]; got != string(enrichedManifest) {
		t.Errorf("manifest content = %q, want %q", got, string(enrichedManifest))
	}

	// Verify file content integrity.
	if got := contents["test-pkg/prompts/review.md"]; got != "review content" {
		t.Errorf("review.md content = %q, want %q", got, "review content")
	}
}

func TestCreateArchive_DeflateCompression(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "prompts", "review.md"), "review content here")

	var buf bytes.Buffer
	err := CreateArchive(&buf, root, "pkg", []string{"prompts/review.md"}, []byte("{}"))
	if err != nil {
		t.Fatalf("CreateArchive() error = %v", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("reading zip: %v", err)
	}

	for _, f := range reader.File {
		if f.Method != zip.Deflate {
			t.Errorf("entry %q uses method %d, want Deflate (%d)", f.Name, f.Method, zip.Deflate)
		}
	}
}

// helpers

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readAll(rc io.Reader) ([]byte, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(rc)
	return buf.Bytes(), err
}
