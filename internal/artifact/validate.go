package artifact

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tjespers/aipkg/internal/manifest"
)

// nameChars validates character set and length: lowercase alphanumeric and hyphens,
// 1-64 characters, cannot start or end with a hyphen.
var nameChars = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,62}[a-z0-9])?$`)

// validArtifactName checks all naming rules including the consecutive hyphen
// restriction (which requires a lookahead not supported by Go's regexp).
func validArtifactName(name string) bool {
	return nameChars.MatchString(name) && !strings.Contains(name, "--")
}

// ValidateAll validates all discovered artifacts against their type-specific
// rules. Returns a combined error with all validation failures, or nil if
// all artifacts are valid.
func ValidateAll(rootDir string, artifacts []manifest.Artifact) error {
	var errs ValidationErrors

	for _, art := range artifacts {
		// Validate artifact name against naming rules (FR-019).
		if !validArtifactName(art.Name) {
			errs.Addf(art.Path, "invalid artifact name %q: must be lowercase alphanumeric and hyphens, 1-64 characters, no consecutive hyphens, cannot start or end with a hyphen", art.Name)
		}

		switch Type(art.Type) {
		case TypeSkill:
			validateSkill(rootDir, art, &errs)
		case TypeMCPServer:
			validateJSON(rootDir, art, &errs)
		case TypePrompt, TypeCommand, TypeAgent, TypeAgentInstructions:
			validateNonEmpty(rootDir, art, &errs)
		}
	}

	return errs.Err()
}

// validateJSON checks that a file parses as valid JSON (FR-017).
func validateJSON(rootDir string, art manifest.Artifact, errs *ValidationErrors) {
	path := filepath.Join(rootDir, art.Path)
	data, err := os.ReadFile(path) //nolint:gosec // path constructed from validated artifact
	if err != nil {
		errs.Addf(art.Path, "reading file: %v", err)
		return
	}
	if !json.Valid(data) {
		errs.Add(art.Path, "invalid JSON")
	}
}

// validateNonEmpty checks that a file has size greater than zero (FR-018).
func validateNonEmpty(rootDir string, art manifest.Artifact, errs *ValidationErrors) {
	path := filepath.Join(rootDir, art.Path)
	info, err := os.Stat(path)
	if err != nil {
		errs.Addf(art.Path, "reading file: %v", err)
		return
	}
	if info.Size() == 0 {
		errs.Add(art.Path, "file must not be empty")
	}
}
