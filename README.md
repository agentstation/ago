# ago

[![CI](https://github.com/agentstation/ago/actions/workflows/ci.yml/badge.svg)](https://github.com/agentstation/ago/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/agentstation/ago.svg)](https://pkg.go.dev/github.com/agentstation/ago)
[![Go Report Card](https://goreportcard.com/badge/github.com/agentstation/ago)](https://goreportcard.com/report/github.com/agentstation/ago)
[![License](https://img.shields.io/badge/license-MIT%20OR%20Apache--2.0-blue)](#license)

A restriction-only linter for Go.

Go permits several forms for some operations. ago lets a project select
accepted forms and gives developers, coding agents, and CI the same executable
policy.

`ago` rejects selected legal Go constructs. It does not add syntax, rewrite
code, or change semantics. Code that passes `ago` is ordinary Go that builds
with the stock toolchain.

```console
$ go tool ago ./...
internal/store/index.go:42:2: naked return in indexAll; name the values you are returning (no-naked-return)
internal/store/index.go:88:9: new() takes a type, not an expression (no-new-expr)
2 violations
```

The name reads as *a-go*, the Go subset an agent may write. It also reads as
*ago*, the Go that earlier toolchains accepted. Read the [design
case](docs/design.md) for the project boundary and evidence.

## Adopt ago in a Go repository

ago requires Go 1.25 or later and a Go module.

1. Add ago as a module tool dependency.

   ```sh
   go get -tool github.com/agentstation/ago/cmd/ago@latest
   ```

   This command records an exact ago version in `go.mod` and `go.sum`.

2. Write the starter rule policy.

   ```sh
   go tool ago -init
   ```

   This command creates `.ago.yml`. Edit the file, then commit it with
   `go.mod` and `go.sum`.

3. Check the module.

   ```sh
   go tool ago ./...
   ```

   A clean run prints nothing and exits with status 0.

The Go module now owns the ago version. Developers, coding agents, and CI can
run `go tool ago` without a global installation or a `PATH` change.

### Other installation methods

Install a global command when one pinned repository does not own the use:

```sh
go install github.com/agentstation/ago/cmd/ago@latest
```

On macOS or Linux with Homebrew:

```sh
brew install agentstation/tap/ago
```

Release archives, checksums, and software bills of materials are available on
the [release page](https://github.com/agentstation/ago/releases).

## Make the policy automatic

Four repository files make the policy repeatable:

| File | Contract |
| --- | --- |
| `go.mod` | Pins the ago command version. |
| `.ago.yml` | Selects the rule policy. |
| `AGENTS.md` | Tells coding agents when and how to run the policy. |
| CI workflow | Rejects a change that violates the policy. |

Add this instruction to the adopting repository's `AGENTS.md`:

```markdown
Run `go tool ago -stale-ignores -format json ./...` after each Go change.
Fix findings in source. Do not change `.ago.yml` or add a suppression only to
make the run pass. Exit status 2 means the check was incomplete.
```

Use the same pinned tool in GitHub Actions:

```yaml
- uses: actions/setup-go@v7
  with:
    go-version: stable
- run: go tool ago -format github ./...
```

CI remains the policy boundary. Agent instructions and the optional
[ago Agent Skill](#optional-agent-skill) improve the local repair loop.

## Run ago

```sh
go tool ago ./...                    # default rule set, current module
go tool ago -list                    # show every rule and which are on
go tool ago -explain no-goto         # print one complete rationale
go tool ago -all ./...               # run every rule
go tool ago -tests ./...             # include _test.go files
go tool ago -stale-ignores ./...     # report unused suppressions
```

Package arguments are [`go/packages`](https://pkg.go.dev/golang.org/x/tools/go/packages)
patterns. With no arguments, ago checks `./...`.

ago always skips `vendor/` and `testdata/`. Third-party code is not yours to
restrict.

| Exit status | Meaning |
| --- | --- |
| `0` | The run completed with no findings or stale ignores. |
| `1` | The run found a rule violation or stale ignore. |
| `2` | ago could not complete a meaningful run. |

## Configure the rule policy

ago reads `.ago.yml` from the working directory or the nearest parent.

```yaml
enable:
  - default
  - no-init-func

disable:
  - no-goto

tests: false
exclude:
  - "*.pb.go"
  - third_party/*
```

`enable` accepts rule names and the meta-names `default` and `all`. `disable`
wins over `enable`. Command flags `-rules` and `-all` override the file.

ago matches each `exclude` pattern against three path shapes:

- the complete slash-separated path.
- each path element.
- each leading path prefix.

Thus, `*.pb.go` matches a file name, `generated` matches that directory at any
depth, and `third_party/*` matches that subtree.

Unknown keys and unknown rule names stop the run. A policy typo cannot disable
a rule silently. Use `-config path` to name a file. Use `-no-config` to ignore
all policy files.

## Fix or suppress a finding

Fix source code when the selected policy applies. Each finding includes the
rule name. Run `go tool ago -explain <rule>` for the full rationale and rule
boundary.

Use a suppression only when the local construct is a justified exception:

```go
//ago:ignore no-goto -- hand-written state machine, see docs/parser.md
goto retry
```

The directive applies to the next line. A top-level `//ago:ignore-file`
directive applies to its file. Both forms accept a comma-separated rule list
or `*`.

Every suppression must name a known rule and include a `--` reason. An invalid
directive suppresses nothing, and `no-invalid-ignore` reports it. Run with
`-stale-ignores` to find a suppression that no longer covers a finding.

## Machine contract for coding agents

ago exposes policy and results as stable data. A coding agent does not need to
parse this README.

Discover the active rules:

```sh
go tool ago -list -format json
```

Each catalogue entry includes its name, analyzer ident, summary, rationale,
default and active state, Go release boundary, severity, and documentation URL.

Read findings and incomplete-run errors:

```sh
go tool ago -stale-ignores -format json ./...
```

```json
{
  "version": "v0.1.1",
  "rules": ["no-dot-import", "no-goto", "no-naked-return"],
  "findings": [
    {
      "rule": "no-naked-return",
      "severity": "error",
      "message": "naked return in indexAll; name the values you are returning",
      "file": "internal/store/index.go",
      "line": 42,
      "column": 2,
      "endLine": 42,
      "endColumn": 8,
      "docURL": "https://github.com/agentstation/ago/blob/main/docs/rules.md#no-naked-return"
    }
  ],
  "staleIgnores": [],
  "errors": []
}
```

ago sorts and deduplicates findings. The same version, policy, and source tree
produce identical JSON bytes. Existing JSON fields keep their names and
meanings. Later versions can add fields.

A package load or parse failure appears in `errors`. ago continues with each
package that it can analyze. Exit status 2 means that no usable result was
available, so an empty finding list is not a clean result.

### Optional Agent Skill

The optional skill teaches compatible coding agents the discovery, repair,
suppression, and verification loop:

```sh
gh skill install agentstation/skills ago --agent codex --scope project
```

Change `--agent` for another supported host. The skill is an aid, not an
enforcement mechanism. The committed Go tool, policy, agent instruction, and
CI workflow remain authoritative.

## Rules

Seven rules are on by default. Their constructs have direct replacements.
Six rules are off by default because they encode a project-specific choice.

| Rule | Default | Restriction |
| --- | --- | --- |
| [`no-self-referential-constraints`](docs/rules.md#no-self-referential-constraints) | on | A generic type cannot name itself in its own type parameter list. |
| [`no-new-expr`](docs/rules.md#no-new-expr) | on | `new` accepts a type, not an expression value. |
| [`no-generic-methods`](docs/rules.md#no-generic-methods) | on | Methods cannot declare type parameters. |
| [`no-naked-return`](docs/rules.md#no-naked-return) | on | A return in a function with named results must name its values. |
| [`no-dot-import`](docs/rules.md#no-dot-import) | on | Imports must keep a package qualifier. |
| [`no-goto`](docs/rules.md#no-goto) | on | `goto` is forbidden. |
| [`no-invalid-ignore`](docs/rules.md#no-invalid-ignore) | on | Each suppression must name known rules and give a reason. |
| [`no-f-bounded-constraints`](docs/rules.md#no-f-bounded-constraints) | off | A type parameter cannot appear inside its own constraint. |
| [`no-generic-decls`](docs/rules.md#no-generic-decls) | off | Functions and types cannot declare type parameters. |
| [`no-redundant-short-decl`](docs/rules.md#no-redundant-short-decl) | off | Use `var` for a short declaration in plain statement position. |
| [`no-embedded-field`](docs/rules.md#no-embedded-field) | off | Struct fields must have names. |
| [`no-init-func`](docs/rules.md#no-init-func) | off | Packages cannot declare `func init()`. |
| [`no-blank-import-outside-main`](docs/rules.md#no-blank-import-outside-main) | off | Only package `main` can use a blank import. |

The [rule reference](docs/rules.md) gives the rationale, replacement, evidence,
and non-findings for each rule. `go tool ago -list -format json` carries the
same rule catalogue in machine-readable form.

## Integrations

### GitHub output and SARIF

Use workflow-command output for inline pull request annotations:

```sh
go tool ago -format github ./...
```

Use SARIF 2.1.0 for GitHub code scanning or another SARIF consumer:

```yaml
- run: go tool ago -format sarif ./... > ago.sarif
  continue-on-error: true
- uses: github/codeql-action/upload-sarif@v4
  with:
    sarif_file: ago.sarif
```

### Go analysis drivers

Every rule is a `*analysis.Analyzer`. A custom analysis command can compose
them with `multichecker`:

```go
package main

import (
	"github.com/agentstation/ago"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(ago.Analyzers()...)
}
```

A binary built with `multichecker` supports `go vet -vettool`. The shipped
`cmd/ago` command uses its own policy, JSON, suppression, and exit contracts.
It is not a vet tool.

Library callers can inspect rules or run the checker directly:

```go
rules := ago.Rules()
rule, ok := ago.Lookup("no-goto")
report, err := ago.Check(ago.Options{Patterns: []string{"./..."}})
```

See the [package documentation](https://pkg.go.dev/github.com/agentstation/ago)
for the complete API.

### golangci-lint

ago ships as a [golangci-lint module
plugin](https://golangci-lint.run/plugins/module-plugins/). Add it to
`.custom-gcl.yml`:

```yaml
version: v2.12.2
plugins:
  - module: github.com/agentstation/ago
    import: github.com/agentstation/ago/plugin/golangci
    version: latest
```

Run `golangci-lint custom`, then enable the `ago` custom linter in
`.golangci.yml`. Pin the plugin version before you commit the configuration.

## Project

- [Design and scope](docs/design.md)
- [Rule reference](docs/rules.md)
- [Contributing](CONTRIBUTING.md)
- [Security policy](SECURITY.md)
- [Changelog](CHANGELOG.md)

## License

ago is available under either license, at your option:

- Apache License 2.0 ([LICENSE-APACHE](LICENSE-APACHE))
- MIT License ([LICENSE-MIT](LICENSE-MIT))

The [dual-license grant and contribution terms](COPYRIGHT) apply to the
repository.
