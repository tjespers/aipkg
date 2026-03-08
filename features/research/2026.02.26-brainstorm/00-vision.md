# aipkg — Vision

## The Problem

AI adoption is accelerating within organizations, but there's no uniform, easy way to share, reuse, and distribute AI artifacts. Developers are creating prompts, skills, agents, commands, and MCP server configurations — but sharing them means copying files, pasting into Slack, or maintaining undiscoverable internal wikis.

Meanwhile, every AI tool (Claude Code, Cursor, Windsurf, etc.) stores its artifacts in different locations with different naming conventions. There's no interoperability layer.

## The Vision

**aipkg** is a package management ecosystem for AI artifacts. Think npm or Composer, but for prompts, skills, agents, commands, and MCP server configurations.

A developer (or a non-technical power user) should be able to:

1. **Package** their AI artifacts with a simple manifest
2. **Publish** them to a repository (GitHub releases, S3, HTTP, or a central registry)
3. **Discover** packages others have created
4. **Install** them with a single command
5. **Use** them immediately — packages are automatically placed where the target tool expects them

## Core Concepts

### AI Artifacts

The initial set of supported artifact types:

| Type | Typical Format | Example |
|---|---|---|
| **Skills** | Markdown | Reusable instruction sets for AI assistants |
| **Prompts** | Markdown | Standalone prompt templates |
| **Commands** | Markdown | Slash commands |
| **Agents** | Markdown | Persona definitions |
| **MCP Servers** | JSON | Server configuration snippets |

Most artifacts are markdown files. MCP servers are the exception (JSON config).

### Packages as Bundles

A package is a **bundle** that can contain multiple artifacts. For example, `@acme/golang-expert` might contain:

- An agent persona (`AGENT.md`)
- A `golang-test-writer` skill
- A `golang-refactoring` skill
- An MCP server config for Go documentation search

Single-artifact packages are just bundles of one. The model stays consistent — no special cases.

### Tool-Agnostic Core, Tool-Specific Adapters

Packages are stored in a canonical, tool-agnostic format. **Adapters** handle placing artifacts where specific tools expect them:

- Claude Code adapter → `~/.claude/skills/`, MCP config merging
- Cursor adapter → `.cursor/rules/`
- Windsurf adapter → `.windsurf/rules/`

Most adapters are thin — they just rename and symlink. Same content, different destination.

### Install Scopes (Project vs Global)

Like npm and Composer, aipkg supports two install scopes:

**Project install (default):** Packages go into `.aipkg/` in the project root. The config file (`aipkg.json`) is committed to version control — teammates clone the repo and run `aipkg install` to get the same setup. `.aipkg/` is gitignored (like `node_modules/`).

**Global install (`--global`):** Packages go into `~/.aipkg/`. Available everywhere, not tied to a project. For personal tools and preferences.

Adapters operate at both levels:
- Project scope → project-level tool dirs (`.claude/` in the repo, `.cursor/rules/`)
- Global scope → user-level tool dirs (`~/.claude/skills/`, user-level Cursor config)

### Namespacing

Always scoped, following npm's `@scope/name` pattern:

```
@shiftbase/golang-expert
@alice/blog-writer-prompt
```

- `@` prefix = package identity (owned namespace)
- No unscoped packages — learned from npm/PyPI's naming mistakes
- Namespace ownership tied to registry accounts (later) or GitHub orgs (v1)
- The `@` also serves as a **parsing disambiguator**: `@scope/name` = package identity (resolve from repos) vs `owner/repo` = GitHub shorthand (fetch directly)

### Manifest Types (Composer-Style)

One file format (`aipkg.json`), two roles determined by a `type` field:

- `"type": "project"` — this is an application/repo that consumes packages. Has `require` and `repositories`.
- `"type": "package"` — this is a publishable package. Has `name`, `version`, `artifacts`, and optionally `require` for its own dependencies.

This avoids the need for two different file formats while keeping roles explicit and validatable.

### Dependencies (v1: Exact Versions Only)

Packages can declare dependencies on other packages from day 1, but only with exact version pinning — no version ranges.

```json
{ "require": { "@shiftbase/go-tools": "1.0.0" } }
```

This gives transitive dependency support (package A needs B needs C) without needing a constraint solver. The hard problem (version range resolution) is deferred. The install algorithm is just: fetch, read dependencies, recurse, error on conflicts.

### Multiple Source Types

Not everyone is a developer. The ops person who writes the best prompts shouldn't need Git to share them.

Supported sources (progressive):
- **GitHub releases** — developer default (v1)
- **HTTP** — any URL that serves the file (v1)
- **S3/GCS** — enterprise cloud storage (v2)
- **Google Drive** — non-technical users (v2-3)
- **Central registry** — aipkg.dev (v3-4)

## Target Audience

**Primary (v1):** Developers familiar with package managers (npm, Composer, Cargo). They already work with AI tools and want to share artifacts within their team or publicly.

**Secondary (v2+):** Non-technical power users who create great AI artifacts but aren't comfortable with Git. They need simpler publishing workflows (drag-and-drop, web UI).

**Tertiary (v3+):** Organizations wanting governed, discoverable AI artifact ecosystems. Central registry, access controls, audit trails.

## Principles

1. **Simple things should be simple** — A valid package is a folder with an `aipkg.json`, zipped. No tooling needed for the basic case.
2. **Tool-agnostic core** — Never tie the package format to a specific tool. Adapters handle the translation.
3. **Always namespaced** — Learned from every ecosystem that started flat and regretted it.
4. **Progressive complexity** — v1 works with just GitHub. Registry, advanced features, and governance come later.
5. **Familiar to developers** — Borrow conventions from npm and Composer. The audience knows these tools.
