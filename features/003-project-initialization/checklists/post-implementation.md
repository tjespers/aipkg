# Post-Implementation Checklist: Project Initialization & Model

**Purpose**: Verify implementation correctness, cross-spec consistency, and documentation completeness before opening the implementation PR. Includes deferred item review from the pre-implementation audit.
**Created**: 2026-03-08
**Completed**: 2026-03-08
**Feature**: [spec.md](../spec.md)

## Implementation Correctness

- [x] POST001 Does `aipkg init` create `aipkg-project.json` with exactly `specVersion: 1` and empty `require: {}`? (FR-015, FR-004)
  - **PASS.** `project.Create()` builds `File{SpecVersion: 1, Require: map[string]string{}}` and writes it with `json.MarshalIndent`. Verified by `TestInitHappyPath` and `TestCreateAndLoad`.
- [x] POST002 Does the created file use two-space JSON indent with trailing newline, matching the `create` command pattern? (DD-005, T003)
  - **PASS.** `json.MarshalIndent(p, "", "  ")` followed by `data = append(data, '\n')`. `TestCreateJSONStructure` asserts trailing newline and exact field count.
- [x] POST003 Does `aipkg init` print a success confirmation message following the `create` command pattern? (DD-005)
  - **PASS.** Prints `"Initialized project in aipkg-project.json"` via `fmt.Fprintln`. Follows the `create` command's pattern (`"Created package %s in %s/"`) with verb + location.
- [x] POST004 Does `aipkg init` NOT create `.aipkg/` or any other directories or files beyond `aipkg-project.json`? (FR-018, FR-010)
  - **PASS.** Only `os.WriteFile` for the project file. `TestInitHappyPath` explicitly asserts `.aipkg/` does not exist after init.
- [x] POST005 Does the re-init guard trigger on file existence only, not content validity? An `aipkg-project.json` with invalid JSON still blocks re-init. (FR-016, Edge Cases)
  - **PASS.** Uses `os.Stat("aipkg-project.json")` which checks file existence, not content. Any file with that name blocks, regardless of content.
- [x] POST006 Does the mutual exclusivity guard trigger on file existence only? An `aipkg.json` with invalid JSON still blocks init. (FR-017, Edge Cases)
  - **PASS.** Uses `os.Stat("aipkg.json")` which checks file existence, not content. The test uses `{"specVersion": 1}` (valid JSON but incomplete manifest) and it still blocks.
- [x] POST007 Is the mutual exclusivity check (aipkg.json) performed before the re-init check (aipkg-project.json)? (FR-002, quickstart.md:41-44)
  - **PASS.** In `runInit()`, the `os.Stat("aipkg.json")` check is at line 27, before `os.Stat("aipkg-project.json")` at line 32.
- [x] POST008 Does the mutual exclusivity error message mention both `aipkg require` and `aipkg install` as alternatives? (FR-017, US3 Acceptance Scenario 3)
  - **PASS.** Error message: `"package manifest (aipkg.json) already exists in this directory; use aipkg require or aipkg install instead"`. `TestInitMutualExclusivityGuard` asserts both substrings.
- [x] POST009 Do error messages follow existing CLI patterns: lowercase, no period, wrapped with context? (DD-005)
  - **PASS.** All three error paths use lowercase, no trailing period: `"package manifest (aipkg.json) already exists..."`, `"project already initialized (aipkg-project.json exists)"`, `"cannot initialize project: %w"`.

## Schema Correctness

- [x] POST010 Does `spec/schema/project.json` use Draft 2020-12 and follow the structure/style of `spec/schema/package.json`? (FR-005)
  - **PASS.** `"$schema": "https://json-schema.org/draft/2020-12/schema"`. Structure follows `package.json` style: `$schema`, `$id`, `title`, `description`, `type`, `required`, `properties`, `additionalProperties`.
- [x] POST011 Does the schema enforce `additionalProperties: false` on the root object, rejecting identity fields and any unknown properties? (FR-003)
  - **PASS.** Root-level `"additionalProperties": false`. `TestValidateProject/extra_fields_rejected` confirms that adding `"name"` fails validation.
- [x] POST012 Does the schema's `require` key pattern match the package manifest's `require` key pattern exactly (same regex)? (CHK018, Assumption 2)
  - **PASS.** Both schemas use identical `propertyNames` pattern: `^@(?!.*--)[a-z0-9]([a-z0-9-]{0,37}[a-z0-9])?/[a-z0-9]([a-z0-9-]{0,62}[a-z0-9])?$`
- [x] POST013 Does the schema accept pre-release versions (`1.0.0-beta.1`, `1.0.0-rc.1`, `1.0.0-0.3.7`) but reject build metadata (`1.0.0+build.1`)? (FR-004, DD-002, R-002)
  - **PASS.** Project schema version pattern includes the optional pre-release group `(?:-(...))?` but no `+` group. Tests confirm: `valid project with pre-release version`, `valid project with alpha pre-release`, `valid project with numeric pre-release`, `valid project with rc pre-release` all pass; `build metadata rejected` fails validation.
- [x] POST014 Does `ValidateProject()` reuse the existing regexp2 engine setup and compiler infrastructure from `internal/schema`? (DD-003)
  - **PASS.** `compiledProject()` follows the same pattern as `compiled()`: `jsonschema.NewCompiler()` with `c.UseRegexpEngine(regexp2Compile)` using the shared `regexp2Compile` function and `regexp2Regexp` adapter type.
- [x] POST015 Is `ProjectSchemaJSON` embedded in `specdata.go` alongside `PackageSchemaJSON`? (T002)
  - **PASS.** `specdata.go` contains both `//go:embed spec/schema/package.json` and `//go:embed spec/schema/project.json` with exported `PackageSchemaJSON` and `ProjectSchemaJSON` variables.

## Documentation Completeness

- [x] POST016 Does `spec/project.md` cover the project file format with field definitions and examples (empty and with dependencies)? (FR-019)
  - **PASS.** Documents both fields (`specVersion`, `require`) with type descriptions, constraints, and examples. Shows minimal example (empty require) and full example with three dependencies including a pre-release version.
- [x] POST017 Does `spec/project.md` document the install directory layout: `.aipkg/` structure, four individual-type subdirectories, two merged root-level files? (FR-006, FR-007, FR-008)
  - **PASS.** "Install directory" section shows full tree: `skills/`, `prompts/`, `commands/`, `agents/` as individual-type subdirectories, plus `mcp.json` and `agent-instructions.md` as merged root-level files.
- [x] POST018 Does `spec/project.md` document the three-segment scoped naming convention (`scope.package-name.artifact-name`) with parsing rules and examples? (FR-012, FR-013, DD-001)
  - **PASS.** "Scoped artifact naming" section documents three-segment format, explains collision prevention and traceability, describes parsing rules ("split on `.`, first = scope, last = artifact, middle = package"), and provides concrete examples.
- [x] POST019 Does `spec/project.md` document the `.gitignore` behavior (content: `*` + `!.gitignore`, git-only, created at install time)? (FR-009, FR-010)
  - **PASS.** ".gitignore" subsection shows the file content (`*` / `!.gitignore`), states it's created "inside a git repository", and clarifies `.gitignore` itself is the only committed file inside `.aipkg/`.
- [x] POST020 Does `spec/project.md` document the merged file ownership model (aipkg-managed, overwritten on install/update, manual edits lost)? (FR-011)
  - **PASS.** "Merged files" subsection: "fully aipkg-managed", "generated and overwritten by install and update operations", "Manual edits to these files will be lost."
- [x] POST021 Does `spec/project.md` document the mutual exclusivity rule in both directions (init refuses when aipkg.json exists, and future package creation commands should refuse when aipkg-project.json exists)? (FR-002, CHK021)
  - **PASS.** "Mutual exclusivity" section documents both directions: "`aipkg init` refuses to create a project file when a package manifest exists. Future commands that create package manifests will refuse when a project file exists."
- [x] POST022 Does `spec/naming.md` document three-segment install directory naming alongside the existing two-segment adapter convention, noting that adapters may use a shorter form? (DD-001, R-001, CHK017)
  - **PASS.** "Dot-notation" section has two subsections: "Install directory naming (three-segment)" with examples and rationale, and "Adapter naming (two-segment)" noting adapters "may use a shorter two-segment form if the target tool has its own namespacing." Summary table also distinguishes both forms.

## Cross-Spec Consistency

- [x] POST023 Are the artifact type directory names in documentation (`skills/`, `prompts/`, `commands/`, `agents/`) consistent with `spec/artifacts.md` and the `type` enum in `spec/schema/package.json`? (CHK020)
  - **PASS.** `spec/project.md` layout uses `skills/`, `prompts/`, `commands/`, `agents/`. `spec/artifacts.md` lists the same four directory names. `spec/schema/package.json` `type` enum: `["skill", "prompt", "command", "agent", "agent-instructions", "mcp-server"]`. The two merged types (`agent-instructions`, `mcp-server`) appear as root-level files in the install layout, consistent with the adapter behavior documented in `spec/artifacts.md`.
- [x] POST024 Is the `specVersion` field definition consistent between the project schema and the package manifest schema (integer, const 1)? (CHK006)
  - **PASS.** Both schemas define `specVersion` as `"type": "integer", "const": 1`.
- [x] POST025 Does `LoadFile()` exist alongside `Create()` in `internal/project`, with rationale documented in DD-006? (DD-006, CHK028)
  - **PASS.** Both functions exist in `internal/project/project.go`. `LoadFile()` reads and parses the file. `Create()` writes a new file. DD-006 rationale is in the feature spec's design decisions.

## Deferred Items Review

These items were deferred during the pre-implementation audit as install-command scope. Confirm the deferral is still appropriate or note if implementation revealed a need to address now.

- [x] POST026 **Git detection method** (CHK010): How to detect "within a git working tree" is unspecified. Did implementation or documentation work reveal that this needs definition now?
  - **Deferral appropriate.** The init command does not create `.aipkg/` or `.gitignore`. The documentation correctly states `.gitignore` is created "when `.aipkg/` is first created inside a git repository", which is install-command scope. The detection method can be defined when implementing the install command.
- [x] POST027 **Merged file formats** (CHK011): `mcp.json` JSON structure and `agent-instructions.md` marker format are unspecified. Still appropriate for deferral to install command?
  - **Deferral appropriate.** The project documentation describes merged files at the conceptual level (ownership model, overwrite behavior). The specific JSON structure and marker format are install-command implementation details.
- [x] POST028 **Installed artifact file extensions** (CHK015): Relationship between scoped name and file extension is unspecified (`.ext` placeholder). Still appropriate for deferral?
  - **Deferral appropriate.** The documentation examples use concrete extensions (`.md` for prompts/commands/agents, directory for skills), but the formal mapping rule belongs with the install command. The examples in `spec/project.md` and `spec/naming.md` are illustrative, not normative.
- [x] POST029 **Installed skill directory structure** (CHK016): Whether installed skills preserve original directory structure is unspecified. Still appropriate for deferral?
  - **Deferral appropriate.** The documentation shows `scope.pkg.skill-name/SKILL.md` in the layout, which implies a flat structure. The exact rules for multi-file skills belong with the install command.

## Test Coverage

- [x] POST030 Do tests cover all four US1 acceptance scenarios (empty dir, existing files, no .aipkg/ created, correct JSON structure)?
  - **PASS.** SC1: `TestInitHappyPath` (empty dir). SC2: `TestInitWithExistingFiles` (existing files unchanged). SC3: `TestInitHappyPath` asserts no `.aipkg/`. SC4: `TestInitHappyPath` + `TestCreateJSONStructure` + `TestCreateAndLoad` verify JSON structure.
- [x] POST031 Do tests cover both US2 acceptance scenarios (error on re-init, existing file unchanged)?
  - **PASS.** `TestInitReInitGuard` covers both: asserts error contains "already initialized" and verifies existing file content is unchanged after failed re-init.
- [x] POST032 Do tests cover all three US3 acceptance scenarios (error on package dir, no files created, error suggests commands)?
  - **PASS.** `TestInitMutualExclusivityGuard` covers all three: asserts error mentions "package manifest"/"aipkg.json", asserts `aipkg-project.json` not created, asserts error contains both "aipkg require" and "aipkg install".
- [x] POST033 Do tests cover the edge cases from spec.md (write permissions, invalid existing JSON for both file types)?
  - **PASS.** Write permissions: `TestInitReadOnlyDir` + `TestCreateReadOnlyDir`. Invalid existing JSON: `TestLoadFileInvalidJSON` covers invalid project JSON. The guards use `os.Stat` (existence only), so invalid JSON content is implicitly covered by any test that has a file present.
- [x] POST034 Do schema validation tests cover SC-005 (valid files pass, invalid files fail: bad names, bad versions, extra fields, missing fields, build metadata rejected)?
  - **PASS.** `TestValidateProject` has 16 subtests: 6 valid cases (empty, with deps, 4 pre-release variants) and 10 invalid cases (missing specVersion, missing require, extra fields, bad name, bad version, build metadata, wrong specVersion, version prefix, consecutive hyphens, invalid JSON).
- [x] POST035 Does `LoadFile()` have roundtrip test coverage (create, load, verify structure)? (DD-006)
  - **PASS.** `TestCreateAndLoad` creates a file, loads it back, and verifies `SpecVersion == 1`, `Require` is non-nil, and `len(Require) == 0`.

## Notes

- Items POST026-POST029 originate from the pre-implementation checklist (CHK010, CHK011, CHK015, CHK016). Resolution notes should explain whether deferral is still appropriate.
- If any deferred item needs action, create a Linear issue and note the issue ID next to the checklist item.
- Cross-spec consistency items (POST023-POST025) ensure the implementation doesn't drift from the existing specification foundation.
- **All 35 items pass.** All tests pass (verified by running `go test` on 2026-03-08). All four deferred items remain appropriate for install-command scope.
