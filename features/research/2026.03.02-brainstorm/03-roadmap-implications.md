# Roadmap Implications

## Revised Feature Sequence

The source-resolved index model changes the dependency graph and
sequencing of existing backlog items significantly.

### Before (assumed in original backlog)

```
AIPKG-1 (manifest) ─> AIPKG-5 (source interface) ─> AIPKG-9 (GitHub source)
                                                  ─> AIPKG-13 (HTTP source)
                                                  ─> AIPKG-10 (install)
```

### After (per-package index + source resolution + Helm deps)

```
                     ┌─ AIPKG-NEW-A (index format spec)
                     │
AIPKG-47 (pack) ────┤
  [DONE, needs       ├─ AIPKG-NEW-B (source resolver + install)
   dep bundling      │    covers: simplified AIPKG-5, AIPKG-10
   expansion]        │    two strategies: dist (native) + source (virtual)
                     │    NO transitive dep resolution (Helm model)
                     │
                     ├─ AIPKG-47b (pack: dependency bundling)
                     │    resolve deps at pack time, bundle in deps/
                     │    author-controlled visibility (internal/public)
                     │
                     └─ AIPKG-NEW-C (ai-interop/packages repo + crawler)
                          central index, automated crawling, bootstrap

                     ┌─ AIPKG-11 (adapter execution)
AIPKG-NEW-B ────────┤    includes: dep visibility (internal vs public)
                     └─ AIPKG-12 (remove & list)

AIPKG-NEW-C ────────── AIPKG-16 (publish, re-scoped)
```

**Key change from Helm dependency model (decision 44):** The pack command
gains significant new scope. It now resolves dependencies from the index
and bundles them as nested `.aipkg` archives in `deps/`. This makes the
install side simpler (no transitive resolution). The adapter layer
(AIPKG-11) gains responsibility for dependency visibility (honoring the
`internal`/`public` setting per bundled dep).


## Re-scoped Existing Issues

### AIPKG-5: Source Type Interface

**Original scope**: Define contracts for GitHub, HTTP, and future source
types with different resolution behaviors.

**New scope**: Define the per-package metadata file schema and the
resolver interface. A "source" is any HTTP endpoint serving per-package
JSON files in the defined format. The interface is trivially simple:
one GET per package resolution.

Might merge entirely into the index format spec (AIPKG-NEW-A).

### AIPKG-9: GitHub Source Type & Resolver

**Original scope**: Implement GitHub API integration for version discovery
and asset download on the install side.

**New scope**: The GitHub API is no longer used for install-time resolution.
GitHub-specific logic moves to two places:
1. The `aipkg publish` command (AIPKG-16): creating releases, uploading assets
2. The virtual package source strategy in the install command: fetching
   repo tarballs at a pinned ref

Re-scope into AIPKG-16 and the source strategy implementation within
AIPKG-NEW-B. This issue can likely be closed or absorbed.

### AIPKG-10: Install Command

**Original scope**: Download, unpack, store, resolve transitive deps.

**New scope**: Two install strategies, but simpler than the original scope:
- Dist: download pre-built archive from URL, verify SHA-256, unpack
- Source: fetch from GitHub at pinned ref, pack locally, then unpack

With Helm-style bundled dependencies (decision 44), the install command
does NOT resolve transitive dependencies. Dependencies are bundled inside
the archive by the author at pack time. Install just extracts what's in
the archive, including bundled deps from `deps/`.

Depends on the index format spec (AIPKG-NEW-A), not on separate source
type implementations.

### AIPKG-13: HTTP Source Type

**Original scope**: Implement HTTP source with URL templates and auth.

**New scope**: Likely unnecessary as a separate item. The index model IS
HTTP-based (per-package metadata over HTTP). Direct URL installs
(`aipkg install https://example.com/package.aipkg`) could be a small
addition to the install command, not a separate source type.

### AIPKG-15: Virtual Package Support

**Original scope**: CLI resolves @virtual/ namespace, fetches recipes,
packages on-the-fly at install time.

**New scope**: Significantly changed. The `@virtual/` namespace is dropped
(decision 34). Virtual packages use the upstream's scope. The "fetch and
package on-the-fly" concept survives as the source install strategy, but
it's driven by the index metadata, not by a separate @virtual resolution
path. The CLI doesn't distinguish virtual from native at a code level;
it just checks whether the version entry has `dist` or `source`.

The virtual package *pipeline* (crawler, index generation) lives in the
`ai-interop/packages` infrastructure (AIPKG-NEW-C), not in the CLI.

This issue can be re-scoped to cover only the source install strategy
within the CLI, or absorbed into AIPKG-NEW-B.

### AIPKG-16: Publish Command

**Original scope**: Upload .aipkg to GitHub Releases.

**New scope**: Upload .aipkg to author's hosting AND register in the
central index. Two paths:
1. For GitHub-hosted packages: create release, upload asset, submit
   metadata entry to `ai-interop/packages` (via PR or API)
2. For other hosting: user provides the archive URL, command generates
   and submits the metadata entry

Absorbs GitHub-specific logic from AIPKG-9.

### AIPKG-46: Global Config

**Unchanged**: Still the place for default repository URLs, GitHub auth
tokens (needed for source-resolved installs to avoid rate limits), cache
settings. Lower priority since the central index provides zero-config
defaults.

### AIPKG-51: Local Package Linking

**Unchanged**: Still the right approach for local development workflows.
Orthogonal to the source resolution model.


## Suggested New Issues

### AIPKG-NEW-A: Package Index Format Specification

**Project**: Specification
**Priority**: High (blocks install)
**Scope**:

- Define the per-package metadata file schema (native and virtual entries)
- Define version entry formats (dist with URL+SHA-256, source with
  repo+path+ref)
- Define resolution algorithm (priority ordering, exact match vs. latest)
- Define caching expectations (TTL, refresh behavior)
- Define how private/enterprise indexes work (same format, different base URL)
- Update the `repositories` field in the project manifest schema

**Depends on**: Nothing (can start immediately)

### AIPKG-NEW-B: Source Resolver & Install Command

**Project**: CLI
**Priority**: High
**Scope**:

- Implement per-package metadata fetching with local caching
- Implement resolution algorithm (project repos > central index)
- Implement dist strategy (HTTP download, SHA-256 verify, unpack)
- Implement source strategy (GitHub tarball fetch, path extraction,
  local pack, unpack)
- Install to `.aipkg/packages/` directory
- Update project `aipkg.json` require section
- Extract bundled deps from `deps/` inside archives (no transitive
  resolution; Helm model means deps are pre-bundled by the author)
- `aipkg install` (no args) for bulk install from require
- Optional: direct URL and local file installs as convenience

**Depends on**: AIPKG-NEW-A (index format spec)

### AIPKG-47b: Pack Command Expansion (Dependency Bundling)

**Project**: CLI
**Priority**: High
**Scope**:

- Resolve dependencies from `require` section against the index
- Download/source-resolve each dependency
- Bundle resolved `.aipkg` archives in `deps/` directory inside the
  parent archive
- Support `aipkg deps update` (or similar) to refresh bundled deps
- Respect dependency visibility settings (`internal`/`public`)
- Validate no circular dependencies

**Depends on**: AIPKG-NEW-A (index format spec), AIPKG-47 (existing pack)

### AIPKG-NEW-C: Central Package Index & Crawler

**Project**: Infrastructure
**Priority**: High (blocks ecosystem launch)
**Scope**:

- Set up `ai-interop/packages` repo with GitHub Pages
- Define per-package file directory structure
- Build automated crawler for SKILL.md discovery on GitHub
- Bootstrap initial catalog (crawl major skill repos)
- Define process for native package registration (PR-based)
- Define scope claiming process
- Optional: consume skillsmp.com data as seed

**Depends on**: AIPKG-NEW-A (index format spec)


## Critical Path

The shortest path to "a user can install a package" is:

```
1. Spec the index format                    (AIPKG-NEW-A)
   │
   ├── 2a. Implement resolver + install     (AIPKG-NEW-B)
   │        in the CLI
   │
   └── 2b. Bootstrap the central index      (AIPKG-NEW-C)
            with initial packages
   │
   └── 3. Ship: users can
           aipkg install @scope/name
```

Steps 2a and 2b can run in parallel. The CLI needs a populated index to
be useful, and the index needs the CLI to be consumable, but development
can proceed simultaneously.

Adapters (AIPKG-11) come next: turning installed packages into active
tool integrations. Without adapters, packages are files in `.aipkg/`
but not wired into Claude Code, Cursor, etc. That's the next milestone
after install works end-to-end.


## The Big Picture

The combination of source-resolved virtual packages and an automated
crawler means the aipkg ecosystem launches with access to the entire
existing skill catalog (350K+ on skillsmp.com) on day one. No author
cooperation required. No manual curation at scale. The cold start
problem is solved by meeting the ecosystem where it already is.

Native packages (author-published, versioned, pre-built archives) provide
the quality tier for serious package authors. Virtual packages (auto-indexed,
source-resolved) provide the breadth tier for everything else. The
graduation path (virtual to native) is seamless: same package name,
same install command, better install performance and versioning.

The Helm-style dependency model (decision 44) means archives are
self-contained. Install is a single download + extract, with zero
transitive resolution. Complexity shifts to pack time (where it runs
once, by the author) rather than install time (where it runs many times,
by every consumer). Dependencies are implementation details, invisible
to the user by default.

This positions aipkg not as "yet another package manager waiting for
content" but as "the install layer for the skill ecosystem that already
exists."
