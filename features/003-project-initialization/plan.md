# Implementation Plan: Project Initialization & Model

**Branch**: `003-project-initialization` | **Date**: 2026-03-08 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/features/003-project-initialization/spec.md`

## Summary

Define the aipkg project model and implement `aipkg init`. The project file (`aipkg-project.json`) declares consumed dependencies; the install directory (`.aipkg/`) provides the layout where installed artifacts live. This feature creates the project file schema, the `init` command, and reference documentation. The install directory is specified but not created by this feature.

## Technical Context

**Language/Version**: Go (per go.mod, currently 1.25)
**Primary Dependencies**: cobra (CLI routing), santhosh-tekuri/jsonschema/v6 (schema validation), dlclark/regexp2 (PCRE regex for name patterns). No new dependencies.
**Storage**: Local filesystem only. `os.Stat` for existence checks, `os.WriteFile` for project file creation.
**Testing**: `go test` with `t.TempDir()` for filesystem isolation. No external service dependencies.
**Target Platform**: Cross-platform CLI (Linux, macOS, Windows)
**Project Type**: CLI
**Performance Goals**: N/A (single file write, sub-millisecond operation)
**Constraints**: No new dependencies. Reuse existing schema validation infrastructure.
**Scale/Scope**: Single command (`init`), one new JSON schema, one new internal package, one spec document.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-design evaluation

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Simplicity and Deferral | PASS | `init` creates one file. No directories, no prompts, no config. `.aipkg/` is specified but deferred to install. |
| II. Core/Adapter Separation | PASS | No adapter logic involved. The project model is core infrastructure. |
| III. Convention Over Invention | PASS | Project file follows npm/Composer patterns (dependency map, lockfile-adjacent concept). `@scope/package-name` naming. Scoped naming in `.aipkg/` uses dot-notation convention from spec. |
| IV. Cold Start First | PASS | `aipkg init` requires zero configuration. Single command, zero arguments. |
| V. Backward-Compatible Evolution | PASS | `specVersion` integer enables future schema migration. No existing behavior modified. |
| Schema Authority | PASS | JSON Schema in `spec/schema/` is the validation source of truth. CLI implements, does not invent. |
| User-facing Documentation | PASS | FR-019 requires `spec/project.md` as a deliverable. Docs ship with the feature. |
| Testability | PASS | Filesystem operations via `os` package. `t.TempDir()` for test isolation. |
| Error Handling | PASS | All errors wrapped with context via `fmt.Errorf` + `%w`. |
| Import Discipline | PASS | `internal/project` depends on `internal/schema`. No circular imports. No adapter imports. |

All gates pass. No violations to justify.

## Project Structure

### Documentation (this feature)

```text
features/003-project-initialization/
├── plan.md              # This file
├── research.md          # Scoped naming decision, semver pattern, schema architecture
├── data-model.md        # Project file, install directory, installed artifact entities
├── quickstart.md        # Developer quickstart for implementation
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
spec/
├── schema/
│   └── project.json         # NEW: JSON Schema for aipkg-project.json (FR-005)
├── project.md               # NEW: Reference docs for project model (FR-019)
└── naming.md                # MODIFIED: Add three-segment install directory naming (DD-001)

internal/
├── project/
│   ├── project.go           # NEW: ProjectFile type, Create(), LoadFile()
│   └── project_test.go      # NEW: Unit tests for project operations
├── schema/
│   ├── validate.go          # MODIFIED: Add ValidateProject() + project schema compilation
│   └── validate_test.go     # MODIFIED: Add project schema validation tests
└── cli/
    ├── root.go              # MODIFIED: Register init command
    ├── init.go              # NEW: newInitCmd() + runInit() command handler
    └── init_test.go         # NEW: Integration tests for init command

specdata.go                  # MODIFIED: Add ProjectSchemaJSON embed
```

**Structure Decision**: Follows the existing pattern. New `internal/project` package mirrors `internal/manifest` (types + file I/O for a specific JSON file). Command handler in `internal/cli/init.go` follows the `create.go` and `pack.go` pattern (factory function + isolated `runX()` for testability).

## Design Decisions

### DD-001: Three-segment scoped naming for `.aipkg/`

Installed artifacts use `scope.package-name.artifact-name` in the `.aipkg/` directory. This satisfies FR-012 (collision prevention) and FR-013 (traceability). The existing two-segment dot-notation (`scope.artifact`) fails both requirements when the same scope has multiple packages with identically named artifacts.

Full analysis in [research.md](research.md#r-001-scoped-artifact-naming-format-for-aipkg).

### DD-002: SemVer pre-release without build metadata

The project file schema accepts semver versions with optional pre-release identifiers but excludes build metadata. This extends the package manifest's strict `MAJOR.MINOR.PATCH` pattern per FR-004 while keeping version comparison simple (build metadata has no ordering semantics).

Full analysis in [research.md](research.md#r-002-semver-pre-release-version-pattern).

### DD-003: Extend `internal/schema` for project validation

Add `ValidateProject()` to the existing schema package rather than creating separate validation in `internal/project`. This reuses the regexp2 engine setup and compiler infrastructure. No per-field validation bridge needed for this feature (no interactive prompts).

Full analysis in [research.md](research.md#r-003-schema-validation-architecture).

### DD-004: No `--force` flag for init

The spec (FR-016) requires refusing to overwrite an existing project file. No `--force` override is provided. This follows the constitution's simplicity principle: the feature is either fully present or fully absent. If users need to re-initialize, they can delete the file manually. Adding `--force` later is cheap if demand materializes.

### DD-005: Error messages and success output follow existing CLI patterns

Error messages use the same style as `create.go` and `pack.go`: lowercase, no period, wrapped with context. The mutual exclusivity error (FR-017) includes actionable suggestions (`aipkg require` / `aipkg install`), matching the spec's acceptance criteria.

Success output follows the `create` command pattern. On successful initialization, `init` prints a confirmation message (e.g., `Initialized project in aipkg-project.json`).

### DD-006: `LoadFile()` included for test verification and future use

`internal/project` ships with both `Create()` and `LoadFile()`. No FR requires loading project files in this feature, but `LoadFile()` is needed for test assertions (roundtrip verification: create a project file, load it back, assert structure) and is the natural counterpart that future commands (`require`, `install`) will depend on. Including it now avoids a test-only utility that gets promoted later.
