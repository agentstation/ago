# Migrate from ago to goago

Version 0.3.0 renames the tool to `goago`, pronounced **go ago**.
The rule names, defaults, configuration schema, JSON fields, and exit statuses stay the same.

## Go module tools

Run these commands in each module that pins the old tool:

```sh
go get -tool github.com/agentstation/goago/cmd/goago@v0.3.0
go mod edit -droptool=github.com/agentstation/ago/cmd/ago
go mod tidy
go tool goago -version
go tool goago -list -format json
go tool goago -stale-ignores -format json ./...
```

Commit `go.mod` and `go.sum`. Change repository checks, CI commands, and agent
instructions from `go tool ago` to `go tool goago`.

For a global installation, run:

```sh
go install github.com/agentstation/goago/cmd/goago@v0.3.0
```

## Policy and suppressions

Rename `.ago.yml` to `.goago.yml`, or `.ago.yaml` to `.goago.yaml`.
Change `//ago:ignore` and `//ago:ignore-file` comments to `//goago:ignore` and
`//goago:ignore-file`. Keep each rule list and reason unchanged.

The previous config filenames and comment prefixes still work during migration.
The nearest directory with a policy supplies the policy, regardless of its name.
Multiple policy files in that directory stop the run. Keep one file or select
it explicitly with `-config`. The `-init` command refuses an existing policy
under either name. New policies use `.goago.yml`.

Use `goago.schema.json` for editor validation. Update the schema URL to
`https://raw.githubusercontent.com/agentstation/goago/main/goago.schema.json`.

## Library and golangci-lint plugin

Change imports from `github.com/agentstation/ago` to
`github.com/agentstation/goago`. The package name is now `goago`.
Update selectors such as `ago.Check` to `goago.Check`.

For a custom golangci-lint build, change the plugin module and import path to
`github.com/agentstation/goago` and `github.com/agentstation/goago/plugin/golangci`.
Enable the custom linter as `goago`, then rebuild the custom binary.

## Homebrew and agent skills

The Homebrew cask is now `agentstation/tap/goago`. Update the tap, then run:

```sh
brew trust --cask agentstation/tap/goago
brew install --cask agentstation/tap/goago
```

The Agent Skill is now `goago` in `agentstation/skills`. Replace an installed
`ago` skill with the `goago` skill. The modern Go skill uses the new command.

## Releases and reports

The repository is now `github.com/agentstation/goago`. GitHub redirects old
repository links. Releases through v0.2.0 contain the old module and executable.
Use v0.3.0 or later for the new module path.

Release archives and the executable use `goago`. SARIF identifies the tool as
`goago`, and diagnostics link to the renamed repository. Historical reports
still identify the executable that produced them.
