# Data Model: Init Command

**Feature**: 001-init-command | **Date**: 2026-02-28

## Entities

### Manifest

The central entity produced by `aipkg init`. Represents a single `aipkg.json` file.

**Fields** (init-relevant subset — full schema in `aipkg-spec`):

| Field | Go Type | Project | Package | Validation |
| ----- | ------- | ------- | ------- | ---------- |
| `type` | `string` | required | required | Enum: `"project"`, `"package"` |
| `name` | `string` | optional | required | Schema regex: `^@(?!.*--)[a-z0-9]([a-z0-9-]{0,37}[a-z0-9])?/[a-z0-9]([a-z0-9-]{0,62}[a-z0-9])?$` |
| `version` | `string` | omitted | required | Schema regex: `^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$` |
| `description` | `string` | optional | optional | Max 255 characters |
| `license` | `string` | omitted | optional | Free-form string (SPDX by convention) |

**Go struct**:

```go
type Manifest struct {
    Type        string `json:"type"`
    Name        string `json:"name,omitempty"`
    Version     string `json:"version,omitempty"`
    Description string `json:"description,omitempty"`
    License     string `json:"license,omitempty"`
}
```

**Serialization rules**:

- `omitempty` on all optional fields — skipped fields produce no JSON key
- `type` is always present (never omitempty)
- 2-space indented JSON via `json.MarshalIndent(m, "", "  ")`
- Trailing newline appended after serialization
- Field order in JSON follows struct field order (Go's default)

### Lifecycle

Init only creates. No update, delete, or read operations.

```
[no file] → aipkg init → [aipkg.json exists]
```

**Guard**: If `aipkg.json` already exists at the start, the command refuses (FR-007). No state transitions on existing files.

## Validation Rules

All validation derives from the `aipkg-spec` JSON schema (`schema/aipkg.schema.json`). The init command exposes per-field validators for use in interactive prompts:

| Validator | Input | Returns error when |
| --------- | ----- | ------------------ |
| `ValidateName` | string | Does not match schema name pattern |
| `ValidateVersion` | string | Does not match schema version pattern |
| `ValidateDescription` | string | Exceeds 255 characters |

License has no schema-enforced format constraint — any non-empty string is accepted.

## Fields Not Handled by Init

These manifest fields exist in the schema but are not prompted, flagged, or written by init:

- `artifacts` — optional in schema, derived at package time (FR-012); enforced at package/publish time
- `require` — added via `aipkg require`
- `repositories` — added manually or via future commands
