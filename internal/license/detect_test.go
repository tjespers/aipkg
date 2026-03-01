package license

import (
	"os"
	"path/filepath"
	"testing"
)

const mitLicense = `MIT License

Copyright (c) 2026 Test Author

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

func readApacheLicense(t *testing.T) []byte {
	t.Helper()
	// Use the project's own Apache-2.0 LICENSE file as test fixture.
	data, err := os.ReadFile(filepath.Join("..", "..", "LICENSE"))
	if err != nil {
		t.Skip("project LICENSE file not available for test")
	}
	return data
}

func TestDetect(t *testing.T) {
	apacheLicense := readApacheLicense(t)

	tests := []struct {
		name     string
		filename string
		content  []byte
		wantID   string
		wantOK   bool
	}{
		{
			name:     "MIT LICENSE file",
			filename: "LICENSE",
			content:  []byte(mitLicense),
			wantID:   "MIT",
			wantOK:   true,
		},
		{
			name:     "Apache-2.0 LICENSE file",
			filename: "LICENSE",
			content:  apacheLicense,
			wantID:   "Apache-2.0",
			wantOK:   true,
		},
		{
			name:     "LICENSE.txt variant",
			filename: "LICENSE.txt",
			content:  []byte(mitLicense),
			wantID:   "MIT",
			wantOK:   true,
		},
		{
			name:     "LICENCE spelling variant",
			filename: "LICENCE",
			content:  []byte(mitLicense),
			wantID:   "MIT",
			wantOK:   true,
		},
		{
			name:   "no license file",
			wantID: "",
			wantOK: false,
		},
		{
			name:     "unrecognizable content",
			filename: "LICENSE",
			content:  []byte("This is not a license."),
			wantID:   "",
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if tt.filename != "" {
				if err := os.WriteFile(filepath.Join(dir, tt.filename), tt.content, 0o644); err != nil {
					t.Fatal(err)
				}
			}

			id, ok := Detect(dir)
			if ok != tt.wantOK {
				t.Errorf("Detect() ok = %v, want %v", ok, tt.wantOK)
			}
			if id != tt.wantID {
				t.Errorf("Detect() id = %q, want %q", id, tt.wantID)
			}
		})
	}
}
