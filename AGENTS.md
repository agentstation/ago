# ago agent instructions

`ago` is a restriction-only Go linter. Every rule rejects a legal Go construct.
No rule may add syntax, rewrite code, or change semantics.

## Technical writing

Use `GLOSSARY.md` for developer-facing prose. Run the installed
`technical-writing` linter on durable technical text. Project mode is
`strict`.

## Before you change a rule

- Read `docs/stdlib-survey.md` and measure. Do not quote a count from memory
  or from the README. The README was wrong before.
- Verify a claim about a Go release against a real toolchain, not recall.
- Build the construct and compile it with the version that supposedly
  introduced it and the version before.
- State what the rule must **not** report. Most defects here were a rule
  catching a neighbouring construct that looked similar.

## Structure

- `rule.go` is the registry. `rule_<topic>.go` holds the analyzers.
- Build every analyzer with `newAnalyzer`. Report through `checkPass.reportf`.
  Never use `pass.Report`. `reportf` is what applies suppression.
- `check.go` is the driver. `config.go` is the `.ago.yml` schema.
  `ignore.go` is the suppression index. `report.go` and `sarif.go` are the
  output formats.
- Analyzers run concurrently. Cross-analyzer state needs `sync/atomic`.
- Rule metadata is public interface. `Name` is the doc anchor and the config
  key. `ago -explain` prints the `Rationale` string verbatim.

## Evidence

Run `make check` before you claim done. Fixtures live in
`testdata/src/<analyzer-ident>/` and must contain deliberate non-findings as
well as findings.

`testdata/` holds intentionally unparseable Go. Never run `gofmt` or `go vet`
across it. The Makefile targets already exclude it.

A rule that reverts a language change is dormant on toolchains older than that
change. Its fixture cannot compile there. Those tests skip themselves.
`no-new-expr` and the two constraint rules need Go 1.26.
`no-generic-methods` needs Go 1.27. Do not "fix" those skips. Set `minGo` on
the fixture instead.
