# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in aipkg, please report it responsibly.

**Do not open a public GitHub issue for security vulnerabilities.**

Instead, please email the maintainers directly. We will acknowledge your report within 48 hours and work with you to understand and address the issue.

## Scope

This policy applies to the `aipkg` CLI tool and its direct dependencies. For vulnerabilities in upstream packages managed by aipkg, please report those to the respective upstream maintainers.

## Package content risks

aipkg manages AI artifacts (skills, prompts, commands, agents, MCP server configs). Unlike traditional package managers where the primary risk is malicious code execution, AI artifacts carry additional risks:

- **Prompt injection**: a skill or prompt could contain instructions designed to override AI tool safety guardrails, exfiltrate context, or trigger unintended actions
- **Malicious agent configs**: agent definitions or MCP server configs could point to untrusted endpoints or request excessive permissions
- **Supply chain attacks**: as with any package ecosystem, a compromised or typosquatted package could deliver harmful content

aipkg provides the distribution mechanism but does not review, audit, or endorse package content. Users should review the artifacts they install, especially skills and agent configs that directly influence AI tool behavior.

## Disclosure

We follow coordinated disclosure. Once a fix is available, we will publish a security advisory through GitHub.
