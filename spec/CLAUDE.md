# CLAUDE.md — aipkg specification

## What this directory is

The aipkg specification. Reference documentation and JSON schemas that define the package format, naming rules, artifact types, and interface contracts for the aipkg ecosystem.

## Specification style

Specs are **reference documentation** (like Composer's schema docs), not formal W3C-style specifications.

- Plain language, field-by-field definitions, concrete examples
- No RFC 2119 keywords — use "required", "optional", "must be" in plain English
- JSON Schema files provide machine-readable validation where applicable
- Formalize later if CNCF submission or multi-implementer needs arise

## Structure

- Spec documents (`.md`) live directly in this directory
- `schema/` contains JSON Schema files for manifest validation, recipe format, etc.

## Project management

Work is tracked in Linear, team AIPKG, project "Specification".
