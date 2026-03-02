# CLI Contract: `aipkg pack`

**Branch**: `002-archive-format-pack` | **Date**: 2026-03-01

## Synopsis

```
aipkg pack [flags]
```

Packages the current directory (or the directory specified by the positional argument) into a distributable `.aipkg` archive.

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--output` | `-o` | string | Current directory, conventional filename | Output path. If a directory, writes the archive there with the conventional name. If a file path, writes the archive to that exact path. |

## Positional Arguments

| Position | Required | Default | Description |
|----------|----------|---------|-------------|
| 0 | No | `.` (current directory) | Path to the package directory to pack |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Archive created successfully |
| 1 | Validation or filesystem error (details printed to stderr) |

## Output Behavior

**Stdout**: None. The pack command does not write to stdout.

**Stderr**: Validation errors, progress messages, and the final archive path summary.

**Files created**:
1. `{scope}--{name}-{version}.aipkg` (zip archive)
2. `{scope}--{name}-{version}.aipkg.sha256` (SHA-256 sidecar)

## Archive Structure

For `@tjespers/test-writer` version `1.0.0`:

```
tjespers--test-writer-1.0.0.aipkg
└── test-writer/
    ├── aipkg.json          # Manifest with generated artifacts array
    ├── skills/
    │   └── test-writer/
    │       └── SKILL.md
    ├── prompts/
    │   └── code-review.md
    └── mcp-servers/
        └── github.json
```

The top-level directory name is the package name (part after `/` in `@scope/name`).

## Pipeline Order

1. Load `aipkg.json` from source directory
2. Validate manifest against `spec/schema/package.json`
3. Load `.aipkgignore` (if exists) + built-in defaults
4. Discover artifacts from well-known directories (filtered by ignore rules)
5. Validate each artifact (type-specific rules)
6. Report all validation errors (if any). Abort without producing an archive.
7. Build enriched manifest (inject `artifacts` array)
8. Validate enriched manifest against schema
9. Create zip archive
10. Write SHA-256 sidecar
11. Print summary to stderr

## Error Reporting Format

Validation errors are collected and reported together:

```
pack: 3 validation errors
skills/bad-skill/SKILL.md: missing required field 'name'
mcp-servers/github.json: invalid JSON: unexpected end of JSON input
prompts/empty.md: file must not be empty
```

Each error line follows the format `{relative-path}: {message}`.

## Examples

### Basic pack (from package directory)

```
$ cd my-package/
$ aipkg pack
tjespers--my-package-1.0.0.aipkg (2 artifacts, 4.2 KB)
```

### Pack with custom output

```
$ aipkg pack --output dist/
dist/tjespers--my-package-1.0.0.aipkg (2 artifacts, 4.2 KB)
```

### Pack from a different directory

```
$ aipkg pack ./packages/my-package
tjespers--my-package-1.0.0.aipkg (2 artifacts, 4.2 KB)
```

### Pack failure (validation errors)

```
$ aipkg pack
pack: 2 validation errors
skills/bad-skill/SKILL.md: frontmatter key 'author' is not allowed
prompts/: no artifact files found
Error: pack failed
```
