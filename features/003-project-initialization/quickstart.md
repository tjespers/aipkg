# Quickstart: Project Initialization & Model

**Feature Branch**: `003-project-initialization`

## What this feature delivers

1. **`aipkg init` command** — Creates `aipkg-project.json` in the current directory
2. **Project file JSON Schema** — `spec/schema/project.json` for validation
3. **Reference documentation** — `spec/project.md` covering the project model
4. **Schema validation** — `ValidateProject()` in `internal/schema`
5. **Project package** — `internal/project` with types and file I/O

## Implementation order

The tasks follow a dependency chain:

1. **Schema first** — Write `spec/schema/project.json` and embed it. Everything else validates against this.
2. **Project package** — `internal/project` with `Create()` and `LoadFile()`. Depends on the schema existing.
3. **Schema validation** — Add `ValidateProject()` to `internal/schema`. Depends on the embedded schema.
4. **Init command** — `internal/cli/init.go`. Depends on `internal/project` for file creation.
5. **Reference docs** — `spec/project.md`. Can be written in parallel with the command, but should reflect the final naming decisions.

## Key patterns to follow

### Command handler pattern (from `create.go`)

```go
func newInitCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "init",
        Short: "Initialize a new aipkg project",
        // ...
        RunE: func(cmd *cobra.Command, args []string) error {
            return runInit(cmd)
        },
    }
    return cmd
}

func runInit(cmd *cobra.Command) error {
    // 1. Check for existing aipkg.json (FR-017)
    // 2. Check for existing aipkg-project.json (FR-016)
    // 3. Create project file (FR-015)
    // ...
}
```

### Schema embedding pattern (from `specdata.go`)

```go
//go:embed spec/schema/project.json
var ProjectSchemaJSON []byte
```

### Project file type (mirrors `internal/manifest`)

```go
type ProjectFile struct {
    SpecVersion int               `json:"specVersion"`
    Require     map[string]string `json:"require"`
}
```

## Testing approach

- **Unit tests** (`internal/project/project_test.go`): Create/Load roundtrip, empty require map, JSON structure
- **Schema tests** (`internal/schema/validate_test.go`): Valid project files pass, invalid ones fail (bad names, bad versions, extra fields, missing fields)
- **Integration tests** (`internal/cli/init_test.go`): Successful init, re-init guard, mutual exclusivity guard, file content verification

All tests use `t.TempDir()` for filesystem isolation.

## Build and verify

```sh
task build          # Verify compilation
task test           # Run all tests
task lint           # Check for lint issues
task check          # Full check (lint + vet + test)
```
