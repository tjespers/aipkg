# aipkg — Project Breakdown

The aipkg ecosystem naturally separates into distinct projects. Each will have its own repository (or monorepo subdirectory), roadmap, and specs.

## Projects Overview

| Project | What | When | Repository |
|---|---|---|---|
| **aipkg-spec** | Standards & specifications | First — everything depends on this | `ai-interop/aipkg-spec` |
| **aipkg-cli** | The CLI tool | v1 core | `ai-interop/aipkg-cli` |
| **aipkg-adapters** | Tool integration adapters | Ships with CLI, grows over time | TBD (bundled or separate) |
| **aipkg-virtual** | Community virtual package recipes | v1-2 (bootstrap) | `ai-interop/aipkg-virtual` |
| **aipkg-registry** | Central registry API & storage | v3-4 | `ai-interop/aipkg-registry` |
| **aipkg.dev** | Website — docs, discovery, registry UI | Docs early, full site v3-4 | `ai-interop/aipkg.dev` |

## Project Details

### aipkg-spec

**Purpose:** Define the standards that all other projects implement. This is the "RFC" layer.

**Scope:**
- Package manifest format (`aipkg.json` schema), including `type` field (`project` vs `package`)
- Artifact type definitions and conventions
- Repository configuration format
- Source type interface contract
- Adapter interface contract
- Namespace rules and validation (including reserved `@virtual` namespace)
- Package archive structure
- Virtual package recipe format

**Why it's first:** Every other project implements these specs. Getting the manifest format right is foundational. Changes here cascade everywhere.

**Deliverables:** Specification documents, JSON schemas, reference examples.

---

### aipkg-cli

**Purpose:** The command-line tool users interact with. The `npm` equivalent.

**Scope (v1):**
- `aipkg init` — scaffold manifest or project config (asks: project or package?)
- `aipkg install [pkg]` — resolve, fetch, unpack, adapt (or install all from `require`)
- `aipkg install --global <pkg>` — global install
- `aipkg require <pkg> <version>` — add dependency and install
- `aipkg remove <pkg>` — uninstall with cleanup
- `aipkg list` — show installed packages
- Source types: `github`, `http`
- Install scopes: project (`.aipkg/`) and global (`~/.aipkg/`)
- Adapter execution (ships with built-in adapters)
- Virtual package support (`@virtual/owner:repo` resolution)
- Exact-version dependency resolution (recursive fetch, conflict detection)

**Scope (v2+):**
- `aipkg publish`, `aipkg update`, `aipkg search`
- Lockfile support
- Version ranges and constraint solving
- Additional source types (S3, etc.)
- Plugin architecture for community adapters

**Language:** TBD (research topic #5). Go and TypeScript are the top candidates.

**Key decisions needed:**
- Language choice
- Distribution method (homebrew, GitHub releases, npm global install, all of the above?)

---

### aipkg-adapters

**Purpose:** Tool-specific integration logic. Knows where each tool expects artifacts and how to place them.

**Built-in adapters (ship with CLI):**
- `claude-code` — skills, commands, MCP config merging
- `cursor` — rules
- `windsurf` — rules

**Architecture question:** Are adapters:
- **A.** Config files (JSON mapping artifact types → paths)? Simplest, but limited.
- **B.** Code plugins loaded by the CLI? More flexible, but adds complexity.
- **C.** Both? Config for simple cases, code for complex ones (like MCP merging).

Option C seems likely — most adapters are just path mappings, but MCP config merging needs logic.

**May live in the CLI repo initially** and split out when community adapters become a thing.

---

### aipkg-virtual

**Purpose:** Community-maintained repository of virtual package recipes. The "DefinitelyTyped" of aipkg.

**Structure:**
```
aipkg-virtual/
├── some-author/
│   └── spec-kit/
│       └── aipkg.json      ← recipe with artifact_mapping
├── another-author/
│   └── awesome-skills/
│       └── aipkg.json
└── README.md
```

**Recipe format (minimal):**
```json
{
  "upstream": "github.com/some-author/spec-kit",
  "version_source": "releases",
  "artifact_mapping": [
    { "name": "code-review", "type": "skill", "path": "skills/code-review/" },
    { "name": "refactoring", "type": "skill", "path": "skills/refactoring/" }
  ]
}
```

**How it works:**
1. `aipkg require @virtual/some-author:spec-kit 2.1.0`
2. CLI fetches recipe from `aipkg-virtual` repo
3. CLI fetches upstream at tag/release `2.1.0`
4. Maps files according to `artifact_mapping`
5. Packages and installs locally like any normal package

**Contributing:** PR to the repo with a recipe file. Low barrier. Future: zero-friction contribution via the CLI (see [wild ideas](05-wild-ideas.md)).

**Lifecycle:** When an upstream author adopts aipkg natively, the virtual recipe is deprecated. The CLI could suggest: "An official package is now available."

---

### aipkg-registry

**Purpose:** Central package registry — discover, publish, and distribute packages. The "npmjs.com" backend.

**Scope:**
- Package storage and versioning
- Namespace claiming and verification
- Search API
- Download API
- Authentication and authorization
- Webhook support (publish notifications, etc.)
- Virtual package contribution endpoint (receive recipes from CLI telemetry)
- AI-based vetting pipeline for contributed recipes

**When:** v3-4. Only needed once the manifest format is stable and there's a community publishing packages. Until then, GitHub releases + custom repos + `aipkg-virtual` are sufficient.

---

### aipkg.dev

**Purpose:** The public face of the ecosystem.

**Phased scope:**
- **Early (v1-2):** Documentation site. How to create packages, install them, write adapters. Hosted via GitHub Pages or similar.
- **Later (v3-4):** Package discovery UI. Browse, search, view readmes, see download stats. Virtual package browser. Think npmjs.com.
- **Business model:** Freemium. Free for public packages. Paid tiers for private packages, team management, audit logs, SLAs.

---

## Dependency Graph

```
aipkg-spec ──→ aipkg-cli ──→ aipkg-registry
     │              │               │
     │              ↓               ↓
     ├────→ aipkg-adapters    aipkg.dev
     │
     └────→ aipkg-virtual
```

- **aipkg-spec** is upstream of everything
- **aipkg-cli** implements the spec and consumes `aipkg-virtual`
- **aipkg-adapters** implement the adapter interface from the spec
- **aipkg-virtual** follows the recipe format from the spec
- **aipkg-registry** implements the source type and publishing specs
- **aipkg.dev** consumes the registry API

## Getting Started — Suggested Order

1. **aipkg-spec:** Nail down the manifest format, artifact types, and virtual package recipe format
2. **aipkg-cli (v1):** Build `init`, `install`, `require`, `remove`, `list` with GitHub + HTTP sources, project + global scopes, exact-version deps
3. **aipkg-adapters:** Ship claude-code and cursor adapters with the CLI
4. **aipkg-virtual:** Seed with recipes for popular repos, accept community PRs
5. **aipkg.dev (docs):** Documentation site so people can learn how to use it
6. **aipkg-cli (v2):** Publish, update, search, lockfile, version ranges
7. **aipkg-registry + aipkg.dev (full):** Central registry, discovery UI, contribution pipeline

## Ecosystem Comparison

| Concept | npm equivalent | Composer equivalent | aipkg equivalent |
|---|---|---|---|
| Manifest | `package.json` | `composer.json` | `aipkg.json` |
| CLI | `npm` | `composer` | `aipkg` |
| Registry | npmjs.com | packagist.org | aipkg.dev (future) |
| Lockfile | `package-lock.json` | `composer.lock` | TBD |
| Scoping | `@scope/name` | `vendor/name` | `@scope/name` |
| Config | `.npmrc` | `composer.json` repositories | `aipkg.json` repositories |
| Community types | DefinitelyTyped (`@types/`) | — | `aipkg-virtual` (`@virtual/`) |
