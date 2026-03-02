# Post-Implementation Checklist: Archive Format & Pack Command

**Purpose**: Verify implementation quality and ensure deferred decisions from the pre-implementation audit are tracked or resolved.
**Created**: 2026-03-02
**Feature**: [spec.md](../spec.md)

## Implementation Correctness

- [ ] POST001 Does `aipkg pack` produce a valid zip archive with deflate compression and UTF-8 filenames? (FR-001)
- [ ] POST002 Does the archive contain only the manifest and artifact content (no stray root files)? (FR-002)
- [ ] POST003 Is the filename convention `{scope}--{name}-{version}.aipkg` correctly implemented? (FR-003)
- [ ] POST004 Does the sidecar file use `sha256sum -c` compatible format (lowercase hex, two spaces, basename, LF)? (FR-006)
- [ ] POST005 Does manifest validation failure abort before artifact discovery? (FR-010)
- [ ] POST006 Are hidden files, symlinks, and nested subdirectories in file-based dirs correctly skipped? (FR-008)
- [ ] POST007 Are entire skill directories (including scripts/, references/, assets/) included in the archive? (FR-012)
- [ ] POST008 Is `aipkg.json` protected from `.aipkgignore` exclusion? (FR-027)
- [ ] POST009 Does the archive filename parsing algorithm appear in `spec/archive.md`? (FR-003, SC-005)

## Cross-Spec Consistency

- [ ] POST010 Does `scaffold.WellKnownDirs` in the codebase match the six directories in FR-008 and `spec/artifacts.md`? (from CHK027)
- [ ] POST011 Are `spec/artifacts.md` updates shipped (relaxed file extensions, compound extension name derivation)?
- [ ] POST012 Is `spec/archive.md` self-contained enough for a third party to produce/consume archives without reading source code? (SC-005)

## Deferred Items Review

These items were consciously deferred during the pre-implementation audit. For each, confirm the deferral is still appropriate or note if the implementation revealed a need to address it now.

- [ ] POST013 **Consumer validation** (CHK003, CHK007): Archive consumer behavior (zero/multiple top-level dirs, dir name vs manifest mismatch) is deferred to unpack/install. Is this still appropriate, or did implementation reveal consumer-side concerns that should be captured as a Linear issue?
- [ ] POST014 **Deep frontmatter validation** (CHK017): Optional SKILL.md fields (license, compatibility, metadata, allowed-tools) are only type-checked by yaml.v3, not semantically validated (e.g., SPDX for license). Is this still appropriate?
- [ ] POST015 **Artifact count limit** (CHK037): No upper bound on artifact count. Did testing with large packages reveal any practical issues?
- [ ] POST016 **File permission preservation** (CHK039): Archive uses Go's default zip permission behavior. Did cross-platform testing reveal any issues?
- [ ] POST017 **Binary file detection** (CHK041): No content-type validation for file-based artifacts beyond non-empty. Is this still acceptable, or should a warning be added?

## Test Coverage

- [ ] POST018 Do tests cover all acceptance scenarios from spec.md User Stories 1-3?
- [ ] POST019 Do tests cover the edge cases listed in spec.md (missing manifest, schema failure, invalid names, empty files, malformed ignore patterns, etc.)?
- [ ] POST020 Is there a test that verifies `sha256sum -c` compatibility of the sidecar file?

## Notes

- Items POST013-POST017 originate from the pre-implementation checklist (CHK003, CHK007, CHK017, CHK037, CHK039, CHK041). Resolution notes explain the original deferral rationale in `checklists/pre-implementation.md`.
- If any deferred item needs action, create a Linear issue and note the issue ID next to the checklist item.
