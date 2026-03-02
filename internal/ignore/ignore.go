package ignore

import (
	"os"
	"path/filepath"

	gitignore "github.com/sabhiram/go-gitignore"
)

// Rules holds compiled ignore patterns for filtering paths during pack.
type Rules struct {
	gi *gitignore.GitIgnore
}

// LoadRules loads ignore rules from .aipkgignore (if present) in rootDir,
// and appends built-in defaults (.git/, .aipkgignore, and the archive output
// path). Returns rules that can filter paths via IsExcluded.
func LoadRules(rootDir, archivePath string) (*Rules, error) {
	patterns := []string{".git", ".aipkgignore"}

	// Exclude the archive and sidecar output from the archive itself.
	if archivePath != "" {
		base := filepath.Base(archivePath)
		patterns = append(patterns, base, base+".sha256")
	}

	// Load .aipkgignore if present.
	ignorePath := filepath.Join(rootDir, ".aipkgignore")
	if data, err := os.ReadFile(ignorePath); err == nil { //nolint:gosec // path is constructed from trusted rootDir
		combined := make([]string, len(patterns))
		copy(combined, patterns)
		combined = append(combined, splitLines(data)...)
		gi := gitignore.CompileIgnoreLines(combined...)
		return &Rules{gi: gi}, nil
	}

	gi := gitignore.CompileIgnoreLines(patterns...)
	return &Rules{gi: gi}, nil
}

// IsExcluded returns true if the given relative path should be excluded
// from the archive. Paths should be relative to the package root using
// forward slashes.
func (r *Rules) IsExcluded(path string) bool {
	if r == nil || r.gi == nil {
		return false
	}
	return r.gi.MatchesPath(path)
}

// splitLines splits file content into individual non-empty, non-comment lines,
// suitable for passing to CompileIgnoreLines.
func splitLines(data []byte) []string {
	var lines []string
	start := 0
	for i := 0; i < len(data); i++ {
		if data[i] == '\n' {
			line := string(data[start:i])
			if line != "" && line[0] != '#' {
				lines = append(lines, line)
			}
			start = i + 1
		}
	}
	// Handle last line without trailing newline.
	if start < len(data) {
		line := string(data[start:])
		if line != "" && line[0] != '#' {
			lines = append(lines, line)
		}
	}
	return lines
}
