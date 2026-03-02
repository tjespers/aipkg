package archive

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteSidecar(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "test.aipkg")

	content := []byte("archive content for hashing")
	if err := os.WriteFile(archivePath, content, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := WriteSidecar(archivePath); err != nil {
		t.Fatalf("WriteSidecar() error = %v", err)
	}

	sidecarPath := archivePath + ".sha256"
	data, err := os.ReadFile(sidecarPath)
	if err != nil {
		t.Fatalf("reading sidecar: %v", err)
	}

	// Verify format: {hash}  {basename}\n
	line := string(data)
	if !strings.HasSuffix(line, "\n") {
		t.Error("sidecar does not end with newline")
	}

	parts := strings.SplitN(strings.TrimSuffix(line, "\n"), "  ", 2)
	if len(parts) != 2 {
		t.Fatalf("sidecar format invalid: %q", line)
	}

	gotHash := parts[0]
	gotName := parts[1]

	// Verify hash correctness.
	wantHash := fmt.Sprintf("%x", sha256.Sum256(content))
	if gotHash != wantHash {
		t.Errorf("hash = %q, want %q", gotHash, wantHash)
	}

	// Verify basename (not full path).
	if gotName != "test.aipkg" {
		t.Errorf("filename = %q, want %q", gotName, "test.aipkg")
	}

	// Verify hash is lowercase.
	if gotHash != strings.ToLower(gotHash) {
		t.Error("hash is not lowercase")
	}
}

func TestWriteSidecar_MissingFile(t *testing.T) {
	err := WriteSidecar("/nonexistent/path/test.aipkg")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
