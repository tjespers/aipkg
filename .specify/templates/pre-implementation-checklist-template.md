# Pre-Implementation Checklist: [FEATURE NAME]

**Purpose**: Validate that the specification and plan are precise enough to implement without ambiguity. Covers domain-specific requirements quality and recurring cross-cutting concerns.
**Created**: [DATE]
**Feature**: [spec.md](../spec.md)

<!--
  ============================================================================
  PRE-IMPLEMENTATION CHECKLIST TEMPLATE

  This checklist audits spec + plan quality before implementation begins.
  It has two parts:

  1. VARIABLE sections (top): Domain-specific categories generated from the
     feature's spec and plan. These change per feature. The generator should
     create 2-4 sections based on the feature's primary domains.

  2. FIXED sections (bottom): Recurring categories that appear in every
     pre-implementation checklist. These are always present.

  Items use PRE### IDs, numbered sequentially across all sections.

  Audience: Implementer or reviewer, before coding starts
  Timing: After plan.md and tasks.md exist, before implementation

  RESOLUTION FORMAT:
  When resolving items, use one of these patterns:

    Resolved:
    - [x] PRE### Item text
      > **Resolved**: Brief explanation of how it was addressed.

    Deferred:
    - [x] PRE### Item text — DEFERRED: reason, does not block this feature

    N/A:
    - [x] PRE### Item text — N/A: brief explanation
  ============================================================================
-->

<!--
  VARIABLE SECTIONS — Domain-Specific (2-4 sections)

  Generate these based on the feature's primary domains. Each section should
  contain 3-8 items checking requirements quality in that domain.

  Examples from previous features:

  002 (Archive Format & Pack):
    - Archive Format Specification Completeness
    - Artifact Discovery & Name Derivation Clarity
    - Type-Specific Validation Completeness
    - File Exclusion Requirements
    - Pipeline & Error Reporting

  003 (Project Initialization):
    - Project File Schema Completeness
    - Install Directory Layout Completeness
    - Scoped Naming Convention Completeness

  Item format:
  - [ ] PRE### Question about requirements quality [Quality Dimension, Spec §FR-XXX]

  Quality dimensions: Completeness, Clarity, Consistency, Coverage, Gap,
  Measurability, Ambiguity, Conflict, Assumption
-->

## [Domain-Specific Section 1]

- [ ] PRE001 [Generated from feature spec/plan]

## [Domain-Specific Section 2]

- [ ] PRE00N [Generated from feature spec/plan]

<!--
  FIXED SECTIONS — Always Present

  The sections below appear in every pre-implementation checklist.
  The generator fills in concrete items based on the feature, but the
  category structure is stable.
-->

## Foundational Validity

<!--
  CRITICAL: This section catches Principle I violations before implementation.

  For each FR that references an external concept (directory convention,
  schema field, file format, behavior from another feature), trace the
  concept back to its source and verify it exists in shipped artifacts:

  VALID foundations:
    - spec/*.md documents (shipped reference documentation)
    - spec/schema/*.json schemas that are defined by a merged spec
    - Merged feature specs (features/NNN/spec.md on main, even if
      not yet implemented as code)
    - Implemented and tested code in internal/

  INVALID foundations (must be flagged):
    - features/research/ brainstorm sessions
    - Unexercised schema fields with no merged spec giving them meaning
    - Design decisions from unmerged branches or draft specs
    - "Future" concepts mentioned only in assumptions or roadmap items

  If an FR builds on an invalid foundation, it is premature and should
  be removed from this spec or deferred until the foundation is merged.

  Incident context: 004-package-install originally specced bundled
  dependency extraction building on a deps/ directory concept that only
  existed in brainstorm notes. Four separate reviews missed it because
  they checked internal consistency but not foundational validity.
-->

- [ ] PRE0XX [Foundational validity items generated from feature context]

## Cross-Spec Consistency

<!--
  Check alignment between this feature's spec/plan and existing spec documents.
  Enumerate spec files by scanning spec/*.md and spec/schema/*.json at
  generation time. Do not rely on a hardcoded list.

  Items should check for:
    - Same concept defined consistently across specs
    - Enum values, regex patterns, or field names that must match
    - Terminology alignment (singular vs plural, naming conventions)
    - Bidirectional rules stated in both directions
-->

- [ ] PRE0XX [Cross-spec items generated from feature context]

## Acceptance Criteria Quality

<!--
  Check that acceptance criteria and error paths are specified precisely
  enough to write deterministic test assertions.

  Items should check for:
    - Error messages specified with enough detail for assertions
    - Success output defined (silent vs confirmation message)
    - All error paths (guards, validation failures) have clear behavior
    - Intentional flexibility vs gaps in message wording
-->

- [ ] PRE0XX [Acceptance criteria items generated from feature context]

## Edge Case Coverage

<!--
  Check that boundary conditions and unusual scenarios are addressed
  in requirements.

  Common edge cases to check (select applicable ones):
    - File system: permissions, symlinks, TOCTOU races, disk full
    - Input: empty, missing, malformed, conflicting
    - State: concurrent access, partial failure, interrupted operations
    - Platform: cross-platform behavior differences
-->

- [ ] PRE0XX [Edge case items generated from feature context]

## Plan Review Findings

<!--
  Check plan.md for scope alignment and architectural decisions.

  Items should check for:
    - Deliverables with no backing FR or acceptance scenario (scope creep risk)
    - Design decisions that need explicit rationale
    - Dependencies or deliverables that should be tracked
    - Consistency between plan's project structure and spec's requirements
-->

- [ ] PRE0XX [Plan review items generated from feature context]

## Notes

- Focus: Spec + plan quality for downstream implementers
- Cross-spec references checked: [list spec files reviewed]
- Items marked DEFERRED must appear in the post-implementation checklist's "Deferred Items Review" section with PRE→POST cross-references
