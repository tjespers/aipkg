# Spec-for-Implementers Checklist: Project Initialization & Model

**Purpose**: Validate that the project file format, install directory layout, and naming convention are specified precisely enough to implement the install command (AIPKG-10) without ambiguity. Includes cross-spec consistency with existing naming rules, artifact types, and manifest schema.
**Created**: 2026-03-08
**Feature**: [spec.md](../spec.md)

## Project File Schema Completeness

- [x] CHK001 - Is the exact JSON structure of an empty `aipkg-project.json` (as created by `init`) fully specified with a concrete example? [Completeness, Spec §FR-004] — data-model.md:28-35
- [x] CHK002 - Are the `require` map key constraints (scoped name format) explicitly cross-referenced to the validation regex in `spec/naming.md`? [Cross-spec Consistency, Spec §FR-004] — data-model.md:20 refs package manifest pattern
- [x] CHK003 - Is the `require` map value pattern (semver with optional pre-release) defined with enough precision to write a JSON Schema regex? [Clarity, Spec §FR-004, Research §R-002] — research.md R-002 has exact regex
- [x] CHK004 - Is it specified whether `require` can be absent vs. must be present-but-empty? The schema says "required" but the spec should state this explicitly for implementers. [Clarity, Spec §FR-004] — data-model.md field table: Required: yes
- [x] CHK005 - Are forbidden fields (`name`, `version`, `description`, `license`) enumerated exhaustively, or is the intent "no fields beyond specVersion and require"? [Clarity, Spec §FR-003] — data-model.md:24-25: forbidden list + additionalProperties: false
- [x] CHK006 - Is the `specVersion` value for v1 documented as a constant (`1`) consistent with the package manifest's `specVersion` definition in `spec/manifest.md`? [Cross-spec Consistency, Spec §FR-004] — data-model.md:16
- [x] CHK007 - Does the spec define behavior when `specVersion` is a value other than `1`? (Relevant for future migration.) [Coverage, Gap] — Schema uses const: 1 (matching package manifest pattern); validation catches it. Human-friendly version mismatch error is a future concern.

## Install Directory Layout Completeness

- [x] CHK008 - Are all six artifact types from `spec/artifacts.md` accounted for in the install directory layout? Four individual types (skills, prompts, commands, agents) get subdirectories; two mergeable types (mcp-servers, agent-instructions) get root-level merged files. [Cross-spec Consistency, Spec §FR-007, §FR-008] — data-model.md:86-93
- [x] CHK009 - Is the `.gitignore` content specified precisely enough to implement (exact file contents: `*` and `!.gitignore`)? [Clarity, Spec §FR-009] — data-model.md:59
- [ ] CHK010 - Is the git detection method defined? "Within a git working tree" could mean checking for `.git/` in the project root or walking up the directory tree. [Clarity, Spec §FR-009] — DEFERRED: install-command scope, does not block 003 tasks
- [ ] CHK011 - Are the merged file formats specified? `mcp.json` implies a JSON structure, but is the schema defined? `agent-instructions.md` implies Markdown, but are the package-identifying markers specified? [Gap, Spec §FR-008, §FR-011] — DEFERRED: install-command design decisions. High-level layout is documented.
- [x] CHK012 - Is the ownership model for merged files (overwritten on install/update, manual edits lost) stated clearly enough that the install command can document this in user-facing output? [Clarity, Spec §FR-011] — data-model.md:78, spec FR-011

## Scoped Naming Convention Completeness

- [x] CHK013 - Is the three-segment naming format (`scope.package-name.artifact-name`) documented with enough examples covering all artifact types (directory-based skills vs. file-based prompts/commands/agents)? [Completeness, Research §R-001] — data-model.md:86-93, research.md R-001
- [x] CHK014 - Is the parsing algorithm for three-segment names unambiguous? (Split on `.`, first = scope, last = artifact, middle = package.) Does the spec explicitly state that scope and package names cannot contain dots, making this split safe? [Clarity, Research §R-001, Cross-spec with spec/naming.md] — data-model.md:95-101, R-001 confirms no-dots invariant
- [ ] CHK015 - For file-based artifacts, is the relationship between the scoped name and the file extension specified? E.g., is the installed file `scope.pkg.artifact.md` or `scope.pkg.artifact` with the original extension preserved? [Gap, Research §R-001] — DEFERRED: install-command scope. data-model.md uses `.ext` placeholder.
- [ ] CHK016 - For directory-based artifacts (skills), is the internal structure of the installed directory specified? Is it a copy of the original `skills/artifact-name/` directory contents, or does the structure change? [Gap, Spec §FR-012] — DEFERRED: install-command scope
- [x] CHK017 - Is the naming convention consistent with the existing dot-notation in `spec/naming.md`? The current spec says `scope.artifact-name`; the plan extends to three segments. Is the update to `spec/naming.md` tracked as a deliverable? [Cross-spec Consistency, Research §R-001] — RESOLVED: plan.md now lists `spec/naming.md` as MODIFIED in project structure.

## Cross-Spec Consistency

- [x] CHK018 - Does the project file `require` key pattern use the same scoped name regex as the package manifest schema (`spec/schema/aipkg.json`)? [Consistency, Spec §Assumption 2] — data-model.md:20 confirms
- [x] CHK019 - Is the project `require` field (installed dependencies, resolved at install time) clearly scoped? Package-level dependencies (bundled at pack time) are a separate concern. [Clarity, Spec §Assumption 5] — Spec Assumption 5 is explicit
- [x] CHK020 - Are the artifact type names used in the install directory layout (`skills/`, `prompts/`, `commands/`, `agents/`) consistent with the `type` enum values in `spec/schema/aipkg.json` (`skill`, `prompt`, `command`, `agent`)? Note the singular-vs-plural difference. [Cross-spec Consistency, Spec §FR-007] — data-model.md:86-91 maps singular types to plural directories, consistent with artifacts.md
- [x] CHK021 - Is the mutual exclusivity of `aipkg.json` and `aipkg-project.json` documented in both directions? The 003 spec covers "init refuses when aipkg.json exists." Does it also address future commands that create `aipkg.json` refusing when `aipkg-project.json` exists? [Coverage, Spec §FR-002] — FR-002 states the bidirectional invariant. This feature implements one direction. spec/project.md should state the full rule.

## Acceptance Criteria Quality

- [x] CHK022 - Are all three error paths (re-init guard, mutual exclusivity guard, write permission failure) specified with enough detail to write deterministic assertions? [Measurability, Spec §US-2, §US-3, §Edge Cases] — DD-005 defines style, quickstart shows check order, spec gives intent
- [x] CHK023 - Is the exact content of the error message for the mutual exclusivity guard specified, or just the intent? FR-017 says "suggest `aipkg require` or `aipkg install`" but does not give the exact wording. Is this intentional flexibility or a gap? [Clarity, Spec §FR-017] — Intentional flexibility. DD-005 says follow existing CLI patterns.
- [x] CHK024 - Is the success output of `aipkg init` specified? (E.g., should it print a confirmation message like `create` does, or be silent?) [Gap, Spec §FR-015] — RESOLVED: DD-005 updated to include success output. Follows `create` command pattern.

## Edge Case Coverage

- [x] CHK025 - Is it specified what happens when both `aipkg.json` and `aipkg-project.json` exist simultaneously (e.g., created manually outside the CLI)? FR-002 says they "MUST NOT coexist" but only `init` enforces this. Is enforcement by other commands in scope for this spec? [Coverage, Spec §FR-002] — Init checks aipkg.json first (quickstart:41-44), hits mutual exclusivity guard. Other commands' enforcement is out of scope.
- [x] CHK026 - Is behavior defined for when the directory becomes read-only between the existence check and the file write (TOCTOU)? The edge cases mention write permissions but not the race condition. [Edge Case, Gap] — N/A: not a practical concern for a CLI tool
- [x] CHK027 - Is behavior defined for symlinked `aipkg-project.json` or `aipkg.json` files? `os.Stat` follows symlinks; is this the intended behavior for the existence checks? [Edge Case, Gap] — N/A: os.Stat following symlinks is the expected behavior

## Plan Review Findings

- [x] CHK028 - `LoadFile()` in `internal/project` is listed as a deliverable (plan.md:66, quickstart.md:11) but no FR or acceptance scenario requires loading project files. Pragmatically justified (test verification, Go package completeness), but the plan should include a one-liner rationale to avoid scope creep questions during code review. [Scope, plan.md] — RESOLVED: DD-006 added with rationale (test verification + future command dependency).

## Notes

- Focus: Spec quality for downstream implementers (especially the install command, AIPKG-10)
- Depth: Standard
- Audience: Implementer/reviewer at PR review time
- Cross-spec references checked: `spec/naming.md`, `spec/artifacts.md`, `spec/schema/aipkg.json`, `spec/manifest.md`
