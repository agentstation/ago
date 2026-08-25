# Glossary

Project terms for developer-facing prose. Code identifiers stay exact even
when they match a row.

| Term | Definition | Avoid | Status | Evidence |
|---|---|---|---|---|
| ago | The restriction-only Go linter this repository builds | | approved | README.md |
| analyzer ident | The hyphen-free `Analyzer.Name` that `go/analysis` requires | | approved | rule.go |
| coding agent | A program that drives `ago` through its command, JSON catalogue, or exit status | | approved | README.md |
| F-bounded | A type parameter that appears inside its own constraint | | approved | README.md, docs/stdlib-survey.md |
| finding | One reported violation of a rule | | approved | finding.go, README.md |
| fixture | A `testdata` package that records expected findings and non-findings | | approved | CONTRIBUTING.md |
| golangci-lint | The Go linter runner that loads `ago` as a module plugin | | approved | README.md, plugin/golangci/plugin.go |
| meta-name | The `enable` values `default` and `all` | | approved | README.md, config.go |
| naked return | A `return` with no operands in a function with named results | | approved | README.md, rule_declarations.go |
| near-miss | A neighbouring construct the rule must not report | | approved | CONTRIBUTING.md |
| non-finding | A construct a fixture includes because the rule must not report it | | approved | CONTRIBUTING.md, AGENTS.md |
| non-test | A `.go` file whose name does not end in `_test.go` | | approved | docs/stdlib-survey.md |
| off by default | A rule that is disabled until configuration enables it | off-by-default | approved | README.md, CONTRIBUTING.md |
| on by default | A rule that is enabled when no configuration selects a rule set | | approved | README.md |
| rationale | The explanation `ago -explain` prints for a rule | | approved | rule.go, README.md |
| restriction-only | Rejects legal Go constructs. Does not add syntax, rewrite code, or change semantics | | approved | README.md, AGENTS.md |
| rule | One named restriction paired with a `go/analysis` analyzer | | approved | rule.go |
| SARIF | Static Analysis Results Interchange Format 2.1.0 | | approved | README.md, sarif.go |
| stale ignore | An `//ago:ignore` directive that suppressed no finding in the run | | approved | README.md, check.go |
| standard library | The Go standard library under `GOROOT/src` | stdlib | approved | README.md, docs/stdlib-survey.md |
| suppression | An `//ago:ignore` or `//ago:ignore-file` directive | | approved | README.md, ignore.go |
