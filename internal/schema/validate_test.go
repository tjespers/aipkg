package schema

import "testing"

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name: "valid minimal",
			json: `{"specVersion": 1, "name": "@alice/blog-writer", "version": "0.1.0"}`,
		},
		{
			name: "valid full",
			json: `{
				"specVersion": 1,
				"name": "@alice/blog-writer",
				"version": "1.0.0",
				"description": "AI blog writing assistant",
				"license": "MIT"
			}`,
		},
		{
			name:    "missing specVersion",
			json:    `{"name": "@alice/blog-writer", "version": "0.1.0"}`,
			wantErr: true,
		},
		{
			name:    "missing name",
			json:    `{"specVersion": 1, "version": "0.1.0"}`,
			wantErr: true,
		},
		{
			name:    "missing version",
			json:    `{"specVersion": 1, "name": "@alice/blog-writer"}`,
			wantErr: true,
		},
		{
			name:    "wrong specVersion",
			json:    `{"specVersion": 2, "name": "@alice/blog-writer", "version": "0.1.0"}`,
			wantErr: true,
		},
		{
			name:    "invalid name pattern",
			json:    `{"specVersion": 1, "name": "no-scope", "version": "0.1.0"}`,
			wantErr: true,
		},
		{
			name:    "invalid version format",
			json:    `{"specVersion": 1, "name": "@alice/blog-writer", "version": "1.0"}`,
			wantErr: true,
		},
		{
			name:    "description too long",
			json:    `{"specVersion": 1, "name": "@alice/blog-writer", "version": "0.1.0", "description": "` + longString(256) + `"}`,
			wantErr: true,
		},
		{
			name: "description at max length",
			json: `{"specVersion": 1, "name": "@alice/blog-writer", "version": "0.1.0", "description": "` + longString(255) + `"}`,
		},
		{
			name:    "unknown field rejected",
			json:    `{"specVersion": 1, "name": "@alice/blog-writer", "version": "0.1.0", "type": "package"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateProject(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name: "valid empty project",
			json: `{"specVersion": 1, "require": {}}`,
		},
		{
			name: "valid project with dependencies",
			json: `{"specVersion": 1, "require": {"@alice/blog-tools": "1.2.0", "@bob/code-review": "0.3.0"}}`,
		},
		{
			name: "valid project with pre-release version",
			json: `{"specVersion": 1, "require": {"@alice/blog-tools": "1.0.0-beta.1"}}`,
		},
		{
			name: "valid project with alpha pre-release",
			json: `{"specVersion": 1, "require": {"@alice/blog-tools": "1.0.0-alpha"}}`,
		},
		{
			name: "valid project with numeric pre-release",
			json: `{"specVersion": 1, "require": {"@alice/blog-tools": "1.0.0-0.3.7"}}`,
		},
		{
			name: "valid project with rc pre-release",
			json: `{"specVersion": 1, "require": {"@alice/blog-tools": "1.0.0-rc.1"}}`,
		},
		{
			name:    "missing specVersion",
			json:    `{"require": {}}`,
			wantErr: true,
		},
		{
			name:    "missing require",
			json:    `{"specVersion": 1}`,
			wantErr: true,
		},
		{
			name:    "extra fields rejected",
			json:    `{"specVersion": 1, "require": {}, "name": "not-allowed"}`,
			wantErr: true,
		},
		{
			name:    "bad package name in require key",
			json:    `{"specVersion": 1, "require": {"no-scope": "1.0.0"}}`,
			wantErr: true,
		},
		{
			name:    "bad version string in require value",
			json:    `{"specVersion": 1, "require": {"@alice/blog-tools": "1.0"}}`,
			wantErr: true,
		},
		{
			name:    "build metadata rejected",
			json:    `{"specVersion": 1, "require": {"@alice/blog-tools": "1.0.0+build.123"}}`,
			wantErr: true,
		},
		{
			name:    "wrong specVersion",
			json:    `{"specVersion": 2, "require": {}}`,
			wantErr: true,
		},
		{
			name:    "version prefix rejected",
			json:    `{"specVersion": 1, "require": {"@alice/blog-tools": "v1.0.0"}}`,
			wantErr: true,
		},
		{
			name:    "consecutive hyphens in package name rejected",
			json:    `{"specVersion": 1, "require": {"@alice/blog--tools": "1.0.0"}}`,
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			json:    `not json at all`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProject([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func longString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}
