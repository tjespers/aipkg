# Session Context: Project Model & Init

**Date**: 2026-03-03
**Trigger**: Attempted to spec AIPKG-10 (Package Install) and discovered
the project model is an unresolved prerequisite. The install command needs
to know: what file declares a project? What directory structure does the
tool own? How does `require` work as a concept?

Same pattern as the 2026-03-02 session where install exposed the missing
resolver architecture.

## What exists today

**Package authoring side (done):**
- `aipkg create` scaffolds a package directory with `aipkg.json` manifest
- `aipkg pack` produces a `.aipkg` archive from a package directory
- Manifest schema covers package identity, artifacts, and has a `require`
  field that does nothing yet

**Consumer side (nothing):**
- No concept of a "project" that consumes packages
- No `aipkg init` command
- No `.aipkg/` directory structure defined
- No installed package layout defined
- `require` field exists in schema but has no semantics

## What we need to answer

1. **Project file**: What declares an aipkg-enabled project? A new file?
   Reuse `aipkg.json` with a type discriminator? Something else entirely?
2. **Project directory layout**: What does `.aipkg/` contain? How are
   installed artifacts organized for tool discovery?
3. **Require semantics**: What does the `require` field look like? How does
   it interact with install/require commands?
4. **Init command**: What does `aipkg init` do? What does it create?
5. **Integration surface**: How do installed artifacts become discoverable
   by tools (Claude Code, Cursor, etc.)? This determines the layout.

## Related decisions from previous sessions

- Decision 44: Helm-style bundled deps (resolved at pack time, bundled in
  archive)
- Decision 49: Dependency visibility default internal (author controls
  whether bundled deps are exposed to the user)
- Decision 56: HTTP-only transport
- Decision 57: AIPKG_REGISTRY env var for registry override

## Linear context

- AIPKG-50: Project Initialization (this session's primary subject)
- AIPKG-10: Package Install (blocked by AIPKG-50)
- AIPKG-7: Recipe Install (blocked by AIPKG-10)
- AIPKG-11: Adapter Execution (future, but informs layout decisions)

## Comparable systems

- **Composer**: `composer.json` serves both library and project.
  `composer init` creates it. `vendor/` holds installed packages.
- **npm**: `package.json` serves both. `npm init` creates it.
  `node_modules/` holds installed packages.
- **Helm**: Charts install to a release namespace. No "project" concept
  per se; the cluster is the project.
- **Claude Code**: Looks for `CLAUDE.md` in project root, skills in
  `.claude/skills/`, MCP configs in `.claude/mcp.json`.
