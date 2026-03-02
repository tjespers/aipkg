# Tasks: Package Foundation

**Input**: Design documents from `/features/001-package-foundation/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/cli-create.md, quickstart.md

**Tests**: Included. Each implementation task includes colocated unit tests (Go `_test.go` convention). User story phases include integration tests for acceptance scenarios.

**Organization**: Tasks grouped by user story. Five stories from spec.md (P1 through P5), with Setup and Foundational phases before story work begins.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1 through US5)
- Exact file paths included in descriptions

---

## Phase 1: Setup

**Purpose**: Project initialization, dependencies, CLI skeleton

- [ ] T001 Create project directory structure per plan.md: `cmd/aipkg/`, `internal/cli/`, `internal/manifest/`, `internal/manifest/testdata/`, `internal/schema/`, `internal/naming/`, `internal/scaffold/`, `internal/license/`
- [ ] T002 Add Go dependencies to go.mod: `github.com/spf13/cobra`, `github.com/charmbracelet/huh`, `github.com/santhosh-tekuri/jsonschema/v6`, `github.com/dlclark/regexp2`, `github.com/google/licensecheck`, `golang.org/x/term`
- [ ] T003 Implement cobra root command in `internal/cli/root.go` with `Execute()` function and wire from `cmd/aipkg/main.go` (replaces existing version stub)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Spec revisions, package-only JSON Schema, embedded assets, and all internal packages that user stories depend on

**CRITICAL**: No user story work can begin until this phase is complete

### Specification revisions

- [ ] T004 Rework `spec/schema/aipkg.json` into package-only schema at `spec/schema/package.json`: remove `type` field and `if/then/else` conditionals, add `specVersion` (integer, const 1), make `artifacts` optional, adjust `required` to `[specVersion, name, version]` per data-model.md
- [ ] T005 [P] Revise `spec/manifest.md` for package-only manifest format: remove `type` field references, document `specVersion`, update field table and examples to reflect package-only schema
- [ ] T006 [P] Revise `spec/artifacts.md` for convention-based directory layout: document well-known directories, artifact type mapping table, structural requirements (e.g. skill dirs must contain SKILL.md), name derivation rules per data-model.md
- [ ] T007 [P] Revise `spec/naming.md` for package-only context: update examples to remove project references, confirm scoped name pattern and reserved scope documentation

### Embedded assets

- [ ] T008 Implement go:embed for `spec/schema/package.json` and `spec/reserved-scopes.txt` per research.md embed strategy: repo-root embed package exporting `PackageSchemaJSON` and `ReservedScopesText` byte slices (go:embed cannot use `..` paths, so the embed directive must live in an ancestor of `spec/`)

### Internal packages (with table-driven tests)

- [ ] T009 [P] Implement `PackageManifest` struct with JSON marshal, `omitempty` for optional fields, and `WriteFile` function in `internal/manifest/manifest.go`; golden file tests against `internal/manifest/testdata/` fixtures in `internal/manifest/manifest_test.go`
- [ ] T010 Implement `ScopedName` struct, `Parse()` function (using regexp2 for lookahead pattern), `String()` method, and reserved scope checking (load from embedded `reserved-scopes.txt`, exact and prefix matching) with table-driven tests in `internal/naming/name.go`, `internal/naming/reserved.go`, and `internal/naming/naming_test.go` (depends on T008)
- [ ] T011 Implement schema compilation from embedded bytes (`jsonschema.UnmarshalJSON`, `Compiler.Compile`) and full-document `Validate(*PackageManifest) error` function with table-driven tests (known-good and known-bad manifests) in `internal/schema/validate.go` and `internal/schema/validate_test.go` (depends on T004, T008)
- [ ] T012 Implement schema bridge: `ValidateField(property string) func(string) error` using `root.Properties[property].Validate()`, `sync.Once` for compilation, kind type-switching for user-friendly error messages, with tests in `internal/schema/bridge.go` and `internal/schema/bridge_test.go` (depends on T011)
- [ ] T013 [P] Implement `Scaffold(targetDir string) error`: create well-known directories (`skills/`, `prompts/`, `commands/`, `agents/`, `agent-instructions/`, `mcp-servers/`), skip existing, with filesystem tests using `t.TempDir()` in `internal/scaffold/scaffold.go` and `internal/scaffold/scaffold_test.go`
- [ ] T014 [P] Implement `DetectLicense(dir string) (string, bool)`: try LICENSE/LICENCE/COPYING filenames, `licensecheck.Scan` with 90% threshold, single-match and non-URL-only filter, return SPDX identifier; table-driven tests in `internal/license/detect.go` and `internal/license/detect_test.go`

**Checkpoint**: All internal packages built and tested. Create command implementation can begin.

---

## Phase 3: User Story 1 - Create a new package from scratch (Priority: P1) MVP

**Goal**: `aipkg create @alice/blog-writer` guides the author through interactive prompts and produces a valid package directory with `aipkg.json` and all well-known artifact directories.

**Independent Test**: Run create with a valid scoped name as positional arg, provide version/description/license through prompts, verify directory structure and manifest content match quickstart.md Scenario 1.

### Tests for User Story 1

- [ ] T015 [US1] Write integration tests for basic create flow using huh accessible mode (`WithInput`/`WithOutput`): name as positional arg skips name prompt, version defaults to 0.1.0, generated `aipkg.json` contains correct fields, all six well-known directories created, no `artifacts` field in output, in `internal/cli/create_test.go`

### Implementation for User Story 1

- [ ] T016 [US1] Implement create command in `internal/cli/create.go`: cobra command with positional name arg, `--name`/`--version`/`--description`/`--license` flags, huh form for prompted fields with schema bridge validators, `PackageManifest` construction and write, scaffold invocation; positional arg takes precedence over `--name` flag per contract
- [ ] T017 [US1] Register create subcommand in `internal/cli/root.go`
- [ ] T018 [US1] Add golden file tests for generated `aipkg.json` output (minimal: specVersion + name + version; full: all fields populated) in `internal/manifest/testdata/`

**Checkpoint**: `aipkg create @alice/blog-writer` works interactively. Quickstart Scenario 1 passes.

---

## Phase 4: User Story 2 - Create in existing directory (Priority: P2)

**Goal**: `--path` flag targets an existing directory, preserving existing files while adding package structure. Refuses if `aipkg.json` already exists.

**Independent Test**: Create temp directory with pre-existing files, run create with `--path`, verify existing files untouched and new structure added. Run again to verify conflict rejection.

### Tests for User Story 2

- [ ] T019 [US2] Write tests for `--path` handling in `internal/cli/create_test.go`: existing files preserved when using `--path`, existing `aipkg.json` triggers error, non-existent `--path` target is created, `--path .` works for current directory, directory derived from package name when `--path` absent

### Implementation for User Story 2

- [ ] T020 [US2] Implement `--path`/`-p` flag in `internal/cli/create.go`: resolve target directory, check for existing `aipkg.json` (exit with error if found), derive directory name from package name when `--path` absent, create target directory if it doesn't exist

**Checkpoint**: Quickstart Scenarios 2, 3, and 4 pass. Existing files preserved, conflicts rejected.

---

## Phase 5: User Story 3 - Validate package name during creation (Priority: P3)

**Goal**: Invalid package names show inline validation errors during prompts with clear messages. Author can correct without losing progress on other fields.

**Independent Test**: Enter invalid names (unscoped, reserved scope, invalid characters, consecutive hyphens) and verify inline error messages match contracts/cli-create.md examples.

### Tests for User Story 3

- [ ] T021 [US3] Write tests for name validation UX in `internal/cli/create_test.go`: unscoped name produces "must be scoped" error, reserved scope produces "scope is reserved" error, invalid characters rejected, consecutive hyphens rejected, version format error shows MAJOR.MINOR.PATCH hint, description length error shows max 255 message

### Implementation for User Story 3

- [ ] T022 [US3] Implement composed name validator in `internal/cli/create.go`: chain schema bridge pattern validation with reserved scope check from naming package, translate raw validation errors into user-friendly messages per contracts/cli-create.md error message examples, wire as huh input `.Validate()` function

**Checkpoint**: Quickstart Scenario 6 passes. Invalid names caught inline with helpful messages.

---

## Phase 6: User Story 4 - Non-interactive package creation for CI (Priority: P4)

**Goal**: All metadata via flags creates package with no prompts. Missing flags without TTY exits with error listing the missing fields.

**Independent Test**: Run with all flags and verify zero prompts. Pipe input (no TTY) with partial flags and verify error lists missing ones.

### Tests for User Story 4

- [ ] T023 [US4] Write tests for non-interactive mode in `internal/cli/create_test.go`: all flags provided skips all prompts and creates package, partial flags with TTY prompts only for missing fields, no TTY with missing required flags exits with error listing them, invalid flag value exits with validation error (no partial creation)

### Implementation for User Story 4

- [ ] T024 [US4] Implement TTY detection via `x/term` and non-interactive paths in `internal/cli/create.go`: detect TTY with `term.IsTerminal(os.Stdin.Fd())`, skip huh form entirely when all fields present via flags, build form with only missing fields when partial, exit with missing-flags error when no TTY and fields missing per behavior matrix in contracts/cli-create.md

**Checkpoint**: Quickstart Scenarios 2 (non-interactive) and 7 pass. CI pipelines can create packages via flags.

---

## Phase 7: User Story 5 - License detection (Priority: P5)

**Goal**: Existing LICENSE file in target directory is auto-detected and suggested as the default value in the license prompt.

**Independent Test**: Place known LICENSE files in target directory, run create, verify license prompt shows detected SPDX identifier as default.

### Tests for User Story 5

- [ ] T025 [US5] Write tests for license detection integration in `internal/cli/create_test.go`: Apache-2.0 LICENSE file detected as prompt default, MIT LICENSE detected correctly, no LICENSE file means no default, ambiguous/unrecognized LICENSE means no default

### Implementation for User Story 5

- [ ] T026 [US5] Wire license detection into create command in `internal/cli/create.go`: call `license.DetectLicense(targetDir)` before building huh form, set detected SPDX identifier as default value for license input field, skip detection when `--license` flag already provided

**Checkpoint**: Quickstart Scenario 5 passes. Detected licenses pre-fill the prompt.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, end-to-end validation, CI readiness

- [ ] T027 [P] Write CLI reference documentation for the `create` command per constitution documentation standard (synopsis, flags, examples, common workflows) in `docs/` or alongside spec docs
- [ ] T028 Run all quickstart.md validation scenarios (1 through 7) end-to-end against built binary and fix any issues found
- [ ] T029 Run `task check` (lint + vet + test + race detector) and verify all passes; confirm cross-platform build with `task build`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies. Start immediately.
- **Foundational (Phase 2)**: Depends on Setup. **Blocks all user stories.**
- **US1 (Phase 3)**: Depends on Foundational. **Blocks US2 through US5** (they extend the create command).
- **US2-US5 (Phases 4-7)**: Depend on US1. Independent of each other. Proceed in priority order.
- **Polish (Phase 8)**: Depends on all user stories being complete.

### User Story Dependencies

- **US1 (P1)**: After Foundational. No story dependencies. This is the MVP.
- **US2 (P2)**: After US1. Adds `--path` handling to the existing create command.
- **US3 (P3)**: After US1. Adds composed name validation with user-friendly error messages.
- **US4 (P4)**: After US1. Adds TTY detection and flag-only execution path.
- **US5 (P5)**: After US1. Wires license detection into the prompt flow.

### Within Each Phase

- Spec doc revisions (T005, T006, T007) can run in parallel with each other
- Internal packages T009, T013, T014 can run in parallel (independent files, no shared deps)
- T010 and T011 depend on T008 (embedded assets)
- T012 depends on T011 (schema bridge needs compiled schema)
- Within user stories: tests before implementation

### Parallel Opportunities

**Phase 2 parallel group** (no dependencies on each other):
- T005, T006, T007 (spec doc revisions)
- T009 (manifest struct)
- T013 (scaffold)
- T014 (license detection)

**Phase 2 sequential chain**:
- T004 (schema) → T008 (embed) → T010 (naming), T011 (schema validate) → T012 (schema bridge)

**Phases 4-7**: Independent of each other, sequential after Phase 3.

---

## Parallel Example: Phase 2 Foundational

```text
# Launch all independent tasks in parallel:
T005: "Revise spec/manifest.md"
T006: "Revise spec/artifacts.md"
T007: "Revise spec/naming.md"
T009: "Implement PackageManifest struct in internal/manifest/"
T013: "Implement scaffold in internal/scaffold/"
T014: "Implement license detection in internal/license/"

# Sequential chain (after T004 completes):
T004 → T008 → T010, T011 → T012
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL, blocks everything)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Run quickstart Scenario 1, verify basic flow
5. This is a usable `aipkg create` command

### Incremental Delivery

1. Setup + Foundational → All internal packages ready
2. Add US1 → Basic interactive create → **MVP**
3. Add US2 → Existing directory support
4. Add US3 → Inline validation UX
5. Add US4 → Non-interactive CI mode
6. Add US5 → License auto-detection
7. Polish → Docs, full validation, CI checks
8. Each story adds capability without breaking previous stories

---

## Notes

- [P] tasks = different files, no dependencies on incomplete tasks
- [Story] label maps task to specific user story for traceability
- All internal packages include table-driven tests (Go `_test.go` convention)
- Golden file tests go in `testdata/` directories
- Filesystem tests use `t.TempDir()` for isolation (no mocking framework)
- huh integration tests use accessible mode (`WithInput`/`WithOutput`) per research.md Strategy B
- Schema is the single source of truth for validation (FR-029)
- Commit after each task or logical group
- Stop at any checkpoint to validate the story independently
