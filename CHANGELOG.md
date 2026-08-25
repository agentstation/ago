# Changelog

This file documents every notable change to this project.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
The project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Rule changes are public interface. Adding a rule that is on by default is a
minor version bump at minimum. Widening what an existing default rule reports
is also a minor bump at minimum. Either change can fail a build that previously
passed.

## [Unreleased]

### Changed

- Developer-facing prose now follows the project's strict technical-writing
  rules. `GLOSSARY.md` and `.agents/technical-writing.toml` define the
  terms and the linter config. `make prose` runs the check.

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

[Unreleased]: https://github.com/agentstation/ago/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/agentstation/ago/releases/tag/v0.1.0
