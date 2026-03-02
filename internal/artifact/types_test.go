package artifact

import "testing"

func TestDirToTypeMappings(t *testing.T) {
	expected := map[string]Type{
		"skills":             TypeSkill,
		"prompts":            TypePrompt,
		"commands":           TypeCommand,
		"agents":             TypeAgent,
		"agent-instructions": TypeAgentInstructions,
		"mcp-servers":        TypeMCPServer,
	}

	if len(DirToType) != len(expected) {
		t.Errorf("DirToType has %d entries, want %d", len(DirToType), len(expected))
	}

	for dir, wantType := range expected {
		got, ok := DirToType[dir]
		if !ok {
			t.Errorf("DirToType missing directory %q", dir)
			continue
		}
		if got != wantType {
			t.Errorf("DirToType[%q] = %q, want %q", dir, got, wantType)
		}
	}
}

func TestTypeToDirMappings(t *testing.T) {
	// Every DirToType entry must have a reverse mapping
	for dir, artType := range DirToType {
		got, ok := TypeToDir[artType]
		if !ok {
			t.Errorf("TypeToDir missing type %q", artType)
			continue
		}
		if got != dir {
			t.Errorf("TypeToDir[%q] = %q, want %q", artType, got, dir)
		}
	}

	// Both maps should have the same number of entries
	if len(TypeToDir) != len(DirToType) {
		t.Errorf("TypeToDir has %d entries, DirToType has %d", len(TypeToDir), len(DirToType))
	}
}

func TestIsDirectoryBased(t *testing.T) {
	tests := []struct {
		artType Type
		want    bool
	}{
		{TypeSkill, true},
		{TypePrompt, false},
		{TypeCommand, false},
		{TypeAgent, false},
		{TypeAgentInstructions, false},
		{TypeMCPServer, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.artType), func(t *testing.T) {
			if got := tt.artType.IsDirectoryBased(); got != tt.want {
				t.Errorf("IsDirectoryBased() = %v, want %v", got, tt.want)
			}
		})
	}
}
