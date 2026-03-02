package artifact

// Type represents the type of an artifact within a package.
type Type string

// Artifact type constants for the six well-known artifact types.
const (
	TypeSkill             Type = "skill"
	TypePrompt            Type = "prompt"
	TypeCommand           Type = "command"
	TypeAgent             Type = "agent"
	TypeAgentInstructions Type = "agent-instructions"
	TypeMCPServer         Type = "mcp-server"
)

// DirToType maps well-known directory names to their artifact types.
var DirToType = map[string]Type{
	"skills":             TypeSkill,
	"prompts":            TypePrompt,
	"commands":           TypeCommand,
	"agents":             TypeAgent,
	"agent-instructions": TypeAgentInstructions,
	"mcp-servers":        TypeMCPServer,
}

// TypeToDir maps artifact types to their well-known directory names.
var TypeToDir = map[Type]string{
	TypeSkill:             "skills",
	TypePrompt:            "prompts",
	TypeCommand:           "commands",
	TypeAgent:             "agents",
	TypeAgentInstructions: "agent-instructions",
	TypeMCPServer:         "mcp-servers",
}

// IsDirectoryBased returns true for artifact types stored as directories.
func (t Type) IsDirectoryBased() bool {
	return t == TypeSkill
}
