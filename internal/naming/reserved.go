package naming

import (
	"bufio"
	"bytes"
	"strings"

	aipkg "github.com/tjespers/aipkg"
)

type reservedEntry struct {
	value    string
	isPrefix bool
}

var reservedScopes []reservedEntry

func init() {
	reservedScopes = parseReservedScopes(aipkg.ReservedScopesText)
}

func parseReservedScopes(data []byte) []reservedEntry {
	var entries []reservedEntry
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasSuffix(line, "*") {
			entries = append(entries, reservedEntry{
				value:    strings.TrimSuffix(line, "*"),
				isPrefix: true,
			})
		} else {
			entries = append(entries, reservedEntry{
				value:    line,
				isPrefix: false,
			})
		}
	}
	return entries
}

// IsReservedScope checks whether the given scope (without @) matches a
// reserved scope entry. Returns the matching rule description if reserved.
func IsReservedScope(scope string) (string, bool) {
	for _, entry := range reservedScopes {
		if entry.isPrefix {
			if strings.HasPrefix(scope, entry.value) {
				return entry.value + "*", true
			}
		} else {
			if scope == entry.value {
				return entry.value, true
			}
		}
	}
	return "", false
}
