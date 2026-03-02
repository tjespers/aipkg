package artifact

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/tjespers/aipkg/internal/manifest"
)

// ExcludeFunc checks whether a relative path should be excluded from discovery.
// Paths use forward slashes relative to the package root.
type ExcludeFunc func(path string) bool

// Discover scans well-known directories under rootDir and returns discovered
// artifacts. It skips hidden entries, symlinks, non-regular files, and nested
// subdirectories in file-based type directories (FR-008). If exclude is non-nil,
// paths matching the exclude function are omitted (FR-028).
func Discover(rootDir string, exclude ExcludeFunc) ([]manifest.Artifact, error) {
	var artifacts []manifest.Artifact

	for dirName, artType := range DirToType {
		dirPath := filepath.Join(rootDir, dirName)

		entries, err := os.ReadDir(dirPath)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", dirName, err)
		}

		for _, entry := range entries {
			name := entry.Name()

			// Skip hidden entries (FR-008).
			if strings.HasPrefix(name, ".") {
				continue
			}

			// Skip symlinks and non-regular entries (FR-008).
			if entry.Type()&fs.ModeSymlink != 0 {
				continue
			}

			if artType.IsDirectoryBased() {
				// Skills: expect direct subdirectories.
				if !entry.IsDir() {
					continue
				}
				relPath := dirName + "/" + name + "/"
				if exclude != nil && exclude(relPath) {
					continue
				}
				artifacts = append(artifacts, manifest.Artifact{
					Name: name,
					Type: string(artType),
					Path: relPath,
				})
			} else {
				// File-based types: expect regular files at top level only.
				if entry.IsDir() {
					continue
				}
				// Verify it's a regular file (not a device, pipe, etc.).
				info, err := entry.Info()
				if err != nil {
					return nil, fmt.Errorf("stat %s/%s: %w", dirName, name, err)
				}
				if !info.Mode().IsRegular() {
					continue
				}

				relPath := dirName + "/" + name
				if exclude != nil && exclude(relPath) {
					continue
				}

				artName := deriveName(name)
				artifacts = append(artifacts, manifest.Artifact{
					Name: artName,
					Type: string(artType),
					Path: relPath,
				})
			}
		}
	}

	return artifacts, nil
}

// deriveName extracts the artifact name from a filename by stripping
// everything from the first dot onwards. If there is no dot, the entire
// filename is the artifact name (FR-019).
func deriveName(filename string) string {
	if idx := strings.Index(filename, "."); idx > 0 {
		return filename[:idx]
	}
	return filename
}
