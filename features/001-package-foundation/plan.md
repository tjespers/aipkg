# Implementation Plan: Package Foundation

**Branch**: `001-package-foundation` | **Date**: 2026-03-01 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/features/001-package-foundation/spec.md`

## Summary

Define the package manifest schema and implement the `aipkg create` command. This is the first CLI feature: it reworks the existing unified manifest schema into a package-only format (dropping the `type` discriminator, adding `specVersion`), revises the spec reference docs to document convention-based directory layout, and builds the interactive scaffolding command that creates valid package directories.

The technical approach uses Go's `go:embed` to bake the JSON Schema and reserved-scopes list into the binary, a schema bridge that extracts per-field validators for use in both interactive (huh) and non-interactive (flag) validation paths, and table-driven tests throughout.

## Technical Context

**Language/Version**: Go 1.25 (per go.mod)
**Primary Dependencies**: cobra (CLI routing), huh/bubbletea (interactive prompts), santhosh-tekuri/jsonschema/v6 (schema validation), google/licensecheck (LICENSE file detection), dlclark/regexp2 (PCRE-style regex for name validation lookaheads), x/term (TTY detection)
**Storage**: N/A (local filesystem only, no database)
**Testing**: `go test` with table-driven tests, `t.TempDir()` for filesystem isolation, golden file tests against `testdata/` fixtures, `go test -race -coverprofile` in CI
**Target Platform**: Cross-platform (linux/darwin/windows, amd64/arm64)
**Project Type**: CLI
**Performance Goals**: N/A (one-shot local command, not performance-sensitive)
**Constraints**: No network access, no external service dependencies, all validation offline, Apache-2.0 license
**Scale/Scope**: Single command (`create`), six internal packages, one JSON Schema, spec doc revisions

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Simplicity and Deferral

- **PASS**: The create command is the smallest useful thing for package authoring. No install, no registry, no resolver. Just scaffolding.
- **PASS**: No partial implementations. The spec schema, validation, and create command are each fully implemented or absent.
- **WATCH**: The schema bridge (extracting per-field sub-schemas) adds complexity. Justified by FR-029: the schema MUST be the single source of truth. Without the bridge, we'd duplicate validation logic.

### II. Core/Adapter Separation

- **PASS**: This feature has no adapter code. All packages are core. No tool-specific references anywhere.
- **PASS**: The manifest, schema, naming, and scaffold packages are all tool-agnostic.

### III. Convention Over Invention

- **PASS**: `aipkg create` mirrors `npm init`, `helm create`, `cargo init`. Scoped naming follows npm convention.
- **PASS**: Manifest fields (name, version, description, license) are standard package manager fare.
- **PASS**: Interactive prompts with flag overrides follow the cobra/helm pattern.

### IV. Cold Start First

- **PASS**: A single command creates a valid package. Zero prior setup.
- **PASS**: Default version `0.1.0` lets authors start immediately without thinking about versioning.

### V. Backward-Compatible Evolution

- **PASS**: The `specVersion` integer field enables future schema migration without breaking existing manifests.
- **PASS**: The package-only schema is a clean break from the unified schema (prior art, not shipped to users yet), so no backward compatibility concern.

### Documentation Standard (v1.1.0)

- **REQUIRED**: The create command MUST ship with a reference page (synopsis, flags, examples, workflow).
- **REQUIRED**: Spec doc revisions (manifest.md, artifacts.md, naming.md) MUST be updated alongside the schema changes.

### Gate Result: PASS

No violations. Schema bridge complexity is justified by FR-029. Documentation deliverables noted.

## Project Structure

### Documentation (this feature)

```text
features/001-package-foundation/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (CLI contract)
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
cmd/aipkg/
└── main.go                  # Entry point, wires cobra root command

internal/
├── cli/
│   ├── root.go              # Root command, version/help flags
│   └── create.go            # Create command: flags, prompt orchestration, execution
├── manifest/
│   ├── manifest.go          # PackageManifest struct, JSON marshal/write
│   └── testdata/            # Golden file fixtures for manifest output
├── schema/
│   ├── embed.go             # go:embed for spec/schema/aipkg.json
│   ├── validate.go          # Full-document schema validation
│   └── bridge.go            # Per-field sub-schema extraction, func(string) error bridge for huh
├── naming/
│   ├── name.go              # Parse and validate scoped package names
│   └── reserved.go          # Reserved scope checking (embedded reserved-scopes.txt)
├── scaffold/
│   └── scaffold.go          # Directory creation, well-known dirs, non-destructive writes
└── license/
    └── detect.go            # LICENSE file detection via licensecheck, SPDX mapping

spec/
├── schema/
│   └── aipkg.json           # Package manifest JSON Schema
├── manifest.md              # Revised: package-only manifest reference
├── artifacts.md             # Revised: convention-based directory layout, structural requirements
├── naming.md                # Revised: updated examples, package-only context
└── reserved-scopes.txt      # Unchanged
```

**Structure Decision**: Standard Go CLI layout. All application code in `internal/` with one package per concern. Tests colocated with source (`_test.go` files in each package). No separate `tests/` directory; Go convention is same-package tests. The `testdata/` directories hold golden file fixtures where needed.

## Post-Design Constitution Re-Check

*Re-evaluated after Phase 1 design completion.*

### I. Simplicity and Deferral

- **PASS (confirmed)**: Six internal packages is the minimum for clean separation. Each has a single responsibility. No unnecessary abstractions.
- **PASS (confirmed)**: Schema bridge complexity validated by research. The jsonschema/v6 library supports property sub-schema validation natively (`root.Properties["name"].Validate(val)`). No custom extraction logic needed; the library does the work.
- **WATCH resolved**: The bridge is simpler than initially expected. It's a thin wrapper, not a complex extraction layer.

### II. Core/Adapter Separation

- **PASS (confirmed)**: No adapter code in the design. The `specdata` embed package, `manifest`, `schema`, `naming`, `scaffold`, and `license` packages are all core. No tool-specific imports.

### III. Convention Over Invention

- **PASS (confirmed)**: CLI contract follows standard cobra patterns. Flag names match npm/helm conventions. The interactive-to-flag fallback pattern is well-established.

### IV. Cold Start First

- **PASS (confirmed)**: License detection from existing LICENSE files reduces friction further. Quickstart validates the zero-to-package flow.

### V. Backward-Compatible Evolution

- **PASS (confirmed)**: `specVersion: 1` enables future schema migration. The package-only schema is a clean start (no existing users of the old unified schema).

### Documentation Standard (v1.1.0)

- **PASS**: Spec doc revisions (manifest.md, artifacts.md, naming.md) are in the source code structure, delivered alongside the implementation. CLI reference documentation for the create command is listed as a deliverable.

### Gate Result: PASS (no changes from pre-research check)

## Complexity Tracking

No constitution violations to justify.
