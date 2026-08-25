# Rule reference

A rule is on by default when its construct has a direct, mechanical
replacement. A rule is off by default when the construct has legitimate uses.
Banning that construct is a project policy choice.

The survey measured these counts against the `go1.26.5` standard library. It
used 3,916 non-test files and excluded `testdata/`, `vendor/`, and `_asm/`.
Reproduce the counts with the method in [the standard library
survey](stdlib-survey.md).

## On by default

### `no-self-referential-constraints`

This rule reverts Go 1.26. Earlier compilers rejected
`type Adder[A Adder[A]] interface{ Add(A) A }` as an invalid recursive type.
Go 1.26 accepts it.

The rule reports the declared type's own name inside its type parameter list.

```go
type Adder[A Adder[A]] interface{ Add(A) A }   // reported
type Node[T Cloneable[T]] struct{ v T }        // not reported: legal since Go 1.18
type Graph[N any, E Edge[N]] struct{}          // not reported: sibling parameter
```

Detection is syntactic and catches direct self-reference. The rule does not
report mutual recursion across two declarations. See
[`no-f-bounded-constraints`](#no-f-bounded-constraints) for the F-bounded
forms that earlier Go releases accepted.

### `no-new-expr`

This rule reverts Go 1.26. `new(f(x))` combines declaration, initialization,
and address-taking in one expression. The two-line form gives the value a name
and reads from left to right.

```go
p := new(f(x))            // reported
v := f(x); p := &v        // replacement
p := new(Foo)             // not reported: Foo is a type
```

The rule uses complete type information. It reports `new(someVariable)` when
a syntactic check cannot distinguish a type from a value.

### `no-generic-methods`

This rule reverts Go 1.27. Replace a generic method with a package function
that accepts the receiver as its first argument. The feature provides call-site
chaining but adds a second place to find a type's operations.

> **Dormant before Go 1.27.** Earlier parsers reject generic methods before
> the rule runs. This behavior is correct. The language already enforces the
> restriction until the toolchain upgrade.

### `no-naked-return`

A bare `return` in a function with named results makes the reader find the
result declarations. It also returns the current result variables. Explicit
operands keep that data at the return statement.

The rule checks a function literal against its own result list. It reports a
naked return in a closure only when that closure declares named results.

### `no-dot-import`

Dot imports make the origin of each identifier in a file ambiguous. This
defeats the reader and `grep`.

The standard library uses this form in production code. The two type checkers
account for almost all cases. `go/types` and `cmd/compile/internal/types2`
each have 26 files that dot-import `internal/types/errors`. That package
contains error-code constants. Outside those 53 files, only standard library
tests use dot imports, across 224 files.

A constants-only package with dense references is the narrow case where this
rule is wrong.

### `no-goto`

Labelled `break` and `continue` cover loop control. A `goto` outside those
cases creates control flow that the reader must simulate.

This rule targets application code. The standard library contains 522 `goto`
statements in non-test files. Generators produce only 19. `syscall` contains
120, `runtime` contains 77, and the two type checkers contain 71. That code is
performance-critical or implements a state machine. Disable this rule for that
kind of code.

### `no-invalid-ignore`

Every `//ago:ignore` directive must name a known rule and give a reason. A
directive that does not parse, names no rule, names an unknown rule, or omits
its reason suppresses nothing. This rule reports that directive.

The rule makes the suppression syntax safe for a coding agent. A reviewer can
read the required reason before approving the exception.

## Off by default

### `no-f-bounded-constraints`

This rule reports F-bounded polymorphism. A type parameter appears inside its
own constraint, as in `type Node[T Cloneable[T]]` or
`func algo[A Adder[A]](x A) A`.

This rule does not revert a version. Go 1.18 made the construct legal. It is an
abstraction-dense type-system form. An ordinary interface and a concrete type
can often express the same design. A project can choose to reject it.

The rule does not report a constraint that refers to a sibling type parameter.
For example, it does not report `type Graph[N any, E Edge[N]]`. The parameter
does not appear inside its own constraint.

### `no-generic-decls`

This rule reverts Go 1.18 for declarations. Functions and types cannot declare
type parameters. Code can still call generic functions such as `slices.Sort`.

This is the strictest rule and rarely fits a general-purpose package. Only 115
non-test files in the standard library declare a generic function or type.
Those packages often exist for callers to import.

### `no-redundant-short-decl`

This rule provides one way to introduce a variable. It reports `:=` only in a
plain statement where `var` is a direct replacement.

The rule does not report a load-bearing short declaration:

```go
switch t := x.(type) { ... }   // var is not allowed in a switch guard
for i, v := range xs { ... }   // no var form exists
```

It also does not report declarations in `if` and `for` headers. Moving those
declarations before the statement would widen their scope.

The standard library uses `:=` in 2,522 of 3,916 non-test files. Enable this
rule only when the project chooses this declaration policy.

### `no-embedded-field`

Embedding promotes methods and fields without an explicit selector. A type's
method set is no longer visible at its declaration. A named field makes each
promotion explicit at the use site.

Embedding is idiomatic Go. The standard library has 680 embedded fields across
273 non-test files. These include 24 bare `sync.Mutex` or `sync.RWMutex`
fields. Enable this rule only for a project that rejects that convention.

### `no-init-func`

`init` runs before `main` in an order that the import graph determines. This
makes an initialization failure hard to localize and test. Avoiding `init`
requires an explicit wiring convention that ago cannot provide.

### `no-blank-import-outside-main`

A blank import registers side effects through the import graph. Program
behavior then depends on which packages the linker includes. Restricting blank
imports to `main` keeps the registration at the composition root.

Some driver-registration designs need a blank import in a library package.
That legitimate use keeps this rule off by default.
