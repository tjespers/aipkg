# Post-Implementation Checklist: [FEATURE NAME]

**Purpose**: Verify implementation correctness, cross-spec consistency, and documentation completeness before opening the implementation PR. Includes deferred item review from the pre-implementation audit.
**Created**: [DATE]
**Feature**: [spec.md](../spec.md)

<!--
  ============================================================================
  POST-IMPLEMENTATION CHECKLIST TEMPLATE

  This checklist verifies that the implementation meets the spec bar.
  It has two parts:

  1. VARIABLE sections (top, optional): Domain-specific correctness checks
     generated from the feature's spec. These change per feature. The
     generator should create 0-2 sections for feature-specific verification.

  2. FIXED sections (bottom): Recurring categories that appear in every
     post-implementation checklist. These are always present.

  Items use POST### IDs, numbered sequentially across all sections.

  Audience: Implementer, for self-verification before opening the PR
  Timing: After implementation is complete, before PR creation

  RESOLUTION FORMAT:
  When checking off items during verification:

    Pass:
    - [x] POST### Item text

    Pass with note:
    - [x] POST### Item text
      - Brief note about how it was verified or any nuance.

    Deferred item resolution:
    - [x] POST### **Item title** (PRE###): Question about deferral status
      - **Still appropriate.** Explanation of why deferral stands.
      OR
      - **Needs action.** Created AIPKG-XX to address this.
  ============================================================================
-->

<!--
  VARIABLE SECTIONS — Domain-Specific (0-2 sections)

  Generate these when the feature has domain-specific correctness concerns
  that don't fit neatly into the fixed categories below.

  Examples from previous features:

  003 (Project Initialization):
    - Schema Correctness (Draft 2020-12, additionalProperties, regex patterns)

  002 (Archive Format & Pack):
    - (No variable sections; all items fit in fixed categories)

  Only create these sections when there's a clear domain that warrants
  its own grouping. Prefer using the fixed sections when items fit.
-->

## Implementation Correctness

<!--
  Verify that functional requirements are correctly implemented.
  Each item should reference at least one FR and be verifiable by
  running the command or reading the code.

  Items should check for:
    - Command behavior matches FR descriptions
    - Output format (JSON indent, trailing newline, file permissions)
    - Success/error messages follow CLI patterns
    - Guard behavior (check order, existence-only vs content validation)
    - No extra files/directories created beyond what's specified
-->

- [ ] POST001 [Implementation correctness items generated from feature spec]

## Documentation Completeness

<!--
  Verify that reference documentation ships with the feature.
  Every feature that introduces user-facing concepts needs spec docs.

  Items should check for:
    - spec/*.md documents cover all new concepts
    - Field definitions with examples (empty and populated)
    - Directory layouts and file structures documented
    - Naming conventions documented with parsing rules
    - Ownership models and lifecycle behavior documented
    - Bidirectional rules stated in both directions
-->

- [ ] POST0XX [Documentation items generated from feature spec]

## Cross-Spec Consistency

<!--
  Verify that the implementation doesn't drift from existing specs.
  Enumerate spec files by scanning spec/*.md and spec/schema/*.json at
  generation time. Do not rely on a hardcoded list.

  Items should check for:
    - Enum values in code match schema definitions
    - Directory names match artifact type conventions
    - Field definitions consistent across schemas
    - Shared patterns (regex, naming) identical where required
-->

- [ ] POST0XX [Cross-spec items generated from feature context]

## Deferred Items Review

<!--
  CRITICAL: This section carries forward items marked DEFERRED in the
  pre-implementation checklist. Every DEFERRED PRE item MUST appear here.

  For each deferred item:
  1. State the original concern with its PRE ID
  2. Ask whether implementation revealed a need to address it now
  3. Resolution is either "Still appropriate" or "Needs action" with a
     Linear issue ID

  Format:
  - [ ] POST0XX **Original concern title** (PRE###): Did implementation
    or documentation work reveal that this needs definition now?

  If the pre-implementation checklist has no DEFERRED items, include this
  section with a note: "No items were deferred during pre-implementation."
-->

- [ ] POST0XX [Deferred items carried forward from pre-implementation checklist]

## Test Coverage

<!--
  Verify that tests cover the spec's acceptance scenarios and edge cases.

  Items should check for:
    - All user story acceptance scenarios have corresponding tests
    - Edge cases from spec.md are covered
    - Schema validation tests (valid passes, invalid fails)
    - Roundtrip tests where applicable (create/load/verify)
    - Error path tests (guards, validation failures)
-->

- [ ] POST0XX [Test coverage items generated from feature spec]

## Notes

- Items POST0XX-POST0XX originate from the pre-implementation checklist (PRE###, PRE###). Resolution notes should explain whether deferral is still appropriate.
- If any deferred item needs action, create a Linear issue and note the issue ID next to the checklist item.
- Cross-spec consistency items ensure the implementation doesn't drift from the existing specification foundation.
