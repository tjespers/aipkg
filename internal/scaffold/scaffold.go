package scaffold

import "os"

// WellKnownDirs are the conventional artifact directories in a package.
var WellKnownDirs = []string{
	"agents",
	"agent-instructions",
	"commands",
	"mcp-servers",
	"prompts",
	"skills",
}

// Create creates the target directory (if needed) and all well-known artifact
// directories inside it. Existing directories are skipped. Returns an error if
// any directory creation fails.
func Create(targetDir string) error {
	if err := os.MkdirAll(targetDir, 0o755); err != nil { //nolint:gosec // 0o755 is standard for directories
		return err
	}
	for _, d := range WellKnownDirs {
		if err := os.MkdirAll(targetDir+"/"+d, 0o755); err != nil { //nolint:gosec // 0o755 is standard for directories
			return err
		}
	}
	return nil
}
