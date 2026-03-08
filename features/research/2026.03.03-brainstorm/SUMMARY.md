# Session Summary: Project Model & Init

**Date**: 2026-03-03
**Status**: Research complete. Decisions 59-71 captured. Ready for spec
work on AIPKG-50 (project init) followed by AIPKG-10 (package install).

## The Core Insight

The install command can't work without a project model. "Where do packages
go?" and "what file tracks dependencies?" are prerequisites for "how do
packages get installed?" This is the same pattern as the 2026-03-02 session
where install exposed the missing resolver architecture.

## What We Designed

**Separate project file** (`aipkg-project.json`) at the project root.
Distinct from the package manifest (`aipkg.json`). No identity fields, no
version, purely operational. v1 schema is just a `require` map of package
names to pinned versions.

**Categorized install layout** under `.aipkg/`. Artifacts are fanned out
into well-known directories (skills/, prompts/, commands/, agents/) rather
than kept as intact packages. This makes the directory structure itself the
integration standard. Tools looking for skills just read `.aipkg/skills/`.
No adapter needed if tools adopt the standard directly.

**Merged files for mergeable types.** MCP server configs merge into
`.aipkg/mcp.json`. Agent instructions merge into
`.aipkg/agent-instructions.md`. The merge is a core responsibility,
performed at install time. No individual per-package copies of mergeable
artifacts.

**Init creates the project file, nothing else.** The `.aipkg/` directory is
created on demand when the first package is installed. If you init with
zero deps, you get one file. A `.gitignore` inside `.aipkg/` ignores
everything (strict by default, loosen later if needed).

## The `.aipkg/` Directory Layout

```text
project-root/
  aipkg-project.json        # committed, tracks dependencies
  .aipkg/                   # gitignored, created on first install
    .gitignore              # ignores everything inside
    mcp.json                # merged MCP config
    agent-instructions.md   # merged agent instructions
    skills/                 # individual skill directories
    prompts/                # individual prompt files
    commands/               # individual command files
    agents/                 # individual agent persona files
```

## Key Decisions (59-71)

| # | Decision | One-liner |
|---|----------|-----------|
| 59 | Separate project file | `aipkg.json` = package, new file = project |
| 60 | Categorized install layout | Fan out into skills/, prompts/, etc. |
| 61 | Project config at root | Visible, conventional, marks project root |
| 62 | Named `aipkg-project.json` | Boring, clear, pairs with `.lock` |
| 63 | Project config = package registry | No separate installed.json needed |
| 64 | Scoped artifact naming | Prevents collisions, exact format TBD in spec |
| 65 | No identity fields | Project file is operational, not publishable |
| 66 | Minimal v1 schema | Just `require`, nothing else |
| 67 | `.aipkg/` on demand | Created on first install, not by init |
| 68 | Strict gitignore | Ignore everything, loosen later if needed |
| 69 | Individual vs. merged types | Skills/prompts/commands/agents = dirs, MCP/instructions = merged files |
| 70 | No individual copies for merged types | Merged output only, no per-package files |
| 71 | No coexistence | `aipkg.json` and `aipkg-project.json` are mutually exclusive per directory |

Full details: [01-design-decisions.md](01-design-decisions.md), [02-coexistence-decision.md](02-coexistence-decision.md)

## Roadmap Impact

**AIPKG-50** (Project Initialization) is now a prerequisite for AIPKG-10.
It covers:
- `aipkg init` command (create `aipkg-project.json`)
- Project file JSON schema
- `.aipkg/` directory structure spec
- Documentation on what an aipkg-enabled project is

**AIPKG-10** (Package Install) depends on the project model. The install
layout, require semantics, and merge behavior defined here are inputs to
the install spec.

**Existing `require` field in `aipkg.json`**: Currently does nothing
(Principle I concern). Should be addressed during AIPKG-50/AIPKG-10
spec work. It either gets removed from the package schema (require belongs
in the project file) or repurposed for declaring bundled dependencies.

## Open Questions (not blocking spec work)

1. **Scoped artifact naming format**: Needs research into what naming
   patterns Claude Code, Cursor, and other tools support. Dots, slashes,
   colons? Will be settled during spec/planning phase.
2. **Interactive dep-adding in init**: Nice-to-have Composer-style
   prompting. Not essential for v1 given the index may not be populated.
3. **Worktree behavior**: `.aipkg/` is gitignored, so new worktrees need
   `aipkg install`. Same as npm/Composer. Tool-specific worktree hooks are
   adapter territory (AIPKG-11).
