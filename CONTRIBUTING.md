# Contributing to ago

Thanks for your interest. This document covers how to build, test, and propose
changes — especially new rules, which have a higher bar than ordinary code.

## Build and test

```sh
make build      # build ./cmd/ago
make test       # go test -race ./...
make lint       # gofmt, go vet, golangci-lint
make check      # everything CI runs
```

You need Go 1.25 or later. One rule, `no-generic-methods`, only has anything to
report on Go 1.27 or later; its test skips itself on older toolchains and CI
runs the matrix that exercises it.

## Proposing a rule

`ago` is restriction-only. A rule must forbid a construct that is legal Go and
must never require adding syntax to work around it. A proposal needs four
things.

**A rationale in terms of the reader.** Say what the construct costs someone
reading the code cold — a second place to look, a scroll upward, an identifier
whose origin is not local. "It is confusing" is not a rationale. The rationale
ships in the binary, so `ago -explain <rule>` prints it verbatim; write it for
somebody who just hit the finding and wants to know whether to care.

**An honest default.** A rule is on by default only when the construct has a
direct, mechanical replacement that no reasonable codebase would miss. If
banning it is a taste call, it ships off by default. Say which you are asking
for and why. We would rather add an off-by-default rule than argue about a
default.

**Evidence.** Measure the construct against the standard library — see
[`docs/stdlib-survey.md`](docs/stdlib-survey.md) for the method. A construct
the standard library uses hundreds of times is not automatically disqualified,
but it is automatically off by default, and the count belongs in the rationale
so a reader can weigh it.

**Precision about what the rule does not report.** Most of the bugs we have
fixed were a rule reporting a neighbouring construct that happened to look
similar. State the near-misses explicitly and put them in the fixture as
non-findings.

Rules that revert a specific Go release set `Reverts` to that version. Rules
that encode house style leave it empty and say so in the rationale — do not
dress a preference up as a revert.

## Implementing a rule

Each rule is one file, `rule_<topic>.go`, and one entry in the registry in
`rule.go`. A rule is an `analysis.Analyzer` built with `newAnalyzer`, which
wires reporting through `checkPass.reportf` so suppression cannot be bypassed.
Report through `reportf` and never through `pass.Report` directly.

Type information is available on `checkPass.TypesInfo`. Use it. Several of the
bugs in this tool's history were a syntactic approximation that could not tell
a type from a value.

Fixtures live in `testdata/src/<analyzer-ident>/` and use `analysistest`
GOPATH-mode layout, with `// want "..."` comments on the lines that should be
reported. Every rule needs both findings and deliberate non-findings.

```sh
go test -run TestRules ./...
```

## Pull requests

- One logical change per pull request.
- `make check` passes.
- New behavior has a test; a bug fix has a regression test.
- Update `CHANGELOG.md` under `## [Unreleased]`.
- Update the README rule section if you added or changed a rule. The anchor
  must match the rule name, because `docURL` is generated from it.
- Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/)
  (`feat:`, `fix:`, `docs:`, `ci:`, …); the release changelog is generated from
  them.

## Licensing

Contributions are dual-licensed under MIT and Apache-2.0, matching the project.
By submitting a pull request you agree to license your contribution under both.
