# Project Model Architecture: Session Briefing

**Date**: 2026-03-03
**Session**: Brainstorm on the aipkg project model, init command, and
install layout. Triggered by attempting to spec AIPKG-10 (Package Install)
and discovering the project model was an unresolved prerequisite.

**Pre-requisite reading**: Familiarity with the package manifest schema
(`aipkg.json`), the `.aipkg` archive format (002-archive-format-pack),
the source resolution architecture (2026-03-02 brainstorm), and the
constitution (`.specify/memory/constitution.md`).

---

## The Problem

The CLI can pack archives but has no concept of a "project" that consumes
them. The install command needs answers to: what file tracks dependencies?
Where do installed artifacts go? How does the directory structure support
tool integration?

The existing `aipkg.json` is a package manifest (authoring). It has a
`require` field that does nothing. There's no project-side equivalent.

## Architecture: What We Designed

### Separate project file

`aipkg-project.json` lives at the project root. Distinct from `aipkg.json`
(package manifest). No identity fields (name, version). Purely operational.

v1 schema:

```json
{
  "require": {
    "@scope/name": "1.2.0"
  }
}
```

Rationale: aipkg artifacts are not vendor code that gets imported. The
authoring schema (package identity, artifacts) and consumption schema
(dependencies, config) are genuinely different concerns. Separate files
keep validation clean and let each evolve independently.

### Categorized install layout

Artifacts are fanned out into well-known directories under `.aipkg/`:

```text
project-root/
  aipkg-project.json
  .aipkg/
    .gitignore
    mcp.json
    agent-instructions.md
    skills/
    prompts/
    commands/
    agents/
```

Individual artifact types (skills, prompts, commands, agents) get their own
directories. Mergeable artifact types (MCP servers, agent instructions)
produce a single merged file at the `.aipkg/` root. The merge is a core
responsibility, performed at install time.

Rationale: if the end state is `.aipkg/` replacing `.claude/`, `.cursor/`,
etc. as the unified AI artifact location, the directory structure IS the
standard. Tools looking for skills just read `.aipkg/skills/`. The cleaner
the layout, the easier the adoption pitch. Adapters (bridging `.aipkg/` to
tool-specific directories) become unnecessary as tools adopt the standard.

### Init flow

1. `aipkg init` creates `aipkg-project.json` with empty `require`
2. Optionally prompt to add dependencies interactively (nice-to-have)
3. If deps added, install them (creates `.aipkg/` as side effect)
4. If no deps, done (just the project file, no `.aipkg/` directory)

The `.aipkg/` directory is created on demand, not eagerly. A `.gitignore`
inside it ignores everything (git repo detection, strict by default).

### Scoped artifact naming

Installed artifacts use a scoped naming convention including the package
identity to prevent collisions (e.g., `skills/tjespers.golang-expert.review/`).
Exact format TBD during spec phase; needs research into what naming patterns
the target tools actually support.

### Project config as package registry

No separate `installed.json`. The project file's `require` field plus the
scoped naming convention provides full traceability. `aipkg remove` reads
the config, applies the naming convention, deletes the right artifacts,
removes the `require` entry.

### No coexistence

`aipkg.json` and `aipkg-project.json` MUST NOT coexist in the same directory.
A follow-up analysis (Decision 71) found that allowing both files creates CLI
ambiguity, directory pollution, and confusing UX with no plausible use case.
The package manifest's `require` field already covers bundled dependencies.

## Decisions

This session produced decisions 59-71. Full rationale in
[01-design-decisions.md](01-design-decisions.md) and
[02-coexistence-decision.md](02-coexistence-decision.md). Highlights:

| # | Decision | Why it matters |
|---|----------|----------------|
| 59 | Separate project file | Clean schema split, no dual-purpose ambiguity |
| 60 | Categorized install layout | Directory structure IS the tool integration standard |
| 62 | Named `aipkg-project.json` | Boring, clear, pairs with `.lock` |
| 66 | Minimal v1 schema | Just `require`, Principle I applied |
| 67 | `.aipkg/` on demand | No empty directories, materializes when needed |
| 69 | Individual vs. merged types | Skills/prompts/commands = dirs, MCP/instructions = merged files |
| 70 | Merged output only | No per-package file duplication for mergeable types |
| 71 | No coexistence | Package and project files are mutually exclusive per directory |

## Next: Spec Work

**AIPKG-50** (Project Initialization) should be specced first:
- Project file JSON schema
- `.aipkg/` directory structure specification
- `aipkg init` command behavior
- Documentation on what an aipkg-enabled project is

**AIPKG-10** (Package Install) follows, building on the project model:
- Resolution and download
- Artifact fan-out into categorized directories
- Merge logic for MCP servers and agent instructions
- `require`/`install` command behavior

## Open Questions (not blocking spec work)

1. Scoped artifact naming format (needs tool research)
2. Interactive dep-adding in init (nice-to-have, not v1)
3. Existing `require` in `aipkg.json` (remove or repurpose for bundled deps?)

---

**Detailed reference docs** (same directory):
- [00-session-context.md](00-session-context.md) -- starting state
- [01-design-decisions.md](01-design-decisions.md) -- decisions 59-70
- [02-coexistence-decision.md](02-coexistence-decision.md) -- decision 71
