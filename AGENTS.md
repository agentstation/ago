# ago agent instructions

`ago` is a restriction-only Go linter. Every rule rejects a legal Go construct.
No rule may add syntax, rewrite code, or change semantics.

## Before you change a rule

- Read `docs/stdlib-survey.md` and measure. Do not quote a count from memory or
  from the README; the README has been wrong before.
- Verify a claim about a Go release against a real toolchain, not recall. Build
  the construct and compile it with the version that supposedly introduced it
  and the version before.
- State what the rule must **not** report. Most defects here have been a rule
  catching a neighbouring construct that looked similar.

## Structure

- `rule.go` is the registry. `rule_<topic>.go` holds the analyzers.
- Build every analyzer with `newAnalyzer`. Report through `checkPass.reportf`,
  never `pass.Report` — `reportf` is what applies suppression.
- `check.go` is the driver, `config.go` the `.ago.yml` schema, `ignore.go` the
  suppression index, `report.go` and `sarif.go` the output formats.
- Analyzers run concurrently. Cross-analyzer state needs `sync/atomic`.
- Rule metadata is public interface: `Name` is the doc anchor and the config
  key, and the `Rationale` string is printed verbatim by `ago -explain`.

## Evidence

`make check` before you claim done. Fixtures live in
`testdata/src/<analyzer-ident>/` and must contain deliberate non-findings as
well as findings.

`testdata/` holds intentionally unparseable Go. Never run `gofmt` or `go vet`
across it; the Makefile targets already exclude it.

`no-generic-methods` is dormant before Go 1.27 and its test skips itself. Do not
"fix" that skip.
