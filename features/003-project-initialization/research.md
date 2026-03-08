# Research: Project Initialization & Model

**Feature Branch**: `003-project-initialization`
**Date**: 2026-03-08

## R-001: Scoped artifact naming format for `.aipkg/`

**Context**: FR-012 requires installed artifacts to use a scoped naming convention incorporating source package identity. FR-013 requires traceability from any installed artifact back to its source package. FR-014 defers the exact format to planning, starting from the dot-notation in `spec/naming.md`.

**Problem**: The current dot-notation is `scope.artifact-name`. This format cannot satisfy FR-013 because it loses the package name. Two packages from the same scope with artifacts sharing a name would collide:

```text
@alice/pkg-a  artifact "test-writer" → alice.test-writer
@alice/pkg-b  artifact "test-writer" → alice.test-writer  ← collision
```

Without the package name in the path, you also cannot trace an artifact back to its source package (only to the scope). FR-013 explicitly requires package-level traceability for removal and update operations.

**Decision**: Use three-segment dot-notation: `scope.package-name.artifact-name`

```text
@alice/pkg-a  artifact "test-writer" → alice.pkg-a.test-writer
@alice/pkg-b  artifact "test-writer" → alice.pkg-b.test-writer
```

This format:
- Eliminates same-scope collisions entirely
- Provides full traceability: given `alice.pkg-a.test-writer`, you can reconstruct `@alice/pkg-a`
- Is parseable: split on `.`, first segment is scope, last segment is artifact name, middle segment(s) are the package name. Since package names cannot contain dots (enforced by naming rules), parsing is unambiguous.
- Extends the existing dot-notation naturally (adds one segment)

**Alternatives considered**:

1. **`scope.artifact-name`** (current spec/naming.md) — Rejected. Fails FR-013 traceability requirement and collides when same scope has duplicate artifact names.
2. **Path-based: `scope/package-name/artifact-name`** — Rejected. Introduces nested directories per package inside type directories, complicating the flat categorized layout defined in FR-007. Also diverges from the established dot-notation convention.
3. **`package-name.artifact-name`** (drop scope) — Rejected. Package names are only unique within a scope. Two scopes could have packages with the same name.

**Compatibility note**: The three-segment format works regardless of how tools consume `.aipkg/`. If tools adopt the directory directly (reading skills from `.aipkg/skills/`, commands from `.aipkg/commands/`), the names are user-facing and parseable. Claude Code's current plugin convention (`/<plugin>:<skill-or-command>`) carries similar information density. If adapters translate to tool-specific locations instead, the names are equally valid as source identifiers. The format is self-sufficient; it does not depend on an adapter layer to be correct. Dots in filenames and directory names are universally supported across filesystems.

**Impact on spec/naming.md**: The existing dot-notation section describes `scope.artifact-name` as the adapter convention. The three-segment format applies to the `.aipkg/` install directory specifically. `spec/naming.md` should be updated to document both: the install directory uses `scope.package.artifact`, while adapters may use a shorter form if their target tool has its own namespacing. The reference documentation (FR-019, `spec/project.md`) will cover the install directory convention.

## R-002: SemVer pre-release version pattern

**Context**: FR-004 specifies that `require` values accept semver versions "with optional pre-release identifiers per the SemVer spec." The package manifest schema uses strict `MAJOR.MINOR.PATCH` only (`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$`). The project file schema needs a wider pattern.

**Decision**: Use the official SemVer 2.0.0 regex for MAJOR.MINOR.PATCH with optional pre-release, excluding build metadata:

```text
^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?$
```

This is the semver.org reference regex with the build metadata group (`+build`) removed.

**Rationale**: Pre-release versions are useful for projects that want to pin to beta or RC releases of packages. Build metadata carries no semantic meaning in SemVer and would complicate exact-match pinning, so it is excluded. The regex is taken directly from the SemVer specification to avoid inventing custom validation.

**Examples of valid versions**:
- `1.0.0` (plain release)
- `1.0.0-beta.1` (numeric pre-release)
- `1.0.0-alpha` (named pre-release)
- `1.0.0-0.3.7` (numeric-only identifiers)
- `1.0.0-rc.1` (release candidate)

**Examples of invalid versions**:
- `1.0.0+build.123` (build metadata not allowed)
- `1.0.0-beta.1+build.456` (build metadata not allowed)
- `1.0` (missing patch)
- `v1.0.0` (no prefix)

**Alternatives considered**:

1. **Same strict pattern as package manifest** — Rejected. Contradicts FR-004 which explicitly requires pre-release support.
2. **Full SemVer including build metadata** — Rejected. Build metadata has no ordering semantics in SemVer. Including it in exact pins would create confusion (is `1.0.0+build.1` the same as `1.0.0+build.2`?). The constitution's simplicity principle (I) supports excluding it.

## R-003: Schema validation architecture

**Context**: The existing `internal/schema` package compiles and validates against the package manifest schema (`spec/schema/aipkg.json`). The project file needs its own schema and validation.

**Decision**: Extend `internal/schema` with a `ValidateProject()` function alongside the existing `Validate()`. Add a second `sync.Once` + compiled schema pair for the project schema. Reuse the existing regexp2 engine setup.

**Rationale**: The `internal/schema` package is the natural home for all JSON Schema validation. Adding a second schema is a small, additive change. Creating a separate validation package would duplicate the regexp2 adapter and compiler setup for no benefit.

The `init` command does not need runtime validation (it creates a known-good file). Schema validation is needed for:
1. Future commands (`require`, `install`) that load existing project files
2. Test coverage (verifying the schema accepts valid and rejects invalid project files, per SC-005)

No `ValidateProjectField()` bridge is needed for this feature. The `init` command has no interactive prompts requiring per-field validation.

**Alternatives considered**:

1. **Separate validation in `internal/project`** — Rejected. Duplicates regexp2 adapter code and compiler setup.
2. **Schema-agnostic refactor of `internal/schema`** — Rejected. Over-engineers for two schemas. If a third schema appears, refactoring then is cheap.
