package license

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// apacheLicenseFromRepo reads the actual Apache-2.0 LICENSE file from
// the repository root for use as test data.
func apacheLicenseFromRepo(t *testing.T) string {
	t.Helper()
	// Determine repo root relative to this test file.
	_, thisFile, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	data, err := os.ReadFile(filepath.Join(repoRoot, "LICENSE")) //nolint:gosec // test fixture path
	if err != nil {
		t.Fatalf("cannot read repo LICENSE for test data: %v", err)
	}
	return string(data)
}

const mitLicense = `MIT License

Copyright (c) 2024 Test

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`

func TestDetect(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string // empty means no LICENSE file
		wantID      string
	}{
		{
			name:        "Apache-2.0 license",
			fileContent: apacheLicenseFromRepo(t),
			wantID:      "Apache-2.0",
		},
		{
			name:        "MIT license",
			fileContent: mitLicense,
			wantID:      "MIT",
		},
		{
			name:        "no LICENSE file",
			fileContent: "",
			wantID:      "",
		},
		{
			name:        "unrecognized content",
			fileContent: "This is not a license. Just some random text.",
			wantID:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if tt.fileContent != "" {
				if err := os.WriteFile(filepath.Join(dir, "LICENSE"), []byte(tt.fileContent), 0o600); err != nil { //nolint:gosec // test fixture
					t.Fatal(err)
				}
			}

			got := Detect(dir)
			if got != tt.wantID {
				t.Errorf("Detect() = %q, want %q", got, tt.wantID)
			}
		})
	}
}
