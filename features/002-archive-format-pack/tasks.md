# Tasks: Archive Format & Pack Command

**Input**: Design documents from `/features/002-archive-format-pack/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/cli-pack.md, quickstart.md

**Tests**: Included. Each implementation task includes colocated unit tests (Go `_test.go` convention). User story phases include integration tests for acceptance scenarios.

**Organization**: Tasks grouped by user story. Three stories from spec.md (P1 through P3), with Setup and Foundational phases before story work begins.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1 through US3)
- Exact file paths included in descriptions

---

## Phase 1: Setup

**Purpose**: New dependencies and project structure

- [X] T001 Add Go dependencies to go.mod: `gopkg.in/yaml.v3`, `github.com/sabhiram/go-gitignore`; run `go mod tidy`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared building blocks that user story implementations depend on. Manifest extensions, frontmatter extraction, artifact type system, and validation error collection.

**CRITICAL**: No user story work can begin until this phase is complete

### Manifest extensions

- [X] T002 Extend `internal/manifest/manifest.go`: add `Artifact` struct (Name, Type, Path with JSON tags), add `Artifacts []Artifact` field to `PackageManifest` with `omitempty`, implement `LoadFile(path string) (*PackageManifest, error)` to read and unmarshal existing `aipkg.json`; table-driven tests for LoadFile (valid manifest, missing file, invalid JSON, missing required fields) and round-trip marshal/unmarshal with artifacts in `internal/manifest/manifest_test.go`

### Frontmatter extraction

- [X] T003 [P] Implement generic frontmatter extraction in `internal/frontmatter/frontmatter.go`: `Extract(content []byte) (yamlBytes []byte, body []byte, err error)` that scans for opening `---` on first line, collects lines until closing `---`, returns YAML bytes and remaining body; table-driven tests for valid frontmatter, missing opening delimiter, unclosed delimiter, empty frontmatter, frontmatter-only (no body) in `internal/frontmatter/frontmatter_test.go`

### Artifact type system

- [X] T004 [P] Implement artifact type definitions in `internal/artifact/types.go`: `ArtifactType` string type with constants for all six types (skill, prompt, command, agent, agent-instructions, mcp-server), `DirToType` mapping from well-known directory names to artifact types, `TypeToDir` reverse mapping; tests for mapping completeness and consistency in `internal/artifact/types_test.go`

### Validation error collection

- [X] T005 [P] Implement `ValidationErrors` collector in `internal/artifact/errors.go`: struct with `Add(path, message string)`, `Addf(path, format string, args ...any)`, and `Err() error` (returns nil when empty, `errors.Join` of all collected errors when non-empty); each error formatted as `{path}: {message}`; tests for empty collector returns nil, single error, multiple errors, path formatting in `internal/artifact/errors_test.go`

**Checkpoint**: All foundational packages built and tested. Pack command implementation can begin.

---

## Phase 3: User Story 1 - Pack a package into a distributable archive (Priority: P1) MVP

**Goal**: `aipkg pack` discovers artifacts, validates them, and produces a `.aipkg` archive with SHA-256 sidecar. The core value of the feature.

**Independent Test**: Create a package directory with skills, prompts, and mcp-server artifacts. Run `aipkg pack`. Verify: archive exists with correct name, contains single top-level directory matching package name, manifest inside archive has correct `artifacts` array, original `aipkg.json` unchanged, sidecar verifies with `sha256sum -c`.

### Implementation for User Story 1

- [X] T006 [P] [US1] Implement artifact discovery in `internal/artifact/discover.go`: `Discover(rootDir string) ([]Artifact, error)` scans each well-known directory, for `skills/` treats each subdirectory as an artifact, for file-based types treats each file as an artifact, derives names per data-model.md rules (directory name for skills, filename stem before first `.` for files), skips empty well-known directories, returns error if a well-known directory exists but cannot be read; table-driven tests with `t.TempDir()` fixtures covering single-type packages, multi-type packages, empty well-known dirs, compound file extensions (`.prompt.md`), and name derivation edge cases in `internal/artifact/discover_test.go`
- [X] T007 [P] [US1] Implement skill-specific validation in `internal/artifact/skill.go`: `SkillFrontmatter` struct (per data-model.md), `validateSkill(rootDir string, art Artifact, errs *ValidationErrors)` that reads `SKILL.md`, calls `frontmatter.Extract`, unmarshals YAML with `yaml.v3 Decoder.KnownFields(true)` to reject unexpected keys (FR-015), validates required `name` and `description` fields (FR-014), validates `name` matches parent directory (FR-016), and validates name/description length constraints; tests covering valid skill, missing SKILL.md, missing name, missing description, name mismatch, unknown frontmatter key, invalid YAML in `internal/artifact/skill_test.go`
- [X] T008 [US1] Implement per-type validation dispatch in `internal/artifact/validate.go`: `ValidateAll(rootDir string, artifacts []Artifact) error` iterates all artifacts, dispatches to type-specific validators (skill -> validateSkill, mcp-server -> validateJSON, file-based types -> validateNonEmpty), validates artifact name against naming rules (FR-019) for all types, collects all errors via `ValidationErrors`, returns combined error or nil; tests covering valid multi-type package, mixed valid/invalid artifacts (errors collected not short-circuited), empty file, invalid JSON, invalid artifact name in `internal/artifact/validate_test.go` (depends on T005, T006, T007)
- [X] T009 [P] [US1] Implement zip archive creation in `internal/archive/zip.go`: `CreateArchive(w io.Writer, rootDir string, topLevelDir string, files []string) error` writes a zip with a single top-level directory, each file path prefixed with `topLevelDir/`, uses `zip.Deflate` compression, accepts a modified `aipkg.json` content (byte slice) to write instead of the on-disk version (FR-009: enriched manifest in archive only); tests with `t.TempDir()` verifying archive structure (top-level dir present, files at correct paths, content integrity) by unzipping and comparing in `internal/archive/zip_test.go`
- [X] T010 [P] [US1] Implement SHA-256 sidecar generation in `internal/archive/sha256.go`: `WriteSidecar(archivePath string) error` computes SHA-256 of the archive file, writes `{hex_hash}  {basename}\n` to `{archivePath}.sha256` in `sha256sum` format (two spaces between hash and filename, basename only); tests verifying hash correctness against known input and `sha256sum` format compliance in `internal/archive/sha256_test.go`
- [X] T011 [US1] Implement pack command orchestration in `internal/cli/pack.go`: `newPackCmd()` returning `*cobra.Command`, `runPack(cmd, args)` implementing the full pipeline from contracts/cli-pack.md (load manifest -> validate against schema -> discover artifacts -> validate artifacts -> build enriched manifest copy -> create zip -> write sidecar -> print summary to stderr), uses `naming.Parse` to extract scope and name for filename convention (`{scope}--{name}-{version}.aipkg`), positional arg for source directory (default `.`); integration tests with `t.TempDir()` covering acceptance scenarios AS-1 through AS-6 from spec.md, plus edge cases (missing aipkg.json, schema validation failure, no artifacts discovered) in `internal/cli/pack_test.go` (depends on T002, T006, T008, T009, T010)
- [X] T012 [US1] Register pack subcommand in `internal/cli/root.go`: add `newPackCmd()` to root command's subcommand list
- [X] T013 [P] [US1] Write archive format reference documentation in `spec/archive.md`: format overview (zip, .aipkg extension), filename convention with `--` separator and reconstructibility explanation, top-level directory naming (matches package name), extraction behavior, SHA-256 sidecar format, worked example per contracts/cli-pack.md archive structure
- [X] T014 [P] [US1] Update `spec/artifacts.md`: document relaxed file extensions for file-based artifact types (not restricted to `.md`, extensions like `.txt`, `.prompt`, `.prompt.md` are valid), update name derivation rules to specify "strip from first dot" for compound extensions, add examples

**Checkpoint**: Core pack command functional. Author can produce distributable archives from valid packages. All validation catches invalid packages before archive creation.

---

## Phase 4: User Story 2 - Exclude files from the archive (Priority: P2)

**Goal**: Authors can control what gets included in the archive via `.aipkgignore` and built-in defaults automatically exclude common non-distributable files.

**Independent Test**: Create a package with a `.aipkgignore` containing `*.log`, add a `.log` file, run `aipkg pack`. Verify: archive produced, `.log` file absent, `.git/` absent, `.aipkgignore` absent. Then test that ignore patterns can exclude well-known directory contents.

**Depends on**: US1 (the pack pipeline must exist before ignore filtering can be integrated)

### Implementation for User Story 2

- [X] T015 [US2] Implement ignore rules loading and path filtering in `internal/ignore/ignore.go`: `LoadRules(rootDir string, archivePath string) (*Rules, error)` loads `.aipkgignore` if present (via `go-gitignore`), appends built-in defaults (`.git/`, `.aipkgignore`, archive output path), exposes `Rules.IsExcluded(path string) bool`; tests covering no ignore file (defaults only), custom patterns, built-in defaults always applied, archive self-exclusion, malformed pattern error in `internal/ignore/ignore_test.go`
- [X] T016 [US2] Integrate ignore filtering into pack pipeline in `internal/cli/pack.go` and `internal/artifact/discover.go`: pass ignore rules to discovery so excluded files are omitted from both artifact list and archive contents, `.aipkgignore` patterns take precedence over well-known directory convention (FR-028); update pack_test.go with acceptance scenarios AS-1 through AS-3 from US2 spec (pattern exclusion, defaults without ignore file, ignore overriding well-known dirs)

**Checkpoint**: Ignore filtering works. Authors have full control over archive contents.

---

## Phase 5: User Story 3 - Control output location (Priority: P3)

**Goal**: Authors can specify where the archive is written via `--output`, supporting both directory and file paths.

**Independent Test**: Run `aipkg pack --output dist/` and verify archive lands in `dist/`. Run `aipkg pack --output dist/custom.aipkg` and verify exact path. Run with existing file and verify silent overwrite.

**Depends on**: US1 (the pack command must exist before output flag can be added)

### Implementation for User Story 3

- [X] T017 [US3] Add `--output` flag and path resolution to pack command in `internal/cli/pack.go`: `-o`/`--output` string flag, resolve output path (if trailing `/` or existing directory: use conventional filename inside that directory; if file path: use as-is), parent directory must exist (no `MkdirAll`), overwrite silently (FR-025); update pack_test.go with acceptance scenarios AS-1 through AS-3 from US3 spec (custom file path, directory path, overwrite existing)

**Checkpoint**: All three user stories complete. Pack command fully functional with ignore support and output control.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: End-to-end validation and quality checks

- [X] T018 Run quickstart.md end-to-end validation: execute all 10 steps from quickstart.md against the built binary, verify expected outputs match
- [X] T019 Run `task check` (lint + vet + test) and fix any issues: ensure all tests pass, no lint warnings, race detector clean

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies, start immediately
- **Foundational (Phase 2)**: Depends on Setup completion, BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Foundational completion
- **US2 (Phase 4)**: Depends on US1 completion (needs the pack pipeline to integrate into)
- **US3 (Phase 5)**: Depends on US1 completion (needs the pack command to add the flag to)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1)**: Can start after Foundational (Phase 2). No dependencies on other stories. This IS the MVP.
- **US2 (P2)**: Depends on US1. The ignore package is independent, but integration requires the pack pipeline. Can start the ignore package in parallel with US1 if desired, but integration waits for US1.
- **US3 (P3)**: Depends on US1. Small scope (flag + path resolution). Can start after US1 is complete.

### Within Each Phase

- Tasks marked [P] can run in parallel within their phase
- Non-[P] tasks have implicit ordering (listed in dependency order)
- All foundational tasks (T002-T005) are parallel
- Within US1: T006, T007 are parallel; T009, T010 are parallel; T013, T014 are parallel; T008 depends on T006+T007; T011 depends on everything above; T012 depends on T011

### Parallel Opportunities

```
Phase 2 (all parallel):
  T002 (manifest)  |  T003 (frontmatter)  |  T004 (types)  |  T005 (errors)

Phase 3 / US1 (two parallel waves, then sequential assembly):
  Wave 1: T006 (discover) | T007 (skill) | T009 (zip) | T010 (sha256) | T013 (spec/archive.md) | T014 (spec/artifacts.md)
  Wave 2: T008 (validate-all, needs T006+T007)
  Wave 3: T011 (pack command, needs T008+T009+T010) → T012 (register)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL, blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Run quickstart steps 1-8, verify archive creation and validation
5. This is a shippable increment: authors can pack valid packages

### Incremental Delivery

1. Setup + Foundational -> building blocks ready
2. Add US1 -> Core pack works, test independently -> shippable MVP
3. Add US2 -> Ignore filtering works, test independently -> better author experience
4. Add US3 -> Output control works, test independently -> CI-friendly
5. Polish -> End-to-end validation, lint clean

---

## Notes

- [P] tasks = different files, no dependencies on incomplete tasks
- [Story] label maps task to specific user story for traceability
- US2 and US3 both depend on US1 (unlike 001 where stories were more independent)
- Documentation tasks (T013, T014) are in US1 per constitution v1.1.0 documentation standard
- Each task includes colocated `_test.go` tests (Go convention)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
