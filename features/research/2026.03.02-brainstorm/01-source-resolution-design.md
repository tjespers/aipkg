# Source Resolution & Package Index Design

## The Problem

The CLI needs to go from a package name (`@scope/name`) to installable
content. This mapping is the core of the install flow and shapes every
downstream feature (update, search, publish, virtual packages).

The mapping is trivial when one GitHub repo = one package. It breaks down
in three real-world scenarios that the design must handle:

1. **Monorepo catalogs**: one repo hosts many individually installable packages
2. **Unpackaged upstream repos**: content exists but has no aipkg awareness
3. **Mixed scopes**: a single author publishes some packages from dedicated
   repos and others from a shared catalog

Convention-based guessing (scope = GitHub owner, name = repo) cannot solve
cases 2 and 3. A discovery layer is required.

## The Four Package Flavors

Every package in the ecosystem falls into one of four categories based on
two dimensions: repository structure (dedicated vs. monorepo) and aipkg
awareness (packaged vs. unpackaged).

```text
                    Dedicated repo          Monorepo / catalog
                 ┌─────────────────────┬─────────────────────────┐
  Has aipkg.json │  A. Native 1:1      │  B. Native catalog      │
  (packaged)     │  dist strategy      │  dist strategy          │
                 ├─────────────────────┼─────────────────────────┤
  No aipkg.json  │  C. Virtual 1:1     │  D. Virtual catalog     │
  (unpackaged)   │  source strategy    │  source strategy (350K+)│
                 └─────────────────────┴─────────────────────────┘
```

### Flavor A: Native 1:1

The simplest case. One repo, one package, author maintains it.

- **Example**: `tjespers/my-skill` repo contains one `aipkg.json` with
  `name: "@tjespers/my-skill"`.
- **Identity**: `@tjespers/my-skill` (scope = GitHub owner, name = repo name)
- **Versioning**: Author tags releases with semver. GitHub Releases hold
  `.aipkg` assets produced by `aipkg pack`.
- **Index entry type**: `native` with archive URL + SHA-256.
- **Maintenance**: Author publishes, author maintains. Zero work for aipkg project.

### Flavor B: Native catalog

One repo, multiple packages. Author manages them together but publishes
individually.

- **Example**: `tjespers/claude-skills` repo contains `skills/some-cool-skill/aipkg.json`
  and `skills/some-other-skill/aipkg.json`.
- **Identity**: Each sub-package has its own name (`@tjespers/some-cool-skill`,
  `@tjespers/some-other-skill`). The repo name does NOT appear as a package name.
- **Versioning**: Author chooses the strategy. Could version all packages
  together (one release tag, multiple `.aipkg` assets) or independently.
  Release assets encode the package identity via the filename convention
  (`tjespers--some-cool-skill-1.0.0.aipkg`).
- **Index entry type**: `native` with archive URL + SHA-256 per sub-package.
- **Maintenance**: Author publishes and submits entries to the index
  (or `aipkg publish` handles it). The aipkg project just hosts the index.

### Flavor C: Virtual 1:1

Upstream repo has useful content but no aipkg awareness. The aipkg project
(or a community contributor) wraps it.

- **Example**: `some-author/cool-prompt-collection` contains a single prompt
  file worth packaging.
- **Identity**: `@some-author/cool-prompt-collection` in the index.
- **Versioning**: Pinned to a specific commit ref. Version derived from
  ref (see Version Derivation below).
- **Index entry type**: `virtual` with source pointer (repo + path + ref).
  No pre-built archive. CLI resolves from source at install time.
- **Maintenance**: Near-zero. Crawler auto-generates the entry. Ref can
  be bumped automatically or manually.

### Flavor D: Virtual catalog

Upstream repo contains many individually useful artifacts. The aipkg project
slices them into separate packages. This is the dominant case in the
current ecosystem (350K+ skills on skillsmp.com).

- **Example**: `ComposioHQ/awesome-claude-skills` contains 30+ skills in
  subdirectories, each with a SKILL.md.
- **Identity**: Each skill gets its own package name (`@composiohq/code-review`,
  `@composiohq/git-commit`, etc.).
- **Versioning**: Each package tracks its own source ref independently.
- **Index entry type**: `virtual` with source pointer per skill.
- **Maintenance**: Crawler discovers skills within upstream repos and
  generates per-skill entries automatically. Each virtual package is a
  single artifact (one skill, one prompt, etc.) to keep things simple.

## Two Install Strategies: Dist vs. Source

Inspired by Composer's `preferred-install` model, the CLI supports two
install strategies depending on the package type.

```
┌─────────────────────────────────────────────────────────────────┐
│  Native packages (dist strategy)                                │
│                                                                 │
│  index entry → archive URL → HTTP GET → verify SHA-256 → unpack │
│                                                                 │
│  Pre-built .aipkg archive. Fast. Author-maintained.             │
├─────────────────────────────────────────────────────────────────┤
│  Virtual packages (source strategy)                             │
│                                                                 │
│  index entry → repo + path + ref → fetch from GitHub → pack     │
│  locally → install                                              │
│                                                                 │
│  No pre-built archive. CLI builds on-the-fly. Auto-indexed.     │
└─────────────────────────────────────────────────────────────────┘
```

**Why two strategies?**

The dist strategy is optimal for native packages: one HTTP download, hash
verified, done. But it requires someone to build and host the archive.

The source strategy eliminates that requirement entirely. For the 350K+
skills already on GitHub, nobody needs to pack, host, or maintain
anything. The CLI fetches the source directory, validates the SKILL.md,
packs it in-memory, and installs. The "build" step takes milliseconds
(it's just creating a zip). The bottleneck is the single HTTP fetch from
GitHub.

This avoids the @types/DefinitelyTyped maintenance trap. No CI pipeline
builds and hosts 350K archives. No human attention needed to keep them
updated. The index is metadata only (repo, path, ref), and the CLI does
the packaging at install time.

**The Composer parallel:**

| Composer | aipkg | When |
|----------|-------|------|
| `--prefer-dist` (default) | Native package | Author published a .aipkg |
| `--prefer-source` | Virtual package | No .aipkg exists, resolve from GitHub |

Unlike Composer, the user doesn't choose the strategy. The index entry
determines it: if it has a `url`, use dist; if it has a `source`, use
source resolution. Transparent to the user.

## Architecture: Per-Package Index on Static Hosting

### Why Not a Single index.json?

At 350K packages, a single index file would be 50-100MB. Unacceptable
for a CLI that needs to resolve one package at a time.

### Per-Package Metadata Files

The central index is a directory of static JSON files, one per package,
served via GitHub Pages:

```
packages.aipkg.dev/
  @composiohq/
    code-review.json
    git-commit.json
    pull-request-review.json
  @tjespers/
    my-skill.json
  ...
```

Resolution is a single HTTP GET:

```
GET https://packages.aipkg.dev/@composiohq/code-review.json
```

The index serves two distinct JSON schemas, discriminated by a `type`
field (Decision 52).

**Package entry** (`type: "package"`): published archives with versioned
dist blocks.

```json
{
  "name": "@tjespers/golang-expert",
  "description": "Expert Go developer skill",
  "type": "package",
  "versions": {
    "1.0.0": {
      "dist": {
        "url": "https://github.com/tjespers/golang-expert/releases/download/v1.0.0/tjespers--golang-expert-1.0.0.aipkg",
        "sha256": "a1b2c3d4..."
      }
    },
    "1.1.0": {
      "dist": {
        "url": "https://github.com/tjespers/golang-expert/releases/download/v1.1.0/tjespers--golang-expert-1.1.0.aipkg",
        "sha256": "e5f6a7b8..."
      }
    }
  }
}
```

**Recipe entry** (`type: "recipe"`): virtual packages. Source pointer,
artifact type, no versions block (Decision 53).

```json
{
  "name": "@composiohq/code-review",
  "description": "Code review skill from Composio's awesome-claude-skills",
  "type": "recipe",
  "artifact": "skill",
  "source": {
    "repo": "ComposioHQ/awesome-claude-skills",
    "path": "skills/code-review/"
  }
}
```

**Package entry fields:**

- `name`: Full scoped package name.
- `description`: Short summary. Used for search/display.
- `type`: `"package"`. Tells the CLI to use the dist install strategy.
- `versions`: Map of version strings to install metadata. Each version
  has a `dist` block with `url` + `sha256`. The `dist` wrapper is kept
  for extensibility (Decision 55); future methods like `oci` could be
  added as siblings.

**Recipe entry fields:**

- `name`: Full scoped package name.
- `description`: Short summary.
- `type`: `"recipe"`. Tells the CLI to use the source install strategy.
- `artifact`: The artifact type being packaged (`"skill"`, `"prompt"`,
  `"command"`, etc.). Tells the CLI what it's looking at so it can
  generate a proper `aipkg.json` manifest during local packing.
- `source`: Where to fetch the content.
  - `repo`: GitHub owner/repo.
  - `path`: Path within the repo to the artifact directory or file.

Recipe entries carry no version information. The CLI resolves the ref at
install time (HEAD of default branch, or a specific ref if the consumer
has pinned one in their `require`). Version tracking for virtual packages
lives on the consumer side (Decision 53).

### Resolution Algorithm

```
RESOLVE(package_name, requested_version):

  1. SOURCES = []

     // Project-level overrides first (highest priority)
     if project aipkg.json has "repositories":
       for each repo in repositories (in declared order):
         SOURCES.append(repo)

     // Central index last (implicit default)
     SOURCES.append(CENTRAL_INDEX)

  2. for each source in SOURCES:

       // Fetch per-package metadata
       entry = HTTP_GET(source.base_url + "/" + package_name + ".json")

       if entry is 404:
         continue  // not in this source

       if entry.type == "package":
         // Versioned package: resolve from versions block
         if requested_version is specified:
           if entry.versions contains requested_version:
             return entry
           else:
             error: "version {requested_version} not found"
         else:
           latest = highest_semver(entry.versions.keys())
           return entry (with resolved version = latest)

       if entry.type == "recipe":
         // Unversioned recipe: return the entry as-is
         // CLI will resolve the ref from GitHub at install time
         return entry

  3. error: "package {package_name} not found in any configured source"
```

**Key properties:**

- Project repos are checked first, in declared order. Allows overrides
  and private packages to shadow the central index.
- The central index is always the last source (implicit, cannot be removed).
- First match wins. If a package appears in multiple sources, the
  highest-priority source wins entirely (no version merging across sources).
- For package entries: exact version match if specified, highest semver
  if not. No ranges, no constraints. (Constitution Principle I.)
- For recipe entries: no version resolution in the index. The CLI resolves
  the ref from GitHub at install time. If the consumer has a pinned ref
  in their `require`, the CLI uses that.
- Each resolution is one HTTP GET per source tried (usually just one:
  the central index).

### The `repositories` Field in Project Manifest

From the existing unified schema (`spec/schema/aipkg.json`), the project
manifest supports a `repositories` array. Refined from this session:

```json
{
  "type": "project",
  "repositories": [
    {
      "url": "https://my-company.github.io/aipkg-packages/",
      "scope": "@mycompany"
    }
  ]
}
```

- `url`: Base URL of the index. Per-package files are at
  `{url}/@scope/name.json`.
- `scope` (optional): When present, this source is only consulted for
  packages matching that scope. When absent, consulted for all packages.
- The `type` field from the original schema is dropped. All sources serve
  the same per-package file format over HTTP.
- The `canonical` field from the original schema is dropped for v1.
  Priority ordering is sufficient.

## Install Flows (End-to-End)

### Package Install (type: "package", dist strategy)

```
aipkg install @tjespers/golang-expert

  1. RESOLVE
     - Fetch https://packages.aipkg.dev/@tjespers/golang-expert.json
     - Entry has type: "package"
     - Select latest version from versions block (e.g., "1.1.0")
     - Extract dist: { url, sha256 }

  2. FETCH
     - HTTP GET the .aipkg archive from dist.url
     - Verify SHA-256 matches dist.sha256
     - Store archive in local cache

  3. UNPACK
     - Extract archive into .aipkg/packages/@tjespers/golang-expert/
     - Strip the top-level directory per archive spec (FR-005)
     - Extract bundled deps from deps/ if present

  4. RECORD
     - Update project aipkg.json require:
       "@tjespers/golang-expert": "1.1.0"

  5. ADAPT (future, AIPKG-11)
     - Run tool-specific adapters
```

### Recipe Install (type: "recipe", source strategy)

```
aipkg install @composiohq/code-review

  1. RESOLVE
     - Fetch https://packages.aipkg.dev/@composiohq/code-review.json
     - Entry has type: "recipe", artifact: "skill"
     - No versions block. CLI will resolve ref from GitHub.

  2. FETCH SOURCE
     - Resolve HEAD ref of the default branch from GitHub
       (or use pinned ref from consumer's require if reinstalling)
     - Download repo archive at that ref
     - Extract only the specified path (skills/code-review/)

  3. PACK LOCALLY
     - Use artifact type from recipe to guide packing
     - For skills: validate SKILL.md (frontmatter, required fields)
     - Generate aipkg.json (name from recipe, version from ref,
       single artifact entry of the declared type)
     - Pack into .aipkg archive (in-memory or temp)
     - Cache the resulting archive locally

  4. UNPACK
     - Extract into .aipkg/packages/@composiohq/code-review/
     - Same as package install from this point on

  5. RECORD
     - Update project aipkg.json require:
       "@composiohq/code-review": "0.1.0+a3f8bc1"
       (version derived from ref: 0.1.0+{short-sha})

  6. ADAPT (future, AIPKG-11)
     - Run tool-specific adapters (same as package install)
```

### Bulk Install (no arguments)

```
aipkg install

  1. Read require from project aipkg.json
  2. For each entry, run RESOLVE > FETCH > UNPACK
     (dist or source depending on entry type)
  3. Each package's bundled deps are extracted alongside it
     (no transitive resolution needed — deps are in the archive)
```


## Dependency Model: Helm-Style Bundled Dependencies

### Why not npm/Composer?

In npm, when express depends on body-parser, you might use body-parser
directly. Dependencies are part of your working set.

In aipkg, when @tjespers/golang-expert depends on @someone/golang-test-writer,
the user installed golang-expert because they want a golang expert. The
test-writer is an implementation detail. Having it appear as a surprise
skill in their IDE is confusing at best.

AI skills aren't software libraries. Dependencies are internal
implementation details, not public APIs.

### The Helm model applied to aipkg

Dependencies are resolved at **pack time** and bundled inside the parent
archive. Install extracts everything from one archive. No transitive
dependency resolution at install time.

**Pack time (author side):**

```
aipkg pack

  1. Read aipkg.json, find require section:
     "@someone/golang-test-writer": "1.0.0"

  2. Resolve each dependency from the index
     (dist or source strategy, same as install)

  3. Include resolved .aipkg archives in deps/ directory
     inside the parent archive

  4. Result: self-contained archive
     tjespers--golang-expert-1.0.0.aipkg
       golang-expert/
         aipkg.json
         skills/golang-expert/SKILL.md
         deps/
           someone--golang-test-writer-1.0.0.aipkg
```

**Install time (consumer side):**

```
aipkg install @tjespers/golang-expert

  1. Resolve from index (one HTTP GET for metadata)
  2. Download .aipkg archive (one HTTP GET)
  3. Verify SHA-256
  4. Extract: main package + all bundled deps
  5. Done. No further network calls.
```

### Why this works for aipkg specifically

- **Offline installs**: Archive is self-contained. Corporate/air-gapped
  environments work naturally.
- **Simpler install**: Zero dependency resolution at install time.
  One download, one extract.
- **Author controls visibility**: Bundled deps can be internal (not
  exposed to the user) or public (explicitly exposed). Default: internal.
- **Slower pack > slower install**: Pack happens once (by the author).
  Install happens many times (by every consumer). Put the complexity
  where it runs less.
- **Virtual deps at pack time**: The author resolves virtual dependencies
  once. Consumers never source-resolve transitive deps.
- **Archive size is negligible**: AI artifacts are text files. A skill
  with 3 bundled deps might be 50KB compressed. Helm bundles entire
  application configs at MB+ scale and nobody cares.

### Dependency visibility

Author declares visibility per dependency in the manifest. Default is
`internal`: the adapter does not create a user-visible skill/prompt/etc.
for bundled deps. Author can mark a dependency as `public` to expose it.

```json
{
  "require": {
    "@someone/golang-test-writer": {
      "version": "1.0.0",
      "visibility": "internal"
    }
  }
}
```

- `internal` (default): dependency is installed but not exposed by
  adapters. The parent skill can still delegate to it internally.
- `public`: dependency is exposed as if the user installed it directly.
  Useful for "starter pack" packages that bundle skills for the user.

Exact adapter behavior is deferred to AIPKG-11.

### Dependency update workflow

Like Helm's `helm dependency update`:

- `aipkg deps update` (or similar) refreshes the deps/ directory by
  re-resolving from the index.
- `aipkg pack` can optionally resolve deps implicitly if deps/ is missing.
- Authors version-pin their deps in the manifest. Updates are deliberate.

## The ai-interop/packages Repository

### Structure

```
ai-interop/packages/
  packages/                  # per-package metadata (GitHub Pages root)
    @composiohq/
      code-review.json
      git-commit.json
    @tjespers/
      my-skill.json
  crawler/                   # automated indexing pipeline
    config.yaml              # which repos/orgs to crawl
    src/                     # crawler implementation
  .github/
    workflows/
      crawl.yml              # scheduled: re-index upstream repos
      publish.yml            # on PR merge: update native entries
```

### Automated Crawler (optional, not on critical path)

**Important (Decision 54):** The crawler is an optional automation layer,
not core architecture. Virtual packages work with manually-written recipe
entries in the index. The crawler automates recipe creation at scale but
is not needed to ship virtual package support.

When built, the crawler would run on a schedule and:

1. Scan configured GitHub repos/orgs for SKILL.md files
   (potentially consuming skillsmp.com data as a seed)
2. Parse SKILL.md frontmatter (name, description)
3. Derive package identity: scope from GitHub owner (lowercased),
   name from SKILL.md `name` field or directory name
4. Generate `type: "recipe"` entries with appropriate `artifact` and
   `source` fields
5. Commit and push to trigger GitHub Pages deployment

The crawler produces recipe entries only. No pre-building, no archive
hosting.

### Native Package Registration

Authors of proper packages submit a PR adding their metadata file:

```json
// packages/@tjespers/my-skill.json
{
  "name": "@tjespers/my-skill",
  "description": "My cool skill",
  "type": "package",
  "versions": {
    "1.0.0": {
      "dist": {
        "url": "https://github.com/tjespers/my-skill/releases/download/v1.0.0/tjespers--my-skill-1.0.0.aipkg",
        "sha256": "abc123..."
      }
    }
  }
}
```

Or `aipkg publish` automates this (creates the release, generates the
metadata, opens a PR to the central index).

### Graduation: Virtual to Native

When an upstream author starts publishing proper packages:

1. Author claims their scope (see Namespace Governance)
2. The metadata file is updated: `source` changes from `"virtual"` to
   `"native"`, version entries switch from `source` to `dist`
3. The package name stays the same. No breaking change for consumers.
4. Existing virtual versions remain available for pinned installs.

## Version Derivation for Virtual Packages

Since upstream repos rarely use semver, virtual package versions need
a derivation strategy.

### Rules

1. **Upstream has semver tags**: Use the tag directly. Map to the commit
   ref for the source pointer.
2. **Upstream has releases (non-semver)**: Crawler assigns a semver
   based on the release sequence (e.g., first indexed = 0.1.0, subsequent
   updates bump patch). Ref pins to the release commit.
3. **Upstream has no releases (most common)**: Use `0.1.0+{short-sha}`.
   The `+{sha}` is semver build metadata: ignored in version comparison
   (per semver spec) but provides traceability to the source commit.

### Version Updates

When the crawler re-indexes and detects upstream changes:

- If content at the tracked path has changed, a new version entry is
  created with the new ref.
- Old versions remain in the metadata (immutable). Consumers pinned to
  an old version continue to get the same content.
- "Latest" resolves to the newest entry.

The crawler can be conservative: only create new versions when the
SKILL.md or its supporting files actually changed (content hash
comparison), not on every upstream commit.

## Namespace & Scope Governance

### Virtual Packages Use Upstream's Scope

Virtual packages are published under the upstream's scope:
`@composiohq/code-review`, not `@virtual/composiohq:code-review`.

**Rationale:**

- Users want `@composiohq/code-review`. The virtual/native distinction
  is an implementation detail.
- Graduation from virtual to native is seamless (package name unchanged).
- The `@virtual/` prefix added complexity to naming, CLI parsing, and
  the mental model without user-facing value.

**Trade-off:** The aipkg project publishes under scopes it doesn't own.
Mitigated by:

- `maintainer` field provides transparency.
- `source: "virtual"` marking in metadata.
- Scope claiming process for upstream authors.

### Scope Claiming Process

1. Author opens an issue on `ai-interop/packages`: "I want to claim @composiohq"
2. Author proves ownership (GitHub org admin, or adds `.aipkg-verify`
   file to their org's `.github` repo)
3. Maintainer transfers entry ownership: `maintainer` changes, `source`
   changes to `"native"`, URLs updated
4. Old virtual versions remain available (no broken installs)

## Caching Strategy

### Per-Package Metadata Cache

- Cache metadata files locally: `~/.aipkg/cache/metadata/@scope/name.json`
- TTL: 1 hour by default (configurable later, AIPKG-46)
- `aipkg install --refresh` forces fresh fetch
- `aipkg update` always refreshes before checking

### Archive Cache

For both native (downloaded) and virtual (locally built) packages:

- Store in `~/.aipkg/cache/archives/`
- Keyed by `{scope}--{name}-{version}.aipkg`
- If the same version is installed in multiple projects, the archive is
  fetched/built once
- Virtual packages: the locally-packed archive is cached so subsequent
  installs of the same version don't re-fetch from GitHub

### Cache Eviction

Not in scope for v1. Manual cleanup via `rm -rf ~/.aipkg/cache/` is fine.

## Open Questions

### Q1: GitHub API for source-resolved installs

The source strategy requires fetching content from GitHub. Options:

- **GitHub zipball/tarball endpoint**: Downloads entire repo at a ref.
  Simple but wasteful for mono-repos with one small skill.
- **GitHub Contents API**: Fetches individual files/directories. More
  precise but multiple API calls for a skill directory with subfiles.
- **git sparse checkout**: Fetches only the needed path. Efficient but
  requires git on the user's machine.

Recommendation: Start with the tarball endpoint (one HTTP GET, extract
the needed path locally). Optimize later if mono-repos with huge
non-skill content cause performance issues.

### Q2: GitHub API rate limits for source installs

Unauthenticated: 60 requests/hr. Each source-resolved install is 1 request
(the tarball download). This limits unauthenticated users to ~60 installs/hr.

Authenticated (with a GitHub token): 5,000 requests/hr. Plenty.

The CLI should support optional GitHub token configuration (via env var
or global config) and show a clear message when rate-limited.

### Q3: Direct URL installs

Should the CLI support `aipkg install https://example.com/package.aipkg`?
Bypasses the index, downloads a known URL. Useful for testing, private
packages, and edge cases. Low effort to implement. Probably worth
including as a convenience.

### Q4: Local archive installs

`aipkg install ./my-package-1.0.0.aipkg` for local file installs.
Useful for testing `aipkg pack` output before publishing. Trivial to
implement (skip the fetch, go straight to unpack).

### Q5: Search and discovery

Per-package metadata files don't support browsing or searching. Search
needs a separate mechanism:

- A search index file (package names + descriptions, much smaller than
  full metadata)
- A search API endpoint (Cloudflare Worker, Vercel Edge Function)
- Integration with skillsmp.com for discovery
- `aipkg search` as a CLI command

This is a separate feature, not blocking install.

### Q6: Consuming skillsmp.com data

The skillsmp.com marketplace already indexes 350K+ skills. The aipkg
crawler could:

- Build its own index from scratch (full control, independent)
- Consume skillsmp.com data as a seed (faster bootstrap, dependency)
- Partner formally (aipkg = install layer, agentskills = discovery layer)

Decision deferred. The crawler architecture supports any of these.

### Q7: Non-skill artifact types in virtual packages

The current ecosystem is heavily skill-focused (SKILL.md standard). But
aipkg supports six artifact types. How does the crawler discover
non-skill artifacts (prompts, commands, agents, etc.)?

For skills: SKILL.md is a clear marker.
For others: less clear. Prompt files don't have a standard marker.
The crawler may need heuristics or explicit configuration per upstream repo.

For v1: focus on skills (where the volume is). Add other types as the
ecosystem signals demand.
