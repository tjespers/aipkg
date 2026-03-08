# Tasks: Project Initialization & Model

**Input**: Design documents from `/features/003-project-initialization/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Create the project file JSON Schema and embed it in the binary. Everything else validates against this.

- [ ] T001 Create JSON Schema for `aipkg-project.json` in `spec/schema/project.json`. Schema uses Draft 2020-12. Two required fields: `specVersion` (integer, const 1) and `require` (object). `require` keys use the same scoped name pattern as `spec/schema/package.json`. `require` values use the SemVer pre-release regex from research.md R-002. `additionalProperties: false`. Reference `spec/schema/package.json` for structure and style. [FR-004, FR-005, DD-002]
- [ ] T002 Add `ProjectSchemaJSON` embed in `specdata.go` alongside `PackageSchemaJSON`. Use `//go:embed spec/schema/project.json`. Follow the existing pattern. [DD-003]

**Checkpoint**: Schema exists and compiles into the binary. Run `task build` to verify.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core packages that MUST be complete before the init command can be implemented.

- [ ] T003 [P] Create `internal/project/project.go` with `ProjectFile` type (`SpecVersion int`, `Require map[string]string`), `Create(dir string) error` (writes `aipkg-project.json` with `specVersion: 1` and empty `require`), and `LoadFile(dir string) (*ProjectFile, error)` (reads and parses). Mirror the patterns in `internal/manifest/manifest.go`: `MarshalIndent()` with trailing newline, `json.MarshalIndent` with two-space indent, `0o644` permissions. [FR-015, DD-006]
- [ ] T004 [P] Add `ValidateProject()` to `internal/schema/validate.go`. Add a second `sync.Once` + compiled schema pair for the project schema (import `aipkg.ProjectSchemaJSON`). Reuse the existing `regexp2Compile` engine. Follow the same pattern as `compiled()` and `Validate()`. [DD-003, SC-005]
- [ ] T005 [P] Write unit tests in `internal/project/project_test.go`: create/load roundtrip in `t.TempDir()`, verify JSON structure (specVersion: 1, empty require), verify trailing newline, verify `LoadFile` returns correct types. [DD-006]
- [ ] T006 [P] Write schema validation tests in `internal/schema/validate_test.go`: valid empty project file passes, valid project with dependencies passes (including pre-release versions like `1.0.0-beta.1`), invalid cases fail (missing specVersion, missing require, extra fields, bad package names, bad version strings, build metadata rejected). [SC-005, DD-002]

**Checkpoint**: `task test` passes. Schema validation correctly accepts valid and rejects invalid project files.

---

## Phase 3: User Story 1 - Initialize a new project (Priority: P1)

**Goal**: A developer runs `aipkg init` and gets a valid `aipkg-project.json` with an empty dependency map.

**Independent Test**: Run `aipkg init` in an empty `t.TempDir()`, verify exactly one file created with correct structure.

### Implementation

- [ ] T007 [US1] Create `internal/cli/init.go` with `newInitCmd()` returning `*cobra.Command` (Use: "init", Short: "Initialize a new aipkg project", Args: cobra.NoArgs) and `runInit(cmd *cobra.Command) error`. Happy path only: call `project.Create(".")` and print success message (e.g., `Initialized project in aipkg-project.json`) per DD-005. Follow the command factory pattern from `internal/cli/create.go`. [FR-015, FR-018, DD-005]
- [ ] T008 [US1] Register init command in `internal/cli/root.go`: add `cmd.AddCommand(newInitCmd())` alongside the existing `create` and `pack` commands.
- [ ] T009 [US1] Write integration tests in `internal/cli/init_test.go` for the happy path: init in empty directory creates `aipkg-project.json`, file contains valid JSON with specVersion 1 and empty require, no `.aipkg/` directory created, no other files created, init in directory with existing non-aipkg files creates only `aipkg-project.json`. [US1 Acceptance Scenarios 1-4]

**Checkpoint**: `task check` passes. `aipkg init` works in an empty directory. US1 acceptance scenarios verified.

---

## Phase 4: User Story 2 - Prevent accidental re-initialization (Priority: P2)

**Goal**: `aipkg init` refuses to overwrite an existing `aipkg-project.json`.

**Independent Test**: Create a project file with dependencies, run `aipkg init`, verify file unchanged and error displayed.

### Implementation

- [ ] T010 [US2] Add re-initialization guard to `runInit()` in `internal/cli/init.go`: before calling `project.Create()`, check if `aipkg-project.json` exists via `os.Stat`. If it exists, return an error (e.g., `project already initialized (aipkg-project.json exists)`). Check is file-existence only, not content validation. [FR-016, DD-004, DD-005]
- [ ] T011 [US2] Add integration tests in `internal/cli/init_test.go` for re-init guard: init in directory with existing `aipkg-project.json` returns error, existing file is not modified (write file with known content, run init, verify content unchanged), error message indicates project already initialized. [US2 Acceptance Scenarios 1-2, SC-006]

**Checkpoint**: `task check` passes. Re-initialization correctly blocked. Existing files never modified.

---

## Phase 5: User Story 3 - Refuse initialization in a package directory (Priority: P2)

**Goal**: `aipkg init` refuses when `aipkg.json` exists and suggests the right commands.

**Independent Test**: Create a directory with `aipkg.json`, run `aipkg init`, verify refusal with actionable error.

### Implementation

- [ ] T012 [US3] Add mutual exclusivity guard to `runInit()` in `internal/cli/init.go`: before the re-init check, check if `aipkg.json` exists via `os.Stat`. If it exists, return an error explaining that a package manifest exists and suggesting `aipkg require` or `aipkg install` as alternatives. Check order: mutual exclusivity first (FR-017), then re-init guard (FR-016). [FR-002, FR-017, DD-005]
- [ ] T013 [US3] Add integration tests in `internal/cli/init_test.go` for mutual exclusivity guard: init in directory with `aipkg.json` returns error, no `aipkg-project.json` created, error message mentions package manifest, error message suggests `aipkg require` or `aipkg install`. [US3 Acceptance Scenarios 1-3, SC-003]

**Checkpoint**: `task check` passes. All three user stories work. All guard paths tested.

---

## Phase 6: Reference Documentation

**Purpose**: Ship user-facing documentation with the feature per constitution standard and FR-019. Documents the complete project model including naming decisions finalized during planning.

- [ ] T014 [P] Write `spec/project.md` covering: project file format (fields, examples, validation), install directory layout (`.aipkg/` structure, subdirectories, merged files), scoped artifact naming convention (three-segment `scope.package.artifact` format with parsing rules and examples), `.gitignore` behavior, ownership model for merged files, and mutual exclusivity rule (FR-002 in both directions). Reference `spec/schema/project.json` for the machine-readable schema. Follow the style of existing spec docs (`spec/manifest.md`, `spec/artifacts.md`). [FR-006 through FR-014, FR-019, DD-001]
- [ ] T015 [P] Update `spec/naming.md` dot-notation section to document the three-segment install directory naming (`scope.package-name.artifact-name`) alongside the existing two-segment adapter convention. Add examples showing how a package like `@alice/blog-tools` with artifact `code-review` maps to `alice.blog-tools.code-review` in `.aipkg/`. Note that adapters may use a shorter form if their target tool has its own namespacing. [DD-001, Research R-001]

**Checkpoint**: `spec/project.md` and `spec/naming.md` are complete. Reference documentation covers all FR-006 through FR-014 requirements.

---

## Phase 7: Validation

**Purpose**: Full build and quality check.

- [ ] T016 Run `task check` (lint + vet + test) to verify all code passes
- [ ] T017 Run `task build` and verify `aipkg init` works end-to-end: init in empty directory, re-init guard, mutual exclusivity guard

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies. Start immediately.
- **Phase 2 (Foundational)**: Depends on Phase 1 (schema must exist for embedding and validation).
- **Phase 3 (US1)**: Depends on Phase 2 (`internal/project` must exist for `Create()`).
- **Phase 4 (US2)**: Depends on Phase 3 (init command must exist to add guard).
- **Phase 5 (US3)**: Depends on Phase 4 (guard ordering matters: mutual exclusivity before re-init).
- **Phase 6 (Docs)**: Can start after Phase 2 (all naming decisions are finalized). Independent of Phases 3-5.
- **Phase 7 (Validation)**: Depends on all previous phases.

### User Story Dependencies

- **US1 (P1)**: Depends on Phase 2 only. No cross-story dependencies.
- **US2 (P2)**: Depends on US1 (adds guard to existing command handler).
- **US3 (P2)**: Depends on US2 (guard ordering: FR-017 check before FR-016 check).

### Parallel Opportunities

**Phase 2** (after schema is embedded):
```text
T003 internal/project/project.go
T004 internal/schema/validate.go (ValidateProject)
T005 internal/project/project_test.go
T006 internal/schema/validate_test.go
```
All four tasks modify different files and can run in parallel.

**Phase 6** (after Phase 2):
```text
T014 spec/project.md
T015 spec/naming.md
```
Both docs tasks can run in parallel with each other and with Phases 3-5.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Schema + embed
2. Complete Phase 2: Project package + validation
3. Complete Phase 3: Init command happy path
4. **STOP and VALIDATE**: `aipkg init` works in empty directory
5. Commit and verify

### Incremental Delivery

1. Phase 1 + 2 → Foundation ready (commit)
2. Phase 3 (US1) → Init works (commit)
3. Phase 4 (US2) → Re-init guard (commit)
4. Phase 5 (US3) → Mutual exclusivity guard (commit)
5. Phase 6 → Reference docs (commit)
6. Phase 7 → Final validation

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] labels map to spec.md user stories: US1 (P1), US2 (P2), US3 (P2)
- US2 and US3 are both P2 but sequential because they modify the same file and guard ordering matters
- No new dependencies introduced. All libraries already in go.mod.
- Commit after each phase or logical group
- Reference docs (Phase 6) can be written in parallel with command implementation (Phases 3-5)
