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

- [ ] T001 Copy manifest schema from `../aipkg-spec/schema/aipkg.json` to `internal/schema/aipkg.schema.json`
- [ ] T002 Add Go dependencies: `go get github.com/spf13/cobra github.com/spf13/viper github.com/charmbracelet/huh github.com/santhosh-tekuri/jsonschema/v6 github.com/google/licensecheck golang.org/x/term`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core packages that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T003 [P] Implement Manifest struct with JSON serialization in `internal/manifest/manifest.go` — struct with 5 fields (Type, Name, Version, Description, License), `omitempty` on all except Type, `MarshalJSON` using `json.MarshalIndent(m, "", "  ")` with trailing newline appended
- [ ] T004 [P] Implement schema embedding and per-field validators in `internal/schema/schema.go` — `go:embed aipkg.schema.json`, compile schema with `sync.Once`, export `ValidateName(string) error`, `ValidateVersion(string) error`, `ValidateDescription(string) error` using regex patterns extracted from the compiled schema, and `ValidateManifest([]byte) error` for full schema validation of assembled JSON
- [ ] T005 [P] Implement LICENSE file SPDX detection in `internal/license/detect.go` — read `LICENSE` file in a given directory, run `licensecheck.Scan()`, return SPDX ID if confidence > 80%, empty string otherwise
- [ ] T006 Write unit tests for Manifest serialization in `internal/manifest/manifest_test.go` — table-driven tests: package with all fields, project with only type, omitempty behavior, 2-space indent, trailing newline, field order (depends on T003)
- [ ] T007 Write unit tests for schema validators in `internal/schema/schema_test.go` — table-driven tests: valid/invalid names (scoped, unscoped, consecutive hyphens, length limits), valid/invalid versions (semver, partial, leading zeros), description length, full manifest validation (depends on T004)
- [ ] T008 Write unit tests for license detection in `internal/license/detect_test.go` — table-driven tests using `os.MkdirTemp`: Apache-2.0 LICENSE file, MIT LICENSE file, no LICENSE file, unrecognized content (depends on T005)
- [ ] T009 Implement root cobra command in `internal/cli/root.go` — `NewRootCmd()` returning `*cobra.Command`, `PersistentPreRunE` for viper initialization, `SilenceErrors: true`, `SilenceUsage: true`, version flag using ldflags variables
- [ ] T010 Update entry point in `cmd/aipkg/main.go` — replace current `fmt.Printf` with `cli.NewRootCmd().Execute()`, call `os.Exit(1)` on error, preserve ldflags variables (version, commit, date)

**Checkpoint**: Foundation ready — all packages tested, root command works, `aipkg --version` functional

---

## Phase 3: User Story 1 + User Story 4 — Package Interactive + Overwrite Guard (Priority: P1) 🎯 MVP

**Goal**: Package author can create a manifest through interactive prompts. Existing manifests are never overwritten.

**Independent Test**: Run `aipkg init` in an empty directory, select "package", provide valid fields → verify `aipkg.json` created. Run again in same directory → verify error and file unchanged.

### Implementation

- [ ] T011 [US1] Implement init command scaffold in `internal/cli/init.go` — `newInitCmd()` returning `*cobra.Command` with `RunE`, define 5 string flags (type, name, version, description, license) bound via viper, check `aipkg.json` existence at start and return `fmt.Errorf("aipkg.json already exists")` if found (US4 guard)
- [ ] T012 [US1] Implement package interactive flow in `internal/cli/init.go` — build `huh.Form` dynamically: type select (project/package), then for package: name input with `ValidateName`, version input with `ValidateVersion` and default `0.1.0`, description input with `ValidateDescription` (optional, skippable), license input with detected default from `license.Detect()` (optional, skippable); skip any field already provided via flag; handle `huh.ErrUserAborted` for Ctrl+C
- [ ] T013 [US1] Implement manifest assembly and file write in `internal/cli/init.go` — build `Manifest` struct from collected values, omit empty optional fields, call `ValidateManifest` on serialized JSON, write to `aipkg.json` with `os.WriteFile` (0644 permissions), print `Created aipkg.json (package)` to stdout, return exit code 0
- [ ] T014 [US1] Register init subcommand in `internal/cli/root.go` — add `rootCmd.AddCommand(newInitCmd())` in `NewRootCmd()`

**Checkpoint**: `aipkg init` creates package manifests interactively, refuses on existing file, handles Ctrl+C

---

## Phase 4: User Story 2 — Project Interactive (Priority: P2)

**Goal**: Project maintainer can create a minimal manifest with optional name and description.

**Independent Test**: Run `aipkg init`, select "project", skip all prompts → verify `aipkg.json` contains only `{"type": "project"}`. Run with name and description → verify both included.

### Implementation

- [ ] T015 [US2] Add project prompt flow to `internal/cli/init.go` — when type is "project": prompt for name (optional, validated with `ValidateName` if non-empty) and description (optional, validated with `ValidateDescription` if non-empty); both skippable via Enter; assemble manifest with only non-empty fields; print `Created aipkg.json (project)` to stdout

**Checkpoint**: Both package and project interactive flows work independently

---

## Phase 5: User Story 3 — Non-Interactive & Hybrid Mode (Priority: P3)

**Goal**: CI pipelines can create manifests via flags without interactive prompts. Partial flags trigger prompts only for missing fields.

**Independent Test**: Run `aipkg init --type package --name @myorg/my-tool --version 0.1.0` → verify no prompts, valid file created. Run with `--type package --name @myorg/my-tool` (missing version) in TTY → verify prompted only for version.

### Implementation

- [ ] T016 [US3] Add TTY detection and non-interactive error handling to `internal/cli/init.go` — use `term.IsTerminal(int(os.Stdin.Fd()))` to detect TTY; when no TTY and required fields missing, print `Error: missing required fields: <field list> (non-interactive mode)` to stderr and exit 1; when all required fields provided via flags, skip all prompts entirely
- [ ] T017 [US3] Add irrelevant flag warnings to `internal/cli/init.go` — when `--type project` is set and `--version` or `--license` flags are provided, print `Warning: --version is ignored for project type` (and/or `--license`) to stderr, then proceed normally

**Checkpoint**: All three modes work: fully interactive, fully non-interactive, hybrid. TTY detection correct.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Help text, integration tests, and end-to-end validation

- [ ] T018 Add usage examples to init command help in `internal/cli/init.go` — set `Example` field on `cobra.Command` with examples from quickstart.md (package interactive, project, non-interactive, hybrid)
- [ ] T019 Write integration tests for init command in `internal/cli/init_test.go` — table-driven tests using `os.MkdirTemp`: package creation via flags (all fields, required only), project creation via flags (minimal, with optional fields), overwrite guard, invalid name/version validation errors, non-interactive missing fields error, irrelevant flag warnings, omitempty behavior for skipped optional fields, non-writable directory error
- [ ] T020 Run `task check` (lint + vet + test) and validate all quickstart.md scenarios manually

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - US1+US4 (Phase 3) must complete before US2 (Phase 4) since project flow extends init.go
  - US2 (Phase 4) must complete before US3 (Phase 5) since non-interactive mode covers both types
- **Polish (Phase 6)**: Depends on all user stories being complete

### Within Each Phase

- Phase 2: T003, T004, T005 are [P] (different files, no deps); T006–T008 depend on T003–T005 respectively (tests after implementations); T009 depends on T002; T010 depends on T009
- Phase 3: T011 → T012 → T013 (sequential within init.go); T014 depends on T011
- Phase 4: T015 depends on Phase 3 completion
- Phase 5: T016 → T017 (sequential, same file)
- Phase 6: T018, T019 are parallel; T020 depends on T019

### Parallel Opportunities

- Phase 2: T003, T004, T005 can run in parallel (3 implementation files); then T006, T007, T008 can run in parallel (3 test files, after their implementations)
- Phase 6: T018 and T019 can run in parallel (different concerns in different files)

---

## Parallel Example: Phase 2 Foundational

```
# Launch 3 implementation packages in parallel:
Task: "Implement Manifest struct in internal/manifest/manifest.go"
Task: "Implement schema validators in internal/schema/schema.go"
Task: "Implement license detection in internal/license/detect.go"

# Then launch 3 test files in parallel (after implementations):
Task: "Unit tests for manifest in internal/manifest/manifest_test.go"
Task: "Unit tests for schema in internal/schema/schema_test.go"
Task: "Unit tests for license in internal/license/detect_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 + 4 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories)
3. Complete Phase 3: US1 + US4 (package interactive + overwrite guard)
4. **STOP and VALIDATE**: `aipkg init` creates package manifests, refuses on existing file
5. This is a usable MVP — package authors can create manifests

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. US1 + US4 → Package interactive works → **MVP!**
3. US2 → Project interactive works → Both types supported
4. US3 → Non-interactive + hybrid → CI/scripting ready
5. Polish → Help text, integration tests, full validation

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- US4 (overwrite guard) is folded into US1 since it's a single check at the start of the init command's RunE
- All user stories share `internal/cli/init.go` — they build on each other sequentially
- Foundational packages (manifest, schema, license) are independently testable via unit tests
- Integration tests (init_test.go) cover the full command via flag-only mode (no TTY needed)
- Commit after each task or logical group
