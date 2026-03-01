# CLAUDE.md -- aipkg CLI

## What this repo is

The `aipkg` command-line tool. A package manager for AI artifacts (skills, prompts, commands, agents, MCP server configs).

## Tech stack

- **Language**: Go
- **CLI framework**: cobra + viper
- **Interactive prompts**: bubbletea / huh
- **Task runner**: Taskfile (`task build`, `task test`, `task lint`, `task check`)
- **Linting**: golangci-lint (config in `.golangci.yml`)
- **Releases**: goreleaser (config in `.goreleaser.yml`)

## Project structure

```
cmd/aipkg/       # main entry point
internal/        # private packages (not importable by other modules)
spec/            # specification (manifest schema, naming rules, artifact types)
```

No `/pkg` directory. This is a CLI, not a library. Everything is in `internal/`.

## Development

```sh
task build          # build binary to dist/
task test           # run tests
task lint           # run golangci-lint
task check          # lint + vet + test (full check)
task fmt            # format code
task tidy           # go mod tidy
```

## Specification

The manifest schema, naming rules, and artifact types live in `spec/`. The JSON schemas there are the source of truth for validation. See `spec/CLAUDE.md` for spec-specific context.

## Project management

Work is tracked in Linear, team AIPKG, project "CLI".

## Commit conventions

- Conventional commits (enforced by pre-commit hook)
- Use `git commit -s` for DCO sign-off (CNCF requirement)
- Use `Closes: AIPKG-XX` trailer to link commits to Linear issues

## Active Technologies

- Go 1.25 (per go.mod) + cobra (CLI routing), gopkg.in/yaml.v3 (SKILL.md frontmatter), sabhiram/go-gitignore (.aipkgignore patterns), santhosh-tekuri/jsonschema/v6 (manifest schema validation). Archive and SHA-256 via stdlib (`archive/zip`, `crypto/sha256`). (002-archive-format-pack)

- Go 1.25 (per go.mod) + cobra (CLI routing), huh/bubbletea (interactive prompts), santhosh-tekuri/jsonschema/v6 (schema validation), google/licensecheck (LICENSE file detection), dlclark/regexp2 (PCRE-style regex for name validation lookaheads), x/term (TTY detection) (001-package-foundation)

- N/A (local filesystem only, no database) (001-package-foundation)

## Recent Changes

- 001-package-foundation: Added Go 1.25 (per go.mod) + cobra (CLI routing), huh/bubbletea (interactive prompts), santhosh-tekuri/jsonschema/v6 (schema validation), google/licensecheck (LICENSE file detection), dlclark/regexp2 (PCRE-style regex for name validation lookaheads), x/term (TTY detection)
