# Pre-Implementation Checklist: Archive Format & Pack Command

**Purpose**: Thorough requirements quality audit before implementation. Weighted toward archive format specification clarity (SC-005: third-party interoperability) with cross-spec consistency checks against the foundation (001).
**Created**: 2026-03-02
**Feature**: [spec.md](../spec.md)

## Archive Format Specification Completeness

- [ ] CHK001 Is the zip compression method explicitly specified (deflate, store, or implementer's choice)? The assumptions mention deflate but FR-001 only says "zip file." A third-party implementer needs this. [Completeness, Spec §FR-001]
- [ ] CHK002 Are zip-level metadata requirements defined (e.g., UTF-8 filename encoding, Unix vs DOS timestamps)? Relevant for cross-platform interoperability. [Gap, Spec §FR-001]
- [ ] CHK003 Is the behavior defined when a consumer encounters an archive with zero top-level directories, or more than one? FR-002 says "a single top-level directory" for producers, but does the spec address consumer validation? [Coverage, Spec §FR-002]
- [ ] CHK004 Does the filename convention in FR-003 account for version strings that contain hyphens? The schema restricts versions to strict semver (no pre-release), but this constraint is not stated in FR-003 itself. A third-party producer reading only the archive spec could reasonably try `1.0.0-beta.1`. [Clarity, Spec §FR-003]
- [ ] CHK005 Is the algorithm for reconstructing scope, name, and version from the filename explicitly documented? FR-003 says they are "fully reconstructible" but does not provide the parsing procedure. SC-005 requires a third party to produce/consume archives from the spec alone. [Completeness, Spec §FR-003]
- [ ] CHK006 Are requirements for the top-level directory name defined as normative (MUST) consistently between FR-004 and FR-021? Both state the same rule in different sections. Is there a risk of divergence if one is updated without the other? [Consistency, Spec §FR-004 vs §FR-021]
- [ ] CHK007 Does FR-005 (extraction behavior) define what happens when the archive top-level directory name does not match the manifest's `name` field? For consumer robustness. [Coverage, Spec §FR-005]
- [ ] CHK008 Does FR-006 specify whether the filename in the sidecar is the basename only or includes a relative/absolute path? The assumptions section clarifies this (basename, two spaces), but FR-006 itself says only "sha256sum format." [Clarity, Spec §FR-006]
- [ ] CHK009 Is the sidecar file encoding specified (UTF-8, ASCII, LF line ending)? Relevant for cross-platform `sha256sum -c` compatibility. [Gap, Spec §FR-006]
- [ ] CHK010 Does the archive format specification address whether non-artifact files (e.g., README.md, LICENSE, CHANGELOG) at the package root should be included in the archive? The spec covers well-known directories and ignore rules, but is silent on loose root files. [Gap, Spec §FR-002]

## Artifact Discovery & Name Derivation Clarity

- [ ] CHK011 Is the name derivation rule "strip from first dot" consistent with the foundation spec's "filename without its extension" in `spec/artifacts.md` line 26? For compound extensions like `code-review.prompt.md`, one produces `code-review` (first dot) and the other could produce `code-review.prompt` (last extension). [Conflict, Spec §FR-019 vs artifacts.md §26]
- [ ] CHK012 Is the behavior defined for files with no extension (e.g., `prompts/review` with no dot)? FR-019's "strip from first dot" rule implies the full filename becomes the name, but this is not explicitly stated. [Clarity, Spec §FR-019]
- [ ] CHK013 Is the behavior defined for hidden files (starting with `.`) in well-known directories (e.g., `prompts/.draft`)? Stripping from the first dot produces an empty string, which violates naming rules. Should these be silently skipped or cause a validation error? [Gap, Spec §FR-019]
- [ ] CHK014 Are requirements specified for what happens when a well-known directory contains subdirectories for a file-based type (e.g., `prompts/drafts/review.md`)? Should nested files be discovered or only top-level files? [Gap, Spec §FR-008]
- [ ] CHK015 Does the spec define discovery behavior for non-regular files (symlinks, device files) in well-known directories? [Gap, Spec §FR-008]

## Type-Specific Validation Completeness

- [ ] CHK016 Are the SKILL.md frontmatter validation rules in FR-014 consistent with the Agent Skills specification referenced in `spec/artifacts.md`? The spec defines its own allowed keys (FR-015) rather than deferring to the external standard. Is this intentional, and is the relationship documented? [Consistency, Spec §FR-015 vs artifacts.md §104]
- [ ] CHK017 Does FR-014 specify validation for the optional frontmatter fields (license, compatibility, metadata, allowed-tools) beyond their presence? For example, should `compatibility` be a list of strings? Should `license` be a valid SPDX identifier? [Completeness, Spec §FR-014]
- [ ] CHK018 Is the behavior defined when SKILL.md has valid frontmatter but zero body content (only the `---` delimiters and YAML)? FR-013 requires "valid YAML frontmatter" and FR-018 checks "non-empty" for file-based types, but skills are directory-based. [Gap, Spec §FR-013]
- [ ] CHK019 Does FR-017 (MCP server JSON validation) specify whether an empty JSON object `{}` or empty array `[]` is considered valid? "Parses as JSON" technically passes for both. [Clarity, Spec §FR-017]
- [ ] CHK020 Is validation behavior defined for skill directories that contain extra files beyond SKILL.md (e.g., `skills/writer/scripts/`, `skills/writer/assets/`)? Should these be included in the archive? The Agent Skills spec allows them, but FR-012 only mentions SKILL.md. [Gap, Spec §FR-012]

## File Exclusion Requirements

- [ ] CHK021 Is the interaction between `.aipkgignore` and the `aipkg.json` manifest file itself defined? Can an author accidentally exclude the manifest? [Gap, Spec §FR-026]
- [ ] CHK022 Does FR-027 define the complete list of built-in defaults? It lists `.git/`, `.aipkgignore`, and the output archive. Are other common candidates intentionally excluded (e.g., `.DS_Store`, `node_modules/`, `features/`, `.specify/`)? [Completeness, Spec §FR-027]
- [ ] CHK023 Is the behavior defined when `.aipkgignore` is a directory instead of a file? [Edge Case, Spec §FR-026]
- [ ] CHK024 Does the spec define how the "output archive itself" is identified for self-exclusion when using `--output` to a custom path? The built-in default needs to know the archive path before the archive is created. [Clarity, Spec §FR-027]

## Cross-Spec Consistency (002 vs Foundation)

- [ ] CHK025 Does `spec/artifacts.md` line 14 ("Single markdown file" for prompts) conflict with FR-018 which allows `.txt`, `.prompt`, and `.prompt.md`? The foundation spec implies markdown-only; the 002 spec relaxes this. Which is authoritative, and does artifacts.md need updating? [Conflict, Spec §FR-018 vs artifacts.md §14]
- [ ] CHK026 Does `spec/naming.md` line 69 ("Must be unique within the package") conflict with the edge case on spec.md line 73 (same name in different well-known dirs is OK, distinguished by type)? If `skills/review/` and `prompts/review.md` both produce name `review`, is uniqueness violated? [Conflict, Spec §Edge Cases vs naming.md §69]
- [ ] CHK027 Is the well-known directory list in FR-008 consistent with `scaffold.WellKnownDirs` in the codebase and the table in `spec/artifacts.md`? All three should be a single source of truth. [Consistency, Spec §FR-008 vs artifacts.md §11]
- [ ] CHK028 Does the artifact naming regex in `spec/schema/package.json` (line 62: `^(?!.*--)[a-z0-9]([a-z0-9-]{0,62}[a-z0-9])?$`) match the prose rules in FR-019 and `spec/naming.md`? The regex allows single-character names; the prose says "1-64 characters." [Consistency, Spec §FR-019 vs schema]
- [ ] CHK029 Does FR-019's name derivation rule need to be reflected in `spec/artifacts.md` section "How artifact names are derived" (line 26)? Currently artifacts.md says "filename without its extension" (singular), which differs from "strip from first dot." [Consistency, Spec §FR-019 vs artifacts.md §26]

## Pipeline & Error Reporting

- [ ] CHK030 Is the order of validation steps (manifest validation before vs after artifact discovery) specified in the requirements? The CLI contract defines a specific pipeline order, but FR-010 only says "validate before packing" without positioning relative to FR-008 (discovery). [Clarity, Spec §FR-010]
- [ ] CHK031 Does FR-020 specify the format of collected validation errors? The CLI contract defines `{path}: {message}` format, but the spec requirement only says "report all errors." Is the error format a requirement or an implementation detail? [Clarity, Spec §FR-020]
- [ ] CHK032 Is the enriched manifest validation (CLI contract step 8, after injecting artifacts) specified as a requirement? The spec has FR-010 (validate before packing) but the pipeline validates twice. Is the second validation intentional and required? [Gap, Spec §FR-010]
- [ ] CHK033 Does the spec define whether validation errors from the initial manifest check (FR-010) prevent artifact discovery, or whether both validation phases run and all errors are collected? [Clarity, Spec §FR-010 vs §FR-020]

## Acceptance Criteria Measurability

- [ ] CHK034 Can SC-005 ("documented clearly enough that a third-party tool could produce or consume") be objectively measured? Is there a concrete test (e.g., "the spec/archive.md reference doc contains all information needed to produce a valid archive without reading source code")? [Measurability, Spec §SC-005]
- [ ] CHK035 Can SC-006 ("author familiar with npm pack or helm package can predict behavior") be objectively measured? This is subjective. Is there a concrete proxy (e.g., "flag names and behaviors match npm/helm conventions documented in SC-006 notes")? [Measurability, Spec §SC-006]
- [ ] CHK036 Does SC-004 specify what "identify the specific problem and location" means concretely? Is a file path sufficient, or should line numbers be included for frontmatter errors? [Clarity, Spec §SC-004]

## Edge Case & Boundary Coverage

- [ ] CHK037 Are requirements defined for the maximum number of artifacts in a package? FR-011 requires at least one; is there an upper bound? [Gap]
- [ ] CHK038 Is the behavior defined when the package name contains characters that are invalid in some filesystems (e.g., the `--` in the archive filename on legacy systems)? [Edge Case, Spec §FR-003]
- [ ] CHK039 Are requirements defined for file permission preservation in the archive? Should files retain their original permissions, or use normalized defaults? [Gap, Spec §FR-001]
- [ ] CHK040 Is the behavior defined when the source directory is the filesystem root or a path outside the user's home directory? [Edge Case, Spec §FR-007]
- [ ] CHK041 Are requirements defined for handling binary files in well-known directories (e.g., an image accidentally placed in `prompts/`)? FR-018 checks "non-empty" but not "valid text." [Gap, Spec §FR-018]

## Notes

- Focus areas: Archive format specification (heavier), pack command behavior, cross-spec consistency with 001
- Depth: Thorough (~30-40 items, catches subtle ambiguities)
- Audience: Author/reviewer pre-implementation gate
- Items reference spec sections and foundation docs where applicable
- `[Conflict]` markers indicate items that may require spec updates before implementation
- `[Gap]` markers indicate missing requirements that should be consciously included or explicitly deferred
