# Data Model: Package Foundation

## Entities

### PackageManifest

The Go representation of `aipkg.json` for packages. Serialized as JSON.

| Field | Go type | JSON key | Required | Validation |
|-------|---------|----------|----------|------------|
| SpecVersion | `int` | `specVersion` | yes | Must be `1` (current spec version) |
| Name | `string` | `name` | yes | Scoped format `@scope/pkg`, per JSON Schema pattern |
| Version | `string` | `version` | yes | Strict semver `MAJOR.MINOR.PATCH`, per JSON Schema pattern |
| Description | `string` | `description,omitempty` | no | Max 255 characters |
| License | `string` | `license,omitempty` | no | SPDX identifier or `"proprietary"` |

Notes:
- No `type` field. This struct describes packages only. Project manifests will use a separate type.
- No `artifacts` field at creation time. That field is populated later by `aipkg pack`.
- No `require` or `repositories` fields at creation time. Those are project-consumption concerns.

### ScopedName

Parsed representation of a scoped package name.

| Field | Go type | Description |
|-------|---------|-------------|
| Scope | `string` | The scope portion without `@` (e.g., `alice`) |
| Package | `string` | The package name portion (e.g., `blog-writer`) |

Derived from parsing the full name string `@scope/package-name`. The `String()` method returns the canonical form `@scope/package-name`.

### ArtifactType

Enumeration of well-known artifact types and their directory mappings.

| Directory | Artifact Type | Structure |
|-----------|--------------|-----------|
| `skills/` | `skill` | Directory with `SKILL.md` |
| `prompts/` | `prompt` | Single markdown file |
| `commands/` | `command` | Single markdown file |
| `agents/` | `agent` | Single markdown file |
| `agent-instructions/` | `agent-instructions` | Single markdown file |
| `mcp-servers/` | `mcp-server` | Single JSON file |

This mapping is used by the scaffold package to create well-known directories. The mapping from directory name to artifact type will be used later by `aipkg pack` to generate the `artifacts` array.

### ReservedScopes

The list of reserved scope prefixes, loaded from `spec/reserved-scopes.txt`.

| Entry type | Example | Matching rule |
|------------|---------|---------------|
| Exact match | `official` | Scope must equal the entry exactly |
| Prefix match (trailing `*`) | `aipkg*` | Scope must start with the prefix |

## Relationships

```
PackageManifest
  └── name: ScopedName (parsed from string)

ScopedName
  └── scope: checked against ReservedScopes

ArtifactType[]
  └── used by scaffold to create well-known directories
  └── used by manifest (future: aipkg pack) to map directories to types
```

## State Transitions

The create command has a linear flow with no persistent state:

```
[Start]
  → Parse flags and positional arg
  → Detect TTY availability
  → Resolve target directory (--path or derive from name)
  → Check for existing aipkg.json (abort if found)
  → Detect LICENSE file (if target dir exists)
  → Build prompt form (skip fields provided via flags)
  → Run prompts (or validate flags if fully non-interactive)
  → Validate all fields (schema bridge + reserved scope check)
  → Write aipkg.json
  → Create well-known directories (skip existing)
  → Print success message
[End]
```

On Ctrl+C or validation failure during prompts, the flow exits without writing any files. The operation is atomic in the sense that either all files are created or none are. If directory creation fails partway through, the already-created directories remain (they're empty and harmless), but aipkg.json is written last to avoid a half-valid package.

## JSON Schema (package-only)

The package JSON Schema (`spec/schema/aipkg.json`) is a package-only schema. It validates `aipkg.json` manifest files. Key design decisions:

1. No `type` field (schema describes packages only).
2. `specVersion` (integer, required, const `1`).
3. No `if/then/else` conditional logic (no type discriminator needed).
4. `artifacts` is optional (not required at creation time, required at pack time; enforced by the pack command).
5. All field-level validation unchanged: name pattern, version pattern, description maxLength.

The `$defs` section contains the artifact definition only.
