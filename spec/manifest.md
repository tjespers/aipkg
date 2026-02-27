# Package Manifest (`aipkg.json`)

The `aipkg.json` file is the core of every aipkg package and project. It describes what the package contains (for packages) or what packages a project depends on (for projects).

A single file format serves both roles. The `type` field determines which one it is.

## Minimal examples

**Package** (what you publish):

```json
{
  "type": "package",
  "name": "@alice/blog-writer",
  "version": "1.0.0",
  "artifacts": [
    { "name": "code-review", "type": "skill", "path": "skill.md" }
  ]
}
```

**Project** (what lives in your repo):

```json
{
  "type": "project",
  "require": {
    "@alice/blog-writer": "1.0.0"
  }
}
```

## Fields

### `type`

**Required.** The role of this manifest.

| Value | Meaning |
|-------|---------|
| `"package"` | This is a publishable package containing artifacts |
| `"project"` | This is a project that consumes packages |

Every other field's behavior depends on this value. See [field availability by type](#field-availability-by-type) for the full matrix.

### `name`

**Required for `package`.** Optional for `project`.

The package's identity. Always scoped: `@scope/package-name`.

```json
"name": "@tjespers/golang-expert"
```

**Format rules:**

- Must match `@scope/package-name`
- **Scope**: lowercase alphanumeric and hyphens, 1-39 characters. No consecutive hyphens, can't start or end with a hyphen.
- **Package name**: lowercase alphanumeric and hyphens, 1-64 characters. Same hyphen rules as scope.
- No dots or underscores (dots are reserved for installed artifact naming).

The name in the manifest is authoritative. If a package is hosted at `github.com/some-other-org/repo`, the `name` field still determines the package's identity. The source URL is just a locator.

When a project manifest includes a `name`, it's informational only (for logging, error messages). It has no effect on dependency resolution.

<!-- TODO: link to docs/naming.md when AIPKG-3 is written -->

### `version`

**Required for `package`.** Ignored for `project`.

Strict semver: `MAJOR.MINOR.PATCH`. No `v` prefix, no pre-release suffixes in v1.

```json
"version": "1.0.0"
```

The CLI resolves `latest` at install time and writes the exact version to the project manifest. The manifest itself always contains a concrete version.

### `description`

**Optional.** A short summary of the package. Plain text, one line.

```json
"description": "Complete Go development assistant with skills and tooling"
```

### `license`

**Optional.** Either an [SPDX license identifier](https://spdx.org/licenses/) or `"proprietary"`.

```json
"license": "Apache-2.0"
```

```json
"license": "proprietary"
```

If omitted, no license is assumed. Package consumers should treat unlicensed packages with caution.

### `artifacts`

**Required for `package`.** Not allowed on `project`.

An array of artifact entries describing what the package contains. A package must have at least one artifact.

```json
"artifacts": [
  { "name": "go-expert", "type": "agent", "path": "agents/go-expert.md" },
  { "name": "go-conventions", "type": "agent-instructions", "path": "instructions/go-conventions.md" },
  { "name": "test-writer", "type": "skill", "path": "skills/test-writer/" },
  { "name": "refactoring", "type": "skill", "path": "skills/refactoring/" },
  { "name": "go-docs", "type": "mcp-server", "path": "mcp/go-docs.json" }
]
```

Each entry has three fields:

#### `artifacts[].name`

**Required.** The artifact's name within the package. Used for installed artifact naming (dot-notation: `scope.artifact-name`).

Lowercase alphanumeric and hyphens, 1-64 characters. Same rules as package names.

Must be unique within the package.

#### `artifacts[].type`

**Required.** The artifact type. Determines how adapters handle the artifact.

| Type | Content | Adapter behavior |
|------|---------|------------------|
| `skill` | Markdown directory | Symlinked to the tool's skill/rule directory |
| `prompt` | Markdown file | Symlinked to the tool's prompt directory |
| `command` | Markdown file | Symlinked to the tool's command directory |
| `agent` | Markdown file | Symlinked to the tool's agent directory |
| `agent-instructions` | Markdown file | Merged into the tool's agent instruction file |
| `mcp-server` | JSON config | Merged into the tool's MCP settings file |

See [Artifact Types](artifacts.md) for detailed format conventions per type.

#### `artifacts[].path`

**Required.** Relative path from the package root to the artifact's file or directory.

- For file-based artifacts: path to the file (e.g., `"skill.md"`, `"mcp/go-docs.json"`)
- For directory-based artifacts: path to the directory with a trailing slash (e.g., `"skills/test-writer/"`)

Paths must use forward slashes. They must not escape the package root (no `../`).

### `require`

**Optional.** Dependencies on other packages.

An object where keys are scoped package names and values are exact semver versions.

```json
"require": {
  "@tjespers/golang-expert": "1.0.0",
  "@alice/blog-writer": "2.0.0"
}
```

Works identically in both `project` and `package` manifests. In a project, these are the packages your team uses. In a package, these are dependencies that get installed transitively.

v1 supports exact versions only. No version ranges (`^1.0.0`, `~2.1`) yet.

### `repositories`

**Optional.** Configures where the CLI looks for packages.

An array of repository source definitions, checked in order. First match wins.

```json
"repositories": [
  {
    "type": "github",
    "url": "github.com/my-org/*",
    "scope": "@my-org",
    "canonical": true
  },
  {
    "type": "http",
    "url": "https://packages.internal.company.com/${name}/${version}.aipkg"
  }
]
```

#### `repositories[].type`

**Required.** The source type. v1 supports:

| Type | Resolution |
|------|------------|
| `github` | Fetches from GitHub Releases API |
| `http` | Fetches from a URL template |

<!-- TODO: link to docs/sources.md when AIPKG-5 is written -->

#### `repositories[].url`

**Required.** The source URL. Format depends on the type:

- **`github`**: A GitHub URL pattern. Use `*` as a wildcard for the repo name (e.g., `github.com/my-org/*`).
- **`http`**: A URL template with `${name}` and `${version}` placeholders.

#### `repositories[].scope`

**Optional.** Restricts this repository to packages under a specific `@scope`. If omitted, the repository can serve any package.

```json
"scope": "@my-org"
```

#### `repositories[].canonical`

**Optional.** Boolean, defaults to `false`.

When `true`, the given scope can only be resolved from this repository. This prevents dependency confusion attacks where a public package impersonates a private one.

```json
{
  "type": "github",
  "url": "github.com/my-org/*",
  "scope": "@my-org",
  "canonical": true
}
```

With this config, any `@my-org/*` package must come from `github.com/my-org/*`. If found elsewhere, the CLI rejects it.

## Field availability by type

| Field | `project` | `package` |
|-------|-----------|-----------|
| `type` | required | required |
| `name` | optional | required |
| `version` | ignored | required |
| `description` | optional | optional |
| `license` | ignored | optional |
| `artifacts` | not allowed | required |
| `require` | optional | optional |
| `repositories` | optional | optional |

## Complete examples

### Multi-artifact package

```json
{
  "type": "package",
  "name": "@tjespers/golang-expert",
  "version": "1.0.0",
  "license": "Apache-2.0",
  "description": "Complete Go development assistant with skills and tooling",
  "artifacts": [
    { "name": "go-expert", "type": "agent", "path": "agents/go-expert.md" },
    { "name": "go-conventions", "type": "agent-instructions", "path": "instructions/go-conventions.md" },
    { "name": "test-writer", "type": "skill", "path": "skills/test-writer/" },
    { "name": "refactoring", "type": "skill", "path": "skills/refactoring/" },
    { "name": "go-docs", "type": "mcp-server", "path": "mcp/go-docs.json" }
  ],
  "require": {
    "@tjespers/go-tools": "1.0.0"
  }
}
```

### Project with custom repositories

```json
{
  "type": "project",
  "require": {
    "@tjespers/golang-expert": "1.0.0",
    "@alice/blog-writer": "2.0.0"
  },
  "repositories": [
    {
      "type": "github",
      "url": "github.com/my-org/*",
      "scope": "@my-org",
      "canonical": true
    },
    {
      "type": "http",
      "url": "https://packages.internal.company.com/${name}/${version}.aipkg"
    }
  ]
}
```
