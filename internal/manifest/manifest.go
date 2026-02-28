package manifest

import "encoding/json"

// Manifest represents an aipkg.json file.
type Manifest struct {
	Type        string `json:"type"`
	Name        string `json:"name,omitempty"`
	Version     string `json:"version,omitempty"`
	Description string `json:"description,omitempty"`
	License     string `json:"license,omitempty"`
}

// Marshal serializes the manifest to JSON with 2-space indentation
// and a trailing newline, matching the aipkg convention.
func (m *Manifest) Marshal() ([]byte, error) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')
	return data, nil
}
