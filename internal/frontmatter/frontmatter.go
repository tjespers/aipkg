package frontmatter

import (
	"bytes"
	"errors"
)

var delimiter = []byte("---")

// Extract separates YAML frontmatter from the body of a document.
// It expects the content to start with a "---" line, followed by YAML,
// followed by a closing "---" line. Returns the YAML bytes (without
// delimiters) and the remaining body bytes.
func Extract(content []byte) (yaml, body []byte, err error) {
	// Normalize \r\n to \n for consistent parsing.
	normalized := bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
	lines := bytes.Split(normalized, []byte("\n"))
	if len(lines) == 0 || !bytes.Equal(lines[0], delimiter) {
		return nil, nil, errors.New("missing opening --- delimiter")
	}

	for i := 1; i < len(lines); i++ {
		if bytes.Equal(lines[i], delimiter) {
			yaml = bytes.Join(lines[1:i], []byte("\n"))
			if i+1 < len(lines) {
				body = bytes.Join(lines[i+1:], []byte("\n"))
			}
			return yaml, body, nil
		}
	}

	return nil, nil, errors.New("missing closing --- delimiter")
}
