package artifact

import (
	"errors"
	"fmt"
)

// ValidationErrors collects multiple validation errors from artifact validation.
// Each error carries its file path as context.
type ValidationErrors struct {
	errs []error
}

// Add appends a validation error with the given path and message.
func (v *ValidationErrors) Add(path, message string) {
	v.errs = append(v.errs, fmt.Errorf("%s: %s", path, message))
}

// Addf appends a formatted validation error with the given path.
func (v *ValidationErrors) Addf(path, format string, args ...any) {
	v.Add(path, fmt.Sprintf(format, args...))
}

// Err returns nil if no errors were collected, or a combined error
// with one error per line.
func (v *ValidationErrors) Err() error {
	return errors.Join(v.errs...)
}

// Len returns the number of collected errors.
func (v *ValidationErrors) Len() int {
	return len(v.errs)
}
