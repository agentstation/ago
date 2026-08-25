// Package ago enforces a restricted subset of Go.
//
// ago only ever rejects language constructs. It never adds syntax, never
// rewrites code, and never changes semantics. Code that passes ago is
// ordinary Go that builds with the stock toolchain.
//
// Every rule is a [golang.org/x/tools/go/analysis.Analyzer]. Run them through
// the ago command, compose them in an analysis driver, or load the
// golangci-lint module plugin.
//
// See [github.com/agentstation/ago/cmd/ago] for the command.
package ago
