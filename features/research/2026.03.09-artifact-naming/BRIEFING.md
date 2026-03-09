# Artifact Naming & Placement: Session Briefing

**Date**: 2026-03-09
**Session**: Design decisions on installed artifact naming, collision
handling, and content transformation. Triggered during the
004-package-install spec revision when three-segment naming
(`scope.package.artifact`) was found to be incompatible with the
Agent Skills specification.

**Pre-requisite reading**: The 2026-03-03 brainstorm (Decisions 59-71,
especially Decision 64: Scoped artifact naming), the Agent Skills
specification at agentskills.io, and `spec/artifacts.md`.

---

## The Problem

Decision 64 (2026-03-03) established that installed artifacts need
scoped naming to prevent collisions, but marked the exact format as
TBD. The 004-package-install spec initially adopted three-segment
naming (`scope.package.artifact`, e.g., `tjespers.golang-expert.debug`)
with dots as separators.

During spec quality review, this was found to conflict with the Agent
Skills specification, which restricts the `name` field in SKILL.md
frontmatter to `[a-z0-9-]` (lowercase alphanumeric and hyphens only,
no dots). The `name` field must match the directory name per the spec.

Research note R-001 from feature 003 had investigated three-segment
naming but only validated filesystem compatibility (Windows reserved
names, path lengths, case sensitivity). It did not check the Agent
Skills `name` field constraint. This gap was caught by the spec
quality checklist (CHK001).

## What We Decided

### Drop three-segment naming for v1

Use original artifact names from the archive manifest. A skill called
`debug` in the archive is placed at `.aipkg/skills/debug/`. No
renaming, no content transformation.

### Collision detection instead of collision prevention

Without scoped naming, two packages shipping a skill called `review`
would conflict. The solution is detection at install time: before
placing any artifact, check if another installed package already has
an artifact of the same type and name. If a collision is found, the
second `aipkg require` fails with an error identifying both packages
and the conflicting artifact name.

### No content transformation during placement

The original spec required updating SKILL.md frontmatter `name` fields,
command frontmatter, MCP server merge keys, and agent-instructions
section markers to reflect three-segment names. With original names
preserved, none of this is needed. Artifacts are placed exactly as
they appear in the archive.

### Artifact ownership via cached manifests

Without a naming prefix, there is no way to look at a file in
`.aipkg/skills/` and know which package it came from. Instead,
artifact ownership is determined by reading the `artifacts` array
from cached archive manifests in `~/.aipkg/cache/`. This is used
for upgrades (FR-014: removing old artifacts before placing new ones)
and collision detection (FR-013: checking what other packages own).

## Decisions

This session produced decisions 72-75. Full rationale in
[01-design-decisions.md](01-design-decisions.md). Summary:

| # | Decision | Why it matters |
|---|----------|----------------|
| 72 | Original artifact names | Agent Skills compatible, no transformation needed |
| 73 | Collision detection at install time | Prevents conflicts without scoped naming |
| 74 | No content transformation | Simpler install, artifacts are placed as-is |
| 75 | Cached manifest for artifact ownership | Enables upgrades and collision checks without naming convention |

## Related: Agent Skills Naming Constraint

The Agent Skills specification restricts skill directory names to
`[a-z0-9-]`. This constraint works for single-source skill directories
but creates tension for package managers distributing skills from
multiple sources into a shared directory.

The aipkg project plans to engage with the Agent Skills community
(agentskills/agentskills#81) to discuss loosening the naming constraint
to support scoped distribution. This is a community engagement effort
that runs in parallel with shipping v1. It does not block any feature
work.

## Open Questions (not blocking)

1. If scoped naming is ever adopted (community alignment or v2), what
   migration path exists for projects using original names?
2. Should `aipkg remove @scope/name` be part of the install feature
   or a separate feature? (Currently not in scope for 004.)
3. Should collision detection consider merged artifacts (MCP server
   keys, agent-instructions sections) in addition to file/directory
   artifacts?

---

**Detailed reference docs** (same directory):
- [01-design-decisions.md](01-design-decisions.md) -- decisions 72-75
