# Quickstart: Archive Format & Pack Command

**Branch**: `002-archive-format-pack` | **Date**: 2026-03-01

This quickstart validates the end-to-end pack workflow from "I have a package directory" to "I have a distributable archive."

## Prerequisites

- `aipkg` binary built from this branch
- A valid package directory (created with `aipkg create` or manually)

## Step 1: Create a test package

```bash
mkdir -p /tmp/pack-test && cd /tmp/pack-test
aipkg create @demo/hello-world
cd hello-world
```

This gives you a directory with `aipkg.json` and empty well-known directories.

## Step 2: Add a skill artifact

```bash
mkdir -p skills/greeting
cat > skills/greeting/SKILL.md << 'EOF'
---
name: greeting
description: A friendly greeting skill for testing the pack command
---

# Greeting Skill

Say hello to the user in a friendly way.
EOF
```

## Step 3: Add a prompt artifact

```bash
cat > prompts/review.md << 'EOF'
Review the provided code for correctness, style, and potential issues.
Provide specific suggestions with line references.
EOF
```

## Step 4: Pack the package

```bash
aipkg pack
```

**Expected output** (stderr):

```
demo--hello-world-0.1.0.aipkg (2 artifacts, ~1 KB)
```

**Expected files created**:

```
demo--hello-world-0.1.0.aipkg
demo--hello-world-0.1.0.aipkg.sha256
```

## Step 5: Verify the archive

```bash
# Check integrity
sha256sum -c demo--hello-world-0.1.0.aipkg.sha256

# Inspect contents
unzip -l demo--hello-world-0.1.0.aipkg
```

**Expected archive contents**:

```
hello-world/aipkg.json
hello-world/skills/greeting/SKILL.md
hello-world/prompts/review.md
```

## Step 6: Verify the manifest in the archive

```bash
unzip -p demo--hello-world-0.1.0.aipkg hello-world/aipkg.json | python3 -m json.tool
```

**Expected**: The `aipkg.json` inside the archive includes the `artifacts` array:

```json
{
  "specVersion": 1,
  "name": "@demo/hello-world",
  "version": "0.1.0",
  "artifacts": [
    { "name": "greeting", "type": "skill", "path": "skills/greeting/" },
    { "name": "review", "type": "prompt", "path": "prompts/review.md" }
  ]
}
```

## Step 7: Verify the original manifest is unchanged

```bash
cat aipkg.json
```

**Expected**: No `artifacts` field. The pack command does not modify the original file.

## Step 8: Test validation (negative case)

```bash
# Create an invalid skill (missing SKILL.md)
mkdir -p skills/broken

aipkg pack
```

**Expected**: Validation error, no archive produced:

```
pack: 1 validation error
skills/broken/: missing required SKILL.md file
Error: pack failed
```

## Step 9: Test ignore file

```bash
# Remove the broken skill, add an ignore file
rmdir skills/broken
echo "*.log" > .aipkgignore
echo "temp output" > build.log

aipkg pack
```

**Expected**: Archive produced. `build.log` is excluded. `.aipkgignore` itself is excluded by built-in defaults.

## Step 10: Custom output location

```bash
mkdir -p dist
aipkg pack --output dist/
ls dist/
```

**Expected**: Archive and sidecar written to `dist/`.
