# Feature Specification: Package Foundation: Spec & Create Command

**Feature Branch**: `001-package-foundation`
**Created**: 2026-03-01
**Status**: Draft
**Input**: User description: "Package Foundation: Spec & Create Command (AIPKG-8)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create a new package from scratch (Priority: P1)

A package author wants to start building an aipkg package. They run `aipkg create @alice/blog-writer` and are guided through an interactive flow that asks for version, description, and license. When complete, a new `blog-writer/` directory exists containing a valid `aipkg.json` and the well-known artifact directories (`skills/`, `prompts/`, `commands/`, `agents/`, `agent-instructions/`, `mcp-servers/`). The author can immediately start adding artifacts to the appropriate directories.

**Why this priority**: This is the primary entry point for package authoring. Without it, authors have no guided way to create correctly structured packages.

**Independent Test**: Can be fully tested by running the create command and verifying the resulting directory structure and manifest are valid.

**Acceptance Scenarios**:

1. **Given** an empty directory, **When** the author runs `aipkg create @alice/blog-writer`, **Then** a `blog-writer/` directory is created containing `aipkg.json` and all well-known artifact directories.
2. **Given** the create command is running, **When** the author provides version, description, and license through prompts, **Then** the generated `aipkg.json` contains exactly those values plus a spec version field indicating the package format version.
3. **Given** the create command completes, **When** the author inspects the generated `aipkg.json`, **Then** no `artifacts` field is present (that field is generated later by `aipkg pack`).

---

### User Story 2 - Create a package in an existing directory (Priority: P2)

A package author has an existing directory with some files (perhaps a README, or some artifact files they wrote manually). They want to turn it into an aipkg package without losing their existing work. They run `aipkg create @alice/blog-writer --path .` (or `--path ./existing-dir`), and the command adds the package structure around their existing files without overwriting anything.

**Why this priority**: Supports the common case of converting existing work into a package. Directly enables adoption for people who already have content.

**Independent Test**: Can be tested by creating a directory with pre-existing files, running create with `--path`, and verifying existing files are untouched while new structure is added.

**Acceptance Scenarios**:

1. **Given** a directory containing a `skills/my-skill/SKILL.md` file, **When** the author runs `aipkg create @alice/my-pkg --path .`, **Then** the `skills/` directory is preserved with its contents, other well-known directories are created, and `aipkg.json` is generated.
2. **Given** a directory already containing an `aipkg.json`, **When** the author runs `aipkg create @alice/my-pkg --path .`, **Then** the command refuses to overwrite and exits with a clear error message.
3. **Given** a non-existent path, **When** the author runs `aipkg create @alice/my-pkg --path ./new-dir`, **Then** the directory is created and populated with the full package structure.

---

### User Story 3 - Validate package name during creation (Priority: P3)

A package author provides an invalid package name (wrong format, reserved scope, missing scope). The command catches the error inline within the interactive prompt and shows the validation error, allowing the author to correct it without losing progress on other fields. The author is never kicked out of the creation flow due to a validation error.

**Why this priority**: Catches errors early, before the author invests time adding artifacts. Inline re-prompting prevents frustration from having to restart the entire flow.

**Independent Test**: Can be tested by providing various invalid names and verifying the prompt redisplays with the error message, allowing correction.

**Acceptance Scenarios**:

1. **Given** the interactive prompt for package name, **When** the author enters `blog-writer` (no scope), **Then** the prompt shows an inline error explaining that packages require a scoped name like `@scope/package-name` and allows the author to re-enter without losing progress.
2. **Given** the interactive prompt, **When** the author enters `@aipkg/my-package` (reserved scope), **Then** the prompt shows an inline error explaining that the `@aipkg` scope is reserved and allows correction.
3. **Given** the interactive prompt, **When** the author enters `@Alice/Blog_Writer` (invalid characters), **Then** the input is normalized to lowercase and underscores are rejected with an inline explanation, allowing correction.

---

### User Story 4 - Non-interactive package creation for CI (Priority: P4)

A package author (or a CI pipeline) needs to create a package without interactive prompts. They provide all metadata via flags: `aipkg create --name @alice/blog-writer --version 1.0.0 --description "AI blog writing assistant" --license MIT`. The command creates the package structure without prompting for anything.

**Why this priority**: Enables automation and CI/CD pipelines. Not the primary entry point but essential for professional workflows.

**Independent Test**: Can be tested by running create with all flags and verifying no prompts appear and the output is correct.

**Acceptance Scenarios**:

1. **Given** all required flags are provided, **When** the author runs `aipkg create --name @alice/blog-writer --version 1.0.0`, **Then** the package is created without any interactive prompts.
2. **Given** some flags are provided but not all, **When** the author runs `aipkg create --name @alice/blog-writer`, **Then** the command prompts interactively only for the missing fields (version, description, license).
3. **Given** a flag value is invalid (e.g., `--version bad`), **When** running non-interactively, **Then** the command exits with a validation error and does not create any files.

---

### User Story 5 - License detection from existing LICENSE file (Priority: P5)

A package author creates a package in a directory that already contains a LICENSE file. The create command detects the license type and suggests it as the default value during the license prompt, saving the author from typing it manually.

**Why this priority**: Quality-of-life improvement that reduces friction. Not critical for the core flow but makes the experience feel polished.

**Independent Test**: Can be tested by placing various LICENSE files in the target directory and verifying the correct SPDX identifier is suggested.

**Acceptance Scenarios**:

1. **Given** a directory with an Apache-2.0 LICENSE file, **When** the author runs `aipkg create --path .`, **Then** the license prompt suggests `Apache-2.0` as the default.
2. **Given** a directory with no LICENSE file, **When** the create command runs, **Then** the license prompt has no default and accepts any valid SPDX identifier or `proprietary`.
3. **Given** a directory with an ambiguous or unrecognized LICENSE file, **When** the create command runs, **Then** the license prompt has no default (does not guess incorrectly).

---

### Edge Cases

- What happens when the author provides a scoped name via the command argument AND the interactive prompt asks for it? The command argument takes precedence; the name prompt is skipped.
- What happens when disk is full or the user lacks write permissions? The command fails with a clear filesystem error before partially creating the structure.
- What happens when the package name argument doesn't include a scope? The create command exits with an error explaining that scoped names are required. It does not attempt to guess or add a scope.
- What happens when `--path` points to a file instead of a directory? The command exits with a clear error.
- What happens when the author cancels the interactive prompts (Ctrl+C)? No files are created; the command exits cleanly.
- What happens when the directory name derived from the package name conflicts with an existing file (not directory) at that location? The command exits with a clear error.
- What happens when the user provides an invalid version like `1.0` during the prompt? The prompt shows an inline error explaining the required format (MAJOR.MINOR.PATCH) and allows correction without restarting.
- What happens when all flags are provided but one value is invalid (e.g., `--name invalid-name`)? The command exits with a validation error (no interactive fallback in non-interactive mode).
- What happens when some flags are provided but not all required ones? The command prompts interactively only for the missing fields, preserving flag values.
- What happens when some flags are missing and no TTY is available (e.g., CI pipe)? The command exits with an error listing the missing flags. No interactive fallback without a TTY.

## Requirements *(mandatory)*

### Functional Requirements

**Package specification:**

- **FR-001**: The specification MUST define a convention-based directory layout for packages, mapping well-known directory names to artifact types.
- **FR-002**: The well-known directories MUST be: `skills/`, `prompts/`, `commands/`, `agents/`, `agent-instructions/`, `mcp-servers/`.
- **FR-003**: The specification MUST define how artifact names are derived from filenames and directory names within the well-known directories.
- **FR-004**: The specification MUST define type-specific structural requirements for each artifact type (e.g., skill directories must contain a `SKILL.md` file).
- **FR-005**: The specification MUST clarify that the `artifacts` array in the manifest is generated by `aipkg pack` from directory conventions, not written manually by authors.
- **FR-006**: The specification MUST define the package manifest schema. The manifest does not contain a `type` field; the schema describes packages only. Project-type manifests are out of scope for this feature and may use a different schema entirely.
- **FR-007**: The package manifest MUST require `name` (scoped), `version` (semver), and a spec version field indicating which version of the package format specification this manifest conforms to. The `artifacts` field MUST NOT be required at creation time.
- **FR-008**: The specification MUST define validation rules for package names (scoped format, character restrictions, length limits, reserved scope rejection).
- **FR-009**: The spec version field MUST use an integer format (e.g., `1`, `2`, `3`) to indicate the package format version. This enables the CLI to detect older manifest formats and support migration between spec versions over time.

**Create command:**

- **FR-010**: The system MUST provide a `create` command that scaffolds a new package directory with the conventional structure.
- **FR-011**: The create command MUST accept a scoped package name as a positional argument (e.g., `aipkg create @scope/package-name`).
- **FR-012**: By default, the create command MUST create a new directory named after the package (derived from the package name portion, e.g., `@alice/blog-writer` creates `blog-writer/`).
- **FR-013**: The create command MUST support a `--path` flag to specify a custom target directory (including `.` for the current directory).
- **FR-014**: The create command MUST be non-destructive: it MUST NOT overwrite existing files or directories.
- **FR-015**: The create command MUST refuse to proceed if an `aipkg.json` already exists in the target directory and exit with a clear error.
- **FR-016**: The create command MUST prompt interactively for package metadata: version, description, and license.
- **FR-017**: When the scoped package name is provided as a command argument, the name prompt MUST be skipped.
- **FR-018**: When the scoped package name is not provided as a command argument, the command MUST prompt for it interactively.
- **FR-019**: The create command MUST validate the package name against the naming rules (scoped format, character rules, reserved scopes) during the prompt, before any files are created.
- **FR-020**: The create command MUST generate a valid `aipkg.json` with `name`, `version`, the spec version field, and optionally `description` and `license`. No `artifacts` field. No `type` field.
- **FR-021**: Every interactive prompt field MUST have a corresponding CLI flag (`--name`, `--version`, `--description`, `--license`) to enable fully non-interactive package creation (e.g., for CI pipelines). When all required fields are provided via flags, the command MUST skip all prompts and create the package directly.
- **FR-022**: When no TTY is available (e.g., piped input, CI environment) and required fields are missing from flags, the command MUST exit with an error listing the missing flags. The command MUST NOT attempt interactive prompts without a TTY.
- **FR-023**: The create command MUST create all well-known artifact directories in the target directory, skipping any that already exist.
- **FR-024**: The create command MUST detect an existing LICENSE file in the target directory and suggest the corresponding SPDX identifier as the default license value.
- **FR-025**: The version field MUST default to `0.1.0` during the interactive prompt.
- **FR-026**: If the user cancels the interactive prompts (e.g., Ctrl+C), no files or directories MUST be created.
- **FR-027**: The create command MUST validate that version input follows strict semver format (MAJOR.MINOR.PATCH, no pre-release suffixes).
- **FR-028**: When the user provides invalid input during an interactive prompt, the command MUST show the validation error inline and re-prompt the same field. The user MUST NOT be ejected from the creation flow or forced to restart. All previously entered values MUST be preserved.
- **FR-029**: The CLI MUST validate the manifest against the specification's JSON schema as the single source of truth. Validation logic MUST NOT be duplicated in CLI code separately from the schema; the schema is the canonical authority to prevent drift and inconsistencies.

### Key Entities

- **Package Manifest** (`aipkg.json`): The identity and metadata file for a package. Contains spec version, name, version, description, license. No `type` field (the schema describes packages only; other manifest types may use different schemas). Does not contain the `artifacts` array at creation time.
- **Spec Version**: An integer field in the manifest indicating which version of the package format specification this manifest conforms to. Enables schema migration and forward compatibility as the format evolves.
- **Well-Known Directory**: A directory with a conventional name that maps to an artifact type. The complete set: `skills/`, `prompts/`, `commands/`, `agents/`, `agent-instructions/`, `mcp-servers/`.
- **Artifact Type Mapping**: The relationship between a well-known directory name and the artifact type it represents (e.g., `skills/` maps to type `skill`, `mcp-servers/` maps to type `mcp-server`).
- **Scoped Package Name**: A name in the format `@scope/package-name` that uniquely identifies a package in the ecosystem.
- **Reserved Scope**: A scope prefix (e.g., `@aipkg`, `@anthropic`) that is reserved and cannot be used for user-created packages.

## Assumptions

- The well-known directory names (`skills/`, `prompts/`, `commands/`, `agents/`, `agent-instructions/`, `mcp-servers/`) match the six artifact types currently defined in the specification. If artifact types are added in the future, new well-known directories would follow the same pattern.
- The default version `0.1.0` follows the convention that new packages start in pre-1.0 development.
- License detection from a LICENSE file uses confidence thresholds; ambiguous files result in no suggestion rather than an incorrect one.
- Existing spec docs (manifest.md, artifacts.md, naming.md) are prior art. This feature has the mandate to revise them as needed, particularly to separate package-specific concerns from project-specific concerns and to document the convention-based directory structure.
- The `require` and `repositories` fields are not included in the create command's interactive prompts (they are project-consumption concerns or advanced configuration added later).
- The package manifest does not include a `type` field. The schema is package-specific. Other manifest types (e.g., projects) will be designed separately and may use a different file or schema entirely.
- The initial spec version is `1`. The spec version only increments when the manifest schema changes in a way that requires migration.
- The specification's JSON schema is the single source of truth for manifest validation. The CLI consumes the schema at build time or runtime; it does not re-implement validation rules independently.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A package author can go from zero to a valid, correctly structured package directory using a single command.
- **SC-002**: 100% of generated `aipkg.json` files pass schema validation for the package manifest.
- **SC-003**: An author familiar with Helm or npm can predict the behavior of `aipkg create` without reading documentation (Convention Over Invention principle).
- **SC-004**: The convention-based directory structure is documented clearly enough that a third-party tool implementer could build their own `create` command from the specification alone.
- **SC-005**: The create command never destroys or modifies existing user files, verified by running create in directories with pre-existing content.
- **SC-006**: All six artifact types have a defined well-known directory, documented structural requirements, and clear name derivation rules in the specification.
