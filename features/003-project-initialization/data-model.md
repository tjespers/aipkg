# Data Model: Project Initialization & Model

**Feature Branch**: `003-project-initialization`
**Date**: 2026-03-08

## Entities

### Project File (`aipkg-project.json`)

The project file is a JSON document at the project root. It declares package dependencies and tracks the schema version for migration support.

**Fields**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `specVersion` | integer | yes | Schema version. Always `1` for v1. |
| `require` | object | yes | Map of scoped package names to pinned semver versions. |

**`require` map**:
- Keys: scoped package names matching `@scope/package-name` (same pattern as package manifest)
- Values: semver versions with optional pre-release identifiers (see R-002 in research.md)

**Constraints**:
- No identity fields (`name`, `version`, `description`, `license` are forbidden)
- `additionalProperties: false` on the root object (strict schema)
- Empty on creation (`specVersion: 1`, `require: {}`)

**Example (empty, as created by `aipkg init`)**:

```json
{
  "specVersion": 1,
  "require": {}
}
```

**Example (with dependencies)**:

```json
{
  "specVersion": 1,
  "require": {
    "@shiftbase/golang-expert": "1.2.0",
    "@alice/code-review": "0.3.0-beta.1"
  }
}
```

**Validation**: JSON Schema at `spec/schema/project.json`, embedded as `aipkg.ProjectSchemaJSON`.

### Install Directory (`.aipkg/`)

Not created by this feature. Documented here for completeness since the reference documentation (FR-019) defines its layout.

**Structure**:

```text
.aipkg/
├── .gitignore              # Ignores all contents (*) except itself (!.gitignore)
├── skills/                 # Individual skill directories
│   ├── scope.pkg.skill-a/  # Three-segment dot-notation (see R-001)
│   │   └── SKILL.md
│   └── scope.pkg.skill-b/
│       └── SKILL.md
├── prompts/                # Individual prompt files
│   └── scope.pkg.code-review.md
├── commands/               # Individual command files
│   └── scope.pkg.commit-msg.md
├── agents/                 # Individual agent persona files
│   └── scope.pkg.go-expert.md
├── mcp.json                # Merged MCP server configs (all packages)
└── agent-instructions.md   # Merged agent instructions (all packages)
```

**Properties**:
- Created on demand when the first package is installed (not by `aipkg init`)
- Fully gitignored via internal `.gitignore` (git repos only)
- Merged files (`mcp.json`, `agent-instructions.md`) are aipkg-managed; manual edits are overwritten on install/update

### Installed Artifact

An artifact placed in the appropriate `.aipkg/` subdirectory using the three-segment scoped name.

**Naming convention**: `scope.package-name.artifact-name`

| Artifact type | Location | Naming |
|---------------|----------|--------|
| skill | `.aipkg/skills/<name>/` | Directory: `scope.pkg.artifact-name/` |
| prompt | `.aipkg/prompts/<name>` | File: `scope.pkg.artifact-name.ext` |
| command | `.aipkg/commands/<name>` | File: `scope.pkg.artifact-name.ext` |
| agent | `.aipkg/agents/<name>` | File: `scope.pkg.artifact-name.ext` |
| mcp-server | `.aipkg/mcp.json` | Merged (keyed by scoped name in JSON) |
| agent-instructions | `.aipkg/agent-instructions.md` | Merged (wrapped in package-identifying markers) |

**Traceability**: Given an installed artifact name like `alice.blog-tools.code-review`, parse as:
- Scope: `alice`
- Package: `blog-tools`
- Artifact: `code-review`
- Source package: `@alice/blog-tools`

Parsing is unambiguous because scope names and package names cannot contain dots.

## Relationships

```text
Project File (aipkg-project.json)
  └── require map
       └── key: @scope/package-name → value: semver version
            └── Install (future feature) produces:
                 └── Installed Artifacts in .aipkg/
                      └── Named: scope.package-name.artifact-name
```

## State Transitions

The project file has a simple lifecycle:

1. **Empty** — Created by `aipkg init` with `specVersion: 1` and empty `require`
2. **With dependencies** — Populated by `aipkg require` (future feature) adding entries to `require`
3. **Updated** — Modified by `aipkg update` (future feature) changing version values in `require`
4. **Removed entries** — Modified by `aipkg remove` (future feature) deleting entries from `require`

This feature only implements transition 1 (creation).
