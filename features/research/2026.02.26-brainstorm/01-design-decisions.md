# aipkg — Design Decisions

Decisions made during the initial brainstorming session. These are directional — details will be refined during specification.

## Settled

### 1. Always Scoped Naming (`@scope/name`)

**Decision:** Every package must have a scoped name: `@namespace/package-name`. No unscoped packages.

**Rationale:** Ecosystems that started with flat namespaces (npm unscoped, PyPI, Cargo) all suffer from name squatting and typosquatting. Ecosystems that required scopes from day 1 (Composer's `vendor/package`, Go's module paths) never had these problems. It's trivial to enforce at the start, nearly impossible to retrofit.

**The `@` prefix** serves double duty:

- Visually distinctive — immediately reads as "a package name"
- **Parsing disambiguation** — the CLI can distinguish `@shiftbase/my-skill` (package identity, resolve from repos) from `shiftbase-com/my-skill` (GitHub `owner/repo` locator)

### 2. Manifest Is Authoritative for Name

**Decision:** The `name` field in `aipkg.json` is the package's identity. The source URL is just a locator.

**Rationale:** Like Composer and npm — the package declares what it's called. The GitHub org/repo can differ from the namespace (e.g., GitHub org `shiftbase-com` publishes under `@shiftbase`). This allows org renames, ugly slugs, and branding flexibility.

**Fallback behavior:** If a package is installed via GitHub shorthand (`aipkg install owner/repo`) and the manifest doesn't declare a namespace, the GitHub owner becomes the fallback namespace.

### 3. Install Scopes (Project + Global)

**Decision:** Two install scopes, mirroring npm/Composer behavior.

**Project install (default):**

- Packages stored in `.aipkg/` in the project root (gitignored, like `node_modules/`)
- Config in `aipkg.json` at the project root (committed to VCS)
- Teammates clone + `aipkg install` to reproduce the setup
- Adapters target project-level tool directories (`.claude/`, `.cursor/rules/` within the repo)

**Global install (`--global`):**

- Packages stored in `~/.aipkg/packages/`
- User-level config in `~/.aipkg/config.json`
- Available everywhere, not tied to a project
- Adapters target user-level tool directories (`~/.claude/skills/`, etc.)

**Rationale:** Different projects need different packages (and versions). The project config is the team contract. Global is for personal tools.

### 4. Adapters Are Thin (Rename + Symlink)

**Decision:** Adapters are minimal — they map artifact types to tool-specific paths and create symlinks. No transformation logic.

**Rationale:** Most AI artifacts are markdown files that work identically across tools. A Claude Code "skill" and a Cursor "rule" are essentially the same file in a different directory. The adapter's job is placement, not conversion.

**Example mapping:**

```
Claude Code adapter:  skill → ~/.claude/skills/{namespace}.{artifact}/
Cursor adapter:       skill → .cursor/rules/{namespace}.{artifact}.md
Windsurf adapter:     skill → .windsurf/rules/{namespace}.{artifact}.md
```

### 5. Dot-Notation at Tool Level

**Decision:** Installed artifacts use dot-notation: `namespace.artifact-name`. The `@` and `/` from the package name do not appear in filenames.

**Rationale:** Prevents collision between packages from different namespaces. Reads clearly in tool UIs (e.g., `/shiftbase.deploy-helper` as a slash command). The dot acts as a reserved namespace separator.

### 6. Packages Are Bundles

**Decision:** A package can contain multiple artifacts of different types. A single-artifact package is just a bundle of one.

**Rationale:** Real AI setups are compositional. A "Go expert" package might include an agent persona, two skills, and an MCP server config. Bundling lets authors curate coherent sets. One version, one release, atomic install.

**Install behavior (v1):** All-or-nothing. Cherry-picking individual artifacts from a bundle is a later feature.

### 7. GitHub-First, Registry-Later

**Decision:** v1 uses GitHub releases as the package source. A central registry comes in v3-4.

**Rationale:** Pushes registry complexity (namespace claiming, verification, hosting, auth) down the roadmap. Gets a working version out with infrastructure developers already have. Repository config with custom sources bridges the gap.

### 8. Multiple Source Types

**Decision:** Support pluggable source types beyond GitHub. v1 ships with `github` and `http`.

**Rationale:** Not all package authors are developers. An ops person who writes great skills shouldn't need Git. `http` as a v1 source type means anyone with a web server (or file hosting with URLs) can be a package source.

### 9. Composer-Style Repository Config with Canonical Scopes

**Decision:** Projects configure repositories in a config file, with priority ordering (first listed wins). Repositories can be marked `canonical` to lock a scope to a source.

**Rationale:**

- **Priority order** (Composer model) — predictable resolution
- **Canonical flag** — prevents dependency confusion attacks (a scope can only come from its designated source)
- **Scope routing** (npm model) — direct `@scope` to a specific repository

### 10. One File, Two Roles (`type` field)

**Decision:** Single file format (`aipkg.json`) with a `type` field that determines its role: `"project"` or `"package"`.

**Rationale:** Avoids needing two different file formats. The `type` field is self-documenting and enables strict validation per role. Familiar to Composer users (`type: "project"` vs `type: "library"`).

**Field availability by type:**

| Field | `project` | `package` |
|---|---|---|
| `type` | required | required |
| `name` | optional | required |
| `version` | ignored | required |
| `artifacts` | n/a | required |
| `require` | yes | yes |
| `repositories` | yes | yes (for dev) |

### 11. Exact-Version Dependencies from Day 1

**Decision:** Support inter-package dependencies from v1, but only with exact version pinning. No version ranges.

**Rationale:** This gives transitive dependency support (the useful part) without needing a constraint solver (the hard part). The install algorithm is a simple recursive fetch. Conflicts are trivially detectable: two packages need different exact versions of the same dep → error.

**What this defers:** Version ranges (`^1.0.0`, `~2.1`), constraint solving, lockfile-as-solver-output. These come when the ecosystem is mature enough to need them.

**What this enables now:**

- `require` works identically in both `project` and `package` manifests
- No behavioral difference between contexts → no confusion
- Packages can declare real dependencies that get installed transitively

### 12. Virtual Packages (`@virtual` Namespace)

**Decision:** Reserved `@virtual` namespace for community-maintained wrappers around upstream repos that don't natively support aipkg.

**Naming convention:** `@virtual/owner:repo` — the upstream source is encoded in the package name itself. The `:` separates owner from repo (since `/` is already used for namespace/package).

**How it works:**

- A "recipe" file (minimal `aipkg.json` with just `artifact_mapping`) lives in the `ai-interop/aipkg-virtual` community repo
- The CLI sees `@virtual/` → fetches the recipe, then fetches the upstream at the requested version
- Packages and installs locally like any normal package
- Version numbers map directly to upstream versions (tags/releases)

**Rationale:** Solves the cold start / bootstrapping problem. Existing popular repos (awesome-lists, skill collections) can be made available through aipkg without their authors lifting a finger. When an author adopts aipkg natively, the virtual package is deprecated in favor of the official one.

**In project config:**

```json
{
  "require": {
    "@virtual/some-author:spec-kit": "2.1.0"
  }
}
```

Teammates run `aipkg install` — the `@virtual/` prefix tells the CLI everything it needs to know. No special flags needed.

### 13. License Field (Optional, SPDX)

**Decision:** Optional `license` field in package manifests. Value is either an SPDX identifier (`MIT`, `Apache-2.0`, `GPL-3.0-only`, etc.) or `proprietary`.

**Rationale:** Important for ecosystem trust and compliance, but not required for v1 adoption. Follows npm and Composer conventions.

### 14. Semver Enforced

**Decision:** Package versions must be valid semver (`MAJOR.MINOR.PATCH`). No `v` prefix, no freeform strings. CLI resolves "latest" at install time and writes the exact version to the manifest.

**Rationale:** Every developer knows semver. Exact pinning in v1 means we don't need range resolution yet, but the format is ready for it when ranges come.

### 15. Archive Format (`.aipkg`)

**Decision:** Packages are zip archives with an `.aipkg` extension. `aipkg.json` must be at the archive root.

**Rationale:** Zip works natively on all platforms. The branded extension (like `.vsix`, `.phar`) makes packages instantly recognizable and enables file type associations. Internally it's just a zip — any tool can inspect it.

### 16. Namespace Validation (GitHub-Compatible)

**Decision:** Scope and package name rules mirror GitHub's constraints:

- **Scope:** lowercase alphanumeric + hyphens, no consecutive hyphens, can't start/end with hyphen, 1-39 chars
- **Package name:** lowercase alphanumeric + hyphens, same hyphen rules, 1-64 chars
- **No dots or underscores** — dots are reserved for tool-level dot-notation (`scope.artifact`)
- **Reserved scopes:** see list below

**Rationale:** GitHub-compatible naming means `@my-org/my-package` maps cleanly to GitHub without transformation. No dots avoids ambiguity with installed artifact naming. Scope claiming deferred until registry.

### 17. Reserved Scopes

**Decision:** The following scopes are reserved from day 1. They cannot be used in package manifests unless claimed by the rightful owner (via registry, once it exists, or by opening an issue on the spec repo).

**Project-owned (prefix-reserved):**

- `@aipkg*` — any scope starting with `aipkg` is reserved (e.g., `@aipkg`, `@aipkg-tools`, `@aipkg-contrib`)
- `@virtual` — community-maintained virtual package recipes
- `@ai-interop` — the organization itself

**Generic / ecosystem terms:**

- `@official`, `@core`, `@std`, `@stdlib`
- `@plugin`, `@plugins`, `@adapter`, `@adapters`
- `@test`, `@example`, `@demo`, `@internal`, `@private`
- `@ai`

**AI / LLM providers** (reserved for them to claim):

- `@anthropic`, `@claude`
- `@openai`, `@chatgpt`
- `@google`, `@alphabet`, `@gemini`, `@deepmind`, `@google-cloud`
- `@microsoft`, `@copilot`, `@github-copilot`, `@azure`
- `@meta`, `@llama`
- `@mistral`, `@mistralai`
- `@cohere`
- `@perplexity`
- `@xai`, `@grok`
- `@stability`, `@stabilityai`
- `@huggingface`
- `@aws`, `@amazon`, `@bedrock`
- `@nvidia`
- `@apple`
- `@samsung`
- `@alibaba`, `@qwen`
- `@baidu`

**AI coding tools** (reserved for them to claim):

- `@cursor`
- `@windsurf`, `@codeium`
- `@replit`
- `@sourcegraph`, `@cody`
- `@tabnine`
- `@continue`
- `@zed`

**Platforms** (reserved for them to claim):

- `@github`, `@gitlab`, `@bitbucket`
- `@vercel`, `@netlify`
- `@cloudflare`

**Rationale:** Protects well-known brands from squatting before the registry exists. Any reserved owner can claim their scope by opening an issue. The list can be extended over time. The CLI should warn (not block) in v1 since there's no enforcement authority yet — hard blocking comes with the registry.

## Deferred (Intentionally Not Decided)

- Lockfile design
- Selective install / cherry-picking syntax
- CLI language choice
- Adapter interface specification (config file vs code)
- Non-technical publishing workflows
- Version range syntax (when ranges are introduced)
- Namespace/scope claiming (requires registry)
