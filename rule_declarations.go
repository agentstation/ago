package ago

import (
	"go/ast"
	"go/token"
)

// RuleNoRedundantShortDecl forbids := where var is a drop-in replacement.
var RuleNoRedundantShortDecl = register(Rule{
	Name:     "no-redundant-short-decl",
	Summary:  "use var except where := is syntactically required",
	Default:  false,
	Severity: Error,
	Rationale: `One way to introduce a variable instead of several.

The rule does not ban := outright, because := is load-bearing where var is a
syntax error:

	switch t := x.(type) { ... }   // var is not allowed in a switch guard
	for i, v := range xs { ... }   // no var form exists

It reports := only in plain statement position, where var is a drop-in
replacement. Hoisting an if or for header declaration out to a preceding var
is legal but widens the variable's scope. The rule treats those as
load-bearing too. A blanket ban would force adding syntax to the language.
The tool will not do that.`,
	Analyzer: newAnalyzer("no-redundant-short-decl",
		"reject := in plain statement position, where var is a drop-in replacement",
		checkRedundantShortDecl),
})

func checkRedundantShortDecl(c *checkPass) {
	loadBearing := map[ast.Node]bool{}
	mark := func(s ast.Stmt) {
		if a, ok := s.(*ast.AssignStmt); ok {
			loadBearing[a] = true
		}
	}
	c.forEachNode(func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.IfStmt:
			mark(n.Init)
		case *ast.ForStmt:
			mark(n.Init)
		case *ast.SwitchStmt:
			mark(n.Init)
		case *ast.TypeSwitchStmt:
			mark(n.Init)
			mark(n.Assign)
		case *ast.CommClause:
			mark(n.Comm)
		}
		return true
	})
	c.forEachNode(func(n ast.Node) bool {
		a, ok := n.(*ast.AssignStmt)
		if !ok || a.Tok != token.DEFINE || loadBearing[a] {
			return true
		}
		c.reportf(a, "short declaration in plain statement position; use var")
		return true
	})
}

// RuleNoNakedReturn forbids bare returns from a signature with named results.
var RuleNoNakedReturn = register(Rule{
	Name:     "no-naked-return",
	Summary:  "return statements must be explicit even with named results",
	Default:  true,
	Severity: Error,
	Rationale: `A bare return in a function with named results forces the reader to scroll up
to learn the returned values. It also silently returns whatever the result
variables hold at that point. Naming the values costs nothing and survives
later edits to the function body.

The rule checks function literals against their own result list, not the
enclosing function's. It reports a naked return inside a closure when the
closure itself declares named results.`,
	Analyzer: newAnalyzer("no-naked-return",
		"reject bare return statements in a signature with named results",
		checkNakedReturn),
})

func checkNakedReturn(c *checkPass) {
	// walk descends a body, attributing each return to the innermost
	// signature that encloses it.
	var walk func(body *ast.BlockStmt, sig *ast.FuncType, owner string)
	walk = func(body *ast.BlockStmt, sig *ast.FuncType, owner string) {
		if body == nil {
			return
		}
		reportHere := namedResults(sig)
		ast.Inspect(body, func(n ast.Node) bool {
			if lit, ok := n.(*ast.FuncLit); ok {
				// The literal has its own result list. Recurse with it. Do
				// not let the enclosing signature govern its returns.
				walk(lit.Body, lit.Type, "the function literal")
				return false
			}
			if r, ok := n.(*ast.ReturnStmt); ok && len(r.Results) == 0 && reportHere {
				c.reportf(r, "naked return in %s; name the values you are returning", owner)
			}
			return true
		})
	}
	c.forEachFuncDecl(func(fd *ast.FuncDecl) {
		walk(fd.Body, fd.Type, fd.Name.Name)
	})
}

// RuleNoInitFunc forbids package initialisation functions.
var RuleNoInitFunc = register(Rule{
	Name:     "no-init-func",
	Summary:  "func init() is forbidden. initialize explicitly",
	Default:  false,
	Severity: Error,
	Rationale: `init runs before main in an order determined by the import graph, which makes
an initialisation failure hard to localise and hard to test. Explicit wiring
from main puts the order in one readable place.

Off by default because avoiding init entirely requires a wiring convention
that a linter cannot supply.`,
	Analyzer: newAnalyzer("no-init-func",
		"reject func init(). Prefer explicit initialization",
		checkInitFunc),
})

func checkInitFunc(c *checkPass) {
	c.forEachFuncDecl(func(fd *ast.FuncDecl) {
		if fd.Recv == nil && fd.Name.Name == "init" {
			c.reportf(fd, "func init(); initialize explicitly from main instead")
		}
	})
}

// RuleNoEmbeddedField forbids struct embedding.
var RuleNoEmbeddedField = register(Rule{
	Name:     "no-embedded-field",
	Summary:  "struct embedding is forbidden. name your fields",
	Default:  false,
	Severity: Error,
	Rationale: `Embedding promotes methods and fields implicitly, so a type's method set stops
being visible at its declaration. Naming the field costs one selector per
use and makes the delegation explicit.

Off by default because embedding is thoroughly idiomatic Go. Turning this on
is a real break with house Go, not a tidy-up.

Interface embedding is not reported. Composing interfaces from smaller ones
is the language's intended mechanism and has no field to name.`,
	Analyzer: newAnalyzer("no-embedded-field",
		"reject embedded struct fields. Name them",
		checkEmbeddedField),
})

func checkEmbeddedField(c *checkPass) {
	c.forEachNode(func(n ast.Node) bool {
		st, ok := n.(*ast.StructType)
		if !ok || st.Fields == nil {
			return true
		}
		for _, f := range st.Fields.List {
			if len(f.Names) == 0 {
				c.reportf(f, "embedded field; give it a name")
			}
		}
		return true
	})
}

// RuleNoGoto forbids goto.
var RuleNoGoto = register(Rule{
	Name:     "no-goto",
	Summary:  "goto is forbidden",
	Default:  true,
	Severity: Error,
	Rationale: `Labelled break and continue cover the loop cases. A goto that jumps anywhere
else is control flow the reader has to simulate by hand.

This is a rule for application code, not a claim that goto is vestigial. The
standard library's goto statements concentrate in the type checkers, the
compiler, and the syscall exec paths. That code is either performance
critical or a hand-written state machine. If you write that kind of code,
turn this rule off rather than working around it.`,
	Analyzer: newAnalyzer("no-goto", "reject goto statements", checkGoto),
})

func checkGoto(c *checkPass) {
	c.forEachNode(func(n ast.Node) bool {
		b, ok := n.(*ast.BranchStmt)
		if !ok || b.Tok != token.GOTO {
			return true
		}
		label := "a label"
		if b.Label != nil {
			label = b.Label.Name
		}
		c.reportf(b, "goto %s; use labelled break or continue, or restructure the control flow", label)
		return true
	})
}
