# Tasks: Init Command

**Input**: Design documents from `/specs/001-init-command/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/cli.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization — dependencies and schema file

- [x] T001 Copy manifest schema from `../aipkg-spec/schema/aipkg.json` to `internal/schema/aipkg.schema.json`
- [x] T002 Add Go dependencies: `go get github.com/spf13/cobra github.com/charmbracelet/huh github.com/santhosh-tekuri/jsonschema/v6 github.com/google/licensecheck golang.org/x/term`

______________________________________________________________________

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core packages that ALL user stories depend on

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T003 [P] Implement Manifest struct with JSON serialization in `internal/manifest/manifest.go` — struct with 5 fields (Type, Name, Version, Description, License), `omitempty` on all except Type, `Marshal()` using `json.MarshalIndent(m, "", "  ")` with trailing newline appended
- [x] T004 [P] Implement schema embedding and per-field validators in `internal/schema/schema.go` — `go:embed aipkg.schema.json`, compile schema with `sync.Once`, export `ValidateName(string) error`, `ValidateVersion(string) error`, `ValidateDescription(string) error` (using `utf8.RuneCountInString` for character count), and `ValidateManifest([]byte) error` for full schema validation
- [x] T005 [P] Implement LICENSE file SPDX detection in `internal/license/detect.go` — read `LICENSE` file in a given directory, run `licensecheck.Scan()`, return SPDX ID if confidence > 80%, empty string otherwise
- [x] T006 Write unit tests for Manifest serialization in `internal/manifest/manifest_test.go` — table-driven tests: package with all fields, project with only type, omitempty behavior, 2-space indent, trailing newline, field order (depends on T003)
- [x] T007 Write unit tests for schema validators in `internal/schema/schema_test.go` — table-driven tests: valid/invalid names (scoped, unscoped, consecutive hyphens, length limits), valid/invalid versions (semver, partial, leading zeros), description length, full manifest validation (depends on T004)
- [x] T008 Write unit tests for license detection in `internal/license/detect_test.go` — table-driven tests using `os.MkdirTemp`: Apache-2.0 LICENSE file, MIT LICENSE file, no LICENSE file, unrecognized content (depends on T005)
- [x] T009 Implement root cobra command in `internal/cli/root.go` — `NewRootCmd()` returning `*cobra.Command`, `SilenceErrors: true`, `SilenceUsage: true`, version flag using ldflags variables
- [x] T010 Update entry point in `cmd/aipkg/main.go` — `cli.NewRootCmd().Execute()`, print errors to stderr via `fmt.Fprintln(os.Stderr, "Error:", err)`, call `os.Exit(1)` on error, preserve ldflags variables (version, commit, date)

**Checkpoint**: Foundation ready — all packages tested, root command works, `aipkg --version` functional

______________________________________________________________________

## Phase 3: User Story 1 + User Story 4 — Package Interactive + Overwrite Guard (Priority: P1) MVP

**Goal**: Package author can create a manifest through interactive prompts. Existing manifests are never overwritten.

**Independent Test**: Run `aipkg init` in an empty directory, select "package", provide valid fields -> verify `aipkg.json` created. Run again in same directory -> verify error and file unchanged.

### Implementation

- [x] T011 [US1] Implement init command scaffold in `internal/cli/init.go` — `newInitCmd()` returning `*cobra.Command` with `RunE`, define 5 string flags (type, name, version, description, license) read via `cmd.Flags().GetString()`, check `aipkg.json` existence at start and return `fmt.Errorf("aipkg.json already exists")` if found (US4 guard)
- [x] T012 [US1] Implement package interactive flow in `internal/cli/init.go` — build `huh.Form` dynamically: type select (project/package), then for package: name input with `Prompt("@ ")` prefix and `ValidateName` (prepending `@` before validation), version input with `ValidateVersion` and default `0.1.0`, description input with `optionalValidator(ValidateDescription)`, license input with detected default from `license.Detect()`; skip any field already provided via flag; handle `huh.ErrUserAborted` for Ctrl+C via shared `runForm()` helper
- [x] T013 [US1] Implement manifest assembly and file write in `internal/cli/init.go` — build `Manifest` struct from collected values (shared Name/Description fields, package-only Version/License), call `ValidateManifest` on serialized JSON, write to `aipkg.json` with `os.WriteFile` (0o600 permissions), print `Created aipkg.json (package)` to stdout
- [x] T014 [US1] Register init subcommand in `internal/cli/root.go` — add `rootCmd.AddCommand(newInitCmd())` in `NewRootCmd()`

**Checkpoint**: `aipkg init` creates package manifests interactively, refuses on existing file, handles Ctrl+C

______________________________________________________________________

## Phase 4: User Story 2 — Project Interactive (Priority: P2)

**Goal**: Project maintainer can create a minimal manifest with optional name and description.

**Independent Test**: Run `aipkg init`, select "project", skip all prompts -> verify `aipkg.json` contains only `{"type": "project"}`. Run with name and description -> verify both included.

### Implementation

- [x] T015 [US2] Add project prompt flow to `internal/cli/init.go` — when type is "project": prompt for name (optional, `Prompt("@ ")` prefix, validated with `optionalValidator(ValidateName)` prepending `@`) and description (optional, validated with `optionalValidator(ValidateDescription)`); both skippable via Enter; assemble manifest with only non-empty fields; print `Created aipkg.json (project)` to stdout

**Checkpoint**: Both package and project interactive flows work independently

______________________________________________________________________

## Phase 5: User Story 3 — Non-Interactive & Hybrid Mode (Priority: P3)

**Goal**: CI pipelines can create manifests via flags without interactive prompts. Partial flags trigger prompts only for missing fields.

**Independent Test**: Run `aipkg init --type package --name @myorg/my-tool --version 0.1.0` -> verify no prompts, valid file created. Run with `--type package --name @myorg/my-tool` (missing version) in TTY -> verify prompted only for version.

### Implementation

- [x] T016 [US3] Add TTY detection and non-interactive error handling to `internal/cli/init.go` — use `term.IsTerminal(int(os.Stdin.Fd()))` to detect TTY; when no TTY and required fields missing, return `fmt.Errorf("missing required flags for package: --name")` and exit 1; when all required fields provided via flags, skip all prompts entirely; version defaults to `0.1.0` if not provided
- [x] T017 [US3] Add irrelevant flag warnings to `internal/cli/init.go` — when `--type project` is set and `--version` or `--license` flags are provided, print `Warning: --version is ignored for project type` (and/or `--license`) to stderr, then proceed normally

**Checkpoint**: All three modes work: fully interactive, fully non-interactive, hybrid. TTY detection correct.

______________________________________________________________________

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Help text, integration tests, and end-to-end validation

- [x] T018 Add usage examples to init command help in `internal/cli/init.go` — set `Example` field on `cobra.Command` with examples from quickstart.md (package interactive, project, non-interactive, hybrid)
- [x] T019 Write integration tests for init command in `internal/cli/init_test.go` — table-driven tests using `os.MkdirTemp`: package creation via flags (all fields, required only, default version), project creation via flags (minimal, with optional fields), overwrite guard, invalid name/version validation errors, non-interactive missing fields error, irrelevant flag warnings, omitempty behavior for skipped optional fields, consecutive hyphens in name, JSON formatting (2-space indent, trailing newline)
- [x] T020 Run `task check` (lint + vet + test) and validate all quickstart.md scenarios manually

______________________________________________________________________

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - US1+US4 (Phase 3) must complete before US2 (Phase 4) since project flow extends init.go
  - US2 (Phase 4) must complete before US3 (Phase 5) since non-interactive mode covers both types
- **Polish (Phase 6)**: Depends on all user stories being complete

### Within Each Phase

- Phase 2: T003, T004, T005 are [P] (different files, no deps); T006-T008 depend on T003-T005 respectively (tests after implementations); T009 depends on T002; T010 depends on T009
- Phase 3: T011 -> T012 -> T013 (sequential within init.go); T014 depends on T011
- Phase 4: T015 depends on Phase 3 completion
- Phase 5: T016 -> T017 (sequential, same file)
- Phase 6: T018, T019 are parallel; T020 depends on T019

### Parallel Opportunities

- Phase 2: T003, T004, T005 can run in parallel (3 implementation files); then T006, T007, T008 can run in parallel (3 test files, after their implementations)
- Phase 6: T018 and T019 can run in parallel (different concerns in different files)

______________________________________________________________________

## Implementation Strategy

### MVP First (User Story 1 + 4 Only)

1. Complete Phase 1: Setup
1. Complete Phase 2: Foundational (CRITICAL — blocks all stories)
1. Complete Phase 3: US1 + US4 (package interactive + overwrite guard)
1. **STOP and VALIDATE**: `aipkg init` creates package manifests, refuses on existing file
1. This is a usable MVP — package authors can create manifests

### Incremental Delivery

1. Setup + Foundational -> Foundation ready
1. US1 + US4 -> Package interactive works -> **MVP!**
1. US2 -> Project interactive works -> Both types supported
1. US3 -> Non-interactive + hybrid -> CI/scripting ready
1. Polish -> Help text, integration tests, full validation

______________________________________________________________________

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- US4 (overwrite guard) is folded into US1 since it's a single check at the start of the init command's RunE
- All user stories share `internal/cli/init.go` — they build on each other sequentially
- Foundational packages (manifest, schema, license) are independently testable via unit tests
- Integration tests (init_test.go) cover the full command via flag-only mode (no TTY needed)
- Commit after each task or logical group
