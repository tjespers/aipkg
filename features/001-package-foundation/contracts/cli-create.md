# CLI Contract: `aipkg create`

## Synopsis

```
aipkg create [@scope/package-name] [flags]
```

## Arguments

| Argument | Position | Required | Description |
|----------|----------|----------|-------------|
| `@scope/package-name` | 1 | no | Scoped package name. If omitted, prompted interactively. |

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--path` | `-p` | string | (derived from name) | Target directory. Use `.` for current directory. |
| `--name` | `-n` | string | (none) | Package name (alternative to positional arg). |
| `--version` | `-v` | string | `0.1.0` | Package version (strict semver). |
| `--description` | `-d` | string | (none) | Short package description. |
| `--license` | `-l` | string | (none) | SPDX license identifier or `proprietary`. |

When the name is provided as both a positional argument and a `--name` flag, the positional argument takes precedence.

## Behavior Matrix

| TTY | All flags provided | Behavior |
|-----|-------------------|----------|
| yes | yes | Create package, no prompts |
| yes | partial | Prompt for missing fields only |
| yes | none | Prompt for all fields |
| no | yes | Create package, no prompts |
| no | partial | Exit with error listing missing flags |
| no | none | Exit with error listing missing flags |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Package created successfully |
| 1 | Validation error (invalid name, version, etc.) |
| 1 | Target directory already contains aipkg.json |
| 1 | Filesystem error (permissions, disk full) |
| 1 | Missing required flags in non-interactive mode |
| 130 | User cancelled (Ctrl+C) |

## Output: Directory Structure

Given `aipkg create @alice/blog-writer`:

```
blog-writer/
├── aipkg.json
├── skills/
├── prompts/
├── commands/
├── agents/
├── agent-instructions/
└── mcp-servers/
```

## Output: `aipkg.json`

```json
{
  "specVersion": 1,
  "name": "@alice/blog-writer",
  "version": "0.1.0",
  "description": "AI blog writing assistant",
  "license": "MIT"
}
```

Fields with empty values (`description`, `license`) are omitted from the output.

## Interactive Prompt Flow

When running interactively, prompts appear in this order:

1. **Name** (only if not provided as argument or flag): text input with inline validation against schema pattern and reserved scope check.
2. **Version**: text input, default `0.1.0`, validated as strict semver.
3. **Description**: text input, optional (empty is valid), max 255 characters.
4. **License**: text input, optional. If a LICENSE file is detected in the target directory, the detected SPDX identifier is pre-filled as the default.

Each field validates inline on submit. Invalid input shows an error message below the field and allows re-entry without losing progress on other fields.

## Validation Rules

All validation is driven by the JSON Schema (`spec/schema/aipkg.json`) via the schema bridge, except reserved scope checking which uses `spec/reserved-scopes.txt`.

| Field | Rule | Source |
|-------|------|--------|
| name | Pattern: `^@(?!.*--)[a-z0-9]([a-z0-9-]{0,37}[a-z0-9])?/[a-z0-9]([a-z0-9-]{0,62}[a-z0-9])?$` | JSON Schema |
| name | Not a reserved scope | reserved-scopes.txt |
| version | Pattern: `^(0\|[1-9]\d*)\.(0\|[1-9]\d*)\.(0\|[1-9]\d*)$` | JSON Schema |
| description | maxLength: 255 | JSON Schema |
| license | type: string (no further schema constraint) | JSON Schema |

## Error Messages

Representative examples of user-facing validation errors:

```
Error: package name must be scoped (e.g., @scope/package-name)
Error: scope "@aipkg" is reserved
Error: version must be in MAJOR.MINOR.PATCH format (e.g., 1.0.0)
Error: description must be at most 255 characters
Error: target directory already contains aipkg.json
Error: cannot create package: permission denied
Error: missing required flags for non-interactive mode: --name, --version
```
