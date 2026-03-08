# Session Context: Source Resolution & Package Index Architecture

**Date**: 2026-03-02
**Trigger**: Attempted to spec AIPKG-10 (Install Command) via speckit-specify-primer. Discovered that the source/resolver architecture is a blocking prerequisite that needs design work before install can be specified.
**Participants**: Tim J + Claude

## Starting State

**Completed features:**
- 001-package-foundation: `aipkg create` command, manifest schema, naming rules, artifact types
- 002-archive-format-pack: `aipkg pack` command, `.aipkg` archive format, SHA-256 sidecar, `.aipkgignore`

**Existing design artifacts referenced:**
- `features/research/2026.02.26-brainstorm/02-architecture.md` (Resolve > Fetch > Unpack & Store > Adapt pipeline)
- `features/research/2026.02.26-brainstorm/01-design-decisions.md` (decisions 1-17, including GitHub-first, Composer-style repo config, multiple source types)
- `features/research/2026.02.28-brainstorm/01-design-decisions.md` (decisions 18-31, including Helm-style archive structure)
- `spec/schema/aipkg.json` (unified schema with `repositories` array: `type`, `url`, `scope`, `canonical`)
- `spec/archive.md` (archive format: `{scope}--{name}-{version}.aipkg`, SHA-256 sidecar)
- `.specify/memory/constitution.md` (five principles, especially I: Simplicity, III: Convention, IV: Cold Start)

**Relevant Linear issues:**
- AIPKG-5: Source Type Interface (spec, backlog)
- AIPKG-9: GitHub Source Type & Resolver (CLI, backlog)
- AIPKG-10: Install Command, Project Scope (CLI, backlog)
- AIPKG-13: HTTP Source Type (CLI, backlog, low priority)
- AIPKG-51: Local Package Linking (CLI, backlog)
- AIPKG-15: Virtual Package Support (CLI, backlog)
- AIPKG-16: Publish Command (CLI, backlog)
- AIPKG-46: Global config file (CLI, backlog, research)

## Key Realizations During Session

1. **Install depends on source resolution, which doesn't exist yet.** AIPKG-10's description assumes a resolver, but the resolver interface (AIPKG-5) and its implementation (AIPKG-9) are unsolved design problems, not just implementation tasks.

2. **"Scope = GitHub owner" convention breaks for monorepos.** The convention `@tjespers/my-skill` > `github.com/tjespers/my-skill` only works when one repo = one package. For catalog repos (one repo hosting many skills), the CLI can't derive which repo to look in from the package name alone.

3. **The ecosystem already has content; it just isn't packaged.** Repos like `ComposioHQ/awesome-claude-skills` contain dozens of skills. The value prop isn't creating content, it's making existing content installable.

4. **AI artifacts aren't like software packages.** Authors don't do semver releases, don't tag versions, don't think in terms of packages. The tooling must meet them where they are.

5. **Cold start requires the aipkg project to do the work.** With zero users, nobody will host indexes or structure repos. The Homebrew model (maintainers curate formulas for upstream software) solves this.

6. **An index-based discovery model solves all cases.** Instead of convention-based repo guessing, a static `index.json` maps package names to download URLs. Works for 1:1 repos, monorepos, and virtual packages identically.

7. **The @types/DefinitelyTyped maintenance trap is avoidable.** Skills don't have API surfaces that break when upstream changes. "Build once per pinned ref, store forever" keeps CI costs near zero. Graduation (authors take over their scope) reduces maintenance over time.

8. **SHA-256 integrity can live in the index.** No need to fetch sidecar files separately during install. The index entry carries the hash. The sidecar from `aipkg pack` remains useful for manual verification.
