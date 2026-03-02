package artifact

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/tjespers/aipkg/internal/frontmatter"
	"github.com/tjespers/aipkg/internal/manifest"
)

// SkillFrontmatter represents the parsed YAML frontmatter of a SKILL.md file.
type SkillFrontmatter struct {
	Name          string         `yaml:"name"`
	Description   string         `yaml:"description"`
	License       string         `yaml:"license,omitempty"`
	Compatibility []string       `yaml:"compatibility,omitempty"`
	Metadata      map[string]any `yaml:"metadata,omitempty"`
	AllowedTools  []string       `yaml:"allowed-tools,omitempty"`
}

func validateSkill(rootDir string, art manifest.Artifact, errs *ValidationErrors) {
	skillPath := filepath.Join(rootDir, art.Path, "SKILL.md")
	relPath := art.Path + "SKILL.md"

	content, err := os.ReadFile(skillPath) //nolint:gosec // path constructed from validated artifact
	if err != nil {
		if os.IsNotExist(err) {
			errs.Add(art.Path, "missing required SKILL.md file")
			return
		}
		errs.Addf(art.Path, "reading SKILL.md: %v", err)
		return
	}

	yamlBytes, _, err := frontmatter.Extract(content)
	if err != nil {
		errs.Addf(relPath, "invalid frontmatter: %v", err)
		return
	}

	var fm SkillFrontmatter
	dec := yaml.NewDecoder(bytes.NewReader(yamlBytes))
	dec.KnownFields(true)
	if err := dec.Decode(&fm); err != nil {
		errs.Addf(relPath, "invalid frontmatter YAML: %v", err)
		return
	}

	if fm.Name == "" {
		errs.Add(relPath, "missing required field 'name'")
	}
	if fm.Description == "" {
		errs.Add(relPath, "missing required field 'description'")
	}

	if fm.Name != "" {
		// Validate name length.
		if len(fm.Name) > 64 {
			errs.Addf(relPath, "name %q exceeds 64 characters", fm.Name)
		}

		// Validate name matches directory name (FR-016).
		dirName := filepath.Base(filepath.Clean(filepath.Join(rootDir, art.Path)))
		if fm.Name != dirName {
			errs.Addf(relPath, "name %q does not match directory name %q", fm.Name, dirName)
		}
	}

	if len(fm.Description) > 1024 {
		errs.Add(relPath, fmt.Sprintf("description exceeds 1024 characters (%d)", len(fm.Description)))
	}
}
