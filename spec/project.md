# Project File (`aipkg-project.json`)

The `aipkg-project.json` file identifies a directory as an aipkg-enabled project. It declares which packages the project depends on and tracks the schema version for future migration support.

Projects and packages are separate concepts. A project *consumes* packages. A package *provides* artifacts. The two files (`aipkg-project.json` and `aipkg.json`) are mutually exclusive: they cannot coexist in the same directory. The CLI refuses to create one when the other already exists, in both directions.

## Minimal example

```json
{
  "specVersion": 1,
  "require": {}
}
```

This is what `aipkg init` creates. An empty project with no dependencies, ready for `aipkg require` to add packages.

## Fields

### `specVersion`

**Required.** The project file format version. Currently `1`.

This field allows the tooling to detect and handle format changes as the specification evolves. Same concept as `specVersion` in the [package manifest](manifest.md).

```json
"specVersion": 1
```

### `require`

**Required.** An object mapping scoped package names to pinned semver versions. Can be empty.

```json
"require": {
  "@shiftbase/golang-expert": "1.2.0",
  "@alice/code-review": "0.3.0-beta.1"
}
```

**Keys** are scoped package names matching the same format as package names: `@scope/package-name`. See [Naming & Namespaces](naming.md) for the full rules.

**Values** are semver versions: `MAJOR.MINOR.PATCH` with optional pre-release identifiers. Pre-release versions follow the [SemVer 2.0.0](https://semver.org/) specification. Build metadata (the `+build` suffix) is not supported because it has no ordering semantics.

Valid version examples:
- `1.0.0` (plain release)
- `1.0.0-beta.1` (numeric pre-release)
- `1.0.0-alpha` (named pre-release)
- `1.0.0-rc.1` (release candidate)

Version ranges and prefixes (`^1.0.0`, `~2.1`, `v1.0.0`) are not supported. All versions are exact pins.

No additional fields are allowed. The project file has no identity fields (`name`, `version`, `description`, `license`). It is purely operational.

## Example with dependencies

```json
{
  "specVersion": 1,
  "require": {
    "@shiftbase/golang-expert": "1.2.0",
    "@alice/code-review": "0.3.0-beta.1",
    "@bob/deploy-tools": "2.0.0"
  }
}
```

## JSON Schema

The machine-readable schema lives at [`schema/project.json`](schema/project.json). It uses JSON Schema Draft 2020-12 and validates both field types and value patterns (package name format, semver regex).

## Initialization

Run `aipkg init` in a project directory to create the file. The command:

1. Checks for an existing `aipkg.json` (package manifest). If found, it refuses with an error suggesting `aipkg require` or `aipkg install` instead.
2. Checks for an existing `aipkg-project.json`. If found, it refuses to overwrite.
3. Creates `aipkg-project.json` with `specVersion: 1` and an empty `require` map.

No directories are created. No prompts are shown. No configuration is needed.

## Install directory (`.aipkg/`)

When packages are installed, their artifacts are placed in the `.aipkg/` directory at the project root. This directory is not created by `aipkg init`. It materializes on demand when the first package is installed.

### Layout

```text
.aipkg/
├── .gitignore              # Ignores all contents except itself
├── skills/                 # Individual skill directories
│   ├── scope.pkg.skill-a/
│   │   └── SKILL.md
│   └── scope.pkg.skill-b/
│       └── SKILL.md
├── prompts/                # Individual prompt files
│   └── scope.pkg.code-review.md
├── commands/               # Individual command files
│   └── scope.pkg.commit-msg.md
├── agents/                 # Individual agent persona files
│   └── scope.pkg.go-expert.md
├── mcp.json                # Merged MCP server configs (all packages)
└── agent-instructions.md   # Merged agent instructions (all packages)
```

Four artifact types (skill, prompt, command, agent) are placed as individual files or directories in type-specific subdirectories. Two mergeable types (mcp-server, agent-instructions) contribute to shared files at the `.aipkg/` root.

### `.gitignore`

When `.aipkg/` is first created inside a git repository, a `.gitignore` file is placed inside it:

```text
*
!.gitignore
```

This prevents accidental commits of installed artifacts. The `.gitignore` itself is the only committed file inside `.aipkg/`.

### Merged files

`mcp.json` and `agent-instructions.md` at the `.aipkg/` root are fully aipkg-managed. They are generated and overwritten by install and update operations. Manual edits to these files will be lost. If you need custom MCP server configs or agent instructions, maintain them outside `.aipkg/`.

## Scoped artifact naming

Installed artifacts use a three-segment dot-notation that incorporates the source package identity:

```text
scope.package-name.artifact-name
```

This naming convention serves two purposes:

1. **Collision prevention.** Two packages from the same scope can both have an artifact named `test-writer`. The three-segment name (`alice.pkg-a.test-writer` vs. `alice.pkg-b.test-writer`) keeps them distinct.
2. **Traceability.** Given any installed artifact name, you can reconstruct the source package. `alice.blog-tools.code-review` traces back to `@alice/blog-tools`.

### Parsing

Split on `.`. The first segment is the scope, the last segment is the artifact name, and the middle segment is the package name. This is unambiguous because scope names and package names cannot contain dots (dots are reserved as the namespace separator).

### Examples

Given a package `@alice/blog-tools` with artifacts `code-review` (prompt) and `test-writer` (skill):

```text
.aipkg/prompts/alice.blog-tools.code-review.md
.aipkg/skills/alice.blog-tools.test-writer/SKILL.md
```

Given a package `@bob/deploy-tools` with an artifact `deploy` (command):

```text
.aipkg/commands/bob.deploy-tools.deploy.md
```

For merged types, the scoped name appears as a key or marker within the merged file rather than as a filename.

## Mutual exclusivity

The project file (`aipkg-project.json`) and the package manifest (`aipkg.json`) cannot coexist in the same directory. They serve different purposes for different personas:

- **Project file**: "I consume packages. Here are my dependencies."
- **Package manifest**: "I am a package. Here is my identity and what I contain."

The CLI enforces this in both directions. `aipkg init` refuses to create a project file when a package manifest exists. Future commands that create package manifests will refuse when a project file exists.
