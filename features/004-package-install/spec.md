# Feature Specification: Package Install (Dist Strategy)

**Feature Branch**: `004-package-install`
**Created**: 2026-03-08
**Status**: Draft
**Input**: User description: "Close the pack-to-install loop. Developers run `aipkg require @scope/name` to add a package and install it, or `aipkg install` (no args) to install everything listed in the project file's `require` field. Covers only `type: \"package\"` entries (pre-built `.aipkg` archives from the registry index). Source/recipe strategy is a separate effort (AIPKG-7)."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Add a package to a project (Priority: P1)

A developer working on a project wants to use an AI skill published to the aipkg registry. They run `aipkg require @tjespers/golang-expert` in their project root. The CLI resolves the package from the registry, downloads the `.aipkg` archive to the global cache, verifies its integrity, places the package's artifacts into the project's categorized directories (`.aipkg/skills/`, `.aipkg/prompts/`, etc.), and records the dependency in `aipkg-project.json`. The developer now has the package's artifacts available in their project.

**Why this priority**: This is the core value proposition. Without the ability to add a package, the CLI is a pack-only tool. Everything else (bulk install, version pinning) builds on this flow.

**Independent Test**: Can be fully tested by setting up a local static file server with a per-package JSON file and a corresponding `.aipkg` archive, pointing `AIPKG_REGISTRY` at it, initializing a project with `aipkg init`, and running `aipkg require @scope/name`. Verify that the package's artifacts are placed in the categorized directories and the project file is updated.

**Acceptance Scenarios**:

1. **Given** an initialized project (with `aipkg-project.json`) and a registry containing `@tjespers/golang-expert` at version `1.0.0` with a `debug` skill artifact, **When** the developer runs `aipkg require @tjespers/golang-expert`, **Then** the archive is cached, the skill artifact is placed at `.aipkg/skills/debug/`, and `aipkg-project.json` contains `"@tjespers/golang-expert": "1.0.0"` in its `require` field.
2. **Given** a successful require of a package containing multiple artifact types, **When** the developer inspects `.aipkg/`, **Then** each artifact is placed in the appropriate categorized directory (`.aipkg/skills/`, `.aipkg/prompts/`, etc.) using its original artifact name, and mergeable artifacts (mcp-server, agent-instructions) are merged into their respective files.
3. **Given** a successful require, **When** the developer inspects `.aipkg/`, **Then** a `.gitignore` file exists that ignores all contents (if the project is inside a git repository).
4. **Given** a package already required at version `1.0.0`, **When** the developer runs `aipkg require @tjespers/golang-expert` and the registry now has `1.1.0` as the latest, **Then** the old artifacts are removed, the new version's artifacts are placed, and `aipkg-project.json` is updated to reflect `1.1.0`.
5. **Given** a package already installed at the latest version, **When** the developer runs `aipkg require @tjespers/golang-expert`, **Then** the command reports that the package is already up to date and makes no changes.

---

### User Story 2 - Install all project dependencies (Priority: P2)

A developer clones or checks out a project that has packages listed in its `aipkg-project.json`. They run `aipkg install` to install all dependencies at their pinned versions. This is the reproducibility story: any developer can set up the same package environment from the project file alone.

**Why this priority**: Without bulk install, collaborators would need to run `aipkg require` for each dependency individually. This command makes project setup a single step.

**Independent Test**: Can be tested by creating an `aipkg-project.json` with multiple entries in `require`, running `aipkg install`, and verifying all packages are installed at the specified versions.

**Acceptance Scenarios**:

1. **Given** an `aipkg-project.json` with `require` containing `"@tjespers/golang-expert": "1.0.0"` and `"@alice/blog-writer": "2.0.0"`, **When** the developer runs `aipkg install`, **Then** both packages are resolved, downloaded, verified, and their artifacts are placed in the project's categorized directories.
2. **Given** some packages are already installed at the correct versions, **When** the developer runs `aipkg install`, **Then** already-installed packages are skipped and only missing or outdated packages are downloaded.
3. **Given** an `aipkg-project.json` with an empty `require` object, **When** the developer runs `aipkg install`, **Then** the command completes successfully with a message indicating no packages to install.

---

### User Story 3 - Pin a specific package version (Priority: P3)

A developer needs a specific version of a package rather than the latest. They run `aipkg require @tjespers/golang-expert@1.0.0` to add and install that exact version.

**Why this priority**: Version pinning is essential for reproducibility, but most developers will use the default (latest) on initial require. Explicit pinning is a more deliberate action.

**Independent Test**: Can be tested by setting up a registry with multiple versions and verifying that `aipkg require @scope/name@1.0.0` installs the requested version rather than the latest.

**Acceptance Scenarios**:

1. **Given** a registry containing `@tjespers/golang-expert` at versions `1.0.0` and `1.1.0`, **When** the developer runs `aipkg require @tjespers/golang-expert@1.0.0`, **Then** version `1.0.0` is installed and recorded in `aipkg-project.json`.
2. **Given** a version that does not exist in the registry, **When** the developer runs `aipkg require @tjespers/golang-expert@9.9.9`, **Then** the command fails with a clear error listing the available versions.

### Edge Cases

- What happens when no `aipkg-project.json` exists? Both `aipkg require` and `aipkg install` refuse with a clear error directing the developer to run `aipkg init` first.
- What happens when the registry is unreachable (network error, DNS failure)? The command fails with an error describing the network issue and the registry URL it tried to reach.
- What happens when the per-package index entry returns a non-200 HTTP status? A 404 means the package is not found. Other errors are reported as registry errors with the HTTP status.
- What happens when the SHA-256 hash of the downloaded archive does not match the hash in the index entry? The command refuses to extract, removes the file from the cache, and reports an integrity verification failure.
- What happens when the archive URL in the dist block is unreachable? The command fails with a download error showing the URL that failed.
- What happens when the developer has no write permission on the project directory? The command fails with a filesystem error.
- What happens when `aipkg require` is run in a package directory (one containing `aipkg.json`)? The command refuses with an error explaining that `require` operates on projects, not packages. This mirrors the mutual exclusivity guard from `aipkg init`.
- What happens when `aipkg install` encounters a failure partway through (e.g., second of three packages fails)? Packages that installed successfully before the failure remain installed. The error reports which package failed and which packages still need to be installed.
- What happens when the `AIPKG_REGISTRY` environment variable is set to an invalid URL? The command fails with an error describing the invalid registry URL.
- What happens when the developer runs `aipkg require` with an invalid package name (no scope, uppercase characters, consecutive hyphens)? The command fails immediately with a validation error describing the naming rules, before attempting any registry request.
- What happens when the downloaded archive passes SHA-256 verification but is not a valid `.aipkg` file (e.g., missing `aipkg.json`, no single top-level directory)? The command refuses to extract and reports a structural validation error.
- What happens when two installed packages contain artifacts of the same type and name (e.g., both provide a skill called `review`)? The second `aipkg require` fails with an error identifying both packages and the conflicting artifact. The first package's artifacts remain installed.
- What happens when the per-package index entry has a `type` other than `"package"`? The command fails with an error indicating the entry type is not supported.

## Requirements *(mandatory)*

### Functional Requirements

**Per-package index format:**

- **FR-001**: The per-package index entry format for `type: "package"` entries MUST be defined and documented. Each entry contains the package name, description, type indicator, and a versions map. Each version entry contains a `dist` block with the archive download URL and its SHA-256 hash. This format is the contract between the registry and the CLI.
- **FR-002**: A JSON Schema for the per-package index entry MUST be provided to enable validation by the CLI and third-party registry tooling.

**Registry resolution:**

- **FR-003**: The CLI MUST resolve packages by fetching a per-package metadata file from the registry. The URL is constructed as `{registry_base_url}/@scope/name.json`, where `{registry_base_url}` defaults to `https://packages.aipkg.dev`.
- **FR-004**: The `AIPKG_REGISTRY` environment variable MUST override the default registry URL. This is the only registry configuration mechanism in v1 (full `repositories` configuration is deferred).
- **FR-005**: When no version is specified, the CLI MUST select the latest version by choosing the highest semver key from the `versions` map in the index entry.
- **FR-006**: When a specific version is requested (via `@version` syntax), the CLI MUST select exactly that version from the `versions` map. If the requested version does not exist, the command MUST fail with an error listing the available versions.

**Archive download and verification:**

- **FR-007**: The CLI MUST download the `.aipkg` archive from the URL specified in the selected version's `dist` block.
- **FR-008**: The CLI MUST verify the SHA-256 hash of the downloaded archive against the `sha256` value in the index entry's `dist` block. If the hash does not match, the archive MUST NOT be extracted and the command MUST fail with an integrity error.
- **FR-009**: After hash verification, the CLI MUST validate that the downloaded archive is a structurally valid `.aipkg` file (valid zip, single top-level directory, contains `aipkg.json`). A structurally invalid archive MUST NOT be extracted.
- **FR-010**: The default registry URL MUST use HTTPS. When following HTTP redirects, the CLI MUST NOT downgrade an HTTPS URL to HTTP. Dist URLs using plain HTTP are permitted to support local development and testing with static file servers.
- **FR-011**: The CLI MUST validate that the per-package index entry's `type` field is `"package"`. If the type is not `"package"` (e.g., a future `"recipe"` type), the CLI MUST fail with a clear error indicating that the entry type is not supported by this command.

**Archive caching:**

- **FR-012**: Downloaded `.aipkg` archives MUST be stored in a global cache directory at `~/.aipkg/cache/`. The cache directory MUST be created on demand. Archives MUST be named using a scheme that includes scope, package name, and version to avoid collisions. If a verified archive for the requested version already exists in the cache, the download MUST be skipped.

**Artifact placement:**

- **FR-013**: After verification, the CLI MUST read the `artifacts` array from the archive's `aipkg.json` manifest and place each artifact into the corresponding categorized directory within `.aipkg/` using the artifact's original name from the manifest. Directory-type artifacts (skills) are placed as directories; file-type artifacts (prompts, commands, agents) are placed as files. Mergeable artifact types (mcp-server, agent-instructions) are merged into their respective project-level files (`.aipkg/mcp.json`, `.aipkg/agent-instructions.md`), keyed by the artifact's original name. Before placing any artifact, the CLI MUST check for name collisions: if another already-installed package has an artifact of the same type and name, the command MUST fail with an error identifying both packages and the conflicting artifact name. No content transformation is applied to artifacts during placement; they are placed exactly as they appear in the archive.
- **FR-014**: If a package is already installed at a different version, the CLI MUST remove all artifacts belonging to the previous version before placing the new version's artifacts. The previous version's artifacts are identified by reading the `artifacts` array from the cached archive's `aipkg.json` manifest. Each artifact is removed from its categorized directory, and the package's contributions to merged files are removed by key.
- **FR-015**: If a package is already installed at the same version that would be installed, the command MUST skip the operation and report that the package is already up to date. The metadata fetch still occurs to determine which version is current. The installed version is determined from the `require` field in `aipkg-project.json`.

**Install directory management:**

- **FR-016**: The `.aipkg/` directory and its categorized subdirectories MUST be created on demand when a package is installed. If the project root is inside a git repository, a `.gitignore` file MUST be placed inside `.aipkg/` that ignores all contents (per the project initialization specification, FR-009).

**Project file management:**

- **FR-017**: `aipkg require @scope/name` MUST add or update the package entry in `aipkg-project.json`'s `require` field with the resolved version.
- **FR-018**: `aipkg install` (no arguments) MUST read the `require` field from `aipkg-project.json` and install all listed packages at their pinned versions.
- **FR-019**: If `aipkg install` fails partway through (e.g., one package in a multi-package install fails to resolve or download), packages that were successfully installed before the failure MUST be retained. The error MUST identify which package failed and which packages remain uninstalled.
- **FR-020**: Both `aipkg require` and `aipkg install` MUST require an existing `aipkg-project.json` in the current directory. If none exists, the command MUST fail with an error directing the developer to run `aipkg init`.
- **FR-021**: Both commands MUST refuse to operate in a directory containing `aipkg.json` (a package manifest), mirroring the mutual exclusivity guard from the project initialization specification.
- **FR-022**: When writing to `aipkg-project.json`, the CLI MUST preserve existing content (other `require` entries, `specVersion`) and only modify the relevant entry. The file MUST be written with consistent formatting (indented JSON).

**Documentation:**

- **FR-023**: The per-package index entry format, resolution algorithm, artifact placement layout within `.aipkg/`, and archive caching behavior MUST be documented as reference material in `spec/`. This documentation is a deliverable of this feature.

### Key Entities

- **Per-Package Index Entry**: A JSON document served by the registry at `{base_url}/@scope/name.json`. Contains the package name, description, type (`"package"` for dist strategy entries), and a versions map where each key is a semver version string and each value contains a dist block with the archive URL and SHA-256 hash.
- **Dist Block**: A structure within a version entry that provides the download URL for the `.aipkg` archive and the expected SHA-256 hash for integrity verification. The URL can point to any HTTP or HTTPS endpoint (GitHub Releases, a CDN, a static file server, a local test server).
- **Installed Package**: A set of artifacts placed in the project's `.aipkg/` categorized directories using their original artifact names. The installed version is tracked in `aipkg-project.json`'s `require` field. The original `.aipkg` archive is preserved in the global cache at `~/.aipkg/cache/`, and its manifest is used to identify which artifacts belong to the package.
- **Archive Cache**: The global directory at `~/.aipkg/cache/` where downloaded and verified `.aipkg` archives are stored. Prevents re-downloading archives that have already been fetched and verified.

## Assumptions

- The project initialization feature (003) has established `aipkg-project.json` as the project file with `specVersion` and `require` fields. This feature reads from and writes to that file. It does not modify the format.
- The archive format (002) is stable. Archives contain a single top-level directory, and extraction strips it.
- Package-level dependencies (bundled at pack time) are a separate concern. When that capability is designed, it will introduce its own schema fields and specification. This feature does not depend on or interact with package-level dependencies.
- Adapter execution (AIPKG-11) is entirely separate. This feature places artifacts into `.aipkg/` categorized directories (the natural aipkg format). Adapters will later bridge from `.aipkg/` to tool-specific directories (`.claude/`, `.cursor/`, etc.) by creating symlinks and tool-specific configurations.
- The central registry at `packages.aipkg.dev` will serve static JSON files. The CLI does not depend on any server-side logic; any static file server can act as a registry.
- Metadata is always fetched fresh from the registry. No metadata caching is implemented in v1. Archive caching (in `~/.aipkg/cache/`) avoids re-downloading archives that have already been verified. No cache eviction or size management is implemented in v1.
- Network behavior (timeouts, retries, progress indication) is not specified in v1. Every operation fetches fresh data with default HTTP client behavior. This is a conscious deferral.
- Package names in the index entry's `name` field follow the established naming rules from `spec/naming.md`. The CLI reuses existing name validation logic.
- The `require` field in `aipkg-project.json` uses strict semver (MAJOR.MINOR.PATCH) for dist strategy packages. Pre-release identifiers in require values are allowed by the project schema but will not appear in practice until the source/recipe strategy is implemented.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can add a package to their project with a single command (`aipkg require @scope/name`) and no prior setup beyond `aipkg init`. The package is immediately available on disk.
- **SC-002**: A developer can reproduce another developer's package environment by cloning a project and running `aipkg install`. All packages are installed at the exact versions specified in the project file.
- **SC-003**: Every installed package has its integrity verified before extraction. A tampered archive is never unpacked.
- **SC-004**: The CLI provides clear, actionable error messages for all failure modes: package not found, version not found, network errors, integrity failures, missing project file.
- **SC-005**: The per-package index entry format and resolution algorithm are documented clearly enough that a third party could implement a compatible registry without access to the aipkg source code.
- **SC-006**: Installing a package that is already present at the correct version completes without re-downloading, confirming idempotent behavior.
- **SC-007**: A developer can test the full require/install flow locally by running a static file server and setting `AIPKG_REGISTRY`, with no dependency on the production registry.
