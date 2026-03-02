package naming

import (
	"fmt"
	"strings"

	"github.com/dlclark/regexp2"
)

// ScopedName is a parsed scoped package name (@scope/package-name).
type ScopedName struct {
	Scope   string
	Package string
}

// String returns the canonical form @scope/package-name.
func (n ScopedName) String() string {
	return "@" + n.Scope + "/" + n.Package
}

// namePattern matches valid scoped package names.
// Uses regexp2 because Go's stdlib regexp doesn't support the (?!.*--) lookahead.
var namePattern = regexp2.MustCompile(
	`^@(?!.*--)[a-z0-9]([a-z0-9-]{0,37}[a-z0-9])?/[a-z0-9]([a-z0-9-]{0,62}[a-z0-9])?$`,
	regexp2.None,
)

// Parse validates and parses a scoped package name string into a ScopedName.
// Returns an error if the name doesn't match the expected format.
func Parse(name string) (ScopedName, error) {
	if !strings.HasPrefix(name, "@") || !strings.Contains(name, "/") {
		return ScopedName{}, fmt.Errorf("package name must be scoped (e.g., @scope/package-name)")
	}

	matched, err := namePattern.MatchString(name)
	if err != nil {
		return ScopedName{}, fmt.Errorf("name validation error: %w", err)
	}
	if !matched {
		return ScopedName{}, fmt.Errorf("invalid package name %q: must match @scope/package-name format with lowercase alphanumeric and hyphens only", name)
	}

	// Split after @ and on /
	withoutAt := name[1:]
	parts := strings.SplitN(withoutAt, "/", 2)

	return ScopedName{
		Scope:   parts[0],
		Package: parts[1],
	}, nil
}
