# Specification Quality Checklist: [FEATURE NAME]

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: [DATE]
**Feature**: [spec.md](../spec.md)

<!--
  ============================================================================
  REQUIREMENTS CHECKLIST TEMPLATE

  This checklist validates that a feature specification is ready for planning.
  All items are fixed (identical across every feature). The /speckit.checklist
  command should produce this checklist verbatim, only filling in the header
  fields and checking items that pass.

  Audience: Spec author, before running /speckit.plan
  Timing: After /speckit.specify and /speckit.clarify, before planning
  ============================================================================
-->

## Content Quality

- [ ] REQ001 No implementation details (languages, frameworks, APIs)
- [ ] REQ002 Focused on user value and business needs
- [ ] REQ003 Written for non-technical stakeholders
- [ ] REQ004 All mandatory sections completed

## Requirement Completeness

- [ ] REQ005 No [NEEDS CLARIFICATION] markers remain
- [ ] REQ006 Requirements are testable and unambiguous
- [ ] REQ007 Success criteria are measurable
- [ ] REQ008 Success criteria are technology-agnostic (no implementation details)
- [ ] REQ009 All acceptance scenarios are defined
- [ ] REQ010 Edge cases are identified
- [ ] REQ011 Scope is clearly bounded
- [ ] REQ012 Dependencies and assumptions identified

## Foundational Validity

- [ ] REQ013 Every FR builds on capabilities that exist in merged specs (`spec/`, merged feature specs) or implemented code, not on brainstorm notes, research sessions, or unimplemented design ideas
- [ ] REQ014 Schema fields referenced by FRs are either exercised by shipped code or defined in a merged spec that establishes them. Forward declarations (schema fields that exist but no merged spec or code has given meaning to) are not valid foundations.
- [ ] REQ015 Directory conventions, file formats, or protocols introduced by FRs either (a) are fully defined in this spec or (b) exist in a merged spec. Concepts from `features/research/` or unmerged branches are not valid foundations.
- [ ] REQ016 No user story depends on a capability that would need to be specced as a separate feature first (Principle I: if no merged spec defines the prerequisite, the requirement is premature)

## Feature Readiness

- [ ] REQ017 All functional requirements have clear acceptance criteria
- [ ] REQ018 User scenarios cover primary flows
- [ ] REQ019 Feature meets measurable outcomes defined in Success Criteria
- [ ] REQ020 No implementation details leak into specification

## Notes

- Check items off as completed: `[x]`
- All items must pass before proceeding to `/speckit.plan`
- If an item does not apply to this feature, mark it with a note explaining why
