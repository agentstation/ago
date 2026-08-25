package ago

import (
	"go/ast"
	"go/types"
)

// RuleNoNewExpr forbids the Go 1.26 new(expression) form.
var RuleNoNewExpr = register(Rule{
	Name:     "no-new-expr",
	Summary:  "new() takes a type, not an expression",
	Reverts:  "1.26",
	Default:  true,
	Severity: Error,
	Rationale: `Go 1.26 lets the built-in new take an expression that supplies the initial
value, so new(yearsSince(born)) allocates and initialises in one step.

The form collapses declaration, initialisation, and address-taking into a
single expression. The two-line version with a named variable says the same
thing and reads left to right:

	age := yearsSince(born)
	p := &age

When type information is available the rule decides exactly whether the
argument is a type. Without it the rule falls back to syntax. It reports only
unambiguous expressions such as new(f(x)). A shadowed new or a variable that
looks like a type name is not reported.`,
	Analyzer: newAnalyzer("no-new-expr",
		"reject new() applied to an expression rather than a type (reverts Go 1.26)",
		checkNewExpr),
})

func checkNewExpr(c *checkPass) {
	c.forEachNode(func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		id, ok := call.Fun.(*ast.Ident)
		if !ok || id.Name != "new" || len(call.Args) != 1 {
			return true
		}
		if !c.isBuiltinNew(id) {
			return true
		}
		if c.argIsType(call.Args[0]) {
			return true
		}
		c.reportf(call,
			"new() applied to an expression; declare a variable and take its address")
		return true
	})
}

// isBuiltinNew reports whether the identifier resolves to the predeclared new.
// Without type information it assumes it does, which matches the behaviour of
// a syntax-only run.
func (c *checkPass) isBuiltinNew(id *ast.Ident) bool {
	if c.TypesInfo == nil {
		return true
	}
	obj, ok := c.TypesInfo.Uses[id].(*types.Builtin)
	if !ok {
		// An unresolved identifier means incomplete type information rather
		// than a shadowed new, so fall back to the syntactic assumption.
		return c.TypesInfo.Uses[id] == nil
	}
	return obj.Name() == "new"
}

// argIsType reports whether the sole argument to new denotes a type.
func (c *checkPass) argIsType(arg ast.Expr) bool {
	if c.TypesInfo != nil {
		if tv, ok := c.TypesInfo.Types[arg]; ok {
			return tv.IsType()
		}
	}
	return looksLikeType(arg)
}

// looksLikeType is the syntax-only approximation used when the run has no
// type information. It errs toward permitting anything that could name a type,
// so a run without types under-reports rather than producing false positives.
func looksLikeType(e ast.Expr) bool {
	switch e := e.(type) {
	case *ast.Ident:
		return true
	case *ast.SelectorExpr:
		_, ok := e.X.(*ast.Ident)
		return ok
	case *ast.ParenExpr:
		return looksLikeType(e.X)
	case *ast.StarExpr:
		return looksLikeType(e.X)
	case *ast.ArrayType, *ast.MapType, *ast.ChanType, *ast.StructType,
		*ast.InterfaceType, *ast.FuncType:
		return true
	case *ast.IndexExpr: // generic instantiation, such as new(Box[int])
		return looksLikeType(e.X)
	case *ast.IndexListExpr:
		return looksLikeType(e.X)
	}
	return false
}
