# Feature Specification: Archive Format & Pack Command

**Feature Branch**: `002-archive-format-pack`
**Created**: 2026-03-01
**Status**: Draft
**Input**: User description: "Archive format specification and aipkg pack command (AIPKG-47)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Pack a package into a distributable archive (Priority: P1)

A package author has a directory containing a valid `aipkg.json` and artifacts in the well-known directories. They run `aipkg pack` and the command scans the well-known directories, discovers all artifacts, validates each one against its type-specific rules, generates the `artifacts` array in the manifest, validates the manifest against the package schema, and produces a `.aipkg` archive plus a `.sha256` sidecar file. The author now has a single distributable file they can share, attach to a GitHub release, or feed into a publishing workflow.

**Why this priority**: This is the core value of the feature. Without it, authors have no way to produce distributable packages. Everything else builds on this.

**Independent Test**: Can be fully tested by creating a package directory with artifacts, running `aipkg pack`, and verifying the archive contents and sidecar checksum.

**Acceptance Scenarios**:

1. **Given** a package directory with a valid `aipkg.json` and a `skills/test-writer/SKILL.md` file, **When** the author runs `aipkg pack`, **Then** a `.aipkg` archive and a `.aipkg.sha256` sidecar file are created in the current directory, named per FR-003 (e.g., `tjespers--test-writer-1.0.0.aipkg`).
2. **Given** the pack command completes, **When** the author inspects the archive, **Then** it contains a single top-level directory named after the package name (e.g., `test-writer/`) with the `aipkg.json` (including the generated `artifacts` array) and all artifact files. The original `aipkg.json` on disk is unchanged.
3. **Given** a package with a skill directory missing its `SKILL.md` file, **When** the author runs `aipkg pack`, **Then** the command fails with a validation error identifying the invalid skill and does not produce any archive.
4. **Given** a package directory with no artifacts in any well-known directory, **When** the author runs `aipkg pack`, **Then** the command fails with a clear error explaining that no artifacts were discovered.
5. **Given** a package with an `mcp-servers/github.json` file containing invalid JSON, **When** the author runs `aipkg pack`, **Then** the command fails with a validation error identifying the invalid file.
6. **Given** a package with artifacts of multiple types (skills, prompts, mcp-servers), **When** the author runs `aipkg pack`, **Then** all artifacts are discovered, validated, and included in the archive with correct type mappings in the `artifacts` array.

---

### User Story 2 - Exclude files from the archive (Priority: P2)

A package author has development files, build outputs, or other content they don't want in the distributable archive. They create a `.aipkgignore` file with gitignore-style patterns to control what gets included. Built-in defaults handle the most common exclusions automatically.

**Why this priority**: Without exclusion support, authors would need to maintain a separate clean directory for packing, which is error-prone and tedious. This is essential for real-world use but not required for a minimal viable pack.

**Independent Test**: Can be tested by creating a package with an `.aipkgignore` file, running pack, and verifying excluded files are absent from the archive.

**Acceptance Scenarios**:

1. **Given** a package with a `.aipkgignore` containing `*.log`, **When** the author runs `aipkg pack`, **Then** no `.log` files appear in the archive.
2. **Given** a package with no `.aipkgignore` file, **When** the author runs `aipkg pack`, **Then** `.git/`, `.aipkgignore` (if present), and the output archive itself are still excluded by built-in defaults.
3. **Given** a `.aipkgignore` with a pattern that would exclude a well-known directory (e.g., `skills/`), **When** the author runs `aipkg pack`, **Then** the matching artifacts are excluded from the archive and the `artifacts` array (`.aipkgignore` wins over convention).

---

### User Story 3 - Control output location (Priority: P3)

A package author wants the archive written to a specific location, such as a `dist/` directory or a CI artifact path, rather than the current directory.

**Why this priority**: Convenience for authors with custom build workflows. The default behavior covers most cases, so this is a refinement.

**Independent Test**: Can be tested by running pack with `--output` pointing to various paths and verifying the archive lands in the right place.

**Acceptance Scenarios**:

1. **Given** a valid package, **When** the author runs `aipkg pack --output dist/my-package.aipkg`, **Then** the archive is written to `dist/my-package.aipkg` and the sidecar to `dist/my-package.aipkg.sha256`.
2. **Given** a valid package, **When** the author runs `aipkg pack --output dist/`, **Then** the archive is written to `dist/` using the conventional filename.
3. **Given** `--output dist/my-package.aipkg` and `dist/my-package.aipkg` already exists, **When** the author runs `aipkg pack`, **Then** the existing file is overwritten silently.

---

### Edge Cases

- What happens when `aipkg.json` is missing from the directory? The command fails with a clear error explaining that this is not a package directory.
- What happens when `aipkg.json` fails schema validation (e.g., missing required fields)? The command fails with validation errors before any artifact discovery happens.
- What happens when an artifact name derived from a filename or directory name violates the naming rules (e.g., `skills/My_Skill/`)? The command fails with a validation error identifying the invalid name and the naming rules.
- What happens when a skill's `SKILL.md` `name` field does not match its parent directory name? The command fails with a validation error explaining the mismatch.
- What happens when a skill's `SKILL.md` contains disallowed frontmatter keys? The command fails with a validation error listing the unexpected keys.
- What happens when a prompt or command file is empty (0 bytes)? The command fails with a validation error identifying the empty file.
- What happens when the `--output` parent directory does not exist? The command fails with a filesystem error rather than creating intermediate directories.
- What happens when the author lacks write permissions for the output location? The command fails with a clear filesystem error.
- What happens when the archive would be extremely large? No size limit is enforced in v1.
- What happens when `.aipkgignore` patterns exclude all artifacts? The command fails with the "no artifacts discovered" error.
- What happens when the same artifact name appears in two different well-known directories (e.g., `skills/review/` and `prompts/review.md`)? Both are included in the archive with different types. Artifact entries are distinguished by their `type` field.
- What happens when an `.aipkgignore` pattern is malformed? The command fails with a parse error identifying the bad pattern and line number.

## Requirements *(mandatory)*

### Functional Requirements

**Archive format specification:**

- **FR-001**: The specification MUST define the archive format as a zip file with `.aipkg` extension, using deflate compression and UTF-8 filename encoding within the archive.
- **FR-002**: The specification MUST define that archives contain a single top-level directory holding the manifest and all artifact files and directories. Only the manifest and artifact content is included; non-artifact files at the package root (e.g., README.md, LICENSE) are not part of the archive in v1.
- **FR-003**: The specification MUST define the filename convention as `{scope}--{name}-{version}.aipkg`, where the `@` prefix is stripped and the `/` is replaced with `--` (double dash). Because both scope and package name forbid consecutive hyphens, the `--` separator is unambiguous. Because versions are strict semver (MAJOR.MINOR.PATCH, no pre-release or build metadata), the version contains dots but no hyphens, so `{name}-{version}` splits unambiguously at the last hyphen preceding a digit-dot sequence. Scope, name, and version are fully reconstructible from the filename alone (e.g., `@tjespers/dummy` v1.2.3 produces `tjespers--dummy-1.2.3.aipkg`). The specification reference document MUST include the parsing algorithm. The manifest inside remains authoritative for package identity.
- **FR-004**: The specification MUST define that the top-level directory inside the archive is named after the `name` portion of the manifest (the part after `/`, without scope or version). For `@tjespers/dummy` v1.2.3, the top-level directory is `dummy/`. This follows Helm's convention where the chart archive's top-level directory matches the chart name.
- **FR-005**: The specification MUST define extraction behavior: the CLI strips the top-level directory when extracting, placing contents directly into the target location.
- **FR-006**: The specification MUST define integrity verification via SHA-256 sidecar files using `sha256sum` format (lowercase hex hash, two spaces, archive basename). The sidecar file MUST use UTF-8 encoding with a LF line ending, terminated by a single newline.

**Pack command:**

- **FR-007**: The system MUST provide a `pack` command that creates a `.aipkg` archive from the current directory (or a specified source directory).
- **FR-008**: The pack command MUST discover artifacts by scanning the six well-known directories (`skills/`, `prompts/`, `commands/`, `agents/`, `agent-instructions/`, `mcp-servers/`). Discovery MUST only consider top-level entries within each well-known directory: regular files for file-based types, direct subdirectories for directory-based types (skills). Hidden entries (names starting with `.`) MUST be silently skipped. Symlinks and other non-regular filesystem entries MUST be silently skipped. Nested subdirectories within file-based type directories (e.g., `prompts/drafts/review.md`) MUST NOT be discovered.
- **FR-009**: The pack command MUST generate the `artifacts` array from discovered artifacts and write a copy of the manifest with this array into the archive. The generated array replaces any existing `artifacts` field in the archived copy. The original `aipkg.json` on disk MUST NOT be modified. This ensures the pack command is atomic and idempotent.
- **FR-010**: The pack command MUST validate the manifest against the package JSON schema before packing. If validation fails, the command MUST refuse to pack and MUST NOT proceed to artifact discovery. Manifest validation errors are reported immediately without collecting further errors.
- **FR-011**: If no artifacts are discovered in any well-known directory (after applying ignore rules), the pack command MUST exit with an error.

**Type-specific artifact validation:**

- **FR-012**: Skill directories MUST contain a `SKILL.md` file. Without it, the pack command MUST reject the directory. The entire skill directory is included in the archive (including optional subdirectories like `scripts/`, `references/`, and `assets/` as defined by the Agent Skills specification).
- **FR-013**: The `SKILL.md` file MUST contain valid YAML frontmatter enclosed in `---` delimiters.
- **FR-014**: The `SKILL.md` frontmatter MUST include `name` (kebab-case, 1-64 characters, no leading/trailing/consecutive hyphens) and `description` (1-1024 characters). Both fields are required.
- **FR-015**: The `SKILL.md` frontmatter MUST only contain allowed keys: `name`, `description`, `license`, `compatibility`, `metadata`, `allowed-tools`. Unexpected keys MUST cause a validation error.
- **FR-016**: The `name` field in `SKILL.md` frontmatter MUST match the parent directory name. A mismatch MUST cause a validation error.
- **FR-017**: MCP server config files MUST be valid JSON. The pack command validates that the file parses as JSON but does not validate the JSON structure beyond that.
- **FR-018**: File-based artifact types (prompts, commands, agents, agent-instructions) MUST be non-empty (file size greater than zero bytes). The file extension does not determine the artifact type; the well-known directory determines type. Files are not restricted to `.md`. Extensions like `.txt`, `.prompt`, and `.prompt.md` are equally valid. Recommended extensions: `.md` for most artifact types, but authors are free to use whatever convention they prefer.
- **FR-019**: Artifact names MUST follow the standard naming rules: lowercase alphanumeric and hyphens, 1-64 characters, no consecutive hyphens, cannot start or end with a hyphen. For file-based artifacts, the name is derived by stripping everything from the first `.` onwards (e.g., `code-review.prompt.md` → `code-review`, `my-prompt.txt` → `my-prompt`). If a filename contains no `.`, the entire filename is the artifact name. For directory-based artifacts (skills), the directory name is the artifact name.
- **FR-020**: If any artifact fails validation, the pack command MUST refuse to produce an archive and MUST report all validation errors (not just the first one).

**Archive creation:**

- **FR-021**: The archive MUST satisfy the top-level directory structure defined in FR-004. Consumers can rely on this for predictable extraction.
- **FR-022**: The pack command MUST generate a `.sha256` sidecar file alongside the archive, containing the SHA-256 hash of the archive in `sha256sum` format.
- **FR-023**: The default output filename MUST follow the convention defined in FR-003: `{scope}--{name}-{version}.aipkg`.
- **FR-024**: The pack command MUST support an `--output` flag for specifying a custom output path (file or directory).
- **FR-025**: If the output file already exists, the pack command MUST overwrite it silently.

**File exclusion:**

- **FR-026**: The pack command MUST support `.aipkgignore` files using gitignore-style pattern syntax.
- **FR-027**: Built-in defaults MUST always exclude `.git/`, `.aipkgignore`, and the output archive itself, regardless of `.aipkgignore` contents. Conversely, the `aipkg.json` manifest MUST always be included in the archive and cannot be excluded by `.aipkgignore`.
- **FR-028**: Author-defined `.aipkgignore` patterns take precedence over convention-based directory discovery. If a pattern excludes an artifact file or directory, it is omitted from the archive and the `artifacts` array.

### Key Entities

- **Archive** (`.aipkg` file): A zip file containing a single top-level directory with the package manifest and all artifact files. The distributable unit of an aipkg package.
- **Sidecar File** (`.sha256`): A companion file containing the SHA-256 checksum of the archive in `sha256sum` format, used for integrity verification after transfer.
- **Ignore File** (`.aipkgignore`): An optional file at the package root using gitignore-style patterns to exclude files from the archive.
- **Artifacts Array**: The `artifacts` field in the manifest, auto-generated by pack from directory conventions. Each entry contains `name`, `type`, and `path`. This field only appears in the archived copy of the manifest; the original `aipkg.json` on disk is never modified by pack.
- **Top-Level Directory**: The single directory inside the archive that wraps all package contents. Named after the `name` portion of the manifest (without scope or version), following Helm's convention.

## Assumptions

- The six well-known directories and their type mappings are defined by the package foundation specification (AIPKG-8). This feature does not add or modify artifact types.
- Artifact name derivation follows the rules established in the foundation spec: directory name for skills, filename stem (everything before the first `.`) for file-based types. This handles compound extensions like `.prompt.md` naturally.
- The `artifacts` array structure in the manifest follows the schema defined in `spec/schema/package.json`. Each entry has `name` (string), `type` (enum), and `path` (string).
- YAML frontmatter parsing for SKILL.md follows the standard `---` delimiter convention. The parser handles frontmatter extraction; it does not validate the full YAML specification beyond what's needed for the allowed fields.
- The `sha256sum` format is: lowercase hex hash, two spaces, filename (e.g., `a1b2c3...  filename.aipkg`). This matches the output of the standard `sha256sum` command-line tool.
- Zip archive creation uses the standard deflate compression method. No optimization of compression level is in scope.
- `.aipkgignore` uses the same pattern syntax as `.gitignore` (as defined by Git's documentation).
- The pack command operates on the filesystem only. No network operations, no registry lookups, no remote validation.
- The spec documentation for the archive format will live in `spec/` as a new reference document alongside the existing manifest.md, artifacts.md, and naming.md.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A package author can produce a distributable `.aipkg` archive from a valid package directory using a single command.
- **SC-002**: 100% of archives produced by pack contain a valid manifest with a correctly generated `artifacts` array that matches the directory contents.
- **SC-003**: The `.sha256` sidecar file correctly verifies the archive when checked with standard tools (e.g., `sha256sum -c`).
- **SC-004**: Invalid packages (missing SKILL.md, bad JSON, empty markdown, missing artifacts) are caught before any archive is produced, with error messages that identify the specific problem and location.
- **SC-005**: The archive format specification is documented clearly enough that a third-party tool could produce or consume `.aipkg` archives from the specification alone.
- **SC-006**: An author familiar with `npm pack` or `helm package` can predict the behavior of `aipkg pack` without reading documentation (Convention Over Invention principle).
