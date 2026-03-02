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

func longString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}
