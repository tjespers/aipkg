package schema

import (
	"errors"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// ValidateField returns a validator function for a single manifest field.
// The returned function validates a string value against the field's
// sub-schema constraints (pattern, maxLength, type, etc.).
// This is used by huh prompts for inline validation.
func ValidateField(property string) func(string) error {
	root, err := compiled()
	if err != nil {
		return func(_ string) error { return fmt.Errorf("schema not available: %w", err) }
	}

	propSchema, ok := root.Properties[property]
	if !ok {
		return func(_ string) error { return fmt.Errorf("unknown schema property: %s", property) }
	}

	return func(value string) error {
		return propSchema.Validate(value)
	}
}

// FormatValidationError converts a jsonschema validation error into a
// user-friendly message for the given field.
func FormatValidationError(field string, err error) string {
	if err == nil {
		return ""
	}

	switch field {
	case "name":
		return "package name must match @scope/package-name format with lowercase alphanumeric and hyphens only"
	case "version":
		return "version must be in MAJOR.MINOR.PATCH format (e.g., 1.0.0)"
	case "description":
		return descriptionError(err)
	default:
		return err.Error()
	}
}

func descriptionError(err error) string {
	if hasKeyword(err, "maxLength") {
		return "description must be at most 255 characters"
	}
	return "invalid description"
}

func hasKeyword(err error, keyword string) bool {
	var ve *jsonschema.ValidationError
	if !errors.As(err, &ve) {
		return false
	}
	if ve.ErrorKind != nil {
		path := ve.ErrorKind.KeywordPath()
		for _, p := range path {
			if strings.Contains(p, keyword) {
				return true
			}
		}
	}
	for _, cause := range ve.Causes {
		if hasKeyword(cause, keyword) {
			return true
		}
	}
	return false
}
