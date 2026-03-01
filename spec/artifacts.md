# Artifact Types

Each artifact in a package has a `type` that determines its format requirements and how adapters handle it. This document defines the conventions for each type.

aipkg does not invent new artifact formats. Where an established standard exists, we adopt it. Where no standard exists, we keep things minimal and let the content speak for itself.

## Package directory layout

Packages use a convention-based directory structure. Six well-known directories at the package root map to artifact types:

| Directory             | Artifact type        | Structure                 |
| --------------------- | -------------------- | ------------------------- |
| `skills/`             | `skill`              | Directory with `SKILL.md` |
| `prompts/`            | `prompt`             | Single markdown file      |
| `commands/`           | `command`            | Single markdown file      |
| `agents/`             | `agent`              | Single markdown file      |
| `agent-instructions/` | `agent-instructions` | Single markdown file      |
| `mcp-servers/`        | `mcp-server`         | Single JSON file          |

When you run `aipkg pack`, the CLI scans these directories and generates the `artifacts` array in the manifest. You don't write artifact entries by hand.

### How artifact names are derived

Artifact names come from filenames and directory names within the well-known directories:

- **File-based types** (prompt, command, agent, agent-instructions, mcp-server): the filename without its extension becomes the artifact name. `prompts/code-review.md` produces artifact name `code-review`. `mcp-servers/go-docs.json` produces `go-docs`.
- **Directory-based types** (skill): the directory name becomes the artifact name. `skills/test-writer/` produces artifact name `test-writer`.

Names must follow the standard naming rules: lowercase alphanumeric and hyphens, 1-64 characters, no consecutive hyphens, can't start or end with a hyphen.

### Example layout

```text
my-package/
├── aipkg.json
├── skills/
│   ├── test-writer/
│   │   └── SKILL.md
│   └── refactoring/
│       └── SKILL.md
├── prompts/
│   └── code-review.md
├── commands/
│   └── commit-msg.md
├── agents/
│   └── go-expert.md
├── agent-instructions/
│   └── go-conventions.md
└── mcp-servers/
    └── go-docs.json
```

Running `aipkg pack` on this layout generates seven artifact entries. You only need to provide `specVersion`, `name`, and `version` in your `aipkg.json`. The tooling handles the rest.

## Overview

| Type                 | Format        | Structure | Adapter | Standard                                             |
| -------------------- | ------------- | --------- | ------- | ---------------------------------------------------- |
| `skill`              | Markdown      | Directory | Symlink | [Agent Skills](https://agentskills.io/specification) |
| `agent`              | Markdown      | File      | Symlink | None                                                 |
| `agent-instructions` | Markdown      | File      | Merge   | [AGENTS.md](https://agents.md/)                      |
| `mcp-server`         | JSON          | File      | Merge   | [MCP](https://modelcontextprotocol.io)               |
| `prompt`             | Markdown/text | File      | Symlink | None                                                 |
| `command`            | Markdown      | File      | Symlink | None                                                 |

## `skill`

Reusable instruction sets that give AI agents new capabilities or domain expertise.

Skills follow the [Agent Skills specification](https://agentskills.io/specification). A skill is a directory containing a `SKILL.md` file with YAML frontmatter, plus optional supporting directories.

### Required structure

Each subdirectory under `skills/` must contain a `SKILL.md` file. Without it, `aipkg pack` will reject the directory.

```text
skills/test-writer/
├── SKILL.md           # Required. Frontmatter + instructions.
├── scripts/           # Optional. Executable code the agent can run.
├── references/        # Optional. Additional docs loaded on demand.
└── assets/            # Optional. Templates, images, data files.
```

### SKILL.md format

The file must have YAML frontmatter with at least `name` and `description`:

```markdown
---
name: test-writer
description: Writes comprehensive test suites for Go packages. Use when creating or updating tests.
---

## Instructions

When writing tests for a Go package:

1. Read the existing source files to understand the API
2. Create table-driven tests for all exported functions
3. Include edge cases and error conditions
...
```

See the [Agent Skills specification](https://agentskills.io/specification) for the full frontmatter schema, naming rules, and optional fields like `license`, `compatibility`, `metadata`, and `allowed-tools`.

### In the manifest

Skills are always referenced as directories (trailing slash):

```json
{ "name": "test-writer", "type": "skill", "path": "skills/test-writer/" }
```

The directory name should match the `name` field in the SKILL.md frontmatter.

## `agent`

Standalone persona definitions that shape how an AI agent behaves. An agent artifact defines the identity, tone, expertise areas, and behavioral guidelines for an AI assistant.

This is different from `agent-instructions` (see below). An agent persona says "here's who you are." Instructions say "here's how to work in this project." Personas are symlinked as individual files. Instructions are merged into the project's agent instruction file.

### Format

A single Markdown file under `agents/`. No required frontmatter or sections. Write whatever helps the agent understand its role.

```markdown
# Go Expert

You are a senior Go developer with deep expertise in the standard library,
concurrency patterns, and idiomatic Go style.

## Behavior

- Always prefer the standard library over third-party packages
- Write table-driven tests
- Use `context.Context` for cancellation and timeouts
- Favor composition over inheritance

## Knowledge

- Go 1.22+ features and conventions
- Common patterns: functional options, middleware chains, error wrapping
- Toolchain: `go test`, `go vet`, `staticcheck`
```

### In the manifest

Agents are single files:

```json
{ "name": "go-expert", "type": "agent", "path": "agents/go-expert.md" }
```

## `agent-instructions`

Project-level instructions that tell AI agents how to work in a specific codebase. Think coding standards, architectural conventions, review checklists, or workflow rules that apply to everyone working in the project.

Agent instructions follow the [AGENTS.md](https://agents.md/) convention: standard Markdown with no required structure.

The key difference from `agent` artifacts: agent instructions are **merged** into the project's agent instruction file (AGENTS.md, CLAUDE.md, .cursorrules, etc.), not symlinked as standalone files. Multiple packages can contribute instructions to the same project. This is the same merge pattern used by `mcp-server` artifacts.

### Format

A single Markdown file under `agent-instructions/`. Write project-relevant guidelines, conventions, or rules.

```markdown
## Go Conventions

- Prefer the standard library over third-party packages when reasonable
- Use table-driven tests for all exported functions
- Always handle errors explicitly; never discard with `_`
- Use `context.Context` for cancellation and timeouts
- Run `go vet` and `staticcheck` before committing
```

### In the manifest

Instructions are single files:

```json
{ "name": "go-conventions", "type": "agent-instructions", "path": "agent-instructions/go-conventions.md" }
```

### Adapter behavior

Like `mcp-server` artifacts, instructions are **merged** rather than symlinked. The adapter:

1. Reads the instruction content
1. Wraps it in package-identifying markers (so it can be cleanly removed later)
1. Appends it to the project's agent instruction file
1. On removal, finds and removes the marked section

The specific target file and marker format are adapter-dependent. For example, one adapter might append to AGENTS.md while another merges into a tool-specific rules file.

## `mcp-server`

Configuration for [Model Context Protocol](https://modelcontextprotocol.io) servers. These are JSON files that tell AI tools how to connect to an MCP server.

### Format

A single JSON file under `mcp-servers/`. Two transport types are supported:

**stdio** (local command):

```json
{
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-github"],
  "env": {
    "GITHUB_TOKEN": "${GITHUB_TOKEN}"
  }
}
```

**http/sse** (remote endpoint):

```json
{
  "type": "http",
  "url": "https://mcp.example.com/api",
  "headers": {
    "Authorization": "Bearer ${API_KEY}"
  }
}
```

Environment variable references use `${VAR_NAME}` syntax. The adapter is responsible for resolving these from the user's environment when merging into the tool's MCP configuration.

### In the manifest

MCP servers are single JSON files:

```json
{ "name": "github-api", "type": "mcp-server", "path": "mcp-servers/github.json" }
```

### Adapter behavior

Unlike other artifact types that get symlinked, MCP server configs are **merged** into the tool's existing MCP configuration file. The adapter reads the JSON and adds the server entry to the appropriate config (e.g., Claude Code's `mcp_servers` in settings, Cursor's MCP config). On removal, the adapter removes the entry.

## `prompt`

Standalone prompt templates. These are general-purpose text files intended to be used as-is or with minor customization.

### Format

A single Markdown or plain text file under `prompts/`. No required structure. Prompts can be anything from a one-liner to a multi-page document with sections and examples.

```markdown
# Code Review

Review the following code for:

1. **Correctness** - Does it do what it claims?
2. **Security** - Any injection risks, auth issues, or data leaks?
3. **Performance** - Obvious bottlenecks or unnecessary allocations?
4. **Readability** - Would a new team member understand this?

Be specific. Reference line numbers. Suggest concrete fixes, not vague improvements.
```

### In the manifest

Prompts are single files:

```json
{ "name": "code-review", "type": "prompt", "path": "prompts/code-review.md" }
```

## `command`

Slash commands that can be invoked by name in AI tools. Commands are short, focused instructions typically triggered with a `/` prefix.

### Format

A single Markdown file under `commands/`. Commands can optionally include YAML frontmatter for metadata:

```markdown
---
description: Generate a conventional commit message for staged changes
argument-hint: "[type]"
---

Look at the staged changes (`git diff --cached`) and generate a commit message
following the Conventional Commits specification.

If an argument is provided, use it as the commit type. Otherwise, infer the
appropriate type from the changes.

Format: `type(scope): description`
```

Supported frontmatter fields:

| Field           | Description                                                          |
| --------------- | -------------------------------------------------------------------- |
| `description`   | Short explanation shown in help/autocomplete                         |
| `argument-hint` | Hint about expected arguments (e.g., `[filename]`, `[issue-number]`) |

### In the manifest

Commands are single files:

```json
{ "name": "commit-msg", "type": "command", "path": "commands/commit-msg.md" }
```

## General conventions

These apply across all artifact types:

- **File encoding**: UTF-8
- **Line endings**: LF (Unix-style)
- **Paths**: Forward slashes only, no `../` escaping the package root
- **Markdown**: Standard Markdown (CommonMark). No tool-specific extensions required.
- **Size**: Keep individual files practical. Skills recommend under 500 lines for the main file; move detailed reference material to separate files.
