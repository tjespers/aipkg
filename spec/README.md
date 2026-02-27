# aipkg Specification

<!-- TODO: tagline TBD -->

The specification for **aipkg** (pronounced "ai-pack"), a package management ecosystem for AI artifacts.

## What is aipkg?

Prompt engineers & Teams are building curated sets of AI artifacts faster than ever: prompt templates, coding skills, agent personas, MCP server configs. But there's no standardized way to version, bundle, distribute, or keep track of any of it:

- The marketing team's best blog-writer prompt lives in Alice's personal Google Drive. When Alice leaves, it's gone
- The ops engineer who built the perfect deployment skill shares it by pasting into Slack
- That awesome-skills repo has 50 entries but you only need 2. Copy-paste them and they go stale, or clone the whole thing and carry 48 you don't want
- Nobody knows what version of which prompt anyone is using, or how to reproduce a working setup
- There's no `package.json` for AI artifacts. No manifest, no dependencies, no install command

**aipkg** is package management for AI artifacts. Version them. Bundle them. Distribute them. Install them with one command.

As a bonus, adapters handle the tooling mess. The same package works across Claude Code, Cursor, Windsurf, and future tools without changes.

```bash
# Install a package into your project
aipkg install tjespers/golang-expert

# Your team clones the repo and runs:
aipkg install
# Same setup, every time.
```

A package is just a zip (`.aipkg`) containing an `aipkg.json` manifest and artifact files. Adapters handle placing them where each tool expects.

### What can you package?

| Artifact type | Format | Example |
|---------------|--------|---------|
| **Skills** | Markdown | Reusable instruction sets for AI assistants |
| **Agents** | Markdown | Persona definitions ("you are a senior Go developer") |
| **Agent Instructions** | Markdown | Project-level rules ("prefer stdlib, use table-driven tests") |
| **Prompts** | Markdown | Standalone prompt templates |
| **Commands** | Markdown | Slash commands |
| **MCP Servers** | JSON | Server configuration snippets |

A single package can bundle multiple artifacts: an agent persona, project-level coding conventions, two skills, and an MCP config, all installed atomically.

### Quick look

**Package manifest** (`aipkg.json`):

```json
{
  "type": "package",
  "name": "@tjespers/golang-expert",
  "version": "1.0.0",
  "license": "Apache-2.0",
  "description": "Complete Go development assistant",
  "artifacts": [
    { "name": "go-expert", "type": "agent", "path": "agents/go-expert.md" },
    { "name": "go-conventions", "type": "agent-instructions", "path": "instructions/go-conventions.md" },
    { "name": "test-writer", "type": "skill", "path": "skills/test-writer/" },
    { "name": "refactoring", "type": "skill", "path": "skills/refactoring/" },
    { "name": "go-docs", "type": "mcp-server", "path": "mcp/go-docs.json" }
  ]
}
```

**Project manifest** (`aipkg.json` in your repo):

```json
{
  "type": "project",
  "require": {
    "@tjespers/golang-expert": "1.0.0",
    "@alice/blog-writer": "2.0.0"
  }
}
```

## What's in this repo?

This repository contains the **aipkg specification**: reference documentation and JSON schemas that define how the ecosystem works.

| Area | What it defines |
|------|-----------------|
| **Package manifest** | The `aipkg.json` schema for both `project` and `package` manifests |
| **Artifact types** | Conventions for each artifact type: skill, agent, agent-instructions, prompt, command, mcp-server |
| **Naming rules** | Scoped naming (`@scope/name`), reserved namespaces, dot-notation for installed artifacts |
| **Package archive** | The `.aipkg` format (zip), internal structure |
| **Source types** | Interface contract for package sources (GitHub, HTTP, and future sources) |
| **Adapters** | Interface contract for tool-specific adapters (Claude Code, Cursor, Windsurf, etc.) |
| **Virtual packages** | Recipe format for the `@virtual/` namespace, community wrappers for upstream repos |

```
docs/       Reference documentation
schema/     JSON Schema files for machine-readable validation
```

## Key design decisions

- **Always scoped** - every package is `@scope/name`, no exceptions
- **Tool-agnostic core** - packages are tool-independent; adapters handle placement
- **Simple things should be simple** - a valid package is a folder with `aipkg.json`, zipped as `.aipkg`
- **Semver from day 1** - versions are strict semver, exact pinning in v1
- **Familiar to developers** - conventions borrowed from npm and Composer
- **GitHub-first** - v1 uses GitHub releases; a central registry comes later

## Status

**Phase 1: Foundation.** The manifest schema, artifact types, and naming rules are being specified first. Everything else builds on these.

## Related repositories

| Repository | Purpose |
|------------|---------|
| [ai-interop/aipkg](https://github.com/ai-interop/aipkg) | CLI tool |

## License

Apache-2.0. See [LICENSE](LICENSE).
