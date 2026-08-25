# Design and scope

ago is a restriction-only linter. It gives a project a mechanically enforced
subset of legal Go.

## Motivation

Go's value was never one feature. A large team could read its code without
negotiating a dialect. Rob Pike's retrospective, *What We Got Right, What We
Got Wrong*, states that Go code looks and works the same across programmers.
The language is largely free of factions that use different subsets.

Recent releases widened the language. Go 1.26 let a generic type refer to
itself in its type parameter list. It also let `new` accept an expression
instead of a type. Go 1.27 let methods declare type parameters. Each construct
has another legal expression.

ago does not claim that these features are bad. A project can benefit from one
way to express each operation. The linter enforces that choice instead of
leaving it to review.

Each ago rule removes an alternative. No rule adds syntax, rewrites code, or
changes semantics.

## Policy boundary

The rules encode one project's taste. They do not claim to encode Rob Pike's
preferences. His retrospective does not list variable declaration syntax as a
mistake. Its generics argument is close to the opposite position. Defining
generic containers in the language without programmer access to that
genericity was arguably an error.

Treat ago as a house policy, not an appeal to authority.

## Why a linter instead of a Go fork

The parser change for one language restriction can be small. For example, Go
parsers rejected generic methods through Go 1.26. Go 1.27 removed those
checks. A fork can restore them in two parser locations.

In `src/cmd/compile/internal/syntax/parser.go`, restore the receiver check in
`funcType`:

```diff
 var tparamList []*Field
 if p.got(_Lbrack) {
+    if context == "method" {
+        p.syntaxErrorAt(typ.pos, "method must have no type parameters")
+    }
     if p.tok == _Rbrack {
```

In `src/go/parser/parser.go`, restore the check in `parseFuncDecl`:

```diff
+if recv != nil && tparams != nil {
+    p.error(tparams.Opening, "method must have no type parameters")
+}
```

Build the fork with `./make.bash`.

The scope boundary is harder. The stock standard library uses constructs that
a project can reject. It uses `:=` in 2,522 of 3,916 non-test files. It declares
generic functions or types in 115 non-test files. A compiler that rejects
those constructs cannot build its own standard library.

A useful compiler fork must restrict first-party packages while it exempts
`GOROOT` and the module cache. ago gets that scope from package loading.

A fork also needs its own distribution and version pin through
`GOTOOLCHAIN=local`. It needs a rebase every six months. A stock `gopls` can
disagree with that compiler unless the project forks the language server too.

Use a compiler fork when a dependency must not contain one or two constructs.
Use ago for source policy in the code that the project owns.

## Measured evidence

The [standard library survey](stdlib-survey.md) records the population,
commands, and current counts behind each rule rationale. A rule proposal must
state its non-findings and measure the construct before it selects a default.
