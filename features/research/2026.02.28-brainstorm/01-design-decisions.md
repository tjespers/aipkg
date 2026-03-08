# aipkg — Design Decisions (2026-02-28)

Decisions made during the CLI scaffold session and archive format spec session. Builds on top of the 2026-02-26 brainstorm.

## Settled

### 18. CLI Constitution Ratified

**Decision:** The CLI has a formal constitution (v1.0.0) defining five core principles in priority order: Simplicity and Deferral > Core/Adapter Separation > Convention Over Invention > Cold Start First > Backward-Compatible Evolution.

**Key constraints:**
- When principles conflict, higher priority wins. Document the conflict and resolution.
- Features are either fully implemented or fully absent. No stubs, feature flags, or "coming soon."
- Core packages must not import adapter packages. Enforced by Go's import graph.
- The spec repo is upstream authority. If the CLI finds a gap, the fix is a spec PR, not a CLI hack.

**Rationale:** Codifies the design ethos from the brainstorm into enforceable rules. The priority ordering resolves common tensions (e.g., simplicity vs convention, cold start vs separation) without case-by-case debate.

### 19. Viper Removed, Cobra-Only Flag Handling

**Decision:** Dropped `viper` from the CLI stack. Flags are handled via cobra's native `cmd.Flags().GetString()`.

**Rationale:** Viper's global state model causes flag collisions when multiple commands share flag names, and its dependency tree (afero, fsnotify, mapstructure, go-toml, gotenv, etc.) added ~2 MB for zero value. Cobra alone covers everything the CLI needs for flag parsing, help generation, and subcommand routing.

### 20. huh for Interactive Prompts (Not Raw Bubbletea)

**Decision:** Use `github.com/charmbracelet/huh` for interactive terminal forms, not raw bubbletea.

**Rationale:** huh provides high-level form abstractions (select, input, text) with built-in validation, keyboard navigation, and cancellation handling. Survey is archived. Promptui is less polished. Raw bubbletea is too low-level for form-style prompts.

**Pattern:** Build form dynamically based on type and pre-filled flags. Each field gets a validation function. Form runs once; values merged with flag values after completion.

### 21. JSON Schema Validation via santhosh-tekuri/jsonschema/v6

**Decision:** Use `github.com/santhosh-tekuri/jsonschema/v6` for manifest validation. Schema is embedded via `go:embed`.

**Rationale:** Supports JSON Schema Draft 2020-12 (which the aipkg-spec schema uses). Provides structured error output with JSON pointers. Alternatives (xeipuv/gojsonschema, qri-io/jsonschema) lack Draft 2020-12 support.

**Pattern:** Compile schema once with `sync.Once`. Per-field validation extracts regex patterns from the schema for inline prompt feedback. Full schema validation runs on the assembled manifest before writing.

### 22. License Detection via google/licensecheck

**Decision:** Use `github.com/google/licensecheck` to detect SPDX identifiers from LICENSE files for `aipkg init` defaults.

**Rationale:** Lightweight, single-file scan, returns SPDX IDs directly. Maintained by Google, used by Go ecosystem tooling. The heavier `go-license-detector` scans whole directory trees, which is overkill.

**Pattern:** Read LICENSE in cwd. If confidence > 80%, use the SPDX ID as the default license value in the init prompt. Otherwise leave blank.

### 23. Artifacts Optional at Init, Required at Pack/Publish

**Decision:** The `aipkg-spec` schema was updated to make `artifacts` optional for package manifests. Artifact presence is enforced at pack/publish time, not at init time.

**Rationale:** `aipkg init` creates a manifest before the author has written any artifacts. Requiring artifacts at init forces either a stub value or a two-step workflow. Making it optional means every lifecycle stage produces a schema-valid manifest.

**Schema change:** `artifacts` removed from the `required` array in the JSON Schema's package conditional.

### 24. Init Command Defers Global Config

**Decision:** `aipkg init` does not read or write a global config file (`~/.aipkg/config.json`). It creates purely local manifests. Global config support is deferred to a future `aipkg config` command.

**Rationale:** Simplicity and Deferral (Principle I). The init command's job is to create a valid `aipkg.json` in the current directory. Global defaults (preferred scope, license, author) are nice-to-have, not v1-critical.

**Tracked as:** AIPKG-46

### 25. Archive Structure (Helm-Style)

**Decision:** The `.aipkg` archive contains a single top-level directory with `aipkg.json` inside it. This is the natural result of zipping a package folder.

```
golang-expert-1.0.0.aipkg (zip)
└── golang-expert/
    ├── aipkg.json
    ├── skills/
    └── ...
```

**Rationale:** Same pattern as Helm charts (`chartname/Chart.yaml` inside the `.tgz`). Zipping a folder is the most natural thing to do. The CLI strips the top-level directory on extraction.

**Prior art:** Helm charts, `.vsix` packages.

### 26. Archive Filename Is Advisory

**Decision:** The recommended filename convention is `{package-name}-{version}.aipkg`, but the CLI does not parse or enforce it. The `aipkg.json` inside the archive is always authoritative for identity and version.

**Rationale:** Helm enforces filename = manifest, but it adds complexity with no real security benefit. The manifest is inside the archive, so you've already downloaded it by the time you'd check. `aipkg pack` produces the right filename automatically; mismatches only happen with hand-rolled archives, and erroring on those is annoying, not helpful.

### 27. Top-Level Directory Name Is Not Significant

**Decision:** The name of the single top-level directory inside the archive does not need to match the package name. The manifest is authoritative.

**Rationale:** Same as Helm. The docs say the directory "should" match the chart name, but it's convention, not enforced. Neither Helm nor aipkg validate the directory name on install.

### 28. SHA-256 Checksums via Sidecar File

**Decision:** Packages use SHA-256 checksums for integrity verification. The checksum is published as a `.sha256` sidecar file alongside the archive (e.g., `golang-expert-1.0.0.aipkg.sha256`), using standard `sha256sum` format.

**Behavior:**
- When available, the CLI verifies the download before extraction. Mismatch = error.
- When not available, the CLI proceeds with a warning. Keeps publishing simple for early adopters.

**Rationale:** SHA-256 is the standard (Docker, Helm, Go modules all use it). The sidecar file is a stopgap for GitHub Releases where there's no index file to embed digests in.

**Future direction:** When a registry or Helm-style index file exists, digests move into the index (like Helm's `index.yaml` with its `digest` field). The sidecar becomes a fallback for bare GitHub-only repos.

### 29. No Version Subfolder in Install Path

**Decision:** Packages install to `.aipkg/@{scope}/{package-name}/` with no version nesting. Only one version of a package can be installed at a time. Installing a different version replaces the existing one.

**Old layout (dropped):**
```
.aipkg/@tjespers/golang-expert/1.0.0/
```

**New layout:**
```
.aipkg/@tjespers/golang-expert/
```

**Rationale:** AI artifacts aren't isolated code modules. They're consumed by agents that read everything in scope. Two versions of the same skill means duplicate instructions flooding the context. Two MCP server configs for the same tool means conflicts. Multiple versions is actively harmful, not a future feature. Combined with exact version pinning (decision #11), there's never a valid case for two versions in v1. Don't design for hypothetical future requirements.

### 30. Signing and Provenance Deferred to v2+

**Decision:** Out of scope for v1. Helm's approach (PGP-signed `.prov` sidecar files) is the likely model when we get there.

**Rationale:** Signing requires key management infrastructure that doesn't exist yet. SHA-256 checksums handle integrity. Authentication (who published this?) can wait until the ecosystem has enough packages to make impersonation a real threat.

### 31. Merge aipkg-spec into aipkg Repo

**Decision:** The separate `aipkg-spec` repo should be merged into the main `aipkg` CLI repo as `docs/` and `schema/` directories.

**Rationale:** No CNCF project with a single implementation maintains a separate spec repo. Helm, Flux, Argo, containerd, Prometheus all keep spec and implementation together. Separate spec repos only make sense for multi-vendor standards with independent implementations (OCI, CloudEvents, OpenTelemetry). The two-repo setup adds overhead with no benefit at this stage. Can always extract later if needed.

**Tracked as:** AIPKG-49

## Prior Art Comparison (Helm)

The archive format closely mirrors Helm's chart packaging model:

| Aspect | Helm | aipkg |
|---|---|---|
| Format | `.tgz` | `.aipkg` (zip) |
| Naming | `{name}-{version}.tgz` | `{name}-{version}.aipkg` |
| Structure | Single top-level dir, `Chart.yaml` inside | Single top-level dir, `aipkg.json` inside |
| Dir name matters? | No | No |
| Manifest is authoritative | Yes | Yes |
| Contents | Bundle of arbitrary K8s resources + templates | Bundle of arbitrary AI artifacts |
| Versioning | SemVer 2 | Strict SemVer |
| Deps | `dependencies` in Chart.yaml | `require` in aipkg.json |
| Repo model | Static files: `index.yaml` + archives, gh-pages friendly | Static files: index + archives, gh-pages friendly |
| Integrity | SHA-256 digest in index + `.prov` sidecar | `.sha256` sidecar, digest-in-index later |
| Signing | PGP via `.prov` | v2+ |
| Pack command | `helm package` | `aipkg pack` |
| Install extracts | Strips top-level dir | Strips top-level dir |
| Type system | Ecosystem-defined (any K8s resource) | Manifest-declared (`type` per artifact) |
| Adapters | Helm internals render templates | Tool-specific adapters place artifacts |

Same architecture. Different domain.

## New Backlog Items

From CLI scaffold session:

- **AIPKG-37:** Homebrew tap repository setup.
- **AIPKG-38:** Scoop bucket repository setup.
- **AIPKG-39:** GitHub issue templates (bug report, feature request).
- **AIPKG-40:** Pull request template.
- **AIPKG-41:** Dependabot configuration.
- **AIPKG-42:** CODEOWNERS file.
- **AIPKG-43:** Security considerations in specification (artifact type constraints, MCP server validation).
- **AIPKG-44:** Package security scanning / audit command (prompt injection detection, unsafe MCP configs, typosquatting). Differentiator; no AI package ecosystem has this yet.
- **AIPKG-45:** GitHub Actions release workflow (goreleaser on tag push).
- **AIPKG-46:** Global config file (`~/.aipkg/config.json`) for user defaults. Deferred from init.
- **AIPKG-47:** Pack command (`aipkg pack`). Depends on AIPKG-4.

From archive format session:

- **AIPKG-48:** `.aipkgignore` support for `aipkg pack` (gitignore-style exclusion). CLI concern, not spec.
- **AIPKG-49:** Merge aipkg-spec into aipkg repo. High priority.
