# Implementation Plan: Archive Format & Pack Command

**Branch**: `002-archive-format-pack` | **Date**: 2026-03-01 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/features/002-archive-format-pack/spec.md`

## Summary

Define the `.aipkg` archive format specification and implement the `aipkg pack` command. The pack command reads an existing package directory, discovers artifacts from well-known directories, validates each artifact against type-specific rules (SKILL.md frontmatter, JSON parse, non-empty check), generates the `artifacts` array, produces a zip archive with a SHA-256 sidecar, and writes a new `spec/archive.md` reference document.

The technical approach builds on 001's patterns: same Go stack, same internal package layout, same testing conventions. New dependencies are `gopkg.in/yaml.v3` (frontmatter parsing) and `github.com/sabhiram/go-gitignore` (ignore file patterns). Multi-error collection uses stdlib `errors.Join`. Archive creation and SHA-256 are stdlib only.

## Technical Context

**Language/Version**: Go 1.25 (per go.mod)
**Primary Dependencies**: cobra (CLI routing), gopkg.in/yaml.v3 (SKILL.md frontmatter), sabhiram/go-gitignore (.aipkgignore patterns), santhosh-tekuri/jsonschema/v6 (manifest schema validation). Archive and SHA-256 via stdlib (`archive/zip`, `crypto/sha256`).
**Storage**: N/A (local filesystem only, no database)
**Testing**: `go test` with table-driven tests, `t.TempDir()` for filesystem isolation, golden file tests against `testdata/` fixtures, `go test -race -coverprofile` in CI
**Target Platform**: Cross-platform (linux/darwin/windows, amd64/arm64)
**Project Type**: CLI
**Performance Goals**: N/A (one-shot local command, not performance-sensitive)
**Constraints**: No network access, no external service dependencies, all validation offline, Apache-2.0 license
**Scale/Scope**: Single command (`pack`), four new internal packages, two new dependencies, spec reference doc, spec doc updates

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Simplicity and Deferral

- **PASS**: The pack command is the smallest useful step toward distributable packages. No install, no registry, no unpack. Just archive creation.
- **PASS**: No partial implementations. Every well-known directory type has complete validation. No "TODO" or stub validators.
- **WATCH**: `ValidationErrors` collector introduces a new pattern (multi-error) not used in 001. Justified by FR-020 which requires reporting all errors. The implementation is ~15 lines using stdlib `errors.Join`.
- **WATCH**: `internal/frontmatter/` is a new package for ~25 lines of code. Justified because both skills (required now) and commands (optional frontmatter per spec) will use it. A shared package avoids duplication.

### II. Core/Adapter Separation

- **PASS**: This feature has no adapter code. All packages are core. No tool-specific references.
- **PASS**: artifact, archive, ignore, and frontmatter packages are all tool-agnostic.

### III. Convention Over Invention

- **PASS**: `aipkg pack` mirrors `npm pack` and `helm package`. Archive naming follows established conventions. The `--output` flag follows the same pattern as `helm package -d`.
- **PASS**: The `sha256sum`-compatible sidecar format means standard CLI tools can verify integrity without aipkg installed.
- **PASS**: `.aipkgignore` follows `.gitignore` syntax. No custom pattern language.
- **PASS**: Archive filename uses `--` separator for scope/name boundary, which is unambiguous because naming rules forbid consecutive hyphens. Top-level directory matches the package name, following Helm's convention.

### IV. Cold Start First

- **PASS**: A single command produces a distributable archive from any valid package directory. Zero configuration needed.
- **PASS**: Built-in defaults handle common exclusions (`.git/`, `.aipkgignore`). Authors don't need to create an ignore file for basic use.

### V. Backward-Compatible Evolution

- **PASS**: The `Artifacts` field on `PackageManifest` uses `omitempty`, so manifests created by `aipkg create` (which have no artifacts) are unaffected.
- **PASS**: The archive format is additive. The spec defines it as a new document (`spec/archive.md`). No existing behavior changes.
- **PASS**: The pack command only reads the manifest; it never modifies the original `aipkg.json` on disk.

### Documentation Standard (v1.1.0)

- **REQUIRED**: The pack command MUST ship with a reference page (synopsis, flags, examples, workflow).
- **REQUIRED**: New `spec/archive.md` reference doc for the archive format specification.
- **REQUIRED**: Updates to `spec/artifacts.md` for relaxed file extensions and compound extension name derivation.

### Gate Result: PASS

No violations. Multi-error collector and frontmatter package complexities are justified by spec requirements. Documentation deliverables noted.

## Project Structure

### Documentation (this feature)

```text
features/002-archive-format-pack/
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
└── main.go                  # Entry point (unchanged)

internal/
├── cli/
│   ├── root.go              # Root command (add pack subcommand)
│   ├── create.go            # Unchanged
│   └── pack.go              # Pack command: flags, orchestration, execution
├── manifest/
│   └── manifest.go          # Add Artifact type, Artifacts field, LoadFile()
├── schema/
│   ├── embed.go             # Unchanged
│   ├── validate.go          # Unchanged (artifacts array validation handled by existing schema)
│   └── bridge.go            # Unchanged
├── naming/
│   ├── name.go              # Unchanged (reuse Parse for scope/name extraction)
│   └── reserved.go          # Unchanged
├── scaffold/
│   └── scaffold.go          # Unchanged (reuse WellKnownDirs constant)
├── license/
│   └── detect.go            # Unchanged
├── artifact/
│   ├── discover.go          # Scan well-known dirs, derive names, map types
│   ├── validate.go          # Per-type validation dispatch, ValidationErrors collector
│   ├── skill.go             # Skill-specific validation (SKILL.md frontmatter)
│   └── types.go             # ArtifactType enum, well-known dir → type mapping
├── archive/
│   ├── zip.go               # Create zip archive (single top-level dir, deflate)
│   └── sha256.go            # SHA-256 sidecar generation in sha256sum format
├── ignore/
│   └── ignore.go            # Load .aipkgignore, built-in defaults, path filtering
└── frontmatter/
    └── frontmatter.go       # Extract YAML frontmatter from --- delimiters, parse via yaml.v3

spec/
├── schema/
│   └── package.json         # Unchanged (artifacts array already defined)
├── archive.md               # NEW: archive format reference doc
├── artifacts.md             # Updated: relaxed file extensions, compound name derivation
├── manifest.md              # Unchanged
├── naming.md                # Unchanged
└── reserved-scopes.txt      # Unchanged
```

**Structure Decision**: Standard Go CLI layout, consistent with 001. Four new packages (`artifact`, `archive`, `ignore`, `frontmatter`) with focused responsibilities. Tests colocated with source. The `artifact` package has multiple source files to separate discovery, validation dispatch, and skill-specific logic.

## Post-Design Constitution Re-Check

*Re-evaluated after Phase 1 design completion.*

### I. Simplicity and Deferral

- **PASS (confirmed)**: Four new packages is the minimum for clean separation. Each has a single responsibility. The `artifact` package is the largest with four files, justified by the distinct concerns of discovery, validation, type-specific rules, and type definitions.
- **PASS (confirmed)**: Multi-error collection is 15 lines of stdlib code. The `Collector` helper is simpler than a custom error type.
- **WATCH resolved**: Frontmatter parsing via yaml.v3 `KnownFields(true)` handles FR-015 (allowed keys only) without custom key-checking code.

### II. Core/Adapter Separation

- **PASS (confirmed)**: No adapter code. All new packages are core. The archive format and pack command are tool-agnostic.

### III. Convention Over Invention

- **PASS (confirmed)**: CLI contract follows cobra patterns. `--output` flag mirrors `helm package -d`. Archive naming convention is documented and unambiguous.

### IV. Cold Start First

- **PASS (confirmed)**: Zero-config packing. Authors run `aipkg pack` and get a distributable archive. Quickstart validates the flow.

### V. Backward-Compatible Evolution

- **PASS (confirmed)**: `Artifacts` field uses `omitempty`. Existing manifests unaffected. The archive format is additive.

### Documentation Standard (v1.1.0)

- **PASS**: `spec/archive.md`, `spec/artifacts.md` updates, and CLI reference are in the source code structure, delivered alongside the implementation.

### Gate Result: PASS (no changes from pre-research check)

## Complexity Tracking

No constitution violations to justify.
