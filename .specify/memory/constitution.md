<!--
Sync Impact Report
==================
Version change: (new) → 1.0.0
Modified principles: N/A (initial constitution)
Added sections:
  - Core Principles (5): Simplicity and Deferral, Core/Adapter Separation,
    Convention Over Invention, Cold Start First, Backward-Compatible Evolution
  - Development Standards
  - Architectural Boundaries
  - Governance
Removed sections: N/A
Templates requiring updates:
  - .specify/templates/plan-template.md — ✅ no update needed
    (Constitution Check section is already a dynamic placeholder)
  - .specify/templates/spec-template.md — ✅ no update needed
    (no constitution-specific references)
  - .specify/templates/tasks-template.md — ✅ no update needed
    (task phases are generic, no principle-specific coupling)
  - .specify/templates/commands/*.md — ✅ no files present
Follow-up TODOs: none
-->

# aipkg CLI Constitution

## Core Principles

Principles are listed in priority order. When principles conflict,
higher-numbered principles yield to lower-numbered ones. See
**Conflict Resolution** in the Governance section for guidance.

### I. Simplicity and Deferral

Every feature decision MUST pass the question: *"What is the smallest
useful thing?"* If a capability is not required for v1 to be useful,
it is deferred — not simplified, not stubbed, but absent.

- The install algorithm is a simple recursive fetch. No constraint
  solver, no version ranges, no resolution strategy beyond exact pins.
- Features are either fully implemented or fully absent. No partial
  implementations, feature flags, or "coming soon" stubs in shipping
  code.
- Complexity MUST be justified against a concrete user need, not a
  hypothetical future requirement. YAGNI is the default posture.
- When in doubt, defer. Adding a feature later is cheap; removing one
  is expensive.

**Rationale**: The ecosystem has zero users today. A small, correct,
understandable tool that ships is worth more than a comprehensive one
that doesn't. Deferred features become v2/v3 layers (see Principle V).

**Testable gate**: Can a new user install a package with a single
command and no prior configuration?

### II. Core/Adapter Separation

The CLI has a core pipeline (resolve → fetch → unpack → store) that
MUST know nothing about specific AI tools. Tool-specific placement
logic lives exclusively in adapters.

- Core packages MUST NOT import adapter packages. The dependency
  direction is one-way: adapters depend on core, never the reverse.
- Adapters MUST implement a defined interface. Adding support for a
  new AI tool MUST NOT require changes to core packages.
- The core pipeline operates on abstract artifact types and storage
  locations. It never references Claude Code, Cursor, Windsurf, or
  any other tool by name.
- This boundary is structural and enforced by Go's package import
  graph — not by convention or code review alone.

**Rationale**: The value proposition of aipkg is tool-agnostic package
management. If the core becomes entangled with specific tools, the
project loses its reason to exist. This separation also enables
community-contributed adapters without touching core code.

**Testable gate**: Can the entire core pipeline build and run without
importing or referencing any adapter package?

### III. Convention Over Invention

aipkg follows established package manager conventions. npm, Composer,
Go modules, and Cargo are the reference points. When a well-known
ecosystem has solved a problem, aipkg MUST adopt that pattern unless
there is a documented reason not to.

- Naming uses `@scope/package-name` — always scoped, no exceptions.
- Identity is manifest-authoritative: the package name comes from
  `aipkg.json`, not from the source URL or repository name.
- Install scopes follow the project/global split: `.aipkg/` (project,
  gitignored) and `~/.aipkg/` (global).
- Versioning follows semver. Dependency pinning uses exact versions in
  v1.
- Familiar is better than clever. If a user who knows npm or Composer
  can predict aipkg's behavior, the design is correct.

**Rationale**: Unfamiliar patterns impose learning costs that kill
adoption. The AI tooling audience includes developers who already have
package manager muscle memory — aipkg MUST feel like home.

**Testable gate**: Would an npm or Composer user recognize the workflow
without explanation?

### IV. Cold Start First

Every design decision MUST be evaluated against: *"Does this help or
hurt a user installing their first package?"* The ecosystem needs
content before it has users, and users before it has content.

- GitHub-first distribution: v1 sources packages from GitHub Releases.
  No registry signup, no publishing workflow, no approval process.
- Virtual packages (`@virtual/owner:repo`) let the community wrap
  existing repos as installable packages without upstream cooperation.
- Zero-config defaults: `aipkg install @scope/name` MUST work without
  prior setup, config files, or auth tokens for public packages.
- The happy path MUST require the fewest possible steps. Every
  additional step loses users.

**Rationale**: The cold start problem is existential. If the first
experience is complicated, there won't be a second. Virtual packages
and GitHub-first are strategic — they seed the ecosystem before a
registry exists.

**Testable gate**: Can a user go from "never heard of aipkg" to
"installed and using a package" in under 5 minutes?

### V. Backward-Compatible Evolution

The version roadmap is an additive layer cake: v1 adds GitHub + exact
pins, v2 adds more source types, v3+ adds the registry. Packages
created for v1 MUST remain valid and installable forever.

- New manifest fields are always optional. Existing `aipkg.json` files
  MUST NOT break when the schema evolves.
- New source types and adapter types are additive. Existing sources
  and adapters MUST continue to work unchanged.
- Breaking changes to the manifest schema or CLI behavior require a
  major version bump with a documented migration path.
- The specification (`spec/`) defines the schema evolution rules.
  The CLI implements them but does not invent them.

**Rationale**: Trust is earned slowly and lost instantly. A package
manager that breaks existing packages on upgrade will not be trusted.
The additive model also lets the project ship v1 fast without painting
itself into a corner.

**Testable gate**: Does this feature work without modifying any
existing v1 behavior?

## Development Standards

These standards are derived from the repo's existing tooling
configuration, not invented independently.

**Language and structure**:

- Go is the implementation language. All application code lives in
  `internal/` (no `pkg/` directory — this is a CLI, not a library).
- Entry point is `cmd/aipkg/`. Binary name is `aipkg`.

**Task runner**: All development workflows use
[Taskfile](https://taskfile.dev). The canonical commands are:

| Command | Purpose |
|---------|---------|
| `task build` | Build binary to `dist/` |
| `task test` | Run tests |
| `task lint` | Run golangci-lint |
| `task check` | Full check: lint + vet + test |
| `task fmt` | Format code (gofmt) |
| `task tidy` | go mod tidy |

**Quality enforcement**: golangci-lint for static analysis, pre-commit
hooks for file hygiene and commit message validation. CI MUST pass
linting and tests before merge. See `.golangci.yml` and
`.pre-commit-config.yaml` for current configuration.

**Commits**: Conventional commits, enforced by pre-commit hook. DCO
sign-off required (`git commit -s`) for CNCF compliance. Linear
issues linked via `Closes: AIPKG-XX` trailer.

**CI**: GitHub Actions runs lint, test (with `-race` and coverage),
and build on push/PR to main. Go version is pinned via `go.mod`.

**Commits**: Conventional commits, enforced by pre-commit hook. DCO
sign-off required (`git commit -s`) for CNCF compliance. Linear
issues linked via `Closes: AIPKG-XX` trailer.

**Releases**: goreleaser handles cross-platform binary builds and
distribution.

**License**: Apache-2.0. All contributions MUST be license-compatible.

## Architectural Boundaries

These are structural constraints on the codebase, not aspirational
guidelines.

**Pluggable interfaces**:

- **Source types**: The mechanism for resolving and fetching packages
  (GitHub Releases in v1, more later). New source types MUST implement
  a defined interface and MUST NOT require changes to existing sources
  or the core pipeline.
- **Adapters**: The mechanism for placing artifacts into tool-specific
  locations. New adapters MUST implement a defined interface and MUST
  NOT require changes to existing adapters or the core pipeline.

**Schema validation**: Manifest validation uses JSON Schema files from
`spec/schema/`. The CLI MUST NOT invent its own validation rules. The
specification is the single source of truth for what constitutes a
valid manifest.

**Error handling**: Functions that can fail return `error` as the last
return value. Errors are wrapped with context using `fmt.Errorf` with
`%w`. Panics are reserved for truly unrecoverable programmer errors,
never for runtime conditions.

**Testability**: Packages MUST be testable in isolation. External
dependencies (filesystem, network, registries) MUST be injectable
via interfaces, not hardcoded. Tests run via `go test ./...` with
no external service dependencies.

**Import discipline**: The `internal/` package tree enforces a
one-way dependency graph. Core packages MUST NOT import adapter or
tool-specific packages. Circular imports are a build failure, not a
code review finding.

## Governance

**Authority hierarchy**:

1. The specification (`spec/`) is the authority for manifest schema,
   naming rules, and artifact type definitions. The CLI implements the
   spec. It does not extend or contradict it.
2. Brainstorm docs in `research/` are the source of truth for settled
   design decisions.
3. This constitution governs CLI-specific development principles,
   standards, and architectural boundaries.
4. Linear (team AIPKG, project "CLI") is the work tracker.

**Principle priority and conflict resolution**: Principles are
ordered I through V by priority. When two principles conflict:

- Apply the higher-priority principle.
- Document the conflict and the resolution in the relevant Linear
  issue or PR description.
- Common tensions and their resolution:
  - *Simplicity (I) vs Convention (III)*: If a convention adds
    complexity without clear v1 value, simplicity wins — defer the
    convention to a later version.
  - *Simplicity (I) vs Cold Start (IV)*: If a simplification makes
    first-time adoption harder, cold start wins — the purpose of
    simplicity is adoption, not minimalism for its own sake.
  - *Core/Adapter Separation (II) vs Cold Start (IV)*: If a shortcut
    would entangle core with a specific tool to ship faster, separation
    wins — the shortcut creates tech debt that compounds.
  - *Spec authority vs CLI implementation*: The spec defines *what*.
    The CLI decides *how*, within the spec's constraints. If the CLI
    needs a spec change, it is proposed upstream — not worked around.

**Key tensions** (acknowledged, not resolved — these are ongoing):

- Simplicity vs completeness: v1 deliberately omits features that
  users will request. The answer is "not yet", not "no".
- Developer audience vs non-technical users: the CLI targets
  developers first, but design choices should not preclude simpler
  UX layers later.
- Spec as upstream vs CLI as implementer: the CLI may discover spec
  gaps during implementation. The fix is a spec PR, not a CLI hack.

**Amendments**: Changes to this constitution require:

1. A Linear issue describing the proposed change and rationale.
2. An update to this file with version bump per semver:
   - MAJOR: principle removal or backward-incompatible redefinition.
   - MINOR: new principle, new section, or material expansion.
   - PATCH: wording clarification, typo fix, non-semantic refinement.
3. PR review and approval.

**Compliance**: All PRs and code reviews SHOULD verify alignment with
these principles. The constitution is a living document — if a
principle consistently creates friction, that is signal to amend it,
not ignore it.

**Version**: 1.0.0 | **Ratified**: 2026-02-27 | **Last Amended**: 2026-02-27
