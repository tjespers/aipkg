# Design Decisions: Source Resolution Session

Continuation of the decision log. Numbers continue from decision 31
(last entry in `2026.02.28-brainstorm/01-design-decisions.md`).

---

### Decision 32: Index-based package discovery

**Context**: The CLI needs to map a package name to a download URL.
Three approaches were considered: (A) GitHub Releases API with
scope-as-owner convention, (B) Helm-style static index served over HTTP,
(C) hybrid of both.

**Decision**: Index-based discovery (option B). The CLI resolves packages
by fetching a static `index.json` from configured sources. No GitHub API
client is required on the install path.

**Rationale**: The scope-as-owner convention breaks for monorepo catalogs
(one repo hosting many packages). An index handles all cases (1:1,
monorepo, virtual) with a single resolution path. The CLI implementation
is simpler (HTTP GET, no API client). Decouples install from GitHub
specifically.

**Rejected**: GitHub API resolution. Rate limits (60/hr unauthenticated),
requires a full API client, only works for the 1:1 case, couples the CLI
to GitHub's API surface.

---

### Decision 33: Central package index hosted by aipkg project

**Context**: With index-based resolution, who hosts the default index?

**Decision**: The aipkg project maintains a central index at
`https://aipkg.github.io/packages/index.json`, served from the
`aipkg/packages` GitHub repo via GitHub Pages. This is the implicit
default source for all CLI installations.

**Rationale**: Homebrew model. The project does the cold-start work so
users get value on day one. No author setup required for packages to be
discoverable. The central index is supplementable via project-level
`repositories` config for private/enterprise packages.

---

### Decision 34: Virtual packages use upstream's scope (no @virtual namespace)

**Context**: Earlier brainstorms proposed `@virtual/owner:repo` for
community-maintained packages wrapping upstream repos. This session
reconsidered.

**Decision**: Virtual packages are published under the upstream's scope
(e.g., `@composiohq/code-review`, not `@virtual/composiohq:code-review`).
The index entry's `source` field distinguishes native from virtual.

**Rationale**: Users want `@composiohq/code-review`, not a namespace they
need to learn about. When the upstream author graduates to native publishing,
the package name stays the same (no breaking change for consumers). The
`@virtual` prefix adds complexity to naming, CLI parsing, and the mental
model without providing user-facing value.

**Trade-off acknowledged**: The aipkg project publishes under scopes it
doesn't own. Mitigated by: `maintainer` field for transparency, a scope
claiming process for upstream authors, and `source: "virtual"` marking.

---

### Decision 35: Source-resolved virtual packages (no pre-building)

**Context**: Virtual packages need to go from upstream repo content to
installed package. Three models considered:
(A) Pre-build archives via CI pipeline and host them (recipe + CI model).
(B) Resolve from source at install time; CLI fetches and packs on-the-fly.
(C) Hybrid: pre-build popular ones, source-resolve the long tail.

**Decision**: Source-resolved (option B). Virtual package index entries
contain a source pointer (repo, path, ref) instead of an archive URL.
The CLI fetches the source directory from GitHub and packs locally at
install time. No pre-built archives, no CI pipeline, no hosted artifacts
for virtual packages.

**Rationale**: The ecosystem has 350K+ skills on GitHub (per skillsmp.com).
Pre-building and hosting that many archives is the DefinitelyTyped trap:
massive CI cost, human attention to maintain, and a scaling wall.
Source-resolved installs eliminate all of that. The index is metadata
only (repo + path + ref). The "build" is a local zip operation that
takes milliseconds. The bottleneck is one HTTP fetch from GitHub per
install, which is acceptable.

**Trade-off**: First install of a virtual package is slightly slower
(fetch from GitHub vs. download a pre-built archive). Mitigated by
local caching; subsequent installs of the same version are instant.

**Rejected**: Pre-build model (option A). Would require CI to build and
host 350K+ archives, human curation of recipes, and ongoing maintenance.
The @types/DefinitelyTyped cautionary tale applies directly.

---

### Decision 36: SHA-256 integrity from the index (not sidecar)

**Context**: The archive spec (002) defines `.sha256` sidecar files. The
install flow could verify integrity using either the sidecar or a hash
embedded in the index.

**Decision**: The index carries SHA-256 hashes per version. The install
flow verifies against the index hash. The sidecar from `aipkg pack`
remains useful for manual verification (e.g., `sha256sum -c`) but is not
used by the install path.

**Rationale**: One fewer HTTP request per install. The index is already
fetched for discovery; it carries the hash for free. Keeps the install
flow simple (fetch index, fetch archive, verify, done).

---

### Decision 37: Resolution priority order

**Context**: A package might appear in multiple sources (project repos,
central index).

**Decision**: Resolution checks sources in this order:
1. Project-level `repositories` (in declared order)
2. Central index (implicit, always last)

First match wins entirely (no version merging across sources).

**Rationale**: Matches Composer's behavior. Project config can shadow the
central index for overrides and private packages. Simple, predictable,
no cross-source conflicts.

---

### Decision 38: Version resolution without version ranges

**Context**: When `aipkg install @scope/name` omits a version, the CLI
must pick one.

**Decision**: Resolve to the highest semver version available in the
first matching source. No version ranges, no constraints, no resolution
strategy beyond "latest" or "exact match."

**Rationale**: Constitution Principle I (Simplicity). Exact pins in
`require`, latest-wins for interactive installs. Version ranges are
deferred to a future version.

---

### Decision 39: Source type interface simplification

**Context**: AIPKG-5 originally envisioned multiple source types (GitHub,
HTTP, S3, etc.) with different resolution behaviors.

**Decision**: For v1, a source is anything that serves per-package
metadata files over HTTP. The source type interface is a single type.
The `repositories` schema field `type` is dropped (all sources use the
same format).

**Rationale**: The per-package metadata abstraction makes the source type
interface trivial. All the complexity (GitHub API, asset discovery, version
derivation) moves to the *publish* side and the *crawler*, not the install
side. The CLI's install path is pure HTTP (metadata fetch) plus one
GitHub fetch for source-resolved packages.

**Implication**: AIPKG-9 (GitHub Source Type & Resolver) is re-scoped.
The GitHub-specific logic becomes part of `aipkg publish` (AIPKG-16),
not part of the resolver. AIPKG-5 (Source Type Interface) becomes a
simpler spec.

---

### Decision 40: Virtual package version derivation

**Context**: Upstream repos without semver tags need version strings.

**Decision**: Virtual package versions are explicitly assigned by the
recipe maintainer. Guidance:
- If upstream has semver tags, use the tag.
- If upstream has no tags, use `0.1.0+{short-sha}` (semver build
  metadata for traceability).
- The maintainer bumps the semver portion when they judge the change
  is meaningful.

**Rationale**: Automatic version derivation is fragile and unpredictable.
Explicit assignment keeps the maintainer in control and versions meaningful.
The `+sha` suffix provides traceability without affecting version ordering
(semver ignores build metadata in comparisons).

---

### Decision 41: Per-package metadata files (not a single index)

**Context**: At 350K+ packages, a single `index.json` would be 50-100MB.
Too large for a CLI that resolves one package at a time.

**Decision**: The central index is a directory of per-package JSON files
served as static files via GitHub Pages. Each package has its own file
at `/@scope/name.json`. Resolution is a single HTTP GET per package.

**Rationale**: Scales to any catalog size. Each resolution is one
sub-kilobyte HTTP GET. No full index download. Works on GitHub Pages
(static hosting, zero infrastructure). Search/discovery is a separate
concern handled by a different mechanism.

**Rejected**: Single `index.json`. Works for <1K packages but doesn't
scale. Also rejected: API server. Adds infrastructure complexity that
static files avoid.

---

### Decision 42: Dist vs. source install strategies (Composer model)

**Context**: Native packages have pre-built archives. Virtual packages
have source pointers. The CLI needs to handle both.

**Decision**: Two install strategies, determined by the index entry:
- **Dist** (native): Download pre-built `.aipkg`, verify SHA-256, unpack.
- **Source** (virtual): Fetch directory from GitHub, validate, pack
  locally, install.

The user doesn't choose. The metadata entry determines the strategy
(presence of `dist` vs `source` in the version entry). If both are
present, prefer dist.

**Rationale**: Mirrors Composer's `--prefer-dist` / `--prefer-source`.
Familiar pattern. Keeps native packages fast (one HTTP download) while
enabling the entire virtual ecosystem without pre-building.

---

### Decision 43: Automated crawler for virtual package indexing

**Context**: With 350K+ skills on GitHub, manual recipe curation doesn't
scale. The index needs to be generated automatically.

**Decision**: An automated crawler scans GitHub for SKILL.md files,
extracts metadata from frontmatter, and generates per-package metadata
files for the central index. Runs on a schedule (weekly or daily).
Produces metadata only (no archives, no builds).

**Rationale**: Eliminates human curation as a bottleneck. The crawler
is lightweight (reads metadata, writes small JSON files). No CI build
pipeline, no archive hosting, no release management. Can potentially
consume skillsmp.com data as a seed for faster bootstrap.

**Scope for v1**: Focus on skills (SKILL.md is a clear marker). Other
artifact types (prompts, commands, agents) lack standard markers and
can be added later as conventions emerge.

---

### Decision 44: Helm-style bundled dependencies (not npm/Composer resolve-at-install)

**Context**: When package A depends on package B, two models exist:
(A) npm/Composer: CLI resolves and fetches B separately at install time.
(B) Helm: B is bundled inside A's archive at pack time. Archive is
self-contained.

**Decision**: Helm model. Dependencies are resolved at pack time and
bundled inside the parent archive in a `deps/` directory as nested
`.aipkg` files. Install extracts everything from one archive. No
transitive dependency resolution at install time.

**Rationale**: AI skills aren't software libraries. Users install a skill
to use it, not to build on top of it. Dependencies are implementation
details (e.g., a golang-expert skill delegates to a test-writer skill).
They shouldn't appear as surprise items in the user's IDE.

The Helm model delivers:
- Simpler install (one download, one extract, zero resolution)
- Offline-capable installs (archive is self-contained)
- Author controls what the consumer sees
- Pack-time resolution happens once (by the author), not on every install
- Virtual dependencies are resolved at pack time (author has network),
  so consumers never need to source-resolve transitive deps

**Trade-off**: Pack is more complex (needs resolver + network). Archives
are slightly larger (bundled deps in compressed form). Dep updates
require re-packing the parent. All acceptable for text-based AI artifacts.

**Rejected**: npm/Composer model. Every consumer resolves the full dep
tree at install time. Slower installs, requires network for transitive
deps, all dependencies become visible to the user (confusing UX for
skills that delegate to other skills).

---

### Decision 45: Virtual packages are single-artifact only

**Context**: Virtual packages are auto-indexed from upstream repos.
Should a single virtual package contain multiple artifacts?

**Decision**: No. One virtual package = one artifact (one skill, one
prompt, one command, etc.). Multi-artifact packages are a native-only
concept (author explicitly structures them).

**Rationale**: Keeps the source-resolve step trivial (fetch one directory
or file, validate, done). Avoids composition complexity in the crawler.
Maps directly to how skills exist in the wild (each skill is a directory).

---

### Decision 46: Scope does not imply repository location

**Context**: Should `@tjespers` automatically check a repository hosted
by the `tjespers` GitHub user?

**Decision**: No. Scopes are identifiers, not repository pointers.
Repositories are explicitly configured in the project manifest or
implicitly the central index. No magic derivation from scope names.

**Rationale**: If `@scope` implied a repository, every scope owner would
need to host one. That's cold-start friction. The central index exists
precisely so that nobody needs to host anything. Matches Composer's model
(Packagist is the default, explicit repos are configured, vendor name
does not imply a registry URL).

---

### Decision 47: GitHub Releases is a hosting recommendation, not architectural requirement

**Context**: The original backlog assumed GitHub Releases as the primary
package source. With the index model, where do native package archives
live?

**Decision**: The index entry's `dist.url` can point anywhere: GitHub
Releases, S3, a CDN, the author's own server. GitHub Releases is a
recommended hosting option (free, easy) but not architecturally special.
The `aipkg publish` command may create a GitHub Release as a convenience,
but the install side never talks to the GitHub API.

**Rationale**: Decouples the install path from any specific hosting
provider. The CLI only needs HTTP GET for dist downloads. Keeps the
door open for alternative hosting without code changes.

---

### Decision 48: Central index hosted at aipkg.dev (not github.io)

**Context**: The `aipkg` GitHub handle is not owned by the project. The
`ai-interop` org is available. The domains `aipkg.dev` and `aipkg.io`
are owned.

**Decision**: The central index URL is `https://packages.aipkg.dev/`
(or similar subdomain), CNAME'd to GitHub Pages or other static hosting
under the `ai-interop` org. Not `aipkg.github.io`.

**Rationale**: Owned domain provides stability and branding. Can be
re-pointed to different hosting without changing the URL baked into
the CLI. GitHub Pages under `ai-interop` is the initial backend.

---

### Decision 49: Dependency visibility controlled by author (default: internal)

**Context**: When a package bundles dependencies (Helm model), should
those dependencies be visible to the user as separate skills in their
IDE?

**Decision**: Author declares visibility per dependency. Default is
`internal` (not exposed by the adapter). The author can mark a
dependency as `public` if they want it user-visible.

**Rationale**: The common case is that dependencies are implementation
details. The golang-expert skill delegates to the test-writer skill
internally; the user shouldn't see `/someone.test-writer` appear in
their IDE. But some packages genuinely want deps exposed (e.g., a
"starter pack" that bundles related skills for the user). Author
control with a safe default handles both.

**Deferred detail**: The exact manifest syntax and adapter behavior
are AIPKG-11 (adapter execution) concerns. The decision here is the
principle: author controls, default internal.

---

### Decision 50: Repositories field design is deferred

**Context**: The existing unified schema has a `repositories` array
with `type`, `url`, `scope`, `canonical` fields. This was an initial
draft; project-type config was deferred.

**Decision**: The `repositories` field in the project manifest needs
fresh design based on the index model. The initial schema draft is not
treated as settled. Key changes from the draft:
- `type` field is dropped (all sources serve same format)
- `canonical` field is dropped for v1
- `scope` field is optional (filter which packages come from which source)
- URL points to the base of a per-package file directory

Full design is deferred to the index format spec (AIPKG-52).

---

### Decision 51: Scope claiming process (sketch)

**Context**: Virtual packages use the upstream author's scope. When the
author wants to take ownership, there needs to be a process.

**Decision**: Scope claiming is a simple verification process:
1. Author opens issue/PR on the index repo requesting their scope
2. Author proves GitHub org/user ownership (e.g., add `.aipkg-verify`
   file to their org's `.github` repo, or verified as org admin)
3. Index entries updated: `maintainer` changes, `source` becomes
   `"native"`, URLs updated to author's hosting
4. Old virtual versions remain available (no broken pinned installs)

**Rationale**: The scope-to-GitHub-owner mapping makes verification
straightforward. Keep the process manual and human-reviewed for v1.
Automate later if volume demands it.

**Deferred**: Dispute resolution (what if two people claim the same
scope?). Handle case-by-case for v1.

---

### Decision 52: Unified index with two entry schemas (package + recipe)

**Context**: The earlier design used `source: "native"` / `source: "virtual"`
with both types sharing the same structure (versions block with per-version
`dist` or `source` fields). This created confusion about what a "recipe"
is versus "index metadata," and forced virtual packages to enumerate
versions redundantly.

**Decision**: The index serves two distinct JSON schemas, discriminated
by a `type` field:

- `"package"`: published archives. Has a `versions` map where each
  version contains a `dist` block (URL + SHA-256).
- `"recipe"`: virtual packages. Has an `artifact` type and a `source`
  pointer (repo + path). No versions block.

Both schemas are served from the same index directory structure at the
same URL pattern (`packages.aipkg.dev/@scope/name.json`). The CLI reads
`type` and branches to the appropriate install strategy.

**Rationale**: Two different things deserve two different shapes. A
published package has a list of versioned archives. A recipe points at
upstream content and says "here's what this is." Forcing both into one
schema created redundant version enumeration for virtual packages and
blurred the conceptual boundary.

The recipe IS an index entry, not a separate artifact that gets
"compiled" into index metadata. This eliminates the earlier confusion
about recipe vs. metadata being separate concepts.

---

### Decision 53: Recipe entries are unversioned (consumer-side pinning)

**Context**: The earlier design had virtual package entries enumerate
versions like `"0.1.0+a3f8bc1": { "source": { "repo": ..., "ref": ... } }`.
For upstream repos with frequent commits, this list grows endlessly
with near-identical entries (same repo, same path, different ref).

**Decision**: Recipe entries carry no `versions` block. The recipe says
WHERE the content lives (repo + path). The CLI resolves the ref at
install time (HEAD of default branch, or a specific ref if the consumer
pins one). The resolved ref is recorded in the consumer's `require`
field. Subsequent `aipkg install` (no args) reinstalls at the pinned ref.

Version tracking for virtual packages lives on the consumer side, not
in the index.

**Rationale**: Virtual package "versions" are just commit refs. Enumerating
them in the index is 99% redundant (same repo, same path, different
SHA). Nobody maintains that list by hand, and generating it automatically
is the crawler work that we don't want to depend on. Letting the CLI
resolve the ref at install time is simpler for everyone.

**Trade-off**: You can't browse available versions of a virtual package
in the index. But with no version ranges (Decision 38), there's no use
case for that anyway. You either get "latest" or pin an exact ref.

---

### Decision 54: Crawler is optional automation, not core architecture

**Context**: Earlier discussion elevated the crawler to a core
architectural component, with AIPKG-7 (recipe format) being "absorbed
by the crawler." This created a false dependency on crawler
infrastructure before virtual packages could work.

**Decision**: The crawler is an optional automation layer that generates
recipe entries at scale. It is not on the critical path. Virtual
packages work with manually-written recipe entries in the index. The
crawler can come later to automate recipe creation for the 350K+ skills
on skillsmp.com.

AIPKG-7 (Virtual Package Recipe Format) is preserved as a spec
deliverable. It defines the `type: "recipe"` schema.

**Rationale**: You don't need automation to prove the concept. A handful
of hand-written recipe entries in the index is enough to test the full
install flow end-to-end. The crawler scales recipe creation but isn't
needed to ship virtual package support.

---

### Decision 55: Keep dist wrapper in version entries for extensibility

**Context**: Should the version entry for a package flatten the URL and
hash to the top level, or nest them under a `dist` key?

**Decision**: Keep the `dist` wrapper. Version entries use
`"dist": { "url": ..., "sha256": ... }` rather than putting `url` and
`sha256` at the top level of the version object.

**Rationale**: The wrapper allows future distribution methods to be added
as siblings. A version entry could eventually carry both
`"dist": { ... }` and `"oci": { ... }` (or other methods). The CLI
picks the preferred strategy. This is Principle V (backward-compatible
evolution) applied cheaply. The nesting costs a few bytes of JSON and
buys extensibility without any code changes later.

---

### Decision 56: HTTP-only transport (no file:// scheme)

**Context**: To support local testing, a `file://` repository scheme was
considered alongside `https://`. This would let the CLI read index files
directly from disk without a running server.

**Decision**: The CLI implements HTTP(S) only. Local testing uses any
off-the-shelf static file server (`python -m http.server`, `npx serve`,
etc.) pointed at a directory of index files and archives.

**Rationale**: Supporting `file://` means two code paths in the resolver
(filesystem reads vs. HTTP fetches). The index format is static JSON
files, which is exactly what static file servers are for. One transport
in the CLI, zero extra tooling for local testing. Keeps the codebase
simple (Principle I).

---

### Decision 57: Environment variable for registry override (defer repositories config)

**Context**: The brainstorm designed a `repositories` field in the project
manifest for configuring multiple package sources with priority ordering.
Implementing this fully requires schema design, resolution priority logic,
scope filtering, and more. Meanwhile, the install command needs a way to
point at a test registry during development.

**Decision**: Use a single environment variable (`AIPKG_REGISTRY` or
similar) to override the default registry URL. Defaults to
`https://packages.aipkg.dev`. The full `repositories` manifest config
is deferred to a later feature.

**Rationale**: An env var is the smallest thing that makes the install
command testable. Set it to `http://localhost:8000` and serve a local
directory; set nothing and get the central index. This is a well-known
pattern (npm's `NPM_CONFIG_REGISTRY`, Docker's registry mirrors). When
`repositories` lands later, the env var can become an override or
fallback. No wasted work.

---

### Decision 58: Feature split by entry type, not by transport

**Context**: The install command was initially scoped as a single large
feature covering both `type: "package"` and `type: "recipe"` entries
plus multiple install sources (local, remote, direct URL). This violated
Principle I ("what is the smallest useful thing?").

**Decision**: Split into two features ordered by entry type:

- **Feature A (package install)**: Installs `type: "package"` entries
  via the dist strategy (resolve from index, download archive, verify
  SHA-256, unpack). Includes the require format, registry env var,
  `aipkg install` for reinstall, and bundled dep extraction.
- **Feature B (recipe install)**: Installs `type: "recipe"` entries via
  the source strategy (fetch from GitHub, pack locally, then the same
  unpack flow). Adds GitHub API integration, ref resolution, and local
  packing.

Feature A is self-contained: pack an archive, publish it to a test
registry, install it. Feature B builds on A by adding the source
strategy for virtual packages.

**Rationale**: Feature A closes the pack-to-install loop with zero
GitHub API dependency. It validates the index format, resolution logic,
archive extraction, and require management. Feature B layers on the
network complexity of fetching from GitHub. Splitting by entry type
rather than transport avoids partially-useful intermediate states.
