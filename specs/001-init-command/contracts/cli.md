# CLI Contract: aipkg init

**Feature**: 001-init-command | **Date**: 2026-02-28

## Command Signature

```
aipkg init [flags]
```

## Flags

| Flag | Type | Default | Description |
| ---- | ---- | ------- | ----------- |
| `--type` | string | (none) | Manifest type: `project` or `package` |
| `--name` | string | (none) | Scoped package/project name (`@scope/name`) |
| `--version` | string | (none) | Package version (semver `MAJOR.MINOR.PATCH`) |
| `--description` | string | (none) | Short description (max 255 chars) |
| `--license` | string | (none) | SPDX license identifier |

All flags are optional. Missing required values trigger interactive prompts (if TTY available).

## Behavior Modes

| Condition | Behavior |
| --------- | -------- |
| No flags, TTY available | Fully interactive — prompts for all fields |
| All required flags provided | Fully non-interactive — no prompts, writes file |
| Some flags provided, TTY available | Hybrid — prompts only for missing required fields |
| Missing required flags, no TTY | Error with list of missing fields (exit 1) |

## Interactive Prompt Order

1. Type selection (project / package) — skipped if `--type` provided
2. **Package flow**: name → version (default: `0.1.0`) → description → license (default: detected from LICENSE file)
3. **Project flow**: name → description

Optional fields can be skipped by pressing Enter (empty input = omit from manifest).

## Exit Codes

| Code | Meaning |
| ---- | ------- |
| 0 | Success — `aipkg.json` created |
| 1 | Error — validation failure, file exists, filesystem error, missing required input, user abort |

## Output

**stdout** (success):

```
Created aipkg.json (package)
```

or

```
Created aipkg.json (project)
```

**stderr** (errors):

```
Error: aipkg.json already exists
```

```
Error: invalid name "bad-name": must match @scope/package-name format
```

**stderr** (warnings):

```
Warning: --version is ignored for project type
```

## Generated File

**Path**: `./aipkg.json` (current working directory)

**Format**: JSON, 2-space indentation, trailing newline.

**Package example**:

```json
{
  "type": "package",
  "name": "@myorg/cool-skill",
  "version": "1.0.0",
  "description": "A cool skill",
  "license": "Apache-2.0"
}
```

**Minimal project example**:

```json
{
  "type": "project"
}
```

## Guards

- Refuses if `aipkg.json` already exists (regardless of flags)
- Refuses if Ctrl+C / user abort — no file written
- Refuses if non-writable directory — filesystem error
