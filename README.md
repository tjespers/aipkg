# aipkg

The unified AI package manager.

aipkg is a package manager for AI artifacts: skills, prompts, commands, agents, and MCP server configs. Think npm or Composer, but for AI tooling.

## Key concepts

- **Scoped naming**: all packages use `@scope/package-name`, no exceptions
- **Bundles**: a single package can contain multiple artifact types
- **Adapters**: tool-agnostic core with thin adapter layers for Claude Code, Cursor, Windsurf, and others
- **Virtual packages**: community-maintained recipes that wrap existing repos as installable packages
- **GitHub-first**: v1 sources packages from GitHub Releases, central registry comes later

## Status

Early development. The specification and CLI are being built in parallel. Not yet usable.

## Development

Requires Go 1.25+, [Task](https://taskfile.dev), and [golangci-lint](https://golangci-lint.run) v2.

```sh
task build          # build binary to dist/
task test           # run tests
task lint           # run golangci-lint
task check          # lint + vet + test (full check)
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full development setup.

## Related repositories

| Repository                                             | Purpose                                             |
| ------------------------------------------------------ | --------------------------------------------------- |
| [aipkg-spec](https://github.com/ai-interop/aipkg-spec) | Manifest schema, naming rules, artifact conventions |
| [aipkg](https://github.com/ai-interop/aipkg)           | CLI tool (this repo)                                |

## Disclaimer

aipkg is a package manager and distribution tool. It does not review, audit, or endorse the content of packages installed through it. AI artifacts such as skills, prompts, and agent configs can contain instructions that influence AI tool behavior in unintended ways. You are responsible for reviewing any packages you install. See [SECURITY.md](SECURITY.md) for more details.

## License

[Apache-2.0](LICENSE)
