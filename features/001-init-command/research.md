# Research: Init Command

**Feature**: 001-init-command | **Date**: 2026-02-28

## CLI Framework: cobra

**Decision**: Use `github.com/spf13/cobra` for command structure. Flags read via cobra's native `cmd.Flags().GetString()`.

**Rationale**: Industry standard for Go CLIs. Provides flag parsing, help generation, usage examples, and subcommand routing out of the box.

**Alternatives considered**:

- `cobra + viper` — originally planned, but viper's global state model causes flag collisions when multiple commands share flag names, and its dependency tree (afero, fsnotify, mapstructure, go-toml, gotenv, etc.) added ~2 MB for zero value. Removed during implementation.
- `urfave/cli` — viable but less ecosystem adoption; cobra is the convention for Go CLIs
- `kong` — declarative style doesn't fit well with dynamic prompt assembly

**Pattern**: Root command in `internal/cli/root.go`. Each subcommand in its own file returning `*cobra.Command`. Entry point in `cmd/aipkg/main.go` calls `cli.NewRootCmd().Execute()`.

## Interactive Prompts: huh

**Decision**: Use `github.com/charmbracelet/huh` for interactive terminal forms.

**Rationale**: Already specified in CLAUDE.md (bubbletea/huh). huh provides high-level form abstractions (select, input, text) with built-in validation, keyboard navigation, and cancellation handling. Returns `huh.ErrUserAborted` on Ctrl+C.

**Alternatives considered**:

- `survey` — archived, no longer maintained
- `promptui` — less polished, fewer features
- Raw bubbletea — too low-level for form-style prompts

**Pattern**: Build form dynamically based on type and pre-filled flags. Each field gets a validation function. Optional fields accept empty input (validator returns nil for empty string). Form runs once; values merged with flag values after completion.

## JSON Schema Validation: jsonschema/v6

**Decision**: Use `github.com/santhosh-tekuri/jsonschema/v6` for schema-based validation.

**Rationale**: Supports JSON Schema Draft 2020-12 (which the aipkg-spec schema uses). Provides structured error output (`BasicOutput()`) with JSON pointers to failing fields. Actively maintained.

**Alternatives considered**:

- `xeipuv/gojsonschema` — no Draft 2020-12 support
- `qri-io/jsonschema` — less mature, fewer features

**Pattern**: Embed `aipkg.schema.json` via `go:embed`. Compile once with `sync.Once`. Use per-field validation by extracting regex patterns from the schema for inline prompt feedback. Also run full schema validation on the assembled manifest before writing. Export validation functions: `ValidateName(s string) error`, `ValidateVersion(s string) error`.

### Schema Gap: `artifacts` — Resolved

The aipkg-spec JSON schema previously required `artifacts` when `type == "package"`. Init deliberately omits artifacts (FR-012).

**Resolution**: The `aipkg-spec` schema will be updated to make `artifacts` optional for packages. Presence of artifacts will be enforced at package/publish time, not at init time. This allows init-generated manifests to pass full schema validation and ensures the manifest is valid at every lifecycle stage.

## License Detection: google/licensecheck

**Decision**: Use `github.com/google/licensecheck` for detecting SPDX identifiers from LICENSE files.

**Rationale**: Lightweight, scans a single file (not a directory tree), returns SPDX IDs directly. Maintained by Google, used by Go ecosystem tooling. Small dependency footprint.

**Alternatives considered**:

- `go-enry/go-license-detector` — heavier, directory-based scanning, overkill for single-file detection
- Manual regex matching — fragile, doesn't cover license text variations

**Pattern**: Read `LICENSE` file in cwd. Run `licensecheck.Scan()`. If confidence > 80% and a match exists, use the SPDX ID as the default license value in the prompt. If no match or low confidence, leave the default blank.

## TTY Detection: golang.org/x/term

**Decision**: Use `golang.org/x/term` for terminal detection.

**Rationale**: Standard Go extended library. `term.IsTerminal(int(os.Stdin.Fd()))` is the canonical way to check if stdin is a TTY in Go.

**Pattern**: Check TTY before building prompts. If not a TTY and required fields are missing, print the missing field names to stderr and exit 1. If TTY, proceed with interactive prompts for missing fields.

## Error Handling Pattern

**Decision**: Wrapped errors with `fmt.Errorf("context: %w", err)`. Errors printed to stderr with `fmt.Fprintf(os.Stderr, ...)`. Exit codes: 0 success, 1 error.

**Rationale**: Matches constitution's error handling boundary. Cobra handles the exit code via `RunE` returning an error.

**Pattern**: Commands return errors via `RunE`. Root command's `Execute()` prints the error and returns it. `main.go` calls `os.Exit(1)` on error. Cobra's `SilenceErrors` and `SilenceUsage` set to control output formatting.
