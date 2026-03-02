# Naming & Namespaces

Every package in the aipkg ecosystem uses scoped names. There are no unscoped packages. This document defines the naming rules for packages, namespaces, and installed artifacts.

## Package names

A package name always takes the form `@scope/package-name`. The `@` prefix makes package identities visually distinct and unambiguous to parse. The CLI uses it to distinguish package names (`@shiftbase/my-skill`) from source locators (`shiftbase-com/my-skill`).

```text
@scope/package-name
│     │
│     └── The package name within the scope
└── The namespace (organization, individual, or reserved prefix)
```

### Scope rules

- Lowercase alphanumeric characters and hyphens only (`a-z`, `0-9`, `-`)
- 1 to 39 characters
- Cannot start or end with a hyphen
- No consecutive hyphens (`--`)
- No dots or underscores

These rules mirror GitHub's username/organization constraints, so `@my-org/my-package` maps to GitHub without transformation.

### Package name rules

- Lowercase alphanumeric characters and hyphens only (`a-z`, `0-9`, `-`)
- 1 to 64 characters
- Cannot start or end with a hyphen
- No consecutive hyphens (`--`)
- No dots or underscores

Dots are reserved for installed artifact naming (see [dot-notation](#dot-notation) below). Underscores are excluded to prevent confusion between `my_package` and `my-package`.

### Full name format

The complete package name, including the `@` and `/`, is written as a single string:

```text
@alice/blog-writer
@tjespers/golang-expert
@my-org/code-review
```

The name in the manifest is authoritative. If a package is hosted at `github.com/some-other-org/repo`, the `name` field in `aipkg.json` still determines the package's identity. See [manifest.md](manifest.md#name) for details.

### Case sensitivity

All names are case-insensitive but must be stored and published in lowercase. The CLI normalizes input to lowercase before validation. `@Alice/Blog-Writer` is treated as `@alice/blog-writer`.

### Validation regex

A valid package name matches:

```text
^@(?!.*--)[a-z0-9]([a-z0-9-]{0,37}[a-z0-9])?/[a-z0-9]([a-z0-9-]{0,62}[a-z0-9])?$
```

The `(?!.*--)` lookahead rejects consecutive hyphens anywhere in the name. This allows single-character scopes and package names (e.g., `@a/b`) while enforcing all the rules above.

## Artifact names

Each artifact within a package also has a name. Artifact names follow the same character rules as package names:

- Lowercase alphanumeric and hyphens only
- 1 to 64 characters
- No consecutive hyphens, can't start or end with a hyphen
- Must be unique within the package for a given artifact type

## Dot-notation

When artifacts are installed, they're placed using **dot-notation**: `scope.artifact-name`. The `@` and `/` from the package name don't appear in installed file or directory names.

This is the adapter's responsibility. The pattern is:

```text
scope.artifact-name
```

### Examples

Given a package `@shiftbase/golang-expert` with an artifact named `test-writer`:

```text
Claude Code:  .claude/skills/shiftbase.test-writer/
Cursor:       .cursor/rules/shiftbase.test-writer.md
Windsurf:     .windsurf/rules/shiftbase.test-writer.md
```

The dot separates the scope from the artifact name, preventing collisions between packages from different namespaces. Two packages can both have a `test-writer` artifact, but `shiftbase.test-writer` and `alice.test-writer` are distinct.

Dot-notation also reads naturally in tool UIs. A slash command from `@shiftbase/deploy-tools` with a `deploy` command artifact becomes `/shiftbase.deploy`.

### Why dots?

The dot is reserved as a namespace separator at the tool level. This is why dots are forbidden in scope names and package names: allowing them would create ambiguity (`is.this.a.scope.or.artifact?`).

## Reserved scopes

Certain scopes are reserved to prevent squatting on well-known names. This includes project-owned prefixes (`@aipkg*`, `@ai-interop*`), generic ecosystem terms (`@official`, `@core`, `@std`, etc.), and scopes for well-known companies and brands (AI providers, coding tools, platforms) so those entities can claim them when they choose to publish packages.

The CLI rejects packages that use a reserved scope. The full list is maintained in [`reserved-scopes.txt`](../reserved-scopes.txt) at the repo root, so both humans and the CLI can consume it directly.

## Summary table

| Component     | Characters            | Length | Additional rules                                                        |
| ------------- | --------------------- | ------ | ----------------------------------------------------------------------- |
| Scope         | `a-z`, `0-9`, `-`     | 1-39   | No leading/trailing/consecutive hyphens                                 |
| Package name  | `a-z`, `0-9`, `-`     | 1-64   | No leading/trailing/consecutive hyphens                                 |
| Artifact name | `a-z`, `0-9`, `-`     | 1-64   | No leading/trailing/consecutive hyphens, unique per type within package |
| Dot-notation  | `scope.artifact-name` | N/A    | Generated by adapters at install time                                   |
