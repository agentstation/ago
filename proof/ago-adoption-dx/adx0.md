# ADX0 baseline and research

Date: 2026-08-24

## Fail-before result

`proof/ago-adoption-dx/verify.sh` reports four failing adoption conditions:

```text
FAIL: README documents go get -tool
FAIL: README library example matches Check signature
FAIL: package docs do not claim the shipped command is a vet tool
FAIL: pkg.go.dev license files contain no pointer file
PASS: README contains no duplicated generics sentence
Summary: 4 failed
```

## Live publication state

- `https://proxy.golang.org/github.com/agentstation/ago/@v/list` returns
  `v0.1.0`.
- pkg.go.dev reports `UNKNOWN, Apache-2.0, MIT` and states that documentation
  is not displayed because of license restrictions.
- The pkgsite detector scans every recognized root license filename. It marks
  a file `UNKNOWN` when known license text covers less than 75 percent. Any
  non-redistributable type makes the module non-redistributable.
- The root `LICENSE` is a dual-license grant and pointer. `LICENSE-APACHE` and
  `LICENSE-MIT` contain the recognized texts. The pointer is the only unknown
  file.
- The GitHub license API reports `NOASSERTION` because it selects the pointer
  as the repository license.

Primary evidence:

- https://pkg.go.dev/license-policy
- https://github.com/golang/pkgsite/blob/master/internal/licenses/licenses.go
- https://pkg.go.dev/github.com/agentstation/ago?tab=licenses

## Go-native adoption

Go 1.24 added module tool dependencies. `go get -tool` records a command in
`go.mod`.

`go tool` runs the pinned command by its final path component. ago requires Go
1.25, so all supported adopters have this feature.

This path is stronger than a global installation for a repository policy:

- The version travels with the repository.
- The command needs no `PATH` setup.
- Local development, coding agents, and CI can run the same command.
- Standard Go module resolution and checksums apply.

Primary evidence:

- https://go.dev/doc/modules/managing-dependencies#tools
- https://go.dev/doc/modules/gomod-ref#tool
- https://go.dev/doc/go1.24

## Agent discovery

`AGENTS.md` is the cross-agent, always-on repository instruction file. An
Agent Skill is a task-specific workflow that compatible agents load when it
matches the task. GitHub CLI can install one skill into project or user scope
for Copilot, Claude Code, Cursor, Codex, Gemini CLI, and other agents.

The ago skill is optional. Enforcement stays in the repository through
`go.mod`, `.ago.yml`, `AGENTS.md`, and CI. The skill improves the remediation
loop but is not a policy boundary.

`llms.txt` is not in scope. Its stated contract is a map for an LLM-readable
website. This project has no separate documentation website, and repository
agents already have direct, stronger discovery artifacts.

Primary evidence:

- https://agents.md/
- https://docs.github.com/en/copilot/concepts/agents/about-agent-skills
- https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-skills
- https://llmstxt.org/

## Integration defects

- `README.md` calls `ago.Check(ctx, ago.Options{...})`, but `Check` accepts
  only `Options`.
- `ago.go` says the checks run through `go vet -vettool`, but the shipped
  `cmd/ago` binary does not implement the vet tool protocol. A driver built
  with `multichecker` can implement that protocol.
- Pull request 2 used the workflow-command property escape set for message
  data. Commit `0ff3ed9` corrected it and added the correct contract test.

Primary evidence:

- https://pkg.go.dev/golang.org/x/tools/go/analysis/unitchecker
- https://pkg.go.dev/golang.org/x/tools/go/analysis/checker
- https://github.com/actions/toolkit/blob/main/packages/core/src/command.ts
