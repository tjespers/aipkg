# Research: Archive Format & Pack Command

**Branch**: `002-archive-format-pack` | **Date**: 2026-03-01

## Gitignore Pattern Matching Library

**Decision**: `github.com/sabhiram/go-gitignore`

**Rationale**: Lightweight (stdlib-only, zero transitive deps), simple API (`CompileIgnoreFile` + `MatchesPath`), and 160 stars with widespread use. The pack command reads a single `.aipkgignore` at the package root and checks paths against it. This is the simplest library for that job.

**Known limitation**: `*` is compiled to `.*` regex internally, which means it matches across path separators. Real gitignore uses a `wildmatch` algorithm where `*` stops at `/`. For the patterns `.aipkgignore` will see in practice (simple globs like `*.log`, `build/`, `dist/`, `*.tmp`), this doesn't cause incorrect behavior. If we ever need full gitignore compliance, the library is swappable behind the `internal/ignore/` package boundary.

**Alternatives considered**:

| Library | Status | Why not |
|---------|--------|---------|
| `go-git/v5` gitignore subpackage | Active | Pulls in all of go-git as transitive dependency. Far too heavy for a single-file pattern matcher. |
| `denormal/go-gitignore` | Dead (2018) | Abandoned, complex API designed for repo traversal. |
| `monochromegane/go-gitignore` | Dead (2020) | Does not support `**` recursive globs. Uses `filepath.Match` directly. |
| `git-pkgs/gitignore` | New (Feb 2026) | Correct wildmatch implementation, stdlib-only. Only 1 star, 6 weeks old. Technically superior but zero adoption track record. Worth re-evaluating later. |

## YAML Frontmatter Parsing

**Decision**: Hand-parse `---` delimiters + `gopkg.in/yaml.v3` with `Decoder.KnownFields(true)`

**Rationale**: The frontmatter extraction is ~25 lines of code (scan for opening `---`, collect lines until closing `---`, pass bytes to yaml decoder). No library adds meaningful value over this, and the dedicated frontmatter libraries have problems:

- `github.com/adrg/frontmatter` (177 stars) is unmaintained (last release Nov 2020), depends on `yaml.v2` (not v3), and pulls in `BurntSushi/toml` since it supports YAML+TOML+JSON formats simultaneously.

`yaml.v3`'s `Decoder.KnownFields(true)` rejects unknown keys when decoding into a struct. This maps directly to FR-015 (only allowed keys in SKILL.md frontmatter). The known bug with `KnownFields` and custom `UnmarshalYAML` methods (go-yaml issue #642) does not apply here since our struct has no custom unmarshalers.

**Struct mapping**:

```go
type SkillFrontmatter struct {
    Name          string         `yaml:"name"`
    Description   string         `yaml:"description"`
    License       string         `yaml:"license,omitempty"`
    Compatibility []string       `yaml:"compatibility,omitempty"`
    Metadata      map[string]any `yaml:"metadata,omitempty"`
    AllowedTools  []string       `yaml:"allowed-tools,omitempty"`
}
```

Hyphenated YAML keys (`allowed-tools`) map to Go fields via struct tags. `omitempty` on optional fields.

## Multi-Error Collection for Validation

**Decision**: `errors.Join()` (stdlib, Go 1.20+) with a lightweight `Collector` helper

**Rationale**: FR-020 requires reporting ALL validation errors, not just the first. `errors.Join()` concatenates errors with `\n` between each, which produces clean one-error-per-line CLI output. No third-party library needed.

`github.com/hashicorp/go-multierror` is effectively superseded by the stdlib; the maintainers updated docs in Nov 2025 to point users toward `errors.Join`.

**Pattern**:

```go
type Collector struct {
    errs []error
}

func (c *Collector) Add(path, message string) {
    c.errs = append(c.errs, fmt.Errorf("%s: %s", path, message))
}

func (c *Collector) Err() error {
    return errors.Join(c.errs...)
}
```

Each error carries its file path as context (e.g., `skills/bad-skill/SKILL.md: missing required field 'name'`). The collector returns `nil` when empty (no errors) since `errors.Join` discards nil inputs. Cobra's `RunE` propagates the non-nil error to a non-zero exit code automatically.

For user-friendly output, the pack command prints a count header to stderr before returning the joined error:

```
pack: 3 validation errors
skills/bad-skill/SKILL.md: missing required field 'name'
mcp-servers/github.json: invalid JSON at line 3
prompts/empty.md: file must not be empty
```

## Zip Archive Creation

**Decision**: stdlib `archive/zip` with `deflate` compression

**Rationale**: The standard library's `archive/zip` package is sufficient. No third-party library needed. The archive structure (single top-level directory containing all files) maps to a straightforward walk-and-add pattern.

Key details:
- Create a `zip.Writer` wrapping an `os.File`
- For each file, call `zip.Writer.Create` with the archive-relative path (prefixed by the top-level directory name)
- The top-level directory name is the package name (the part after `/` in the scoped name)
- Use `zip.Deflate` method (standard compression)
- After closing the zip writer, compute SHA-256 over the finished file and write the sidecar

## SHA-256 Sidecar Generation

**Decision**: stdlib `crypto/sha256` with `sha256sum`-compatible format

**Rationale**: Hash the archive file, write `{hex_hash}  {filename}\n` (two spaces between hash and filename) to `{archive}.sha256`. This matches the output format of the `sha256sum` CLI tool, making verification trivial: `sha256sum -c archive.aipkg.sha256`.

## Manifest Loading

**Decision**: Add `LoadFile(path)` to `internal/manifest/`

**Rationale**: The existing `PackageManifest` struct only writes (via `WriteFile` and `MarshalIndent`). The pack command needs to read an existing `aipkg.json`, add the `Artifacts` field, and serialize the enriched copy into the archive. Adding `LoadFile` keeps manifest read/write in one package. The `Artifacts` field is added to the struct with `omitempty` so existing `create` output (which has no artifacts) is unaffected.
