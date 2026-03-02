package archive

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// WriteSidecar computes the SHA-256 hash of the file at archivePath and
// writes a sidecar file at archivePath + ".sha256" in sha256sum format:
// lowercase hex hash, two spaces, basename, newline (LF).
func WriteSidecar(archivePath string) error {
	f, err := os.Open(archivePath) //nolint:gosec // path comes from pack pipeline
	if err != nil {
		return fmt.Errorf("opening archive: %w", err)
	}
	defer f.Close() //nolint:errcheck // read-only file

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("hashing archive: %w", err)
	}

	hexHash := fmt.Sprintf("%x", h.Sum(nil))
	basename := filepath.Base(archivePath)
	line := hexHash + "  " + basename + "\n"

	sidecarPath := archivePath + ".sha256"
	if err := os.WriteFile(sidecarPath, []byte(line), 0o644); err != nil { //nolint:gosec // 0o644 is intentional for non-secret checksum
		return fmt.Errorf("writing sidecar: %w", err)
	}

	return nil
}
