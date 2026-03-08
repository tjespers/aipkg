# Feature Specification: Project Initialization & Model

**Feature Branch**: `003-project-initialization`
**Created**: 2026-03-03
**Status**: Draft
**Input**: User description: "Define the aipkg project model and implement `aipkg init`. A project is the consumer side of the ecosystem: it declares dependencies on packages and provides the directory structure where installed artifacts live."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Initialize a new project (Priority: P1)

A developer wants to start using AI packages in their project. They run `aipkg init` in their project root and get a project file (`aipkg-project.json`) with an empty dependency map. This is all they need to start adding packages later with `aipkg require`. No directories are created, no prompts are shown, no configuration is needed.

**Why this priority**: This is the core value of the feature. Without project initialization, there is no way to declare dependencies or provide a home for installed artifacts. Everything else in the install pipeline depends on this file existing.

**Independent Test**: Can be fully tested by running `aipkg init` in an empty directory and verifying that exactly one file is created with the correct structure.

**Acceptance Scenarios**:

1. **Given** an empty directory, **When** the developer runs `aipkg init`, **Then** a file named `aipkg-project.json` is created in the current directory containing an empty `require` map.
2. **Given** a directory with existing files (source code, configs, etc.) but no `aipkg-project.json` or `aipkg.json`, **When** the developer runs `aipkg init`, **Then** `aipkg-project.json` is created without affecting any existing files.
3. **Given** a successful init, **When** the developer inspects the directory, **Then** only `aipkg-project.json` exists as a new file. No `.aipkg/` directory or other files are created.
4. **Given** a successful init, **When** the developer opens `aipkg-project.json`, **Then** the file contains valid JSON with a `specVersion` field indicating the schema version and a `require` field set to an empty object.

---

### User Story 2 - Prevent accidental re-initialization (Priority: P2)

A developer runs `aipkg init` in a directory that already has an `aipkg-project.json`. The command refuses to overwrite the existing project file and tells the developer the project is already initialized.

**Why this priority**: Protecting existing project configuration from accidental overwrites is essential for trust. Without this guard, a developer could lose their dependency declarations.

**Independent Test**: Can be tested by creating a project file with dependencies, running `aipkg init`, and verifying the file is unchanged.

**Acceptance Scenarios**:

1. **Given** a directory containing an `aipkg-project.json` (with or without dependencies), **When** the developer runs `aipkg init`, **Then** the command exits with an error message indicating the project is already initialized.
2. **Given** the re-initialization guard triggers, **When** the command exits, **Then** the existing `aipkg-project.json` is unchanged.

---

### User Story 3 - Refuse initialization in a package directory (Priority: P2)

A developer runs `aipkg init` in a directory that contains an `aipkg.json` (a package manifest). This is almost certainly a mistake: they likely meant to run `aipkg install` or `aipkg require`. The command refuses and tells them this is a package directory, pointing them to the right commands.

**Why this priority**: Package directories and project directories are mutually exclusive concepts. A package manifest and a project file cannot coexist in the same directory because the CLI would have no way to disambiguate which context commands like `require` target. Blocking this early prevents confusing state.

**Independent Test**: Can be tested by creating a directory with an `aipkg.json`, running `aipkg init`, and verifying the command refuses with a clear error.

**Acceptance Scenarios**:

1. **Given** a directory containing an `aipkg.json`, **When** the developer runs `aipkg init`, **Then** the command exits with an error explaining that a package manifest already exists and that `aipkg init` cannot be used in a package directory.
2. **Given** the error is displayed, **Then** no `aipkg-project.json` is created and no files are modified.
3. **Given** the error is displayed, **Then** the error message suggests `aipkg require` or `aipkg install` as likely intended commands.

---

### Edge Cases

- What happens when the developer lacks write permissions for the directory? The command fails with a clear filesystem error.
- What happens when `aipkg-project.json` exists but contains invalid JSON? The re-initialization guard still triggers based on file existence, not file validity. The file is not modified or validated during init.
- What happens when `aipkg.json` exists but contains invalid JSON? The mutual exclusivity guard still triggers based on file existence. Init does not validate the package manifest.
- What happens when the developer runs `aipkg init` inside the `.aipkg/` directory? The command operates on the current directory. It would create `aipkg-project.json` inside `.aipkg/`, which is a valid (if unusual) location. No special handling is needed.

## Requirements *(mandatory)*

### Functional Requirements

**Project file specification:**

- **FR-001**: The project file MUST be named `aipkg-project.json` and located at the project root. It is the marker that identifies a directory as an aipkg-enabled project.
- **FR-002**: The project file and the package manifest (`aipkg.json`) are mutually exclusive. They MUST NOT coexist in the same directory. The two files serve different concerns and different personas: the project file declares consumed dependencies, the package manifest declares package identity and authored artifacts. The CLI MUST refuse to create one when the other already exists.
- **FR-003**: The project file MUST NOT contain identity fields (`name`, `version`, `description`, `license`). It is purely operational.
- **FR-004**: The v1 project file schema MUST contain a `specVersion` field and a `require` field. `specVersion` is an integer indicating which version of the project file specification was used to generate the file (v1 sets this to `1`), matching the format used in the package manifest schema. This enables automated migration if the schema evolves. `require` is an object mapping scoped package names to semver versions. Keys follow the established package naming rules (e.g., `@scope/package-name`). Values are semver versions (MAJOR.MINOR.PATCH, with optional pre-release identifiers per the SemVer spec, e.g. `1.0.0-beta.1`). Ranges and prefixes are not supported.
- **FR-005**: A JSON Schema for the project file MUST be provided to enable validation by the CLI and third-party tools.

**Install directory layout specification:**

- **FR-006**: The specification MUST define `.aipkg/` at the project root as the install directory for all packages consumed by the project.
- **FR-007**: The install directory MUST use a categorized layout with well-known subdirectories for individual artifact types: `skills/`, `prompts/`, `commands/`, `agents/`. Each subdirectory holds artifacts of the corresponding type from all installed packages.
- **FR-008**: Mergeable artifact types MUST produce a single merged file at the `.aipkg/` root level rather than individual per-package copies. MCP server configurations merge into `mcp.json`. Agent instructions merge into `agent-instructions.md`. The merge is a core responsibility performed at install time, not an adapter concern.
- **FR-009**: When `.aipkg/` is first created inside a git repository, a `.gitignore` file MUST be placed inside it that ignores all contents (`*` with `!.gitignore` exception). This prevents accidental commits of installed artifacts. The git detection and `.gitignore` creation only apply when the project root is within a git working tree.
- **FR-010**: The `.aipkg/` directory MUST NOT be created by `aipkg init`. It materializes on demand when the first package is installed.
- **FR-011**: Merged files in `.aipkg/` (`mcp.json`, `agent-instructions.md`) are fully aipkg-managed. They are generated and overwritten by install and update operations. Manual edits to these files will be lost. The reference documentation (FR-019) MUST make this ownership model explicit, including the expected behavior on install/update and the fact that manual modifications are not preserved.

**Scoped artifact naming:**

- **FR-012**: Installed artifacts MUST use a scoped naming convention that incorporates the source package identity. This prevents name collisions when multiple packages contribute artifacts of the same type to the same install directory.
- **FR-013**: The naming convention MUST enable traceability from any installed artifact back to its source package, so that removal and update operations can identify which files belong to which package.
- **FR-014**: The exact scoped naming format is deferred to the planning phase, pending research into naming patterns supported by target tools (Claude Code, Cursor, etc.). The existing dot-notation convention defined in `spec/naming.md` serves as the starting point.

**Init command:**

- **FR-015**: `aipkg init` MUST create `aipkg-project.json` in the current directory with an empty `require` object.
- **FR-016**: If `aipkg-project.json` already exists in the current directory, the command MUST refuse to proceed and display an error message. The existing file MUST NOT be modified.
- **FR-017**: If `aipkg.json` exists in the current directory, the command MUST refuse to proceed and display an error message explaining that a package manifest already exists and that a project file cannot be created in a package directory. The error MUST suggest `aipkg require` or `aipkg install` as likely intended commands.
- **FR-018**: `aipkg init` MUST NOT create the `.aipkg/` directory or any subdirectories.

**Documentation:**

- **FR-019**: The project model (project file, install directory layout, scoped naming) MUST be documented as reference material in `spec/`. This documentation is a deliverable of this feature, not a follow-up task.

### Key Entities

- **Project File** (`aipkg-project.json`): A JSON file at the project root that declares package dependencies via a `require` map and tracks the schema version via `specVersion`. Contains no identity fields. Serves as both the dependency declaration and the installed-package registry (the `require` map combined with the scoped naming convention provides full traceability of installed artifacts).
- **Install Directory** (`.aipkg/`): A categorized directory at the project root where installed artifacts live. Contains well-known subdirectories for individual artifact types and fully aipkg-managed merged files for mergeable types. Created on demand, gitignored by default.
- **Installed Artifact**: A package artifact placed in the appropriate `.aipkg/` subdirectory using a scoped name derived from the source package identity. Individual types (skills, prompts, commands, agents) are placed as separate files or directories. Mergeable types (mcp-servers, agent-instructions) contribute to shared files at the `.aipkg/` root.

## Assumptions

- The six artifact types and their classification into individual vs. mergeable categories are defined by the package foundation specification (001) and `spec/artifacts.md`. This feature does not add or modify artifact types.
- The `require` field in the project file uses the same scoped name format (`@scope/package-name`) as the existing `require` field in the package manifest schema (`spec/schema/package.json`). The name validation pattern is reusable. The version pattern needs extending: the package manifest uses strict MAJOR.MINOR.PATCH, while the project file also accepts optional pre-release identifiers per FR-004 (e.g. `1.0.0-beta.1`).
- Dot-notation as defined in `spec/naming.md` provides the basis for scoped artifact naming. The planning phase will finalize whether the format needs to include the package name in addition to scope and artifact name, based on collision analysis and tool compatibility research.
- The `.aipkg/` directory contains no committed files in v1. The strict `.gitignore` is sufficient; no exceptions are needed.
- The `require` field in the package manifest (`aipkg.json`) and the `require` field in the project file (`aipkg-project.json`) are separate concerns. The package `require` handles bundled dependencies resolved at pack time. The project `require` handles installed dependencies resolved at install time. This feature does not modify the package manifest schema.
- The mutual exclusivity check (FR-017) is a simple file-existence guard with no interactive component. It behaves identically in interactive and non-interactive environments.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can initialize a project with a single command (`aipkg init`) and no prior setup or configuration.
- **SC-002**: Project initialization produces exactly one file (`aipkg-project.json`). No directories, no config files, no noise.
- **SC-003**: A developer who accidentally runs `aipkg init` in a package directory is blocked with a clear error. No files are created, and the error points them to the right commands (`aipkg require` / `aipkg install`).
- **SC-004**: The project file format, install directory layout, and scoped naming requirements are documented clearly enough that the install command (AIPKG-10) can implement them without ambiguity or additional design work.
- **SC-005**: The project file JSON Schema is complete and validates both valid and invalid project files correctly. A valid project file with dependencies passes validation; a file with malformed package names or invalid version strings fails.
- **SC-006**: An existing `aipkg-project.json` is never overwritten or modified by running `aipkg init`.
