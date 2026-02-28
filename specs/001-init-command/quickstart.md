# Quickstart: aipkg init

**Feature**: 001-init-command | **Date**: 2026-02-28

## Create a Package Manifest (Interactive)

```sh
$ aipkg init
? Type: Package
? Name: @myorg/golang-expert
? Version (0.1.0): 1.0.0
? Description: Expert Go code reviewer and test writer
? License (Apache-2.0): Apache-2.0
Created aipkg.json (package)
```

Result:

```json
{
  "type": "package",
  "name": "@myorg/golang-expert",
  "version": "1.0.0",
  "description": "Expert Go code reviewer and test writer",
  "license": "Apache-2.0"
}
```

## Create a Project Manifest (Interactive)

```sh
$ aipkg init
? Type: Project
? Name (@scope/name, optional):
? Description (optional):
Created aipkg.json (project)
```

Result:

```json
{
  "type": "project"
}
```

## Non-Interactive (CI/Scripting)

```sh
# Package — all required fields via flags
aipkg init --type package --name @myorg/my-skill --version 0.1.0

# Project — type is the only relevant flag
aipkg init --type project

# Package with all fields
aipkg init --type package \
  --name @myorg/my-skill \
  --version 1.0.0 \
  --description "My skill" \
  --license MIT
```

## Hybrid Mode

Provide some flags, get prompted for the rest:

```sh
$ aipkg init --type package --name @myorg/my-skill
? Version (0.1.0):
? Description (optional):
? License (Apache-2.0):
Created aipkg.json (package)
```

## Error Cases

```sh
# File already exists
$ aipkg init
Error: aipkg.json already exists

# Invalid name
$ aipkg init --type package --name bad-name --version 1.0.0
Error: invalid name "bad-name": must match @scope/package-name format

# No TTY, missing required fields
$ echo | aipkg init --type package
Error: missing required fields: name, version (non-interactive mode)
```
