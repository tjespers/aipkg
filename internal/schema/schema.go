package schema

import (
	"bytes"
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/dlclark/regexp2"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed aipkg.schema.json
var schemaJSON []byte

var (
	compiledSchema *jsonschema.Schema
	compileOnce    sync.Once
	compileErr     error
)

// Patterns extracted from the embedded schema.
// Note: the spec schema uses a negative lookahead (?!.*--) which Go's regexp
// engine does not support. We handle the consecutive-hyphen check separately.
var (
	namePattern    = regexp.MustCompile(`^@[a-z0-9]([a-z0-9-]{0,37}[a-z0-9])?/[a-z0-9]([a-z0-9-]{0,62}[a-z0-9])?$`)
	versionPattern = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$`)
)

// regexp2Regexp adapts dlclark/regexp2 to the jsonschema.Regexp interface.
type regexp2Regexp regexp2.Regexp

// MatchString implements jsonschema.Regexp.
func (re *regexp2Regexp) MatchString(s string) bool {
	matched, _ := (*regexp2.Regexp)(re).MatchString(s)
	return matched
}

// String implements jsonschema.Regexp.
func (re *regexp2Regexp) String() string {
	return (*regexp2.Regexp)(re).String()
}

func regexp2Compile(s string) (jsonschema.Regexp, error) {
	re, err := regexp2.Compile(s, regexp2.ECMAScript)
	if err != nil {
		return nil, err
	}
	return (*regexp2Regexp)(re), nil
}

func compile() (*jsonschema.Schema, error) {
	compileOnce.Do(func() {
		doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaJSON))
		if err != nil {
			compileErr = fmt.Errorf("schema: unmarshal: %w", err)
			return
		}
		c := jsonschema.NewCompiler()
		c.UseRegexpEngine(regexp2Compile)
		if err := c.AddResource("aipkg.schema.json", doc); err != nil {
			compileErr = fmt.Errorf("schema: add resource: %w", err)
			return
		}
		compiledSchema, compileErr = c.Compile("aipkg.schema.json")
		if compileErr != nil {
			compileErr = fmt.Errorf("schema: compile: %w", compileErr)
		}
	})
	return compiledSchema, compileErr
}

// ValidateName checks if s is a valid scoped package name.
func ValidateName(s string) error {
	if strings.Contains(s, "--") {
		return fmt.Errorf("invalid name %q: must match @scope/package-name format", s)
	}
	if !namePattern.MatchString(s) {
		return fmt.Errorf("invalid name %q: must match @scope/package-name format", s)
	}
	return nil
}

// ValidateVersion checks if s is a valid semver MAJOR.MINOR.PATCH.
func ValidateVersion(s string) error {
	if !versionPattern.MatchString(s) {
		return fmt.Errorf("invalid version %q: must be MAJOR.MINOR.PATCH (e.g., 1.0.0)", s)
	}
	return nil
}

// ValidateDescription checks if s is within the max length.
func ValidateDescription(s string) error {
	if n := utf8.RuneCountInString(s); n > 255 {
		return fmt.Errorf("description too long: %d characters (max 255)", n)
	}
	return nil
}

// ValidateManifest validates a complete JSON manifest against the schema.
func ValidateManifest(data []byte) error {
	sch, err := compile()
	if err != nil {
		return err
	}
	v, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("schema: invalid JSON: %w", err)
	}
	if err := sch.Validate(v); err != nil {
		return fmt.Errorf("schema: validation failed: %w", err)
	}
	return nil
}
