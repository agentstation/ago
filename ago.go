// Package ago enforces a restricted subset of Go.
//
// ago only ever rejects language constructs. It never adds syntax, never
// rewrites code, and never changes semantics. Code that passes ago is
// ordinary Go that builds with the stock toolchain.
//
// Every rule is a [golang.org/x/tools/go/analysis.Analyzer], so the same
// checks run three ways: through the ago command, through go vet with
// -vettool, and through a golangci-lint module plugin.
//
// The command is documented at [github.com/agentstation/ago/cmd/ago].
package ago

// Version is the module version, overwritten at link time by the release
// build. It is "dev" for builds from source.
var Version = "dev"
