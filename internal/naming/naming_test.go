package naming

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantScope string
		wantPkg   string
		wantErr   bool
	}{
		{
			name:      "valid simple",
			input:     "@alice/blog-writer",
			wantScope: "alice",
			wantPkg:   "blog-writer",
		},
		{
			name:      "single char scope and pkg",
			input:     "@a/b",
			wantScope: "a",
			wantPkg:   "b",
		},
		{
			name:      "numeric scope",
			input:     "@123/my-pkg",
			wantScope: "123",
			wantPkg:   "my-pkg",
		},
		{
			name:    "no scope prefix",
			input:   "blog-writer",
			wantErr: true,
		},
		{
			name:    "no slash",
			input:   "@alice",
			wantErr: true,
		},
		{
			name:    "uppercase letters",
			input:   "@Alice/Blog-Writer",
			wantErr: true,
		},
		{
			name:    "consecutive hyphens in scope",
			input:   "@al--ice/blog",
			wantErr: true,
		},
		{
			name:    "consecutive hyphens in pkg",
			input:   "@alice/blog--writer",
			wantErr: true,
		},
		{
			name:    "leading hyphen in scope",
			input:   "@-alice/blog",
			wantErr: true,
		},
		{
			name:    "trailing hyphen in pkg",
			input:   "@alice/blog-",
			wantErr: true,
		},
		{
			name:    "underscores not allowed",
			input:   "@alice/blog_writer",
			wantErr: true,
		},
		{
			name:    "dots not allowed",
			input:   "@alice/blog.writer",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got.Scope != tt.wantScope {
				t.Errorf("Parse(%q).Scope = %q, want %q", tt.input, got.Scope, tt.wantScope)
			}
			if got.Package != tt.wantPkg {
				t.Errorf("Parse(%q).Package = %q, want %q", tt.input, got.Package, tt.wantPkg)
			}
		})
	}
}

func TestScopedNameString(t *testing.T) {
	n := ScopedName{Scope: "alice", Package: "blog-writer"}
	if got := n.String(); got != "@alice/blog-writer" {
		t.Errorf("String() = %q, want %q", got, "@alice/blog-writer")
	}
}

func TestIsReservedScope(t *testing.T) {
	tests := []struct {
		scope    string
		wantOK   bool
		wantRule string
	}{
		{"aipkg", true, "aipkg*"},
		{"aipkg-tools", true, "aipkg*"},
		{"official", true, "official"},
		{"anthropic", true, "anthropic"},
		{"claude", true, "claude"},
		{"alice", false, ""},
		{"my-org", false, ""},
		{"test", true, "test"},
		{"github", true, "github"},
	}

	for _, tt := range tests {
		t.Run(tt.scope, func(t *testing.T) {
			rule, ok := IsReservedScope(tt.scope)
			if ok != tt.wantOK {
				t.Errorf("IsReservedScope(%q) ok = %v, want %v", tt.scope, ok, tt.wantOK)
			}
			if rule != tt.wantRule {
				t.Errorf("IsReservedScope(%q) rule = %q, want %q", tt.scope, rule, tt.wantRule)
			}
		})
	}
}
