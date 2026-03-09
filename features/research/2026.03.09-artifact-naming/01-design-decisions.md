# Design Decisions: Artifact Naming & Placement

Decisions from the 2026-03-09 session on installed artifact naming
and collision handling. Continues the numbering from the 2026-03-03
session (last: 71).

---

## Decision 72: Use original artifact names in install directories

**Context:** Decision 64 established scoped artifact naming but left the
exact format TBD. The 004-package-install spec adopted three-segment
naming (`scope.package.artifact`) with dots as separators. During spec
quality review, this was found to conflict with the Agent Skills
specification, which restricts the `name` field in SKILL.md frontmatter
to `[a-z0-9-]` only. The `name` field must match the directory name.

**Options considered:**
- Three-segment naming with dots (e.g., `tjespers.golang-expert.debug`).
  Incompatible with Agent Skills spec.
- Three-segment naming with hyphens (e.g., `tjespers-golang-expert-debug`).
  Ambiguous parsing since hyphens are valid in artifact names.
- Push to loosen the Agent Skills naming constraint. Possible but depends
  on community consensus and timing.
- Original artifact names with collision detection. Simple, compatible,
  ships now.

**Decision:** Installed artifacts use their original names from the
archive manifest. A skill called `debug` in the archive is placed at
`.aipkg/skills/debug/`, not `.aipkg/skills/tjespers.golang-expert.debug/`.

**Rationale:** Original names are fully compatible with the Agent Skills
specification and every other tool that reads artifact directories.
No content transformation is needed. The trade-off is that collisions
become possible when two packages ship same-named artifacts. This is
handled by Decision 73.

The Agent Skills naming constraint (`[a-z0-9-]`) is sensible for
single-source skill directories. The tension only arises when a package
manager distributes skills from multiple sources into a shared namespace.
Rather than fight the constraint with workarounds, we work within it and
address collisions directly.

---

## Decision 73: Collision detection at install time

**Context:** With original artifact names (Decision 72), two packages
shipping a skill called `review` would overwrite each other in
`.aipkg/skills/review/`. This needs to be handled.

**Options considered:**
- Scoped naming (prevention). Ruled out by Decision 72.
- Artifact selection at install time (e.g., choose which artifacts from a
  package to install). Too complex for v1 with no multi-artifact packages
  in the wild yet.
- Error on collision. Simple, clear, sufficient for the current ecosystem.

**Decision:** Before placing any artifact, the CLI checks if another
installed package already has an artifact of the same type and name. If a
collision is found, the `aipkg require` command fails with an error that
identifies both packages and the conflicting artifact name.

**Rationale:** Collisions in practice will be rare. Most packages ship
one or two artifacts with distinctive names. When collisions do occur,
an error is the right response because silent overwriting would break
the first package. The error message gives the developer enough context
to decide how to proceed (remove one package, contact the author, etc.).

This is deliberately simple. If the ecosystem grows to the point where
collisions are frequent, scoped naming or artifact selection can be
revisited with real data about usage patterns.

---

## Decision 74: No content transformation during artifact placement

**Context:** Three-segment naming (before Decision 72) required
transforming artifact content at install time: updating SKILL.md
frontmatter `name` fields to match the scoped directory name, rewriting
MCP server merge keys to scoped names, updating command frontmatter,
and setting agent-instructions section markers.

**Decision:** No content transformation is applied during artifact
placement. Artifacts are placed exactly as they appear in the archive.

**Rationale:** With original artifact names, there is nothing to
transform. The skill directory name matches the SKILL.md `name` field
because neither was changed. MCP server entries use their original keys.
Agent-instructions sections use their original markers.

This is simpler to implement, simpler to debug, and eliminates an
entire class of bugs (malformed frontmatter after transformation,
encoding issues, edge cases in YAML/JSON rewriting). It also means
`aipkg pack` output and `aipkg require` output are identical in content;
the archive is just a transport wrapper.

---

## Decision 75: Cached manifest for artifact ownership tracking

**Context:** Without scoped naming prefixes, there is no way to look at
a file in `.aipkg/skills/` and determine which package installed it.
Artifact ownership is needed for two operations: removing old artifacts
during version upgrades (FR-014), and checking for collisions before
placing new artifacts (FR-013).

**Options considered:**
- Naming convention (prefix-based identification). Ruled out by
  Decision 72.
- Separate installed-packages manifest (a file tracking which artifacts
  belong to which packages). Adds a new file to maintain and keep in
  sync.
- Cached archive manifests. The archive in `~/.aipkg/cache/` already
  contains `aipkg.json` with the full `artifacts` array. Reading it
  provides the artifact list for any installed package.

**Decision:** Artifact ownership is determined by reading the `artifacts`
array from the cached archive's `aipkg.json` manifest. For each package
listed in `aipkg-project.json`'s `require` field, the CLI reads the
corresponding cached archive to discover which artifacts it owns.

**Rationale:** This avoids introducing a new tracking file. The data
already exists in the cached archives. The `require` field tells the CLI
which packages are installed, and the cached manifests tell the CLI which
artifacts each package contains. The combination is sufficient.

The risk is that a manually deleted cache entry would break upgrade
and collision detection. This is acceptable for v1: the cache is an
implementation detail, and manually modifying it is unsupported. If this
proves fragile in practice, a dedicated tracking file can be introduced
later without changing any user-facing behavior.
