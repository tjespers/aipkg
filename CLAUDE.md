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

## Spec dependency

The manifest schema, naming rules, and artifact types are defined in `ai-interop/aipkg-spec`. The JSON schemas in that repo are the source of truth for validation.

## Project management

Work is tracked in Linear, team AIPKG, project "CLI".

## Commit conventions

- Conventional commits (enforced by pre-commit hook)
- Use `git commit -s` for DCO sign-off (CNCF requirement)
- Use `Closes: AIPKG-XX` trailer to link commits to Linear issues

## Active Technologies

- Go 1.25.7 (pinned in go.mod) (001-init-command)
- Filesystem only (reads/writes `aipkg.json` in cwd) (001-init-command)

## Recent Changes

- 001-init-command: Added Go 1.25.7 (pinned in go.mod)
