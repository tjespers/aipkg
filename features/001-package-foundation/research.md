# Research: Package Foundation

## Schema Bridge (jsonschema/v6 per-field validation)

**Decision**: Use `santhosh-tekuri/jsonschema/v6` property sub-schemas for per-field validation in huh prompts.

**Rationale**: The compiled `*Schema` struct exposes `Properties map[string]*Schema`. Each property sub-schema has its own `Validate(v any) error` method that checks all constraints (type, pattern, minLength, maxLength, enum). This means we compile the root schema once, then call `root.Properties["name"].Validate(input)` to validate a single field. No need to construct fake documents or extract constraints manually.

**Alternatives considered**:
- Hand-written Go validators per field: rejected because it duplicates schema logic and violates FR-029 (schema as single source of truth).
- Construct minimal full documents and validate the whole thing: rejected as unnecessarily complex when property sub-schemas work directly.

**Key API details**:

Loading from embedded bytes:
```go
doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaBytes))
c := jsonschema.NewCompiler()
c.DefaultDraft(jsonschema.Draft2020)
c.AddResource("package.json", doc)
root, err := c.Compile("package.json")
```

Per-field validation:
```go
nameSchema := root.Properties["name"]
err := nameSchema.Validate("@myorg/cool-skill")
```

Error types from the `kind/` subpackage allow type-switching for user-friendly messages:
- `kind.Pattern` -> "invalid format"
- `kind.MinLength` -> "must be at least N characters"
- `kind.MaxLength` -> "must be at most N characters"
- `kind.Enum` -> "must be one of: ..."

Each `ErrorKind` also implements `LocalizedString(*message.Printer)` as a fallback.

**Bridge design**: `ValidateField(property string) func(string) error` compiles the schema once (via `sync.Once`), looks up the property sub-schema, and returns a closure that validates a string value. This closure is passed directly to `huh.NewInput().Validate(...)`.

**Reserved scope handling**: Reserved scope checking lives outside the schema bridge. The schema validates format (pattern, length), but the reserved scopes list is a separate concern. The `naming` package handles this: first the bridge validates format, then the naming package checks against the embedded reserved-scopes.txt. The create command composes both checks into a single validator function for the name field.

## huh Testing Patterns

**Decision**: Use the Elm architecture approach (Strategy A) as the primary testing pattern. Use accessible mode (Strategy B) for integration tests.

**Rationale**: Strategy A is fast, deterministic, needs no TTY, and tests the actual TUI interaction. Strategy B (accessible mode with `WithInput`/`WithOutput`) is simpler for end-to-end flow tests.

**Strategy A (unit tests, no Run)**:
- Construct the form, call `f.Update(f.Init())` to initialize.
- Simulate keystrokes with `tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'H'}}`.
- Navigate with `tea.KeyMsg{Type: tea.KeyEnter}`, `tea.KeyTab`, etc.
- Read values via `field.GetValue()` or bound value pointers.
- Inspect rendered output with `ansi.Strip(f.View())`.
- Process batch commands with a `batchUpdate` helper that recursively executes returned `tea.Cmd` values.

**Strategy B (integration tests, accessible mode)**:
```go
f := huh.NewForm(huh.NewGroup(huh.NewInput().Title("Name:"))).
    WithAccessible(true).
    WithOutput(&buf).
    WithInput(strings.NewReader("@alice/my-pkg\n"))
err := f.Run()
```

Accessible mode uses line-based I/O. Input for fields is newline-terminated. Select fields use 1-based index numbers.

**Strategy C (abort/timeout tests)**:
```go
f.WithInput(nil).WithOutput(io.Discard)
```
Used for testing Ctrl+C abort and timeout behavior.

**Practical implication**: For the `aipkg create` command, the form logic (field definitions, validation functions, value binding) should be separated from the cobra command wiring. This lets unit tests exercise the form directly via Strategy A, while integration tests run the full cobra command.

## License Detection (licensecheck)

**Decision**: Use `google/licensecheck` v0.3.1 with a 90% coverage threshold and single-match requirement.

**Rationale**: The library returns SPDX identifiers directly for all common licenses (MIT, Apache-2.0, BSD-2-Clause, GPL-3.0-only, etc.). No SPDX mapping layer needed. The 90% threshold is conservative: clean LICENSE files hit 95-100%, and suggesting the wrong license is worse than suggesting nothing.

**Alternatives considered**:
- Manual regex matching against known license headers: fragile, hard to maintain, doesn't handle license variations.
- go-license-detector (another Google library): heavier dependency, designed for entire repositories rather than single files.
- Embedding a list of LICENSE file hashes: brittle, doesn't handle copyright line variations.

**API**:
```go
data, _ := os.ReadFile(filepath.Join(dir, "LICENSE"))
cov := licensecheck.Scan(data)
```

Returns `Coverage{Percent float64, Match []Match}`. Each `Match` has `ID string` (SPDX identifier), `Type` (Notice, ShareProgram, etc.), byte offsets, and `IsURL bool`.

**Detection rules**:
1. Try common filenames: LICENSE, LICENSE.txt, LICENSE.md, LICENCE, COPYING.
2. Scan with `licensecheck.Scan(data)`.
3. Require `cov.Percent >= 90` (clean file, single license).
4. Require exactly one match (`len(cov.Match) == 1`).
5. Skip URL-only matches (`m.IsURL`).
6. Skip unknown type (`m.Type == licensecheck.Unknown`).
7. Return `m.ID` as the SPDX identifier.

Any failure returns `("", false)`, and the create command shows no default for the license prompt.

**Non-SPDX IDs**: A small number of licenses (Anti996, CommonsClause, etc.) have non-SPDX IDs. These are rare enough that they'll naturally fall through. No special handling needed.

## go:embed Strategy

**Decision**: Embed the package JSON Schema and reserved-scopes.txt using `//go:embed`. Spec reference docs are not embedded.

**Rationale**: The schema and reserved scopes list are required at runtime for validation. Embedding avoids filesystem lookups and makes the binary self-contained. The spec reference docs have no runtime use in this feature.

**Placement**: The `//go:embed` directives will reference files in `spec/schema/` and `spec/`. Go's embed rules require the directive to be in a package whose source directory is an ancestor of the embedded files, or the files must be in or below the package directory. Since `internal/schema/` is not an ancestor of `spec/`, we have two options:

1. Place a thin embed package at the repo root (e.g., `internal/specdata/embed.go`) that exports the embedded bytes.
2. Copy the schema file into the package directory at build time.

Option 1 is cleaner and avoids build-time file copying. The `specdata` package would simply expose `var SchemaBytes []byte` and `var ReservedScopesBytes []byte`. The `schema` and `naming` packages import from `specdata`.

## regexp2 vs stdlib regex

**Decision**: Use `dlclark/regexp2` for package name validation where PCRE features (lookaheads) are needed.

**Rationale**: The schema's name pattern uses `(?!.*--)` (negative lookahead) to reject consecutive hyphens. Go's `regexp` package does not support lookaheads. `regexp2` implements .NET-compatible regex with full PCRE support.

**Usage scope**: Limited to the `naming` package for validating scoped package names when operating outside the JSON Schema validator. The schema bridge handles this automatically through jsonschema/v6 (which uses its own regex engine internally), but the `naming` package needs it for standalone name parsing (e.g., extracting scope and package-name parts, which is beyond what schema validation does).

## cobra Command Organization

**Decision**: One file per command in `internal/cli/`. Root command in `root.go`, create command in `create.go`.

**Rationale**: This matches the patterns used by helm, flux, and other Go CLIs at scale. At the current size (one command), more elaborate patterns (command factories, command registries) add complexity without benefit. The structure scales by adding files, not by refactoring.

**Command wiring**: `main.go` calls `cli.Execute()`. `root.go` defines the root command and adds subcommands. `create.go` defines the create command, its flags, and calls into the business logic packages.

**Flag-to-prompt pattern**: The create command collects values from flags first, then prompts for any missing values (if TTY is available). This means the form builder receives pre-filled values from flags and only shows prompts for empty fields. The same validation functions (from the schema bridge) apply to both paths.
