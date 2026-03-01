package license

import (
	"os"
	"path/filepath"

	"github.com/google/licensecheck"
)

var licenseFilenames = []string{
	"LICENSE",
	"LICENSE.txt",
	"LICENSE.md",
	"LICENCE",
	"LICENCE.txt",
	"LICENCE.md",
	"COPYING",
}

const coverageThreshold = 90.0

// Detect looks for a LICENSE file in dir and returns the SPDX identifier if
// exactly one license is detected with high confidence. Returns ("", false) if
// no license file is found, the file is ambiguous, or confidence is too low.
func Detect(dir string) (string, bool) {
	for _, name := range licenseFilenames {
		data, err := os.ReadFile(filepath.Join(dir, name)) //nolint:gosec // filename from fixed list
		if err != nil {
			continue
		}
		return identify(data)
	}
	return "", false
}

func identify(data []byte) (string, bool) {
	cov := licensecheck.Scan(data)
	if cov.Percent < coverageThreshold {
		return "", false
	}

	var match licensecheck.Match
	found := 0
	for _, m := range cov.Match {
		if m.IsURL {
			continue
		}
		if m.ID == "" {
			continue
		}
		match = m
		found++
	}
	if found != 1 {
		return "", false
	}
	return match.ID, true
}
