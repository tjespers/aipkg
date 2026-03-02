package artifact

import (
	"strings"
	"testing"
)

func TestValidationErrors_Empty(t *testing.T) {
	var v ValidationErrors
	if err := v.Err(); err != nil {
		t.Errorf("empty collector should return nil, got %v", err)
	}
	if v.Len() != 0 {
		t.Errorf("Len() = %d, want 0", v.Len())
	}
}

func TestValidationErrors_Single(t *testing.T) {
	var v ValidationErrors
	v.Add("skills/broken/SKILL.md", "missing required field 'name'")

	err := v.Err()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if v.Len() != 1 {
		t.Errorf("Len() = %d, want 1", v.Len())
	}

	want := "skills/broken/SKILL.md: missing required field 'name'"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestValidationErrors_Multiple(t *testing.T) {
	var v ValidationErrors
	v.Add("skills/broken/SKILL.md", "missing required field 'name'")
	v.Add("mcp-servers/bad.json", "invalid JSON")
	v.Addf("prompts/empty.md", "file must not be %s", "empty")

	err := v.Err()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if v.Len() != 3 {
		t.Errorf("Len() = %d, want 3", v.Len())
	}

	msg := err.Error()
	lines := strings.Split(msg, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %q", len(lines), msg)
	}

	wantLines := []string{
		"skills/broken/SKILL.md: missing required field 'name'",
		"mcp-servers/bad.json: invalid JSON",
		"prompts/empty.md: file must not be empty",
	}
	for i, want := range wantLines {
		if lines[i] != want {
			t.Errorf("line[%d] = %q, want %q", i, lines[i], want)
		}
	}
}
