# aipkg — Research Topics

Open questions that need further exploration before or during specification. Roughly ordered by priority.

## High Priority (Blocks v1 Design)

### 1. Manifest Format Details
Beyond the settled fields (`type`, `name`, `version`, `artifacts`, `require`, `repositories`), what else is needed? Description, license, keywords, author, homepage? What's required vs optional per `type`?

### 2. Package Archive Format
Zip? tar.gz? Both? What's the expected structure inside the archive? Should there be a convention for the archive filename (e.g., `package-name-1.0.0.zip`)?

### 3. Versioning Strategy
Semver from day 1? AI artifacts don't have traditional APIs — what does a "breaking change" mean for a prompt or skill? With exact pinning in v1, this is less urgent but still shapes the format. Does the manifest enforce semver validation?

### 4. Adapter Interface Specification
What contract must an adapter implement? Is it a config file mapping artifact types to filesystem paths? Or actual code (a plugin)? How does the CLI discover available adapters? MCP server config merging needs logic beyond simple symlinking — how is that handled?

### 5. CLI Language Choice
Go (single binary, fast, great CLI ecosystem), TypeScript (fast to prototype, audience knows it), Rust (performance, but steeper curve). Trade-offs between time-to-v1 and long-term maintainability.

## Medium Priority (Shapes v1, Blocks v2)

### 6. Source Type Interface
What contract must a source type implement? Just URL resolution, or also listing/search? How are credentials handled per source type? Plugin architecture or built-in?

### 7. Namespace Validation Rules
Allowed characters, length limits, reserved names (`aipkg`, `core`, `official`, etc.). Should namespaces be case-sensitive? Normalization rules?

### 8. GitHub Source Type Details
Which GitHub APIs to use? How to handle private repos (token auth)? How to map package versions to GitHub release tags? Monorepo support (one repo, multiple packages)?

### 9. Non-Technical Publishing Workflow
How does someone without Git/CLI create a valid package? Web UI uploader? Template generator? Is a valid package literally just "folder with aipkg.json, zipped"?

### 10. Conflict Detection Details
What happens when two configured repos provide the same package name? Error? Priority order? What about two bundles shipping artifacts with the same name? How are exact-version dependency conflicts reported to the user?

### 11. Project vs Global Interaction
What happens when the same package is installed both globally and in a project? Which takes precedence? Can adapters show both? How does `aipkg list` display this?

## Lower Priority (v2+ Concerns)

### 12. Version Range Syntax
When ranges are introduced, what syntax? Semver ranges (`^1.0.0`, `~2.1`)? Composer-style constraints? npm-style? This shapes the constraint solver design.

### 13. Selective Install / Cherry-Picking
Syntax for installing specific artifacts from a bundle. Dependency warnings for intra-bundle references. E.g., `aipkg install @scope/pkg:artifact` or `--only skills`.

### 14. Lockfile Design
When to introduce? What does it capture? Hash verification? With exact versions, a lockfile is mostly about integrity hashes and source URLs — still useful but not critical.

### 15. Dependency Confusion Prevention
Canonical repos, namespace-to-source binding. How to prevent a public package from impersonating a private one. Security model for mixed public/private sources.

### 16. Trust Model & Prefix Behavior
Should configured/trusted scopes install artifacts without namespace prefix? Multi-scope collision handling. Default scope concept (`aipkg config set default-scope @myorg`).

### 17. Namespace Governance (Registry Era)
How does namespace claiming work when the central registry arrives? GitHub org verification? Domain verification? Trademark disputes?

### 18. MCP Server Config Merging
MCP server artifacts need to be merged into tool config files (not just symlinked). How to handle merge conflicts? How to cleanly unmerge on uninstall? Schema validation?

### 19. `aipkg link` — Local Development
Like `npm link` — develop a package locally and have it available as if installed. Symlink from source directory to the install scope's package store.

### 20. Monorepo Support
One repository containing multiple packages. How are they discovered? Subdirectory conventions? Separate manifests per package?

### 21. Update/Upgrade Flow
How to handle breaking changes in AI artifacts. Migration guides? Changelogs? Side-by-side versions? With exact pinning, updates are explicit — but how does the user discover that new versions exist?

### 22. Online Package Management UI
Web-based interface for browsing, discovering, and managing packages. The "npmjs.com" equivalent. Business model considerations (freemium: free public, paid private/teams).

### 23. Virtual Package Recipe Format
The recipe format in `aipkg-virtual` needs specification. How expressive should `artifact_mapping` be? How does `version_source` work (tags, releases, branches)? What happens when upstream restructures between versions — do recipes support version-conditional mappings?

### 24. Virtual Package Version Discovery
How does the CLI discover available versions for a `@virtual/` package? GitHub Releases API? Tags? What about upstreams that don't use semver tags? Mapping conventions between upstream versioning and aipkg version requirements.

### 25. Virtual-to-Official Migration Path
When an upstream author adopts aipkg natively, how does the transition work? CLI detection ("an official package exists for this virtual package")? Automatic migration command? What happens to version history?

### 26. Contribution Pipeline & Telemetry
The zero-friction "share with the world?" flow needs design: what data is collected? Where does it go (endpoint)? Privacy implications. AI vetting pipeline specification — what does "good quality" mean for a recipe? Auto-publish criteria vs human review triggers.

### 27. Bulk Ingestion Pipeline
Server-side scanning of entire repos (awesome-lists, skill collections). GitHub API rate limiting strategy. License compliance checking. Quality filtering (not everything in a repo is worth packaging). Author notification policy. Deduplication across multiple sources.
