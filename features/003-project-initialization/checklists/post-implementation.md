# Post-Implementation Checklist: Project Initialization & Model

**Purpose**: Verify implementation correctness, cross-spec consistency, and documentation completeness before opening the implementation PR. Includes deferred item review from the pre-implementation audit.
**Created**: 2026-03-08
**Feature**: [spec.md](../spec.md)

## Implementation Correctness

- [ ] POST001 Does `aipkg init` create `aipkg-project.json` with exactly `specVersion: 1` and empty `require: {}`? (FR-015, FR-004)
- [ ] POST002 Does the created file use two-space JSON indent with trailing newline, matching the `create` command pattern? (DD-005, T003)
- [ ] POST003 Does `aipkg init` print a success confirmation message following the `create` command pattern? (DD-005)
- [ ] POST004 Does `aipkg init` NOT create `.aipkg/` or any other directories or files beyond `aipkg-project.json`? (FR-018, FR-010)
- [ ] POST005 Does the re-init guard trigger on file existence only, not content validity? An `aipkg-project.json` with invalid JSON still blocks re-init. (FR-016, Edge Cases)
- [ ] POST006 Does the mutual exclusivity guard trigger on file existence only? An `aipkg.json` with invalid JSON still blocks init. (FR-017, Edge Cases)
- [ ] POST007 Is the mutual exclusivity check (aipkg.json) performed before the re-init check (aipkg-project.json)? (FR-002, quickstart.md:41-44)
- [ ] POST008 Does the mutual exclusivity error message mention both `aipkg require` and `aipkg install` as alternatives? (FR-017, US3 Acceptance Scenario 3)
- [ ] POST009 Do error messages follow existing CLI patterns: lowercase, no period, wrapped with context? (DD-005)

## Schema Correctness

- [ ] POST010 Does `spec/schema/project.json` use Draft 2020-12 and follow the structure/style of `spec/schema/package.json`? (FR-005)
- [ ] POST011 Does the schema enforce `additionalProperties: false` on the root object, rejecting identity fields and any unknown properties? (FR-003)
- [ ] POST012 Does the schema's `require` key pattern match the package manifest's `require` key pattern exactly (same regex)? (CHK018, Assumption 2)
- [ ] POST013 Does the schema accept pre-release versions (`1.0.0-beta.1`, `1.0.0-rc.1`, `1.0.0-0.3.7`) but reject build metadata (`1.0.0+build.1`)? (FR-004, DD-002, R-002)
- [ ] POST014 Does `ValidateProject()` reuse the existing regexp2 engine setup and compiler infrastructure from `internal/schema`? (DD-003)
- [ ] POST015 Is `ProjectSchemaJSON` embedded in `specdata.go` alongside `PackageSchemaJSON`? (T002)

## Documentation Completeness

- [ ] POST016 Does `spec/project.md` cover the project file format with field definitions and examples (empty and with dependencies)? (FR-019)
- [ ] POST017 Does `spec/project.md` document the install directory layout: `.aipkg/` structure, four individual-type subdirectories, two merged root-level files? (FR-006, FR-007, FR-008)
- [ ] POST018 Does `spec/project.md` document the three-segment scoped naming convention (`scope.package-name.artifact-name`) with parsing rules and examples? (FR-012, FR-013, DD-001)
- [ ] POST019 Does `spec/project.md` document the `.gitignore` behavior (content: `*` + `!.gitignore`, git-only, created at install time)? (FR-009, FR-010)
- [ ] POST020 Does `spec/project.md` document the merged file ownership model (aipkg-managed, overwritten on install/update, manual edits lost)? (FR-011)
- [ ] POST021 Does `spec/project.md` document the mutual exclusivity rule in both directions (init refuses when aipkg.json exists, and future package creation commands should refuse when aipkg-project.json exists)? (FR-002, CHK021)
- [ ] POST022 Does `spec/naming.md` document three-segment install directory naming alongside the existing two-segment adapter convention, noting that adapters may use a shorter form? (DD-001, R-001, CHK017)

## Cross-Spec Consistency

- [ ] POST023 Are the artifact type directory names in documentation (`skills/`, `prompts/`, `commands/`, `agents/`) consistent with `spec/artifacts.md` and the `type` enum in `spec/schema/package.json`? (CHK020)
- [ ] POST024 Is the `specVersion` field definition consistent between the project schema and the package manifest schema (integer, const 1)? (CHK006)
- [ ] POST025 Does `LoadFile()` exist alongside `Create()` in `internal/project`, with rationale documented in DD-006? (DD-006, CHK028)

## Deferred Items Review

These items were deferred during the pre-implementation audit as install-command scope. Confirm the deferral is still appropriate or note if implementation revealed a need to address now.

- [ ] POST026 **Git detection method** (CHK010): How to detect "within a git working tree" is unspecified. Did implementation or documentation work reveal that this needs definition now?
- [ ] POST027 **Merged file formats** (CHK011): `mcp.json` JSON structure and `agent-instructions.md` marker format are unspecified. Still appropriate for deferral to install command?
- [ ] POST028 **Installed artifact file extensions** (CHK015): Relationship between scoped name and file extension is unspecified (`.ext` placeholder). Still appropriate for deferral?
- [ ] POST029 **Installed skill directory structure** (CHK016): Whether installed skills preserve original directory structure is unspecified. Still appropriate for deferral?

## Test Coverage

- [ ] POST030 Do tests cover all four US1 acceptance scenarios (empty dir, existing files, no .aipkg/ created, correct JSON structure)?
- [ ] POST031 Do tests cover both US2 acceptance scenarios (error on re-init, existing file unchanged)?
- [ ] POST032 Do tests cover all three US3 acceptance scenarios (error on package dir, no files created, error suggests commands)?
- [ ] POST033 Do tests cover the edge cases from spec.md (write permissions, invalid existing JSON for both file types)?
- [ ] POST034 Do schema validation tests cover SC-005 (valid files pass, invalid files fail: bad names, bad versions, extra fields, missing fields, build metadata rejected)?
- [ ] POST035 Does `LoadFile()` have roundtrip test coverage (create, load, verify structure)? (DD-006)

## Notes

- Items POST026-POST029 originate from the pre-implementation checklist (CHK010, CHK011, CHK015, CHK016). Resolution notes should explain whether deferral is still appropriate.
- If any deferred item needs action, create a Linear issue and note the issue ID next to the checklist item.
- Cross-spec consistency items (POST023-POST025) ensure the implementation doesn't drift from the existing specification foundation.
