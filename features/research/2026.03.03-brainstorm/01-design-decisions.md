# Design Decisions: Project Model & Init

Decisions from the 2026-03-03 brainstorm session on the aipkg project
model. Continues the numbering from the 2026-03-02 session (last: 58).

---

## Decision 59: Separate project file from package manifest

**Context:** The existing `aipkg.json` describes a package (name, version,
artifacts). The install command needs a place to declare dependencies and
project settings. Traditional package managers (Composer, npm) use a single
file for both authoring and consumption.

**Decision:** Use a separate file for project configuration. `aipkg.json`
remains the package manifest. A new file handles the consumer/project side.

**Rationale:** aipkg artifacts are not traditional vendor code that gets
imported into userspace. The authoring schema (package identity, artifacts)
and the consumption schema (dependencies, layout, future adapter config)
are genuinely different concerns. Separate files keep schema validation
clean, avoid conditional field logic, and let each file evolve
independently. A package directory can coexist with a project file without
ambiguity (a package that also consumes dependencies).

---

## Decision 60: Categorized install layout

**Context:** Installed packages need to land somewhere in the project. Two
approaches: keep packages intact (`.aipkg/packages/@scope/name/`) or fan
out artifacts into well-known directories (`.aipkg/skills/`,
`.aipkg/prompts/`, etc.).

**Decision:** Categorized layout. Install fans out artifacts into
well-known directories under `.aipkg/`:

```
.aipkg/
  skills/
  prompts/
  commands/
  agents/
  agent-instructions/
  mcp-servers/
```

**Rationale:** If the end state is `.aipkg/` replacing `.claude/`,
`.cursor/`, etc. as the unified AI artifact location, the directory
structure IS the standard. Tools looking for skills just read
`.aipkg/skills/`. They don't need to know about aipkg's package model.
This makes tool adoption trivial and means adapters (which bridge between
`.aipkg/` and tool-specific directories) are a temporary measure, not a
permanent requirement. A flat package layout would be more opinionated and
harder to sell to tool vendors.

---

## Decision 61: Project config at project root

**Context:** The project configuration file could live inside `.aipkg/`
(clean root, single directory footprint) or at the project root (visible,
conventional).

**Decision:** Project config lives at the project root, not inside
`.aipkg/`.

**Rationale:** Root placement follows the npm/Composer convention developers
already know. The file is visible, discoverable, and easy to review in PRs.
The dependency list is front-and-center. Root placement also gives
unambiguous project root detection (walk up until you find the file).
Hidden files solve a problem we don't have.

---

## Decision 62: Project file named `aipkg-project.json`

**Context:** The project file needs a name that is clearly namespaced to
aipkg, won't clash with other tools, and isn't hidden.

**Options considered:**
- `aipkg.json` — already taken (package manifest)
- `project.json` — too generic, will clash
- `aipkg.project.json` — double extension, possible tooling confusion
- `.aipkgrc.json` — "rc" implies runtime config, not dependency manifest
- `aipkg-project.json` — descriptive, no ambiguity

**Decision:** `aipkg-project.json`.

**Rationale:** Boring, clear, zero edge cases. No double-extension
confusion for editors or schema associations. Naturally pairs with
`aipkg-project.lock` if a lockfile is ever needed. Same naming pattern
works for the package side (`aipkg.json` / `aipkg.lock`).

---

## Decision 63: Project config IS the installed package registry

**Context:** With a categorized install layout, we need a way to track
which artifacts came from which package (for remove, update, etc.). A
separate `installed.json` was considered.

**Decision:** The project config's `require` field combined with the
artifact naming convention provides full traceability. No separate
installed package registry needed.

**Rationale:** If `require` lists `@tjespers/golang-expert@1.2.0` and
the naming convention dictates that its "reviewer" skill lands at
`.aipkg/skills/tjespers.golang-expert.reviewer/` (exact format TBD),
then the project config plus the convention is enough to reconstruct
what's installed and where. `aipkg remove @scope/name` reads the config,
applies the naming convention, deletes the right files, removes the
`require` entry. Adding a separate registry file would duplicate
information the project config already holds.

---

## Decision 64: Scoped artifact naming in install directories

**Context:** When artifacts from different packages are fanned out into
shared directories (e.g., `.aipkg/skills/`), name collisions are possible.
Two packages could ship a skill called "reviewer."

**Decision:** Installed artifacts use a scoped naming convention that
includes the package identity. Exact format TBD during spec phase (needs
research into what naming patterns the target tools support).

**Rationale:** Namespacing prevents collisions and provides traceability
back to the source package. The exact format (dots, slashes, colons) needs
testing against real tools (Claude Code, Cursor, etc.) and will be settled
during the spec/planning phase for AIPKG-50 and AIPKG-10.

---

## Decision 65: Project file has no identity fields

**Context:** Traditional package managers include `name`, `version`, and
other identity fields in their config files because the same file serves
both authoring and consumption. With a separate project file (Decision 59),
these fields have no purpose.

**Decision:** `aipkg-project.json` contains no `name`, `version`, or other
identity fields. It is purely operational.

**Rationale:** The project file exists to declare dependencies and (in the
future) configuration. It is not a publishable artifact. Adding identity
fields would be cargo-culting from dual-purpose systems without serving any
need. Principle I: if it does nothing, don't ship it.

---

## Decision 66: Minimal v1 project file schema

**Context:** The project file needs a schema. The question is how much to
include in v1.

**Decision:** v1 schema contains only a `require` field:

```json
{
  "require": {
    "@scope/name": "1.2.0"
  }
}
```

**Rationale:** This is the minimum needed for `aipkg require` and
`aipkg install` to function. Future versions can add `repositories`
(registry config), adapter preferences, or other settings. Principle I
applied: ship the smallest useful thing.

---

## Decision 67: `.aipkg/` created on demand, not by init

**Context:** `aipkg init` creates the project file. Should it also create
the `.aipkg/` directory and its subdirectories?

**Decision:** `aipkg init` only creates `aipkg-project.json`. The `.aipkg/`
directory is created on demand when the first package is installed (via
`aipkg require` or `aipkg install`).

**Rationale:** An empty `.aipkg/` directory with empty subdirectories is
noise. If you init with zero deps, you get one file and nothing else. The
directory structure materializes when there's something to put in it.

---

## Decision 68: Strict `.aipkg/.gitignore`, loosen later if needed

**Context:** The `.aipkg/` directory contains installed artifacts
(generated, should not be committed). Future versions might add committable
config files inside `.aipkg/`.

**Decision:** When `.aipkg/` is first created, drop a `.gitignore` inside
it that ignores everything (`*` with `!.gitignore` exception). Only do this
if the project is inside a git repository. Loosen with explicit `!`
exceptions if committable files are ever added.

**Rationale:** Start strict, loosen later. Ignoring everything prevents
accidental commits of installed artifacts. Adding exceptions is easy;
tightening after people have already committed files is harder. Git
detection is straightforward (existing codebase already handles git-related
logic in the pack command).

---

## Decision 69: Two artifact layout strategies (individual vs. merged)

**Context:** The six artifact types have different consumption models.
Skills, prompts, commands, and agents are individual items (one file or
directory per artifact). MCP servers and agent instructions are mergeable
(multiple packages contribute entries to a single config/document).

**Decision:** Individual artifact types get their own directories under
`.aipkg/` (skills/, prompts/, commands/, agents/). Mergeable artifact types
produce a single merged file at the `.aipkg/` root level:

```text
.aipkg/
  mcp.json                # merged MCP config from all installed packages
  agent-instructions.md   # merged instructions from all installed packages
  skills/                 # individual skill directories
  prompts/                # individual prompt files
  commands/               # individual command files
  agents/                 # individual agent persona files
```

The merge is performed at install time as a core responsibility (not an
adapter concern). Merged files are regenerated deterministically whenever
a package is added or removed.

**Rationale:** This makes `.aipkg/` look like a native tool directory. A
tool adopting the standard reads `mcp.json` for MCP servers, reads
`agent-instructions.md` for instructions, scans `skills/` for skills. No
adapter or additional processing needed. The merge being core (not adapter)
is intentional: the merged output IS the standard format aipkg proposes.

---

## Decision 70: No individual-file copies for mergeable types

**Context:** For mergeable artifact types (mcp-servers, agent-instructions),
we could keep both the individual source files per package AND the merged
output, or just the merged output.

**Decision:** Only the merged output exists in `.aipkg/`. No individual
per-package files for mergeable types (no `.aipkg/mcp-servers/` or
`.aipkg/agent-instructions/` directories).

**Rationale:** Two representations of the same data is clutter. The merge
is deterministic from `require` + installed package contents. If you need
to trace which package contributed an MCP entry, re-read the package. This
can always be extended later if real use cases demand it; stripping files
that are already out there is harder than adding them.
