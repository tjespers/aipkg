# Quickstart: Package Foundation

This quickstart validates that the `aipkg create` command works end-to-end after implementation.

## Prerequisites

- `aipkg` binary built: `task build`
- Empty working directory for testing

## Scenario 1: Create a package interactively

```bash
$ aipkg create @alice/blog-writer
```

Expected: interactive prompts for version (default 0.1.0), description, and license. On completion:

```bash
$ ls blog-writer/
agents/  agent-instructions/  aipkg.json  commands/  mcp-servers/  prompts/  skills/

$ cat blog-writer/aipkg.json
{
  "specVersion": 1,
  "name": "@alice/blog-writer",
  "version": "0.1.0"
}
```

## Scenario 2: Create a package non-interactively

```bash
$ aipkg create --name @alice/blog-writer --version 1.0.0 --description "AI blog writing" --license MIT
```

Expected: no prompts, package created immediately.

```bash
$ cat blog-writer/aipkg.json
{
  "specVersion": 1,
  "name": "@alice/blog-writer",
  "version": "1.0.0",
  "description": "AI blog writing",
  "license": "MIT"
}
```

## Scenario 3: Create in an existing directory

```bash
$ mkdir my-pkg && echo "# My Package" > my-pkg/README.md
$ aipkg create @alice/my-pkg --path ./my-pkg
```

Expected: README.md preserved, well-known directories added, aipkg.json created.

```bash
$ ls my-pkg/
agents/  agent-instructions/  aipkg.json  commands/  mcp-servers/  prompts/  README.md  skills/
```

## Scenario 4: Reject existing package

```bash
$ aipkg create @alice/blog-writer
# (completes successfully)
$ aipkg create @alice/blog-writer
Error: target directory already contains aipkg.json
```

## Scenario 5: License detection

```bash
$ mkdir licensed-pkg
$ cp /path/to/apache-2.0-license licensed-pkg/LICENSE
$ aipkg create @alice/licensed-pkg --path ./licensed-pkg
```

Expected: license prompt shows `Apache-2.0` as the default value.

## Scenario 6: Invalid name validation

```bash
$ aipkg create blog-writer
Error: package name must be scoped (e.g., @scope/package-name)

$ aipkg create @aipkg/my-tool
Error: scope "aipkg" is reserved
```

## Scenario 7: Non-interactive without TTY

```bash
$ echo "" | aipkg create --name @alice/blog-writer
Error: missing required flags for non-interactive mode: --version
```

## Validation checklist

- [ ] Generated aipkg.json passes `spec/schema/package.json` validation
- [ ] All six well-known directories are created
- [ ] Existing files in --path target are preserved
- [ ] Ctrl+C during prompts leaves no files behind
- [ ] Reserved scope names are rejected
- [ ] Invalid semver is rejected with helpful error
- [ ] --path to existing aipkg.json is rejected
- [ ] No-TTY mode exits cleanly with missing flags listed
