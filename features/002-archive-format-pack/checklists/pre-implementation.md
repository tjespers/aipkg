# Pre-Implementation Checklist: Archive Format & Pack Command

**Purpose**: Thorough requirements quality audit before implementation. Weighted toward archive format specification clarity (SC-005: third-party interoperability) with cross-spec consistency checks against the foundation (001).
**Created**: 2026-03-02
**Resolved**: 2026-03-02
**Feature**: [spec.md](../spec.md)

## Archive Format Specification Completeness

- [X] CHK001 Is the zip compression method explicitly specified (deflate, store, or implementer's choice)? The assumptions mention deflate but FR-001 only says "zip file." A third-party implementer needs this. [Completeness, Spec §FR-001]
  > **Resolved**: FR-001 now explicitly specifies "deflate compression and UTF-8 filename encoding."

- [X] CHK002 Are zip-level metadata requirements defined (e.g., UTF-8 filename encoding, Unix vs DOS timestamps)? Relevant for cross-platform interoperability. [Gap, Spec §FR-001]
  > **Resolved**: FR-001 now specifies UTF-8 filename encoding. Timestamps are left to the implementation (Go's archive/zip uses reasonable defaults). The archive.md reference doc will note this for third-party implementers.

- [X] CHK003 Is the behavior defined when a consumer encounters an archive with zero top-level directories, or more than one? FR-002 says "a single top-level directory" for producers, but does the spec address consumer validation? [Coverage, Spec §FR-002]
  > **Deferred**: Consumer validation is out of scope for 002 (no unpack/install command). Will be addressed when the install/unpack feature is specified.

- [X] CHK004 Does the filename convention in FR-003 account for version strings that contain hyphens? The schema restricts versions to strict semver (no pre-release), but this constraint is not stated in FR-003 itself. A third-party producer reading only the archive spec could reasonably try `1.0.0-beta.1`. [Clarity, Spec §FR-003]
  > **Resolved**: FR-003 now explicitly states "versions are strict semver (MAJOR.MINOR.PATCH, no pre-release or build metadata)" and explains why this makes the filename unambiguous.

- [X] CHK005 Is the algorithm for reconstructing scope, name, and version from the filename explicitly documented? FR-003 says they are "fully reconstructible" but does not provide the parsing procedure. SC-005 requires a third party to produce/consume archives from the spec alone. [Completeness, Spec §FR-003]
  > **Resolved**: FR-003 now describes the parsing logic (split on `--` for scope, match version as digit-dot sequence from the right) and requires the specification reference document to include the full parsing algorithm.

- [X] CHK006 Are requirements for the top-level directory name defined as normative (MUST) consistently between FR-004 and FR-021? Both state the same rule in different sections. Is there a risk of divergence if one is updated without the other? [Consistency, Spec §FR-004 vs §FR-021]
  > **Resolved**: FR-021 now references FR-004 instead of restating the rule. Single source of truth.

- [X] CHK007 Does FR-005 (extraction behavior) define what happens when the archive top-level directory name does not match the manifest's `name` field? For consumer robustness. [Coverage, Spec §FR-005]
  > **Deferred**: Consumer validation is out of scope for 002. The pack command always produces correct archives. Consumer-side mismatch detection will be addressed in the install/unpack feature.

- [X] CHK008 Does FR-006 specify whether the filename in the sidecar is the basename only or includes a relative/absolute path? The assumptions section clarifies this (basename, two spaces), but FR-006 itself says only "sha256sum format." [Clarity, Spec §FR-006]
  > **Resolved**: FR-006 now explicitly says "archive basename" instead of just "filename."

- [X] CHK009 Is the sidecar file encoding specified (UTF-8, ASCII, LF line ending)? Relevant for cross-platform `sha256sum -c` compatibility. [Gap, Spec §FR-006]
  > **Resolved**: FR-006 now specifies "UTF-8 encoding with a LF line ending, terminated by a single newline."

- [X] CHK010 Does the archive format specification address whether non-artifact files (e.g., README.md, LICENSE, CHANGELOG) at the package root should be included in the archive? The spec covers well-known directories and ignore rules, but is silent on loose root files. [Gap, Spec §FR-002]
  > **Resolved**: FR-002 now explicitly states "Only the manifest and artifact content is included; non-artifact files at the package root (e.g., README.md, LICENSE) are not part of the archive in v1." This is a conscious design choice. The manifest's `license` field carries the SPDX identifier. Including additional root files can be revisited in a future version.

## Artifact Discovery & Name Derivation Clarity

- [X] CHK011 Is the name derivation rule "strip from first dot" consistent with the foundation spec's "filename without its extension" in `spec/artifacts.md` line 26? For compound extensions like `code-review.prompt.md`, one produces `code-review` (first dot) and the other could produce `code-review.prompt` (last extension). [Conflict, Spec §FR-019 vs artifacts.md §26]
  > **Already resolved**: artifacts.md was already updated to say "stripping everything from the first `.` onwards" with a compound extension example. Consistent with FR-019.

- [X] CHK012 Is the behavior defined for files with no extension (e.g., `prompts/review` with no dot)? FR-019's "strip from first dot" rule implies the full filename becomes the name, but this is not explicitly stated. [Clarity, Spec §FR-019]
  > **Resolved**: FR-019 now explicitly states "If a filename contains no `.`, the entire filename is the artifact name."

- [X] CHK013 Is the behavior defined for hidden files (starting with `.`) in well-known directories (e.g., `prompts/.draft`)? Stripping from the first dot produces an empty string, which violates naming rules. Should these be silently skipped or cause a validation error? [Gap, Spec §FR-019]
  > **Resolved**: FR-008 now explicitly states "Hidden entries (names starting with `.`) MUST be silently skipped."

- [X] CHK014 Are requirements specified for what happens when a well-known directory contains subdirectories for a file-based type (e.g., `prompts/drafts/review.md`)? Should nested files be discovered or only top-level files? [Gap, Spec §FR-008]
  > **Resolved**: FR-008 now explicitly states "Discovery MUST only consider top-level entries within each well-known directory" and "Nested subdirectories within file-based type directories (e.g., `prompts/drafts/review.md`) MUST NOT be discovered."

- [X] CHK015 Does the spec define discovery behavior for non-regular files (symlinks, device files) in well-known directories? [Gap, Spec §FR-008]
  > **Resolved**: FR-008 now explicitly states "Symlinks and other non-regular filesystem entries MUST be silently skipped."

## Type-Specific Validation Completeness

- [X] CHK016 Are the SKILL.md frontmatter validation rules in FR-014 consistent with the Agent Skills specification referenced in `spec/artifacts.md`? The spec defines its own allowed keys (FR-015) rather than deferring to the external standard. Is this intentional, and is the relationship documented? [Consistency, Spec §FR-015 vs artifacts.md §104]
  > **Intentional**: FR-015 defines aipkg's own allowed keys as a deliberate subset of the Agent Skills spec. This is consistent with Principle I (Simplicity and Deferral). The artifacts.md reference to the Agent Skills spec is for reader context, not normative delegation.

- [X] CHK017 Does FR-014 specify validation for the optional frontmatter fields (license, compatibility, metadata, allowed-tools) beyond their presence? For example, should `compatibility` be a list of strings? Should `license` be a valid SPDX identifier? [Completeness, Spec §FR-014]
  > **Deferred**: v1 validates optional fields for correct YAML types only (via yaml.v3 struct unmarshaling). Deep validation (SPDX license checks, compatibility value constraints) is deferred. The research.md `SkillFrontmatter` struct definition establishes the type contracts (`[]string` for compatibility, etc.).

- [X] CHK018 Is the behavior defined when SKILL.md has valid frontmatter but zero body content (only the `---` delimiters and YAML)? FR-013 requires "valid YAML frontmatter" and FR-018 checks "non-empty" for file-based types, but skills are directory-based. [Gap, Spec §FR-013]
  > **Acceptable**: A SKILL.md with valid frontmatter but no body is valid for aipkg purposes. The SKILL.md itself is non-empty (it has frontmatter). Body content quality is the skill author's responsibility. FR-018's non-empty check applies to file-based types, not skill directories.

- [X] CHK019 Does FR-017 (MCP server JSON validation) specify whether an empty JSON object `{}` or empty array `[]` is considered valid? "Parses as JSON" technically passes for both. [Clarity, Spec §FR-017]
  > **Already resolved**: FR-017 explicitly says "validates that the file parses as JSON but does not validate the JSON structure beyond that." This is intentional per Principle I (Simplicity and Deferral). `{}` and `[]` are valid JSON and pass validation.

- [X] CHK020 Is validation behavior defined for skill directories that contain extra files beyond SKILL.md (e.g., `skills/writer/scripts/`, `skills/writer/assets/`)? Should these be included in the archive? The Agent Skills spec allows them, but FR-012 only mentions SKILL.md. [Gap, Spec §FR-012]
  > **Resolved**: FR-012 now explicitly states "The entire skill directory is included in the archive (including optional subdirectories like `scripts/`, `references/`, and `assets/` as defined by the Agent Skills specification)."

## File Exclusion Requirements

- [X] CHK021 Is the interaction between `.aipkgignore` and the `aipkg.json` manifest file itself defined? Can an author accidentally exclude the manifest? [Gap, Spec §FR-026]
  > **Resolved**: FR-027 now explicitly states "the `aipkg.json` manifest MUST always be included in the archive and cannot be excluded by `.aipkgignore`."

- [X] CHK022 Does FR-027 define the complete list of built-in defaults? It lists `.git/`, `.aipkgignore`, and the output archive. Are other common candidates intentionally excluded (e.g., `.DS_Store`, `node_modules/`, `features/`, `.specify/`)? [Completeness, Spec §FR-027]
  > **Intentional**: The built-in defaults are intentionally minimal. The archive only contains the manifest and artifact content (FR-002), so `.DS_Store`, `node_modules/`, etc. at the package root are never included anyway. The `.aipkgignore` file is for excluding content within skill directories or specific artifact files.

- [X] CHK023 Is the behavior defined when `.aipkgignore` is a directory instead of a file? [Edge Case, Spec §FR-026]
  > **Implementation detail**: If `.aipkgignore` is a directory, the pack command treats it as if no ignore file exists (the file read will fail, falling back to defaults only). This is standard filesystem behavior and doesn't need spec-level definition.

- [X] CHK024 Does the spec define how the "output archive itself" is identified for self-exclusion when using `--output` to a custom path? The built-in default needs to know the archive path before the archive is created. [Clarity, Spec §FR-027]
  > **Implementation detail**: The output path is resolved from flags before file collection begins. The resolved absolute path is added to the exclusion list. This is straightforward implementation logic, not a spec requirement.

## Cross-Spec Consistency (002 vs Foundation)

- [X] CHK025 Does `spec/artifacts.md` line 14 ("Single markdown file" for prompts) conflict with FR-018 which allows `.txt`, `.prompt`, and `.prompt.md`? The foundation spec implies markdown-only; the 002 spec relaxes this. Which is authoritative, and does artifacts.md need updating? [Conflict, Spec §FR-018 vs artifacts.md §14]
  > **Already resolved**: The checklist misquotes artifacts.md. The table says "Single file" (not "Single markdown file") and the Format column says "Markdown/text". The prompts section (line 247) says "A single Markdown or plain text file." These are already consistent with FR-018.

- [X] CHK026 Does `spec/naming.md` line 69 ("Must be unique within the package") conflict with the edge case on spec.md line 73 (same name in different well-known dirs is OK, distinguished by type)? If `skills/review/` and `prompts/review.md` both produce name `review`, is uniqueness violated? [Conflict, Spec §Edge Cases vs naming.md §69]
  > **Already resolved**: naming.md line 69 says "Must be unique within the package **for a given artifact type**" (emphasis added). A skill named `review` and a prompt named `review` have different types. No conflict.

- [X] CHK027 Is the well-known directory list in FR-008 consistent with `scaffold.WellKnownDirs` in the codebase and the table in `spec/artifacts.md`? All three should be a single source of truth. [Consistency, Spec §FR-008 vs artifacts.md §11]
  > **Verify during implementation**: The spec and artifacts.md both list the same six directories. Consistency with `scaffold.WellKnownDirs` will be verified when implementing the artifact discovery package (which imports from scaffold).

- [X] CHK028 Does the artifact naming regex in `spec/schema/aipkg.json` (line 62: `^(?!.*--)[a-z0-9]([a-z0-9-]{0,62}[a-z0-9])?$`) match the prose rules in FR-019 and `spec/naming.md`? The regex allows single-character names; the prose says "1-64 characters." [Consistency, Spec §FR-019 vs schema]
  > **Already resolved**: The regex matches exactly 1-64 characters of `[a-z0-9-]` with no consecutive hyphens, no leading/trailing hyphens. Single-char names (e.g., `a`) are valid by the prose rules too ("1-64 characters"). No inconsistency.

- [X] CHK029 Does FR-019's name derivation rule need to be reflected in `spec/artifacts.md` section "How artifact names are derived" (line 26)? Currently artifacts.md says "filename without its extension" (singular), which differs from "strip from first dot." [Consistency, Spec §FR-019 vs artifacts.md §26]
  > **Already resolved**: artifacts.md was already updated to say "stripping everything from the first `.` onwards" with examples including compound extensions. Consistent with FR-019.

## Pipeline & Error Reporting

- [X] CHK030 Is the order of validation steps (manifest validation before vs after artifact discovery) specified in the requirements? The CLI contract defines a specific pipeline order, but FR-010 only says "validate before packing" without positioning relative to FR-008 (discovery). [Clarity, Spec §FR-010]
  > **Already resolved**: The CLI contract (contracts/cli-pack.md) defines the precise 11-step pipeline order. FR-010 now explicitly states that manifest validation failure prevents artifact discovery.

- [X] CHK031 Does FR-020 specify the format of collected validation errors? The CLI contract defines `{path}: {message}` format, but the spec requirement only says "report all errors." Is the error format a requirement or an implementation detail? [Clarity, Spec §FR-020]
  > **Already resolved**: The error format is defined in the CLI contract (the authoritative source for CLI behavior). FR-020 specifies the requirement (report all errors); the contract specifies the format. This separation is appropriate.

- [X] CHK032 Is the enriched manifest validation (CLI contract step 8, after injecting artifacts) specified as a requirement? The spec has FR-010 (validate before packing) but the pipeline validates twice. Is the second validation intentional and required? [Gap, Spec §FR-010]
  > **Acceptable**: FR-010 says "validate the manifest against the package JSON schema before packing." The enriched manifest (with artifacts array) is the manifest that gets packed. Validating it before packing satisfies FR-010. The two-phase validation is documented in the CLI contract.

- [X] CHK033 Does the spec define whether validation errors from the initial manifest check (FR-010) prevent artifact discovery, or whether both validation phases run and all errors are collected? [Clarity, Spec §FR-010 vs §FR-020]
  > **Resolved**: FR-010 now explicitly states "MUST NOT proceed to artifact discovery" on manifest validation failure and "Manifest validation errors are reported immediately without collecting further errors."

## Acceptance Criteria Measurability

- [X] CHK034 Can SC-005 ("documented clearly enough that a third-party tool could produce or consume") be objectively measured? Is there a concrete test (e.g., "the spec/archive.md reference doc contains all information needed to produce a valid archive without reading source code")? [Measurability, Spec §SC-005]
  > **Acceptable**: Measurable by inspection: the deliverable spec/archive.md reference doc must be self-contained (all format details, parsing algorithm, sidecar format) such that reading it alone is sufficient to implement a producer or consumer. Review during implementation.

- [X] CHK035 Can SC-006 ("author familiar with npm pack or helm package can predict behavior") be objectively measured? This is subjective. Is there a concrete proxy (e.g., "flag names and behaviors match npm/helm conventions documented in SC-006 notes")? [Measurability, Spec §SC-006]
  > **Acceptable**: SC-006 is a design principle, not a measurable test. It guides design decisions (flag naming, default behavior, output format) rather than serving as a pass/fail gate. This is documented in the constitution check (plan.md §III Convention Over Invention).

- [X] CHK036 Does SC-004 specify what "identify the specific problem and location" means concretely? Is a file path sufficient, or should line numbers be included for frontmatter errors? [Clarity, Spec §SC-004]
  > **Already resolved**: The CLI contract defines error format as `{relative-path}: {message}`. File paths are sufficient for v1. Line numbers for frontmatter errors would be a nice-to-have enhancement but are not required by the spec.

## Edge Case & Boundary Coverage

- [X] CHK037 Are requirements defined for the maximum number of artifacts in a package? FR-011 requires at least one; is there an upper bound? [Gap]
  > **Deferred**: No upper limit in v1. Practical limits are bounded by filesystem constraints and zip format limits. An explicit cap can be added if needed based on real-world usage.

- [X] CHK038 Is the behavior defined when the package name contains characters that are invalid in some filesystems (e.g., the `--` in the archive filename on legacy systems)? [Edge Case, Spec §FR-003]
  > **Not a concern**: `--` is valid on all target filesystems (ext4, NTFS, APFS, HFS+). The naming rules restrict scope and package names to `[a-z0-9-]`, and archive filenames to `{scope}--{name}-{version}.aipkg`. All characters in this set are valid on all major filesystems.

- [X] CHK039 Are requirements defined for file permission preservation in the archive? Should files retain their original permissions, or use normalized defaults? [Gap, Spec §FR-001]
  > **Deferred**: v1 uses Go's `archive/zip` default behavior (which stores basic permission info). For text-based artifacts (markdown, JSON), permissions are not meaningful. Normalized permissions can be specified in a future version if cross-platform consistency becomes important.

- [X] CHK040 Is the behavior defined when the source directory is the filesystem root or a path outside the user's home directory? [Edge Case, Spec §FR-007]
  > **Not a concern**: The pack command operates on any directory containing a valid aipkg.json. There is no restriction to the user's home directory. The filesystem root is an unlikely but valid source directory. No spec change needed.

- [X] CHK041 Are requirements defined for handling binary files in well-known directories (e.g., an image accidentally placed in `prompts/`)? FR-018 checks "non-empty" but not "valid text." [Gap, Spec §FR-018]
  > **Deferred**: v1 does not validate file content beyond non-empty (file-based types) and valid JSON (mcp-servers). Authors are responsible for placing appropriate files in well-known directories. Content-type validation can be added in a future version if misuse becomes common.

## Notes

- Focus areas: Archive format specification (heavier), pack command behavior, cross-spec consistency with 001
- Depth: Thorough (~30-40 items, catches subtle ambiguities)
- Audience: Author/reviewer pre-implementation gate
- Items reference spec sections and foundation docs where applicable
- `[Conflict]` markers indicate items that may require spec updates before implementation
- `[Gap]` markers indicate missing requirements that should be consciously included or explicitly deferred

### Resolution Summary (2026-03-02)

- **Spec updated (15)**: CHK001, CHK002, CHK004, CHK005, CHK006, CHK008, CHK009, CHK010, CHK012, CHK013, CHK014, CHK015, CHK020, CHK021, CHK033
- **Already resolved (11)**: CHK011, CHK019, CHK025, CHK026, CHK028, CHK029, CHK030, CHK031, CHK032, CHK036, CHK038
- **Intentional design (4)**: CHK016, CHK018, CHK022, CHK040
- **Deferred to future (7)**: CHK003, CHK007, CHK017, CHK027, CHK037, CHK039, CHK041
- **Implementation detail (2)**: CHK023, CHK024
- **Design principle (2)**: CHK034, CHK035
