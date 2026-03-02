# Archive Format (`.aipkg`)

The `.aipkg` archive is the distributable unit of an aipkg package. It is a standard zip file containing a single top-level directory with the package manifest and all artifact content.

## Format

- Standard zip format with deflate compression
- UTF-8 filename encoding within the archive
- File extension: `.aipkg`
- Companion sidecar: `.aipkg.sha256`

Any tool that reads zip files can inspect an `.aipkg` archive. The format is intentionally simple so third-party tools can produce or consume archives without depending on aipkg.

## Filename convention

Archive filenames follow this pattern:

```
{scope}--{name}-{version}.aipkg
```

The `@` prefix from the scoped package name is stripped, and the `/` separator is replaced with `--` (double dash). For example, `@tjespers/dummy` version `1.2.3` produces:

```
tjespers--dummy-1.2.3.aipkg
```

### Why this works

Both scope and package name forbid consecutive hyphens (per the [naming rules](naming.md)), so `--` never appears inside either component. This makes the scope/name boundary unambiguous.

Versions are strict semver (MAJOR.MINOR.PATCH) with no pre-release or build metadata, so the version contains dots but no hyphens. The name-version boundary splits at the last hyphen preceding a digit-dot sequence.

### Parsing algorithm

Given a filename like `tjespers--dummy-1.2.3.aipkg`:

1. Strip the `.aipkg` extension.
2. Split on `--` to separate scope from the rest: `tjespers` and `dummy-1.2.3`.
3. Find the last `-` followed by a digit. Split there: `dummy` and `1.2.3`.
4. Reconstruct: `@tjespers/dummy` version `1.2.3`.

The manifest inside the archive remains authoritative for package identity. The filename is a convenience for humans and tooling, not the source of truth.

## Archive structure

A single top-level directory wraps all package contents. The directory name is the package name (the part after `/` in `@scope/name`), without scope or version. This follows Helm's convention where the chart archive's top-level directory matches the chart name.

For `@tjespers/test-writer` version `1.0.0`:

```
tjespers--test-writer-1.0.0.aipkg
└── test-writer/
    ├── aipkg.json
    ├── skills/
    │   └── test-writer/
    │       └── SKILL.md
    ├── prompts/
    │   └── code-review.md
    └── mcp-servers/
        └── github.json
```

### What gets included

The archive contains:

- The `aipkg.json` manifest (with the generated `artifacts` array)
- All discovered artifact files and directories

Non-artifact files at the package root (README.md, LICENSE, etc.) are not included in the archive in v1. Only content within the well-known artifact directories makes it into the archive.

### Enriched manifest

The `aipkg.json` inside the archive contains a generated `artifacts` array that describes the package contents. This array is produced by `aipkg pack` during archive creation. The original `aipkg.json` on disk is never modified.

See [Package Manifest](manifest.md) for the `artifacts` field format.

## Extraction behavior

When extracting an archive, the CLI strips the top-level directory and places contents directly into the target location. If you extract `tjespers--test-writer-1.0.0.aipkg` into `/some/path/`, the result is:

```
/some/path/
├── aipkg.json
├── skills/
│   └── test-writer/
│       └── SKILL.md
├── prompts/
│   └── code-review.md
└── mcp-servers/
    └── github.json
```

## Integrity verification

Each archive is accompanied by a `.sha256` sidecar file containing the SHA-256 hash in `sha256sum` format:

```
a1b2c3d4e5f6...  tjespers--test-writer-1.0.0.aipkg
```

The format is: lowercase hex hash, two spaces, archive basename (not the full path). The file uses UTF-8 encoding with a single LF-terminated line.

This matches the output format of the standard `sha256sum` command-line tool, so verification is straightforward:

```sh
sha256sum -c tjespers--test-writer-1.0.0.aipkg.sha256
```

## Creating archives

Use `aipkg pack` to create archives:

```sh
# Pack the current directory
aipkg pack

# Pack a specific directory
aipkg pack ./my-package

# Write to a custom location
aipkg pack --output dist/
```

The pack command validates all artifacts before creating the archive. If any validation fails, no archive is produced. See `aipkg pack --help` for full usage.
