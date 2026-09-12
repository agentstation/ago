// Package goago enforces one way to write Go across a codebase.
//
// A project selects the Go constructs that it accepts. Developers and coding
// agents use the same rule policy, and CI enforces it.
//
// goago only ever rejects language constructs. It never adds syntax, never
// rewrites code, and never changes semantics. Code that passes goago is
// ordinary Go that builds with the stock toolchain.
//
// Every rule is a [golang.org/x/tools/go/analysis.Analyzer]. Run them through
// the goago command, compose them in an analysis driver, or load the
// golangci-lint module plugin.
//
// See [github.com/agentstation/goago/cmd/goago] for the command.
package goago
