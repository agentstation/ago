# ago

[![CI](https://github.com/agentstation/ago/actions/workflows/ci.yml/badge.svg)](https://github.com/agentstation/ago/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/agentstation/ago.svg)](https://pkg.go.dev/github.com/agentstation/ago)
[![Go Report Card](https://goreportcard.com/badge/github.com/agentstation/ago)](https://goreportcard.com/report/github.com/agentstation/ago)
[![License](https://img.shields.io/badge/license-MIT%20OR%20Apache--2.0-blue)](#license)

**One Go dialect for the whole codebase.**

Go's productive kind of boring comes from familiarity. The original design put
simplicity, readability, and one common form ahead of extra choice. Uniform Go
code also helps teams read and maintain code across authors. See [Go design
history](https://go.dev/talks/2015/gophercon-goevolution.slide) and [Go at
Google](https://go.dev/talks/2012/splash.article).

As Go gains more legal forms, ago lets a project select the forms it accepts.
Developers and coding agents run the same policy, and CI enforces it. Where
`gofmt` standardizes presentation, ago standardizes the accepted language
subset.

`ago` rejects selected legal Go constructs. It does not add syntax, rewrite
code, or change semantics. Code that passes `ago` is ordinary Go that builds
with the stock toolchain.

```console
$ go tool ago ./...
internal/store/index.go:42:2: naked return in indexAll; name the values you are returning (no-naked-return)
internal/store/index.go:88:9: new() takes a type, not an expression (no-new-expr)
2 violations
```

The name reads as *a-go*, the Go subset that a project accepts. It also refers
to an earlier, smaller Go language. Read the [design case](docs/design.md) for
the project boundary and evidence.

## Adopt ago in a Go repository

ago requires Go 1.25 or later and a Go module.

1. Add ago as a module tool dependency.

   ```sh
   go get -tool github.com/agentstation/ago/cmd/ago@latest
   ```

   This command pins the ago version in `go.mod`. It records module checksums
   in `go.sum`.

2. Check the module.

   ```sh
   go tool ago ./...
   ```

   A clean run prints nothing and exits with status 0.

The Go module now owns the ago version. Developers, coding agents, and CI can
run `go tool ago` without a global installation or a `PATH` change.

ago does not require a config file. The pinned ago version supplies the default
rule policy. Add `.ago.yml` only when the project needs a different policy.
Run the same `go get -tool` command later to upgrade ago deliberately.

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

Use these repository files to give each contributor the same command and
policy:

| File | Purpose | When needed |
| --- | --- | --- |
| `go.mod` and `go.sum` | Pin the ago command and its module graph. | Always |
| `.ago.yml` | Change or record the built-in rule policy. | Only for a custom policy |
| `AGENTS.md` | Tell coding agents when and how to run ago. | Repositories that use coding agents |
| CI workflow | Reject a change that violates the policy. | Repositories that enforce ago |

Add this instruction to the adopting repository's `AGENTS.md`:

```markdown
Run `go tool ago -stale-ignores -format json ./...` after each Go change.
Fix findings in source. Do not add or change `.ago.yml` only to make the run
pass. Do not add a suppression only to make the run pass. Exit status 2 means
the check was incomplete.
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

Configuration is optional. With no config file, ago runs the default rules from
the version pinned in `go.mod`.

Create a minimal policy only when the project needs one:

```sh
go tool ago -init
```

The command writes `.ago.yml` at the nearest `go.mod` or `go.work` root. It
refuses to create a second policy when a parent policy already applies.

```yaml
version: 1
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

Choose the policy form that matches the project:

| Form | Upgrade behavior |
| --- | --- |
| No `.ago.yml` | Use the defaults in the pinned ago version. |
| `enable: [default]` | Record a policy file and use the defaults in the pinned version. |
| Explicit rule names | Keep the named rule set until the project edits the file. |

ago matches each `exclude` pattern against three path shapes:

- the complete slash-separated path.
- each path element.
- each leading path prefix.

Thus, `*.pb.go` matches a file name, `generated` matches that directory at any
depth, and `third_party/*` matches that subtree.

Unknown keys and unknown rule names stop the run. A policy typo cannot disable
a rule silently. Use `-config path` to name a file. Use `-no-config` to ignore
all policy files. The [JSON Schema](ago.schema.json) supplies editor validation.
ago also accepts unversioned files created by v0.1.

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

The document includes a schema version and the resolved policy source. It also
reports the config path, test setting, and exclude patterns. Each rule entry
includes its name, analyzer ident, summary, rationale, default and active state,
Go release boundary, severity, and documentation URL.

`policy.ruleSource` is `built-in`, `config`, or `flags`. `configDisabled` is
true when the command used `-no-config`.

Read findings and incomplete-run errors:

```sh
go tool ago -stale-ignores -format json ./...
```

```json
{
  "schemaVersion": 1,
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
produce the same JSON document. Existing JSON fields keep their names and
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

Change `--agent` for another supported host. The skill guides the local repair
loop. The pinned Go tool, optional policy, agent instruction, and CI workflow
remain authoritative.

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
- [Config schema](ago.schema.json)
- [Contributing](CONTRIBUTING.md)
- [Security policy](SECURITY.md)
- [Changelog](CHANGELOG.md)

## License

ago is available under either license, at your option:

- Apache License 2.0 ([LICENSE-APACHE](LICENSE-APACHE))
- MIT License ([LICENSE-MIT](LICENSE-MIT))

The [dual-license grant and contribution terms](COPYRIGHT) apply to the
repository.
