# Spec-for-Implementers Checklist: Project Initialization & Model

**Purpose**: Validate that the project file format, install directory layout, and naming convention are specified precisely enough to implement the install command (AIPKG-10) without ambiguity. Includes cross-spec consistency with existing naming rules, artifact types, and manifest schema.
**Created**: 2026-03-08
**Feature**: [spec.md](../spec.md)

## Project File Schema Completeness

- [ ] CHK001 - Is the exact JSON structure of an empty `aipkg-project.json` (as created by `init`) fully specified with a concrete example? [Completeness, Spec §FR-004]
- [ ] CHK002 - Are the `require` map key constraints (scoped name format) explicitly cross-referenced to the validation regex in `spec/naming.md`? [Cross-spec Consistency, Spec §FR-004]
- [ ] CHK003 - Is the `require` map value pattern (semver with optional pre-release) defined with enough precision to write a JSON Schema regex? [Clarity, Spec §FR-004, Research §R-002]
- [ ] CHK004 - Is it specified whether `require` can be absent vs. must be present-but-empty? The schema says "required" but the spec should state this explicitly for implementers. [Clarity, Spec §FR-004]
- [ ] CHK005 - Are forbidden fields (`name`, `version`, `description`, `license`) enumerated exhaustively, or is the intent "no fields beyond specVersion and require"? [Clarity, Spec §FR-003]
- [ ] CHK006 - Is the `specVersion` value for v1 documented as a constant (`1`) consistent with the package manifest's `specVersion` definition in `spec/manifest.md`? [Cross-spec Consistency, Spec §FR-004]
- [ ] CHK007 - Does the spec define behavior when `specVersion` is a value other than `1`? (Relevant for future migration.) [Coverage, Gap]

## Install Directory Layout Completeness

- [ ] CHK008 - Are all six artifact types from `spec/artifacts.md` accounted for in the install directory layout? Four individual types (skills, prompts, commands, agents) get subdirectories; two mergeable types (mcp-servers, agent-instructions) get root-level merged files. [Cross-spec Consistency, Spec §FR-007, §FR-008]
- [ ] CHK009 - Is the `.gitignore` content specified precisely enough to implement (exact file contents: `*` and `!.gitignore`)? [Clarity, Spec §FR-009]
- [ ] CHK010 - Is the git detection method defined? "Within a git working tree" could mean checking for `.git/` in the project root or walking up the directory tree. [Clarity, Spec §FR-009]
- [ ] CHK011 - Are the merged file formats specified? `mcp.json` implies a JSON structure, but is the schema defined? `agent-instructions.md` implies Markdown, but are the package-identifying markers specified? [Gap, Spec §FR-008, §FR-011]
- [ ] CHK012 - Is the ownership model for merged files (overwritten on install/update, manual edits lost) stated clearly enough that the install command can document this in user-facing output? [Clarity, Spec §FR-011]

## Scoped Naming Convention Completeness

- [ ] CHK013 - Is the three-segment naming format (`scope.package-name.artifact-name`) documented with enough examples covering all artifact types (directory-based skills vs. file-based prompts/commands/agents)? [Completeness, Research §R-001]
- [ ] CHK014 - Is the parsing algorithm for three-segment names unambiguous? (Split on `.`, first = scope, last = artifact, middle = package.) Does the spec explicitly state that scope and package names cannot contain dots, making this split safe? [Clarity, Research §R-001, Cross-spec with spec/naming.md]
- [ ] CHK015 - For file-based artifacts, is the relationship between the scoped name and the file extension specified? E.g., is the installed file `scope.pkg.artifact.md` or `scope.pkg.artifact` with the original extension preserved? [Gap, Research §R-001]
- [ ] CHK016 - For directory-based artifacts (skills), is the internal structure of the installed directory specified? Is it a copy of the original `skills/artifact-name/` directory contents, or does the structure change? [Gap, Spec §FR-012]
- [ ] CHK017 - Is the naming convention consistent with the existing dot-notation in `spec/naming.md`? The current spec says `scope.artifact-name`; the plan extends to three segments. Is the update to `spec/naming.md` tracked as a deliverable? [Cross-spec Consistency, Research §R-001]

## Cross-Spec Consistency

- [ ] CHK018 - Does the project file `require` key pattern match the package manifest `require` key pattern exactly (same regex)? [Consistency, Spec §Assumption 2, spec/schema/package.json]
- [ ] CHK019 - Is the distinction between package `require` (bundled dependencies, resolved at pack time) and project `require` (installed dependencies, resolved at install time) clear enough that implementers won't confuse the two? [Clarity, Spec §Assumption 5]
- [ ] CHK020 - Are the artifact type names used in the install directory layout (`skills/`, `prompts/`, `commands/`, `agents/`) consistent with the `type` enum values in `spec/schema/package.json` (`skill`, `prompt`, `command`, `agent`)? Note the singular-vs-plural difference. [Cross-spec Consistency, Spec §FR-007]
- [ ] CHK021 - Is the mutual exclusivity of `aipkg.json` and `aipkg-project.json` documented in both directions? The 003 spec covers "init refuses when aipkg.json exists." Does it also address future commands that create `aipkg.json` refusing when `aipkg-project.json` exists? [Coverage, Spec §FR-002]

## Acceptance Criteria Quality

- [ ] CHK022 - Are all three error paths (re-init guard, mutual exclusivity guard, write permission failure) specified with enough detail to write deterministic assertions? [Measurability, Spec §US-2, §US-3, §Edge Cases]
- [ ] CHK023 - Is the exact content of the error message for the mutual exclusivity guard specified, or just the intent? FR-017 says "suggest `aipkg require` or `aipkg install`" but does not give the exact wording. Is this intentional flexibility or a gap? [Clarity, Spec §FR-017]
- [ ] CHK024 - Is the success output of `aipkg init` specified? (E.g., should it print a confirmation message like `create` does, or be silent?) [Gap, Spec §FR-015]

## Edge Case Coverage

- [ ] CHK025 - Is it specified what happens when both `aipkg.json` and `aipkg-project.json` exist simultaneously (e.g., created manually outside the CLI)? FR-002 says they "MUST NOT coexist" but only `init` enforces this. Is enforcement by other commands in scope for this spec? [Coverage, Spec §FR-002]
- [ ] CHK026 - Is behavior defined for when the directory becomes read-only between the existence check and the file write (TOCTOU)? The edge cases mention write permissions but not the race condition. [Edge Case, Gap]
- [ ] CHK027 - Is behavior defined for symlinked `aipkg-project.json` or `aipkg.json` files? `os.Stat` follows symlinks; is this the intended behavior for the existence checks? [Edge Case, Gap]

## Notes

- Focus: Spec quality for downstream implementers (especially the install command, AIPKG-10)
- Depth: Standard
- Audience: Implementer/reviewer at PR review time
- Cross-spec references checked: `spec/naming.md`, `spec/artifacts.md`, `spec/schema/package.json`, `spec/manifest.md`
