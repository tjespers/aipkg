package license

import (
	"os"
	"path/filepath"

	"github.com/google/licensecheck"
)

// Detect reads the LICENSE file in dir and returns the SPDX identifier
// if confidence exceeds 80%. Returns empty string if no LICENSE file
// exists, is unreadable, or no match is found.
func Detect(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "LICENSE")) //nolint:gosec // dir is trusted caller input
	if err != nil {
		return ""
	}
	cov := licensecheck.Scan(data)
	if cov.Percent < 80 || len(cov.Match) == 0 {
		return ""
	}
	return cov.Match[0].ID
}
