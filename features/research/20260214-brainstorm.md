---
type: brainstorm
date: 2026-02-14 19:27 CET
participants:
  - name: Tim J
    role: founder
decisions: 9
tasks: 3
---

# Brainstorm Session -- AI Interoperability & AI Package Manager

## Agenda

Tim J used a solo Google Meet session as an experiment in monologue-to-spec workflow. Two main concepts were explored, with the second building directly on the first.

1. [AI interoperability and fragmentation](#1-ai-interoperability-and-fragmentation)
2. [Organizational shared skill library gap](#2-organizational-shared-skill-library-gap)
3. [AI package manager concept](#3-ai-package-manager-concept)
4. [Existing solutions and shortcomings (APM)](#4-existing-solutions-and-shortcomings-apm)
5. [AI package specification design](#5-ai-package-specification-design)
6. [CLI tool and operations](#6-cli-tool-and-operations)
7. [Interoperability via transformers](#7-interoperability-via-transformers)
8. [Registry and SpecKit as use case](#8-registry-and-speckit-as-use-case)
9. [Standardized distribution vision](#9-standardized-distribution-vision)
10. [Monologue transcription experiment](#10-monologue-transcription-experiment)

---

## Discussion & Decisions

### 1. AI interoperability and fragmentation

**Context**: Tim J opened by describing the current state of AI tooling: prompts, agents, skills, and MCP server configs are all fundamentally Markdown (often with YAML frontmatter), but every tool stores them in different directories (`.claude/`, `.cursor/skills/`, `.cursor/commands/`, etc.).

**Discussion**: Despite near-identical formats, portability is broken by directory conventions. Most tools support two layers: project-level (e.g., `.claude/` in the repo) and global/developer-level (e.g., `~/.claude/`). The content is the same Markdown; only the file paths differ per tool.

**Decisions**:
- [D1](#decision-log-quick-reference)

---

### 2. Organizational shared skill library gap

**Context**: Beyond project-level and developer-level skill storage, there is no mechanism for an organizational shared library of skills.

**Discussion**: Tim J walked through three current approaches for sharing skills across repos: (a) publish to a GitHub repo and have each developer manually clone into their home directory -- not scalable; (b) embed as Git submodules in every project -- more scalable but high maintenance overhead; (c) copy the exact skill files into every repo -- completely unmaintainable, updates require PRs to every repo. None of these are acceptable at scale.

**Decisions**:
- [D2](#decision-log-quick-reference)

---

### 3. AI package manager concept

**Context**: After exhausting current approaches, the realization emerged that what's actually needed is a package manager for AI artifacts.

**Discussion**: The "wild west" state of AI tooling means different developers use different IDEs that don't natively share skill files. A simple CLI that sim-links files between directories still felt hacky. The real solution is a central registry with versioned packages (skills, prompts, MCP server configs) that can be declared as project dependencies, similar to `npm install` pulling in JavaScript dependencies. The key insight: if you have a shared library that needs custom configuration per IDE, you're building a package manager whether you call it that or not.

**Decisions**:
- [D3](#decision-log-quick-reference)

---

### 4. Existing solutions and shortcomings (APM)

**Context**: Tim J investigated whether this already exists and found APM (Agent Package Manager) by a Microsoft developer.

**Discussion**: APM uses a YAML manifest to specify prompts/skills from external repos and supports `apm install`. However, it has significant gaps: no lock file (always pulls latest or whatever the link resolves to, no version pinning), requires custom runtimes for IDE-specific file transformation, and only solves the receiving/consuming end. The distribution side remains fragmented with no central registry, no standardized package format, and no way to know what you're cloning looks like (could be an entire repo with one skill buried inside).

**Decisions**:
- [D4](#decision-log-quick-reference)

---

### 5. AI package specification design

**Context**: To structure the solution, Tim J proposed defining an "AI package" specification analogous to Node modules or Composer packages.

**Discussion**: An AI package needs: a name, a version, a source (GitHub repo, website, etc.), and content. Content types identified: skill.md files, agent.md files, MCP server JSON configurations, and plain prompts. These are all either plain JSON or Markdown -- no compilation needed, runtime-agnostic. A generic prompt works across OpenAI, Anthropic, Cursor, VS Code, Claude Code. Packages can declare dependencies on other AI packages, creating a dependency tree (e.g., a "GitHub issue triager" package could depend on a package containing the GitHub MCP server config).

**Decisions**:
- [D5](#decision-log-quick-reference)
- [D6](#decision-log-quick-reference)

---

### 6. CLI tool and operations

**Context**: A CLI tool is needed to manage AI packages within projects.

**Discussion**: First-iteration operations: add, update, remove. Each operation updates the manifest and/or lock file and interacts with the filesystem. The manifest declares desired packages (like `package.json`), the lock file pins exact versions for reproducibility in CI/CD. Language candidates: Golang (modularity) or Python (ecosystem alignment).

**Decisions**:
- [D7](#decision-log-quick-reference)

**Follow-up**:
- [T2](#action-items)

---

### 7. Interoperability via transformers

**Context**: Once packages are downloaded, they need to end up in the right IDE-specific directories.

**Discussion**: Packages download into a vendor folder (e.g., `.aai/vendor/`) which is git-ignored. From there, a set of "transformers" handle IDE-specific placement: for Cursor, sim-link or copy skills to `.cursor/skills/`, agents to `.cursor/agents/`, etc. For Claude Code, link to `.claude/`. The copied/moved files must also be excluded from version control. This preserves developer freedom (use whatever IDE you want) while ensuring everyone on the project uses the same AI dependencies.

**Decisions**:
- [D8](#decision-log-quick-reference)

---

### 8. Registry and SpecKit as use case

**Context**: A central registry is needed for discovery and sharing at scale.

**Discussion**: Without a registry, everyone recreates standard items like GitHub API integrations. Tim J cited GitHub's SpecKit framework as a concrete example of the problem: SpecKit creates a `specify/` folder in your project root with vendor templates that get committed to version control. Upgrading requires backing up modified files, running the upgrade (which overwrites), then restoring -- a clear signal these are dependency files that shouldn't be under VCS. If SpecKit were distributed as a package, its internal files would live in the git-ignored dependency directory, and only user-customizable files would be in the codebase. The alternative (just git-ignoring `specify/`) doesn't scale: you'd end up with `.cursor/`, `specify/`, and a new folder for every AI tool, bloating the repo.

**Decisions**:
- [D9](#decision-log-quick-reference)

**Follow-up**:
- [T1](#action-items)

---

### 9. Standardized distribution vision

**Context**: The end-state vision for the ecosystem.

**Discussion**: Settle on a single directory (suggested `.aai`), standardize the distribution format, and unify the fragmented naming conventions. AI workflows are evolving from one-off prompts into full frameworks. Package management is a proven technique across all programming languages and frameworks -- AI tooling should follow the same path. The goal is to free skill/prompt authors from worrying about cross-vendor, cross-runtime compatibility.

---

### 10. Monologue transcription experiment

**Context**: This session itself was an experiment in voice-to-spec workflow.

**Discussion**: Tim J was walking outside without a computer, brainstorming into a Google Meet monologue transcribed by Gemini. The hypothesis: if spoken ideas can be transcribed, fed into an LLM for structuring, and then piped into spec-driven development tooling (like SpecKit), the bottleneck of translating fast-paced thinking into structured specs could be eliminated. "I can think faster than my body can execute" -- the gap between ideation speed and structured output is the productivity bottleneck this workflow targets.

**Follow-up**:
- [T3](#action-items)

---

## Decision Log (Quick Reference)

| # | Topic | Decision |
|---|-------|----------|
| D1 | [AI interoperability](#1-ai-interoperability-and-fragmentation) | The core problem is directory fragmentation -- content is portable, file paths are not |
| D2 | [Shared library gap](#2-organizational-shared-skill-library-gap) | Current sharing methods (manual clone, submodules, copy-paste) are all unacceptable at scale |
| D3 | [Package manager concept](#3-ai-package-manager-concept) | Build a dedicated AI package manager with a central registry and versioned packages |
| D4 | [APM shortcomings](#4-existing-solutions-and-shortcomings-apm) | APM is insufficient -- build a new solution that includes lock files, standardized distribution, and a registry |
| D5 | [Package spec](#5-ai-package-specification-design) | AI packages must have: name, version, source, content, and dependency declarations |
| D6 | [Package spec](#5-ai-package-specification-design) | Four content types on day one: skills, agents, prompts, MCP server configurations |
| D7 | [CLI tool](#6-cli-tool-and-operations) | CLI supports add, update, remove with manifest + lock file for reproducibility |
| D8 | [Transformers](#7-interoperability-via-transformers) | Packages download to a git-ignored vendor folder; transformers handle IDE-specific placement |
| D9 | [Registry](#8-registry-and-speckit-as-use-case) | A central registry is required for package discovery and sharing |

---

## Action Items

| # | Source | Task | Assignee |
|---|--------|------|----------|
| T1 | [Package spec](#5-ai-package-specification-design) | Define the AI package specification (schema): name, version, source, content types (skill.md, agents.md, MCP server JSON, prompts) | Tim J |
| T2 | [CLI tool](#6-cli-tool-and-operations) | Experiment with building the AI package manager on a small scale | Tim J |
| T3 | [Transcription experiment](#10-monologue-transcription-experiment) | Build agents/skills for transcribing and extracting ideas into structured spec-driven development input | Tim J |
