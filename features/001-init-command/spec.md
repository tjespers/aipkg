# Feature Specification: Init Command

**Feature Branch**: `001-init-command`
**Created**: 2026-02-28
**Status**: Draft
**Input**: User description: "Package authors and project maintainers need a guided way to create their first aipkg.json manifest. Running aipkg init walks the user through an interactive flow — choosing between project or package type, then prompting for the relevant fields. For projects: an optional name and description. For packages: a scoped name (@scope/name) with validation, version (semver), description, and license. The command generates a valid aipkg.json in the current directory. All prompted fields can also be passed as flags for non-interactive/CI use. If aipkg.json already exists, the command refuses with a clear error message. No network operations, no adapter execution, no artifact scaffolding — artifacts are derived at package time, not during init. This is the first command implemented in the CLI, so it also establishes the foundational patterns all future commands build on: cobra command structure, help/usage system, structured error handling with wrapped errors, and logging."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Package Author Creates Manifest Interactively (Priority: P1)

A package author wants to publish a set of AI artifacts (skills, prompts, etc.) as an aipkg package. They run `aipkg init` in their project directory and are guided through an interactive flow. The command first asks whether they are creating a project or a package. They select "package." The command then prompts for the required and optional fields: a scoped name, a version, a description, and a license. All input is validated against the `aipkg-spec` manifest JSON schema. After all prompts are answered, the command writes a valid `aipkg.json` to the current directory.

**Why this priority**: Package creation is the primary use case for the ecosystem. Without packages, there is nothing to install. This is the core path.

**Independent Test**: Can be fully tested by running `aipkg init` in an empty directory, selecting "package", entering valid field values, and verifying the resulting `aipkg.json` contains all fields with correct values.

**Acceptance Scenarios**:

1. **Given** an empty directory with no `aipkg.json`, **When** user runs `aipkg init` and selects "package" and provides name `@myorg/cool-skill`, version `1.0.0`, description `"A cool skill"`, and license `Apache-2.0`, **Then** a valid `aipkg.json` is created with `type: "package"` and all provided fields.
1. **Given** user enters a name that does not pass schema validation (e.g., `cool-skill`), **When** the name is submitted, **Then** the command shows a validation error and re-prompts for a valid name.
1. **Given** user enters a version that does not pass schema validation (e.g., `1.0`), **When** the version is submitted, **Then** the command shows a validation error and re-prompts for a valid version.

______________________________________________________________________

### User Story 2 - Project Maintainer Creates Manifest Interactively (Priority: P2)

A project maintainer wants to consume aipkg packages in their project. They run `aipkg init` and select "project." The command prompts for an optional name and an optional description. Both can be skipped by pressing Enter. The command writes a minimal `aipkg.json` with `type: "project"` and any provided optional fields.

**Why this priority**: Projects consume packages and are essential for the ecosystem, but a project manifest is simpler (fewer fields, all optional beyond type) so the flow is less complex.

**Independent Test**: Can be fully tested by running `aipkg init` in an empty directory, selecting "project", skipping all optional prompts, and verifying a valid `aipkg.json` with just `{ "type": "project" }` is created.

**Acceptance Scenarios**:

1. **Given** an empty directory, **When** user runs `aipkg init`, selects "project", and skips all optional prompts, **Then** a valid `aipkg.json` is created containing only `{ "type": "project" }`.
1. **Given** an empty directory, **When** user runs `aipkg init`, selects "project", and provides name `@myteam/my-project` and description `"My AI project"`, **Then** a valid `aipkg.json` is created with all three fields.
1. **Given** user provides a project name, **When** the name is submitted, **Then** it is validated against the manifest JSON schema.

______________________________________________________________________

### User Story 3 - Non-Interactive Manifest Creation via Flags (Priority: P3)

A CI pipeline or automation script needs to create an `aipkg.json` without interactive prompts. The user passes all values as command-line flags (e.g., `aipkg init --type package --name @myorg/my-skill --version 1.0.0 --description "Automated" --license MIT`). When all required fields for the chosen type are provided via flags, the command runs without any interactive prompts and writes the manifest file.

**Why this priority**: Non-interactive use enables CI/CD and scripting, which are important for adoption but secondary to the interactive experience most users will encounter first.

**Independent Test**: Can be fully tested by running `aipkg init` with all required flags and verifying no interactive prompts appear and the output file is correct.

**Acceptance Scenarios**:

1. **Given** an empty directory, **When** user runs `aipkg init --type package --name @myorg/my-tool --version 0.1.0`, **Then** a valid `aipkg.json` is created without any interactive prompts.
1. **Given** flags provide some but not all required values for a package (e.g., `--type package --name @myorg/my-tool` without `--version`), **When** running in a terminal, **Then** the command prompts interactively only for the missing required fields.
1. **Given** flags provide values that do not pass schema validation (e.g., `--name bad-name`), **When** the command runs, **Then** it exits with a validation error and does not create a file.

______________________________________________________________________

### User Story 4 - Prevent Overwriting Existing Manifest (Priority: P1)

A user accidentally runs `aipkg init` in a directory that already contains an `aipkg.json`. The command detects the existing file and refuses to proceed, displaying a clear error message. No file is modified or overwritten.

**Why this priority**: Data loss prevention is critical. This guard must be in place from the start.

**Independent Test**: Can be fully tested by creating a dummy `aipkg.json`, running `aipkg init`, and verifying the original file is untouched and the error message is displayed.

**Acceptance Scenarios**:

1. **Given** a directory containing an existing `aipkg.json`, **When** user runs `aipkg init`, **Then** the command exits with an error message indicating the file already exists and no changes are made.
1. **Given** a directory containing an existing `aipkg.json`, **When** user runs `aipkg init` with flags, **Then** the same guard applies — the command refuses regardless of flags.

______________________________________________________________________

### Edge Cases

- What happens when the user cancels the interactive flow mid-way (e.g., Ctrl+C)? No file should be written; the directory should remain unchanged.
- What happens when the current directory is not writable? The command should report a clear file-system error.
- What happens when running in a non-interactive terminal (no TTY) without providing all required flags? The command should exit with an error indicating the missing fields rather than hanging.
- What happens when the user provides an empty string for a required field (package name or version)? The command should re-prompt or error, not write an invalid manifest.
- What happens when a project name is provided that does not pass schema validation? The same schema rules apply regardless of type.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The command must prompt the user to choose between "project" and "package" type as the first step of the interactive flow.
- **FR-002**: For package type, the command must prompt in this order: scoped name (required), version (required), description (optional), license (optional).
- **FR-003**: For project type, the command must prompt for: name (optional) and description (optional).
- **FR-004**: All user-provided field values must be validated against the `aipkg-spec` manifest JSON schema before the manifest is written.
- **FR-005**: Validation must cover all schema-enforced constraints for the chosen type, including name format and version format. License is accepted as-is (SPDX by convention, not enforced by schema).
- **FR-006**: The command must generate a valid `aipkg.json` file in the current working directory containing only the fields relevant to the chosen type.
- **FR-007**: If an `aipkg.json` file already exists in the current directory, the command must refuse to proceed and display a clear error message.
- **FR-008**: All prompted fields must also be settable via command-line flags for non-interactive use.
- **FR-009**: When all required fields are provided via flags, the command must run without any interactive prompts.
- **FR-010**: When some but not all required fields are provided via flags, the command must prompt interactively only for the missing fields (hybrid mode).
- **FR-011**: The command must not perform any network operations.
- **FR-012**: The generated manifest must not include an `artifacts` field — artifacts are derived at package time, not during init. The `aipkg-spec` schema will be updated to make `artifacts` optional for packages, with presence enforced at package/publish time.
- **FR-013**: The command must provide clear, actionable validation error messages that explain what is wrong and what format is expected.
- **FR-014**: The command must provide a `--help` flag that documents all available flags, their defaults, and usage examples.
- **FR-015**: If the interactive flow is cancelled (e.g., user interrupt), no file must be written and the directory must remain unchanged.
- **FR-016**: Optional fields skipped during the interactive flow must be omitted from the generated manifest (not written as empty strings or null values).
- **FR-017**: The command must exit with code 0 on success and code 1 on any error (validation failure, file exists, filesystem error, missing required input).
- **FR-018**: On success, the command must print a short confirmation message to stdout indicating the file was created and the chosen type (e.g., `Created aipkg.json (package)`).
- **FR-019**: When flags irrelevant to the chosen type are provided (e.g., `--version` with `--type project`), the command must print a warning that the flag is ignored and proceed normally.
- **FR-020**: Interactive prompts must offer sensible default values where applicable: version defaults to `0.1.0`, license defaults to the identifier detected from a LICENSE file in the current directory (or blank if none found). Name always requires explicit input.

### Key Entities

- **Manifest (`aipkg.json`)**: The package/project configuration file. Structure, required/optional fields, and validation rules are defined by the `aipkg-spec` manifest JSON schema. The init command prompts for and writes only the fields relevant to the chosen type.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can create a valid manifest through the interactive flow.
- **SC-002**: 100% of generated manifest files pass the aipkg.json schema validation.
- **SC-003**: Users can create manifests entirely non-interactively by providing flags, with zero interactive prompts appearing.
- **SC-004**: All input that fails schema validation is caught and communicated before any file is written.
- **SC-005**: Running `aipkg init` in a directory with an existing `aipkg.json` never modifies the existing file.
- **SC-006**: The help output for the command clearly documents all flags and provides at least one usage example.

## Clarifications

### Session 2026-02-28

- Q: What exit codes should the command use? → A: Standard Unix: 0 for success, 1 for all errors.
- Q: What should the command output on successful manifest creation? → A: Short confirmation, e.g., `Created aipkg.json (package)` or `Created aipkg.json (project)`.
- Q: What should happen when type-irrelevant flags are provided (e.g., `--type project --version 1.0.0`)? → A: Warn but proceed — print a warning that the flag is ignored for the chosen type, then continue.
- Q: Should prompted fields offer default/suggested values? → A: Defaults where sensible — version defaults to `0.1.0`, license defaults to detected LICENSE file or blank, name always requires input.
- Q: In what order should the package fields be prompted? → A: name → version → description → license (required fields first, then optional).
- Q: How should we resolve the `artifacts` requirement for package manifests created by init? → A: Update `aipkg-spec` schema to make `artifacts` optional for packages. Enforce presence at package/publish time, not at init time. This allows init-generated manifests to pass full schema validation without including artifacts.

## Assumptions

- The `aipkg-spec` manifest JSON schema is the single source of truth for field definitions, required/optional status per type, and all validation rules. The init command does not implement its own validation logic — it delegates to the schema.
- The `require` and `repositories` fields are not prompted during init. They are added later via other commands (e.g., `aipkg require`).
- The generated JSON is formatted with 2-space indentation for readability and includes a trailing newline.
- This is the first command in the CLI and establishes the foundational patterns (command structure, flag handling, help system, error handling, logging) that all subsequent commands will follow.
