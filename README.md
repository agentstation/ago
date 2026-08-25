# ago

[![CI](https://github.com/agentstation/ago/actions/workflows/ci.yml/badge.svg)](https://github.com/agentstation/ago/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/agentstation/ago.svg)](https://pkg.go.dev/github.com/agentstation/ago)
[![Go Report Card](https://goreportcard.com/badge/github.com/agentstation/ago)](https://goreportcard.com/report/github.com/agentstation/ago)
[![License](https://img.shields.io/badge/license-MIT%20OR%20Apache--2.0-blue)](#license)

A restriction-only linter for Go.

The name reads two ways: *a-go* (agent Go, the subset an agent may write)
and *ago* (the Go we had before, restored).

`ago` only ever rejects language constructs. It never adds syntax, rewrites
code, or changes semantics. Code that passes `ago` is ordinary Go. That Go
builds with the stock toolchain.

```console
$ ago ./...
internal/store/index.go:42:2: naked return in indexAll; name the values you are returning (no-naked-return)
internal/store/index.go:88:9: new() takes a type, not an expression (no-new-expr)
2 violations
```

## Install

```sh
brew install agentstation/tap/ago
```

```sh
go install github.com/agentstation/ago/cmd/ago@latest
```

Or download a signed archive from [releases](https://github.com/agentstation/ago/releases).
Every archive ships with `checksums.txt` and an SBOM.

## Quick start

```sh
ago ./...                    # default rule set, current module
ago -init                    # write a starter .ago.yml
ago -list                    # show every rule and which are on
ago -explain no-goto         # full rationale for one rule
ago -all ./...               # every rule, ignoring config
ago -tests ./...             # include _test.go files
```

Package arguments are [`go/packages`](https://pkg.go.dev/golang.org/x/tools/go/packages)
patterns. They are the same ones `go build` accepts. With no arguments `ago`
checks `./...`.

| Exit | Meaning |
| ---- | ------- |
| `0`  | no violations |
| `1`  | at least one violation |
| `2`  | `ago` could not complete the run |

`vendor/` and `testdata/` are always skipped. Third-party code is not yours to
restrict.

## Motivation

Go's value was never any single feature. A large team could read each other's
code without negotiating a dialect.
Pike's own summary of what the project got right ends on this: Go code looks
and works the same regardless of who wrote it. It is largely free of factions
that use different subsets of the language.

Recent releases widened the language faster than that property can absorb. Go
1.26 let a generic type refer to itself in its own type parameter list. Go
1.26 also let `new` take an expression rather than a type. Go 1.27 let methods
declare their own type parameters. Each change is defensible in isolation.
Together they add ways to express things that were already expressible.

That is the specific kind of growth that produces dialects.

`ago` does not argue those features are bad. It argues that a given codebase
benefits from picking one way to say each thing. Enforce that choice. Do not
leave it to review. Every rule here removes an alternative. None adds one.

## Rules

A rule is **on by default** when the construct it forbids has a direct,
mechanical replacement that no reasonable codebase would miss. It is **off by
default** when the construct has legitimate uses. Banning it is then a genuine
policy choice rather than a cleanup.

The survey measured counts below against the `go1.26.5` standard library, over
non-test files, excluding `testdata/`, `vendor/`, and `_asm/`. That population
is 3,916 files. Reproduce them with the script in
[`docs/stdlib-survey.md`](docs/stdlib-survey.md).

### On by default

#### `no-self-referential-constraints`

Reverts Go 1.26. Before Go 1.26 the compiler rejected `type Adder[A Adder[A]]
interface{ Add(A) A }` as an invalid recursive type. Go 1.26 accepts it.

This rule reports exactly that construct: the declared type's own name appearing
inside its own type parameter list.

```go
type Adder[A Adder[A]] interface{ Add(A) A }   // reported
type Node[T Cloneable[T]] struct{ v T }        // not reported — legal since Go 1.18
type Graph[N any, E Edge[N]] struct{}          // not reported — sibling parameter
```

Detection is syntactic and catches direct self-reference. Mutual recursion
across two declarations is not reported. For the F-bounded shapes that were
always legal, see [`no-f-bounded-constraints`](#no-f-bounded-constraints).

#### `no-new-expr`

Reverts Go 1.26. `new(f(x))` collapses declaration, initialization, and
address-taking into one expression. The two-line form with a named variable says
the same thing. A reader can follow it left to right.

```go
p := new(f(x))            // reported
v := f(x); p := &v        // the replacement
p := new(Foo)             // not reported — Foo is a type
```

The rule uses full type information. So the rule reports `new(someVariable)`
even where a purely syntactic check could not tell a type from a value.

#### `no-generic-methods`

Reverts Go 1.27. Rewrite any such method as a package-level function that
takes the receiver as its first argument. The feature buys call-site chaining.
It costs a second place to look for a type's operations.

> **Dormant before Go 1.27.** On earlier toolchains `go/parser` rejects generic
> methods itself, so the construct never reaches this rule. That is correct.
> The language is already enforcing it. The rule does nothing for you until
> you upgrade.

#### `no-naked-return`

A bare `return` in a function with named results forces the reader to scroll up
to learn the returned values. It also silently returns whatever the result
variables hold at that point. Naming the values costs nothing and survives
later edits to the body.

The rule checks function literals against their own result list rather than the
enclosing function's. It reports a naked return inside a closure when the
closure itself declares named results.

#### `no-dot-import`

Dot imports make every identifier in a file ambiguous as to origin. That
defeats both the reader and `grep`.

The standard library does use them in production code. The pattern is
essentially one: `go/types` and `cmd/compile/internal/types2` (26 files each)
dot-import `internal/types/errors`, a package of nothing but error-code
constants. That is the narrow case where this rule is wrong. A constants-only
package referenced densely enough that qualifying every use is pure noise.
Outside those 53 files, only tests in the standard library use dot imports
(224 files).

#### `no-goto`

Labelled `break` and `continue` cover the loop cases. A `goto` that jumps
anywhere else does control flow the reader has to simulate by hand.

This is a rule for application code. It is not a claim that `goto` is vestigial.
The standard library contains 522 `goto` statements in non-test files, only 19
of them generated. They concentrate in `syscall` (120), `runtime` (77), and the
two type checkers (71). That code is either performance-critical or a
hand-written state machine. If you write that kind of code, turn this off.

#### `no-invalid-ignore`

Every `//ago:ignore` directive must name a known rule and give a reason. See
[Suppressing a finding](#suppressing-a-finding). This rule is what makes the
suppression syntax safe to hand to an agent. A directive that does not parse,
names no rule, names a missing rule, or omits its reason suppresses nothing.
This rule reports that directive.

### Off by default

#### `no-f-bounded-constraints`

Reports F-bounded polymorphism. A type parameter appears inside its own
constraint, as in `type Node[T Cloneable[T]]` or `func algo[A Adder[A]](x A) A`.

This is **not** a version revert. Generics in Go 1.18 made this construct
legal. It is the most abstraction-dense shape the type system allows. An
ordinary interface plus a concrete type usually reaches the same shape, so
some codebases choose to ban it. That is a house-style decision. That is why
the rule is off by default.

A constraint that refers to a *sibling* type parameter, such as `type Graph[N
any, E Edge[N]]`, is not reported. The parameter does not appear in its own
constraint.

#### `no-generic-decls`

Reverts Go 1.18 entirely: no type parameters on any func or type declaration.
The strictest rule here and the least likely to be right for you. Reasonable for
a codebase with no container libraries of its own. Hostile otherwise. Only 115
non-test files in the standard library declare a generic func or type. The
ones that do are the ones you import.

It does not stop you *calling* generic functions such as `slices.Sort`.

#### `no-redundant-short-decl`

One way to introduce a variable instead of several. 2,522 of the standard
library's 3,916 non-test files use `:=`. Turning this on is a decision about
your code, not a defect count.

The rule does not ban `:=` outright, because `:=` is load-bearing where `var` is
a syntax error:

```go
switch t := x.(type) { ... }   // var is not allowed in a switch guard
for i, v := range xs { ... }   // no var form exists
```

It reports `:=` only in plain statement position, where `var` is a drop-in
replacement. Hoisting an `if` or `for` header declaration out to a preceding
`var` is legal but widens the variable's scope. The rule treats those as
load-bearing too. A blanket ban would force *adding* syntax to the language.
`ago` will not do that.

#### `no-embedded-field`

Embedding promotes methods and fields implicitly. A type's method set stops
being visible at its declaration. Naming the field costs one selector per use.

Off by default because embedding is thoroughly idiomatic. The standard library
has 680 embedded fields across 273 non-test files, including 24 bare
`sync.Mutex` or `sync.RWMutex` embeds. Turning this on is a real break with
house Go, not a tidy-up.

#### `no-init-func`

`init` runs before `main` in an order determined by the import graph. That
makes initialization failures hard to localize and hard to test. Off by default
because avoiding it entirely requires an explicit wiring convention the tool
cannot supply.

#### `no-blank-import-outside-main`

Blank imports register side effects through the import graph. Program behavior
then depends on which packages the linker includes. Confining them
to `main` keeps that explicit. Off by default because some driver-registration
patterns legitimately need them in library packages.

## Configuration

`ago` reads `.ago.yml` from the working directory or the nearest parent. Write a
starter with `ago -init`.

```yaml
enable:
  - no-goto
  - no-naked-return
  - no-new-expr

disable:
  - no-dot-import

tests: false        # also check _test.go files
exclude:            # vendor and testdata are always skipped
  - "*.pb.go"
  - third_party/*
```

`enable` also accepts the meta-names `default` (the default rule set) and `all`
(every rule). So `enable: [default]` plus `disable: [no-goto]` is the usual
shape. `disable` wins over `enable`.

`ago` matches an `exclude` pattern against the whole slash-separated path,
against each path element, and against each leading path prefix. So `*.pb.go`
matches by name. `generated` excludes any directory with that name at any
depth. `third_party/*` excludes a whole subtree.

Unknown keys are an error rather than a silent no-op. A typo in your config
fails the run instead of quietly disabling a rule.

Precedence, strongest first: `-rules` / `-all` → `.ago.yml` → defaults. Use
`-config path` to point at a specific file. Use `-no-config` to ignore any file
on disk.

## Suppressing a finding

```go
//ago:ignore no-goto -- hand-written state machine, see docs/parser.md
goto retry
```

The directive applies to the following line. `//ago:ignore-file` at the top of a
file applies to the whole file. Both take a comma-separated rule list, or `*`
for every rule.

**The reason is mandatory.** A directive with no `--` reason, with no rule named,
or naming a rule that does not exist suppresses nothing.
[`no-invalid-ignore`](#no-invalid-ignore) reports it. Suppression that costs
nothing gets used reflexively. This one costs a sentence.

Run `ago -stale-ignores ./...` to find directives that no longer suppress
anything. That is how a suppression gets deleted after you fix the code it
covered.

## Output formats

```sh
ago -format text ./...     # default, one finding per line
ago -format json ./...     # stable schema, for tools and agents
ago -format sarif ./...    # SARIF 2.1.0, for GitHub code scanning
ago -format github ./...   # ::error workflow commands, inline PR annotations
```

The JSON schema is stable:

```json
{
  "version": "1.0.0",
  "rules": ["no-dot-import", "no-goto", "no-naked-return"],
  "findings": [
    {
      "rule": "no-naked-return",
      "severity": "error",
      "message": "naked return in indexAll; name the values you are returning",
      "file": "internal/store/index.go",
      "line": 42,
      "column": 2,
      "endLine": 42,
      "endColumn": 8,
      "docURL": "https://github.com/agentstation/ago#no-naked-return"
    }
  ],
  "staleIgnores": [],
  "errors": []
}
```

`version` is the `ago` version that produced the report. `rules` is the rule set
that actually ran. Only `-stale-ignores` fills `staleIgnores`.

`ago` sorts findings by file, line, column, then rule, and deduplicates them.
The output is byte-identical across runs and safe to diff in CI.

A package that fails to load or parse lands in `errors`. It does **not** stop
the run. You still get every finding from every package that did load. The
exit status is 2 to tell you the run was incomplete.

## For coding agents

`ago` serves an agent harness as much as a person.

**Discover the rules.** `ago -list -format json` emits the full catalogue.
It includes name, summary, complete rationale, default state, whether it
reverts a specific Go release, severity, and doc URL. An agent can decide
what a rule means without fetching this page.

```sh
ago -list -format json | jq -r '.rules[] | select(.enabled) | .name'
```

**Read the findings.** `ago -format json` gives one flat array with a `docURL`
per finding. `ago -explain <rule>` prints the same rationale a human would read.

**Trust the exit code.** `0` clean, `1` violations, `2` incomplete. An agent can
branch on `2` to report a broken build rather than an empty success.

**Put the policy in the repo.** A committed `.ago.yml` means an agent reads the
house style from the working tree instead of from a prompt. The same rules
apply to every contributor, human or not.

**Suppress honestly.** The mandatory `--` reason means an agent has to state why
it silences a rule. `-stale-ignores` finds the ones that outlived their
cause.

Recommended `AGENTS.md` snippet for a repo that adopts `ago`:

```markdown
Run `ago ./...` before you finish. It enforces this repo's Go subset.
Run `ago -list` to see the active rules and `ago -explain <rule>` for the
rationale behind one. Do not suppress a finding without a `--` reason, and do
not edit `.ago.yml` to make a finding go away.
```

## Use as a library

Every rule is a `*analysis.Analyzer`. So `ago` plugs into anything built on
`golang.org/x/tools/go/analysis`.

```go
import "github.com/agentstation/ago"

func main() {
	multichecker.Main(ago.Analyzers()...)
}
```

```go
rules := ago.Rules()                 // full catalogue with metadata
rule, ok := ago.Lookup("no-goto")    // by kebab name or analyzer ident
report, err := ago.Check(ctx, ago.Options{Patterns: []string{"./..."}})
```

See the [package documentation](https://pkg.go.dev/github.com/agentstation/ago)
for the full surface.

### golangci-lint

`ago` ships as a [module plugin](https://golangci-lint.run/plugins/module-plugins/).
Add it to `.custom-gcl.yml`:

```yaml
version: v2.6.0
plugins:
  - module: github.com/agentstation/ago
    import: github.com/agentstation/ago/plugin/golangci
    version: latest
```

Then `golangci-lint custom` and enable `ago` in your `.golangci.yml`.

## CI

```yaml
- uses: actions/setup-go@v6
  with:
    go-version: stable
- run: go install github.com/agentstation/ago/cmd/ago@latest
- run: ago -format github ./...
```

For code scanning, emit SARIF and upload it:

```yaml
- run: ago -format sarif ./... > ago.sarif
  continue-on-error: true
- uses: github/codeql-action/upload-sarif@v4
  with:
    sarif_file: ago.sarif
```

## Scope

These rules encode one project's taste, not Rob Pike's. His 2024 retrospective
*What We Got Right, What We Got Wrong* does not list variable declaration syntax
as a mistake. On generics it argues close to the opposite of this tool. Defining
generic containers in the language without giving programmers access to that
genericity was arguably an error. Treat `ago` as a house style, not an appeal
to authority.

## The fork alternative

If you want compiler-level enforcement, the patch is genuinely small. Both Go
parsers rejected generic methods up through 1.26. Go 1.27 deletes those
checks. Restoring them is a two-site revert.

`src/cmd/compile/internal/syntax/parser.go`, in `funcType`:

```diff
 var tparamList []*Field
 if p.got(_Lbrack) {
+    if context == "method" {
+        p.syntaxErrorAt(typ.pos, "method must have no type parameters")
+    }
     if p.tok == _Rbrack {
```

`src/go/parser/parser.go`, in `parseFuncDecl`:

```diff
+if recv != nil && tparams != nil {
+    p.error(tparams.Opening, "method must have no type parameters")
+}
```

Then `./make.bash`.

**Why this repo is a linter instead.** The stock standard library uses the
things you would ban. 2,522 of 3,916 non-test files use `:=`. 115 declare
generic funcs or types. A compiler that rejects those constructs cannot build
its own standard library. So a real fork needs scoping.

Restrictions apply to first-party packages only, exempting `GOROOT` and the
module cache. That is the part that stops being a small patch. It is what a
linter gives you for free.

A fork also means a second toolchain to distribute and pin
(`GOTOOLCHAIN=local`). It means a rebase every six months. It means a `gopls`
that disagrees with your compiler unless you fork that too.

The reasonable split: fork for the one or two constructs a rogue dependency
could sneak into your binary. Use `ago` for everything that is house policy.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). New rules need a rationale that says
what the construct costs a reader, a `testdata` fixture, and an honest default.
If banning it is a taste call, it ships off by default.

## License

Dual-licensed under either of

- Apache License, Version 2.0 ([LICENSE-APACHE](LICENSE-APACHE))
- MIT license ([LICENSE-MIT](LICENSE-MIT))

at your option.
