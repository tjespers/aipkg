# Implementation Plan: Init Command

**Branch**: `001-init-command` | **Date**: 2026-02-28 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-init-command/spec.md`

## Summary

Implement the `aipkg init` command — the first CLI command that creates `aipkg.json` manifests through an interactive flow or via flags. This also establishes the foundational patterns for all future commands: cobra command structure, interactive prompts with huh, schema-driven validation, structured error handling, and the `internal/` package layout.

## Technical Context

**Language/Version**: Go 1.25.7 (pinned in go.mod)
**Primary Dependencies**:
- `github.com/spf13/cobra` — CLI command framework
- `github.com/charmbracelet/huh` — Interactive terminal forms
- `github.com/santhosh-tekuri/jsonschema/v6` — JSON Schema validation (Draft 2020-12)
- `github.com/google/licensecheck` — LICENSE file SPDX detection
- `golang.org/x/term` — TTY detection
**Storage**: Filesystem only (reads/writes `aipkg.json` in cwd)
**Testing**: `go test ./...` via `task test`; table-driven tests, filesystem via `os.MkdirTemp`
**Target Platform**: Cross-platform (linux/darwin/windows amd64/arm64)
**Project Type**: CLI
**Performance Goals**: N/A — single-shot command, sub-second execution
**Constraints**: Offline-only (FR-011), no adapter logic, no artifact scaffolding
**Scale/Scope**: Single command, 4 new packages in `internal/`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Simplicity and Deferral | ✅ Pass | Init is the smallest useful first command. No `require`, `repositories`, or `artifacts` fields — those come later. Viper removed as unnecessary complexity — cobra flags are sufficient. |
| II. Core/Adapter Separation | ✅ Pass | Init has zero adapter logic. No tool-specific code. |
| III. Convention Over Invention | ✅ Pass | Follows `npm init` / `composer init` conventions. Scoped names, semver, SPDX — all standard. |
| IV. Cold Start First | ✅ Pass | Creates the manifest needed before any other operation. Interactive flow with `@ ` prompt prefix reduces friction — users type `scope/name` without the `@`. |
| V. Backward-Compatible Evolution | ✅ Pass | First command — no backward compat concerns yet. |
| Schema validation boundary | ✅ Pass | Field validation uses patterns from the `aipkg-spec` JSON schema. CLI does not invent validation rules. |
| Error handling | ✅ Pass | Wrapped errors with `fmt.Errorf("context: %w", err)`. Exit 0/1. Errors printed to stderr via `main.go`. |
| Testability | ✅ Pass | All packages testable in isolation. Filesystem via temp dirs, prompts bypassed in tests via flag-only mode. |
| Import discipline | ✅ Pass | Everything in `internal/`. No circular imports. One-way dependency graph. |

No violations. Complexity tracking not needed.

## Project Structure

### Documentation (this feature)

```text
specs/001-init-command/
├── plan.md              # This file
├── research.md          # Phase 0: dependency decisions
├── data-model.md        # Phase 1: manifest entity model
├── quickstart.md        # Phase 1: usage examples
├── contracts/
│   └── cli.md           # Phase 1: CLI interface contract
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
cmd/aipkg/
    main.go                  # Entry point → cli.NewRootCmd().Execute(), error printing

internal/
    cli/
        root.go              # Root cobra command, version flag
        init.go              # aipkg init command (flags, prompt orchestration, file write)
        init_test.go         # Integration tests for init command
    manifest/
        manifest.go          # Manifest struct, JSON marshal with 2-space indent + trailing newline
        manifest_test.go     # Unit tests for manifest serialization
    schema/
        schema.go            # Embedded JSON schema, per-field validation functions
        schema_test.go       # Validation tests (valid/invalid names, versions, etc.)
        aipkg.schema.json    # Copied from aipkg-spec (go:embed)
    license/
        detect.go            # LICENSE file SPDX detection
        detect_test.go       # Detection tests
```

**Structure Decision**: Flat `internal/` package layout with clear single-responsibility packages. No nested hierarchies — this is a CLI with 4 concerns: command wiring (cli), data (manifest), validation (schema), license detection (license). The `cli` package is the composition root that wires the others together.

## Key Design Decisions

### No Viper

Originally planned with viper for config/flag binding. Removed during implementation — viper's global state model would cause flag collisions when adding future commands, and its full dependency tree (afero, fsnotify, mapstructure, go-toml, gotenv, etc.) added ~2 MB to the binary for zero value. Cobra's native `cmd.Flags().GetString()` is sufficient.

### Schema Validation Strategy

The `aipkg-spec` JSON schema was updated to make `artifacts` optional for package-type manifests (enforced at package/publish time instead). This resolves the previous gap where init-generated manifests would have failed full schema validation.

**Decision**: Use per-field validation during interactive prompting (extract name regex, version regex from the compiled schema). Additionally, run full schema validation on the assembled manifest before writing, since init output will now be schema-valid with `artifacts` optional.

**Rationale**: With the upstream schema change, the manifest is valid at every lifecycle stage. Per-field validation provides inline feedback during prompts; full validation provides a final safety net before writing.

### Interactive Prompt Architecture

Use `huh` forms with per-field validation functions. The form is assembled dynamically based on the manifest type and which flags are already provided:

1. If `--type` not provided → prompt for type selection
2. Based on type, build field list → subtract fields already provided via flags
3. Assemble huh form with remaining fields → each field gets a schema-derived validator
4. Run form → collect values → merge with flag values
5. Build manifest struct → serialize to JSON → write file

This handles all three modes (fully interactive, fully non-interactive, hybrid) with a single code path.

**Name prompt UX**: Interactive name input uses `Prompt("@ ")` so the `@` appears as a fixed prefix. Users type `scope/package-name` without the `@`. The `@` is prepended programmatically before validation and storage. The `--name` flag still accepts the full `@scope/name` format for scripting.

### Error Output

All errors go to stderr. Confirmation message goes to stdout (FR-018). Warnings (irrelevant flags) go to stderr. This follows Unix convention and allows stdout to be piped/redirected cleanly.

Cobra's `SilenceErrors` and `SilenceUsage` are both true — error printing is handled by `main.go` with `fmt.Fprintln(os.Stderr, "Error:", err)` before `os.Exit(1)`.

### TTY Detection

When stdin is not a TTY and required fields are missing, the command errors immediately with a list of missing fields (edge case from spec). Use `term.IsTerminal(int(os.Stdin.Fd()))` from `golang.org/x/term`.

## Complexity Tracking

No constitution violations to justify.
