# Changelog

This file documents every notable change to this project.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
The project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Rule changes are public interface. Adding a rule that is on by default is a
minor version bump at minimum. Widening what an existing default rule reports
is also a minor bump at minimum. Either change can fail a build that previously
passed.

## [Unreleased]

### Added

- Report the rule source, config path, test setting, and exclude patterns in
  `-list`. JSON reports and the rule catalogue now include `schemaVersion: 1`.
- Publish `ago.schema.json` for `.ago.yml` editor validation. New generated
  policies declare config schema version 1. Unversioned v0.1 policies remain
  valid.
- Add command tests for help output, policy discovery, and config creation.

### Changed

- Make the zero-config Go module path the primary adoption flow. A project can
  run the pinned default policy after `go get -tool` without `.ago.yml`.
- Make `-init` an optional customization command. It writes a minimal policy
  at the nearest `go.mod` or `go.work` root and refuses a competing child
  policy.
- Lead the README, package docs, command help, and release notes with ago's
  shared-dialect goal.

### Fixed

- Print `-h` once on stdout, including the complete flag list.
- Reject an invalid `exclude` glob instead of ignoring it.

## [0.1.1] - 2026-08-25

### Security

- Require `golang.org/x/mod` v0.40.0, which fixes GO-2026-6179 and
  GO-2026-6180. `ago` does not call the vulnerable `sumdb` symbols. Scorecard
  still flags the module until this bump.
- Escape `file=` and `title=` in `-format github`. A finding path can no
  longer split a GitHub Actions workflow command.
- Reject a `.ago.yml` larger than 1 MiB. Reject an `exclude` list longer
  than 1024 patterns. Reject an empty exclude pattern and a pattern that
  contains NUL. A config that excludes every package is an error, not a
  clean run.
- Add native Go fuzz tests for config load, exclude matching, ignore
  directives, and GitHub output. CI runs them for 20s each.
- Run `govulncheck` in CI.

### Fixed

- `.github/CODEOWNERS` named `@agentstation/maintainers`, a team that does
  not exist. It now names `@jackspirou` and `@savkat`.
- pkg.go.dev treated the dual-license grant in `LICENSE` as an unknown
  license and hid the package documentation. `COPYRIGHT` now holds the grant.
  The recognized `LICENSE-APACHE` and `LICENSE-MIT` files remain unchanged.
- Commands built through `go tool` or `go install` reported version `dev`.
  They now read the pinned module version from Go build information.

### Changed

- Checkout steps set `persist-credentials: false`.
- Developer-facing prose now follows the project's strict technical-writing
  rules. `GLOSSARY.md` and `.agents/technical-writing.toml` define the
  terms and the linter config. `make prose` runs the check.
- The primary adoption path now pins ago as a Go module tool dependency.
  The README separates adoption, coding-agent contracts, and integrations.
  Detailed rule and design references now live under `docs/`.

## [0.1.0] - 2026-08-22

Initial public release.

### Added

- Thirteen rules, seven on by default. Each is a `go/analysis` analyzer that
  runs with full type information.
- `no-f-bounded-constraints`, off by default, for F-bounded polymorphism
  written through a helper interface.
- Suppression through `//ago:ignore` and `//ago:ignore-file`, with a mandatory
  reason. `no-invalid-ignore` reports directives that suppress nothing.
- `-stale-ignores` to report suppressions that no longer cover a finding.
- `.ago.yml` configuration with `enable`, `disable`, `tests`, and `exclude`.
  Also the `default` and `all` meta-names, parent-directory search, and strict
  unknown-key rejection.
- Four output formats: `text`, `json`, `sarif`, and `github`.
- `-list -format json`, a machine-readable rule catalogue for coding agents.
- A `golangci-lint` module plugin at `plugin/golangci`.
- Homebrew distribution through `agentstation/tap`.

### Fixed

Relative to the unreleased single-file prototype:

- `no-self-referential-constraints` reported two constructs legal since Go
  1.18. One is a constraint naming a sibling type parameter
  (`Graph[N any, E Edge[N]]`). The other is F-bounded polymorphism written
  through a helper interface (`Node[T Cloneable[T]]`). Neither is what Go
  1.26 changed. The rule now reports only a type naming itself in its own
  type parameter list. The helper-interface shape moved to the new, off by
  default `no-f-bounded-constraints`.
- `no-naked-return` never descended into function literals. A naked return
  inside a closure with its own named results went unreported.
- A parse error in any file aborted the run and discarded every finding
  already collected. Load and parse failures now appear beside the findings
  from the packages that did load.
- `no-new-expr` could not tell a type from a value, so `new(someVariable)` went
  unreported. It now uses `types.Info`.
- Passing the same package twice produced duplicate findings.

[Unreleased]: https://github.com/agentstation/ago/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/agentstation/ago/releases/tag/v0.1.1
[0.1.0]: https://github.com/agentstation/ago/releases/tag/v0.1.0
