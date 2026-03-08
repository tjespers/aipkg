# aipkg — Architecture Concepts

## System Overview

```
┌─────────────────────────────────────────────────────────┐
│                      aipkg CLI                          │
│                                                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────┐  │
│  │  Resolve  │→│  Fetch    │→│  Unpack   │→│ Adapt  │  │
│  │          │  │          │  │ & Store   │  │        │  │
│  └──────────┘  └──────────┘  └──────────┘  └────────┘  │
│       ↑                                        │        │
│       │                                        ↓        │
│  ┌──────────┐                          ┌────────────┐   │
│  │  Source   │                          │  Adapter    │   │
│  │  Types    │                          │  Registry   │   │
│  └──────────┘                          └────────────┘   │
└─────────────────────────────────────────────────────────┘
        ↑                                        │
        │                                        ↓
 ┌──────────────┐                     ┌──────────────────┐
 │ GitHub       │                     │ ~/.claude/skills/ │
 │ HTTP server  │                     │ .cursor/rules/    │
 │ S3 bucket    │                     │ .windsurf/rules/  │
 │ Registry     │                     │ MCP configs       │
 └──────────────┘                     └──────────────────┘
```

## Core Pipeline

### 1. Resolve

Takes a package identifier and figures out where to download it.

**Two input modes:**
- `@scope/name` — package identity. Searches configured repositories by manifest name.
- `owner/repo` — GitHub shorthand. Goes directly to `github.com/owner/repo`.

**Resolution order:**
1. Check configured repositories (priority order, first match wins)
2. Canonical repositories short-circuit: if a scope is canonical, only that repo is checked
3. For GitHub shorthand: go directly to GitHub releases

### 2. Fetch

Downloads the package archive from the resolved source.

**Source types** each implement one contract: given a package name and version, return a download URL.

| Source Type | Resolution Logic | Auth |
|---|---|---|
| `github` | GitHub Releases API → find asset by tag | GitHub token (optional for public) |
| `http` | URL template with `${name}` and `${version}` placeholders | Basic auth, headers, or none |
| `s3` | Construct S3 object URL from bucket + key pattern | AWS credentials |
| `gdrive` | Google Drive API → resolve file by folder + name | OAuth |
| `registry` | aipkg.dev API → resolve download URL | API key |

### 3. Unpack & Store

Extracts the archive and stores it in the install scope's package directory.

### 4. Adapt

Reads the manifest's artifact list and creates tool-specific symlinks/copies.

**Adapter operation per artifact:**
1. Read artifact type and name from manifest
2. Look up target path for the active tool adapter
3. Create symlink from package store → tool directory
4. For JSON artifacts (MCP servers): merge into tool config file

## Install Scopes

Packages install into one of two scopes, determined by the `--global` flag:

### Project Scope (default)

```
my-project/
├── aipkg.json              ← project config (committed to VCS)
├── .aipkg/                 ← installed packages (gitignored)
│   └── @shiftbase/
│       └── golang-expert/
│           └── 1.0.0/
│               ├── aipkg.json
│               ├── skills/
│               └── mcp/
├── .claude/                ← adapter symlinks (project-level)
│   └── skills/
│       └── shiftbase.golang-test-writer → ../../.aipkg/@shiftbase/golang-expert/1.0.0/skills/test-writer/
└── src/
```

- `.aipkg/` is gitignored — teammates run `aipkg install` to populate
- `aipkg.json` is committed — it's the team contract
- Adapters create symlinks in project-level tool directories

### Global Scope (`--global`)

```
~/.aipkg/
├── packages/
│   └── @alice/
│       └── my-personal-tool/
│           └── 1.0.0/
│               └── ...
└── config.json              ← user-level defaults (repos, credentials)

~/.claude/
└── skills/
    └── alice.my-personal-tool → ~/.aipkg/packages/@alice/my-personal-tool/1.0.0/skills/...
```

- For personal tools available everywhere
- Adapters create symlinks in user-level tool directories

## Package Format

### Manifest (`aipkg.json`)

**As a package:**

```json
{
  "type": "package",
  "name": "@shiftbase/golang-expert",
  "version": "1.0.0",
  "description": "Complete Go development assistant with skills and tooling",
  "artifacts": [
    {
      "name": "golang-expert-agent",
      "type": "agent",
      "path": "agent/"
    },
    {
      "name": "golang-test-writer",
      "type": "skill",
      "path": "skills/test-writer/"
    },
    {
      "name": "golang-refactoring",
      "type": "skill",
      "path": "skills/refactoring/"
    },
    {
      "name": "golang-docs-search",
      "type": "mcp-server",
      "path": "mcp/golang-docs.json"
    }
  ],
  "require": {
    "@shiftbase/go-tools": "1.0.0"
  }
}
```

**As a project:**

```json
{
  "type": "project",
  "require": {
    "@shiftbase/golang-expert": "1.0.0",
    "@alice/blog-writer": "2.0.0"
  },
  "repositories": [
    {
      "type": "github",
      "url": "github.com/shiftbase-com/*",
      "scope": "@shiftbase",
      "canonical": true
    },
    {
      "type": "http",
      "url": "https://ai-packages.internal.company.com/${name}/${version}.zip"
    }
  ]
}
```

### Single-Artifact Package (Simplest Case)

```json
{
  "type": "package",
  "name": "@alice/blog-writer",
  "version": "1.0.0",
  "description": "Thorough code review prompt",
  "artifacts": [
    {
      "name": "code-review",
      "type": "prompt",
      "path": "review.md"
    }
  ]
}
```

A valid package is just this JSON file + the referenced files, zipped.

### Artifact Types (v1)

| Type | Typical Content | Adapter Action |
|---|---|---|
| `skill` | Markdown directory or file | Symlink to tool's skill/rule directory |
| `prompt` | Markdown file | Symlink to tool's prompt directory |
| `command` | Markdown file | Symlink to tool's command directory |
| `agent` | Markdown file | Symlink to tool's agent directory |
| `mcp-server` | JSON config | Merge into tool's MCP settings file |

## Dependency Resolution (v1: Exact Versions)

v1 uses exact version pinning only. No version ranges, no constraint solver.

### Install Algorithm

```
install(package, version):
  1. Resolve source for package
  2. Fetch package@version archive
  3. Unpack to install scope's .aipkg/ directory
  4. Read package's aipkg.json
  5. For each entry in require:
     a. If already installed at the exact version → skip
     b. If installed at a different version → ERROR (conflict)
     c. Otherwise → install(dependency, exact_version)  // recurse
  6. Run adapter for each artifact
```

This is a simple depth-first traversal. No SAT solver needed.

### Conflict Detection

The only possible conflict: two packages require different exact versions of the same dependency.

```
@acme/tool-a requires @shared/utils 1.0.0
@acme/tool-b requires @shared/utils 2.0.0
→ ERROR: version conflict for @shared/utils (1.0.0 vs 2.0.0)
```

Resolution: the user must align versions manually. Version ranges (later) will make this more flexible.

## Repository Configuration

### Resolution Priority

Repositories are checked in order. First match wins. This follows Composer's model.

**Canonical scope behavior:** When a repository declares `"canonical": true` with a scope, that scope can ONLY be resolved from that repository. This prevents dependency confusion attacks where a malicious public package impersonates a private one.

### User-Level Config (`~/.aipkg/config.json`)

Default repositories and credentials that apply when no project-level config is present:

```json
{
  "repositories": [
    {
      "type": "github",
      "url": "github.com/*"
    }
  ],
  "auth": {
    "github.com": { "token": "ghp_..." }
  }
}
```

## CLI Commands (Vision)

### v1 (Core)

| Command | Description |
|---|---|
| `aipkg init` | Scaffold a new package manifest or project config (asks: project or package?) |
| `aipkg install [pkg]` | Install a package, or install all from `require` if no argument |
| `aipkg install --global <pkg>` | Install a package globally |
| `aipkg remove <pkg>` | Uninstall a package (remove from store + cleanup symlinks) |
| `aipkg list` | Show installed packages and their artifacts |

### v2 (Publishing & Versioning)

| Command | Description |
|---|---|
| `aipkg publish` | Package and publish to a source (GitHub release, etc.) |
| `aipkg update [pkg]` | Update package(s) to latest version |
| `aipkg info <pkg>` | Show detailed package metadata |
| `aipkg search <query>` | Search configured repositories |

### v3+ (Registry & Ecosystem)

| Command | Description |
|---|---|
| `aipkg login` | Authenticate with the central registry |
| `aipkg register` | Claim a namespace on the registry |
| `aipkg audit` | Check installed packages for known issues |
| `aipkg adapter list` | Show available adapters |
| `aipkg adapter add` | Install a community adapter |

## Adapter Mapping Examples

```
Artifact: { "name": "golang-test-writer", "type": "skill" }
Package:  @shiftbase/golang-expert (installed in project scope)

Claude Code adapter:
  symlink .aipkg/@shiftbase/golang-expert/1.0.0/skills/test-writer/
       → .claude/skills/shiftbase.golang-test-writer/

Cursor adapter:
  symlink .aipkg/@shiftbase/golang-expert/1.0.0/skills/test-writer/skill.md
       → .cursor/rules/shiftbase.golang-test-writer.md
```

```
Same artifact, installed globally:

Claude Code adapter:
  symlink ~/.aipkg/packages/@shiftbase/golang-expert/1.0.0/skills/test-writer/
       → ~/.claude/skills/shiftbase.golang-test-writer/
```
