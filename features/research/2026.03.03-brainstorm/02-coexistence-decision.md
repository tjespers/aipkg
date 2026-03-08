# Package + Project Coexistence: Can They Share a Directory?

## The Question

Can `aipkg.json` (package manifest) and `aipkg-project.json` (project file) live in the same directory? If so, under what conditions? If not, what guardrails do we need?

## Why This Matters

The current spec (003) says they can coexist (FR-002) and adds a confirmation prompt when `aipkg init` encounters an existing `aipkg.json` (FR-017). But before we commit to this, we need to think through what it actually means for the CLI to operate in a directory that has both files.

## The Two Files

| | aipkg.json | aipkg-project.json |
|---|---|---|
| **Role** | Package identity + authored artifacts | Consumed dependencies |
| **Created by** | `aipkg create` | `aipkg init` |
| **Identity fields** | Yes (name, version, license) | No |
| **`require` field** | Bundled deps (resolved at pack time) | Installed deps (resolved at install time) |
| **Install dir** | N/A (packaged into .aipkg archive) | `.aipkg/` directory |
| **`specVersion`** | Integer, `1` | Integer, `1` |

## Scenarios to Think Through

### 1. "Package that consumes other packages"

A package author wants their package to depend on other packages at development time. For example, a skill that builds on another skill during authoring. The `aipkg.json` already has a `require` field for bundled dependencies. Does the author also need a project-level `require`?

### 2. "CLI ambiguity"

When both files exist, how does the CLI know which context to operate in? `aipkg require @scope/pkg` could mean:
- Add a bundled dependency to the package manifest (author intent)
- Add an installed dependency to the project file (consumer intent)

### 3. "Install directory inside a package"

If a project file exists alongside a package manifest, `.aipkg/` would be created in the package root. When `aipkg pack` runs, should it include `.aipkg/`? Probably not (and `.aipkgignore` could handle this), but it's a footgun.

### 4. "Confusion risk"

The user's original feedback: confusing `init` with `create` is less likely than confusing `init` with `install`/`require`. The real confusion vector is a developer who wants to install packages in their project but accidentally runs `init` in a package directory they're developing.

### 5. "Ecosystem parallels"

How do other package managers handle this?
- **npm**: `package.json` serves both roles (author + consumer). No separate project file.
- **Cargo**: `Cargo.toml` is both the package definition and the dependency declaration.
- **Go modules**: `go.mod` is both module identity and dependency declaration.
- **Composer**: `composer.json` is both package identity and dependency declaration.

Most ecosystems use a single file. aipkg has a deliberate split because the package manifest and project file serve fundamentally different personas (author vs. consumer). But this split creates the coexistence question.

## Open Questions

1. Is there a real use case for a directory being both a package AND a project?
2. If yes, how does the CLI disambiguate commands that apply to both contexts?
3. If no, should the CLI hard-block `init` when `aipkg.json` exists (and vice versa)?
4. What about monorepo patterns where a package lives inside a larger project?
5. Should `.aipkg/` (install dir) be excluded from `aipkg pack` by default?

---

## Decision 71: No Coexistence

**Context:** The spec (003) initially allowed `aipkg.json` and `aipkg-project.json` to coexist with a confirmation prompt. Revisiting this before committing to the design.

**Decision**: `aipkg.json` and `aipkg-project.json` MUST NOT coexist in the same directory.

**Rationale**: The dual-file split (package vs. project) is justified and clean. But allowing both files in the same directory solves no real problem while creating several:

- **No plausible use case.** The package manifest's `require` field already covers the "package depends on other packages" scenario. Development tool installation is a project-level concern that belongs at the repo root, not the package root.
- **CLI ambiguity.** Commands like `aipkg require` would need to disambiguate which file to target. Unnecessary complexity.
- **Directory pollution.** `.aipkg/` (install directory) sitting alongside authored artifacts is a footgun for `aipkg pack`.
- **Confusing UX.** Two `require` fields in two files in the same directory, meaning different things.

Being the odd one out with a dual-file model is fine (justified by independent schema evolution). Being the even odder one out by also allowing coexistence is complexity for no gain.

**Impact on spec**: US3 changes from a confirmation prompt to a hard block. FR-002 drops the coexistence allowance. Related items (FR-017, AS6, SC-003) updated accordingly.
