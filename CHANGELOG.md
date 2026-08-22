# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Rule changes are versioned as part of the public interface: adding a rule that
is on by default, or widening what an existing default rule reports, is a minor
version bump at minimum, because it can fail a build that previously passed.

## [Unreleased]

## [0.1.0] - 2026-08-22

Initial public release.

### Added

- Thirteen rules, seven on by default, each one a `go/analysis` analyzer that
  runs with full type information.
- `no-f-bounded-constraints`, off by default, for F-bounded polymorphism
  written through a helper interface.
- Suppression through `//ago:ignore` and `//ago:ignore-file`, with a mandatory
  reason and a `no-invalid-ignore` rule that reports directives which suppress
  nothing.
- `-stale-ignores` to report suppressions that no longer cover a finding.
- `.ago.yml` configuration with `enable`, `disable`, `tests`, and `exclude`,
  the `default` and `all` meta-names, parent-directory search, and strict
  unknown-key rejection.
- Four output formats: `text`, `json`, `sarif`, and `github`.
- `-list -format json`, a machine-readable rule catalogue for coding agents.
- A `golangci-lint` module plugin at `plugin/golangci`.
- Homebrew distribution through `agentstation/tap`.

### Fixed

Relative to the unreleased single-file prototype:

- `no-self-referential-constraints` reported two constructs that have been
  legal since Go 1.18: a constraint naming a sibling type parameter
  (`Graph[N any, E Edge[N]]`) and F-bounded polymorphism written through a
  helper interface (`Node[T Cloneable[T]]`). Neither is what Go 1.26 changed.
  The rule now reports only a type naming itself in its own type parameter
  list; the helper-interface shape moved to the new, off-by-default
  `no-f-bounded-constraints`.
- `no-naked-return` never descended into function literals, so a naked return
  inside a closure with its own named results went unreported.
- A parse error in any file aborted the run and discarded every finding
  already collected. Load and parse failures are now reported alongside the
  findings from the packages that did load.
- `no-new-expr` could not tell a type from a value, so `new(someVariable)` went
  unreported. It now uses `types.Info`.
- Passing the same package twice produced duplicate findings.

[Unreleased]: https://github.com/agentstation/ago/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/agentstation/ago/releases/tag/v0.1.0
