package archive

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// CreateArchive writes a zip archive to w. All files under rootDir matching
// the provided paths are added under a single top-level directory (topLevelDir).
// If manifestJSON is non-nil, it replaces the content of aipkg.json in the archive.
func CreateArchive(w io.Writer, rootDir, topLevelDir string, paths []string, manifestJSON []byte) error {
	zw := zip.NewWriter(w)
	defer zw.Close() //nolint:errcheck // final Close error is acceptable; data is flushed per entry

	// Write the enriched manifest first.
	if manifestJSON != nil {
		header := &zip.FileHeader{
			Name:   topLevelDir + "/aipkg.json",
			Method: zip.Deflate,
		}
		fw, err := zw.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("creating manifest entry: %w", err)
		}
		if _, err := fw.Write(manifestJSON); err != nil {
			return fmt.Errorf("writing manifest: %w", err)
		}
	}

	for _, relPath := range paths {
		absPath := filepath.Join(rootDir, relPath)

		info, err := os.Lstat(absPath)
		if err != nil {
			return fmt.Errorf("stat %s: %w", relPath, err)
		}

		if info.IsDir() {
			// Walk directory recursively (for skill directories).
			err := filepath.WalkDir(absPath, func(path string, d fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if d.IsDir() {
					return nil
				}
				// Skip non-regular files.
				info, err := d.Info()
				if err != nil {
					return err
				}
				if !info.Mode().IsRegular() {
					return nil
				}

				rel, err := filepath.Rel(rootDir, path)
				if err != nil {
					return err
				}
				// Normalize to forward slashes for zip entries.
				zipPath := topLevelDir + "/" + filepath.ToSlash(rel)
				return addFileToZip(zw, path, zipPath)
			})
			if err != nil {
				return fmt.Errorf("walking %s: %w", relPath, err)
			}
		} else {
			zipPath := topLevelDir + "/" + filepath.ToSlash(relPath)
			if err := addFileToZip(zw, absPath, zipPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func addFileToZip(zw *zip.Writer, absPath, zipPath string) error {
	// Skip aipkg.json since we write the enriched version separately.
	if strings.HasSuffix(zipPath, "/aipkg.json") {
		return nil
	}

	header := &zip.FileHeader{
		Name:   zipPath,
		Method: zip.Deflate,
	}

	fw, err := zw.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("creating entry %s: %w", zipPath, err)
	}

	f, err := os.Open(absPath) //nolint:gosec // path comes from validated artifact paths
	if err != nil {
		return fmt.Errorf("opening %s: %w", absPath, err)
	}
	defer f.Close() //nolint:errcheck // read-only file

	if _, err := io.Copy(fw, f); err != nil {
		return fmt.Errorf("writing %s: %w", zipPath, err)
	}

	return nil
}
