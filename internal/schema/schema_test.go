package schema

import "testing"

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid scoped name", "@myorg/cool-skill", false},
		{"valid single-char scope and name", "@a/b", false},
		{"valid with numbers", "@org123/tool456", false},
		{"valid with hyphens", "@my-org/my-tool", false},
		{"unscoped name", "cool-skill", true},
		{"missing scope", "/cool-skill", true},
		{"missing name", "@myorg/", true},
		{"missing slash", "@myorg", true},
		{"uppercase", "@MyOrg/Cool-Skill", true},
		{"spaces", "@my org/cool skill", true},
		{"consecutive hyphens in scope", "@my--org/tool", true},
		{"consecutive hyphens in name", "@myorg/my--tool", true},
		{"leading hyphen in scope", "@-myorg/tool", true},
		{"trailing hyphen in scope", "@myorg-/tool", true},
		{"leading hyphen in name", "@myorg/-tool", true},
		{"trailing hyphen in name", "@myorg/tool-", true},
		{"scope too long (39 chars)", "@abcdefghijklmnopqrstuvwxyz0123456789abcd/x", true},
		{"scope max length (39 chars)", "@abcdefghijklmnopqrstuvwxyz012345678901/x", false},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid semver", "1.0.0", false},
		{"valid with zeros", "0.1.0", false},
		{"valid large numbers", "100.200.300", false},
		{"missing patch", "1.0", true},
		{"missing minor and patch", "1", true},
		{"leading zero major", "01.0.0", true},
		{"leading zero minor", "1.01.0", true},
		{"leading zero patch", "1.0.01", true},
		{"with prefix v", "v1.0.0", true},
		{"with prerelease", "1.0.0-beta", true},
		{"with build metadata", "1.0.0+build", true},
		{"empty string", "", true},
		{"non-numeric", "a.b.c", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVersion(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateDescription(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"short description", "A cool skill", false},
		{"empty description", "", false},
		{"max length (255)", string(make([]byte, 255)), false},
		{"over max length (256)", string(make([]byte, 256)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fill byte slices with valid chars for readability
			err := ValidateDescription(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDescription(len=%d) error = %v, wantErr %v", len(tt.input), err, tt.wantErr)
			}
		})
	}
}

func TestValidateManifest(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "valid package",
			json:    `{"type":"package","name":"@myorg/tool","version":"1.0.0"}`,
			wantErr: false,
		},
		{
			name:    "valid package with all fields",
			json:    `{"type":"package","name":"@myorg/tool","version":"1.0.0","description":"A tool","license":"MIT"}`,
			wantErr: false,
		},
		{
			name:    "valid minimal project",
			json:    `{"type":"project"}`,
			wantErr: false,
		},
		{
			name:    "valid project with name",
			json:    `{"type":"project","name":"@myteam/my-project"}`,
			wantErr: false,
		},
		{
			name:    "missing type",
			json:    `{"name":"@myorg/tool"}`,
			wantErr: true,
		},
		{
			name:    "invalid type",
			json:    `{"type":"library"}`,
			wantErr: true,
		},
		{
			name:    "package missing name",
			json:    `{"type":"package","version":"1.0.0"}`,
			wantErr: true,
		},
		{
			name:    "package missing version",
			json:    `{"type":"package","name":"@myorg/tool"}`,
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			json:    `{not json}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateManifest([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateManifest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
