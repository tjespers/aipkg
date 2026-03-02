package schema

import "testing"

func TestValidateField_Name(t *testing.T) {
	validate := ValidateField("name")

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "@alice/blog-writer", false},
		{"valid single char", "@a/b", false},
		{"no scope", "blog-writer", true},
		{"uppercase", "@Alice/Blog", true},
		{"consecutive hyphens", "@alice/blog--writer", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validate(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateField_Version(t *testing.T) {
	validate := ValidateField("version")

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "1.0.0", false},
		{"valid zero", "0.1.0", false},
		{"missing patch", "1.0", true},
		{"v prefix", "v1.0.0", true},
		{"prerelease", "1.0.0-beta", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validate(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateField_Description(t *testing.T) {
	validate := ValidateField("description")

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid short", "A blog writing assistant", false},
		{"empty string", "", false},
		{"at max length", longString(255), false},
		{"over max length", longString(256), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validate(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateField_Unknown(t *testing.T) {
	validate := ValidateField("nonexistent")
	err := validate("anything")
	if err == nil {
		t.Error("expected error for unknown property")
	}
}

func TestFormatValidationError(t *testing.T) {
	validate := ValidateField("version")
	err := validate("bad")
	if err == nil {
		t.Fatal("expected error")
	}

	msg := FormatValidationError("version", err)
	if msg == "" {
		t.Error("expected non-empty message")
	}
}
