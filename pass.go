package ago

import (
	"go/ast"
	"reflect"

	"golang.org/x/tools/go/analysis"
)

// reflectTypeOfIgnoreIndex is the ResultType of ignoresAnalyzer.
var reflectTypeOfIgnoreIndex = reflect.TypeOf((*ignoreIndex)(nil))

// newAnalyzer builds the analyzer for a rule. Every rule analyzer requires
// ignoresAnalyzer, so that suppression is enforced in one place rather than
// re-implemented per rule.
func newAnalyzer(name, doc string, run func(*checkPass)) *analysis.Analyzer {
	a := &analysis.Analyzer{
		Name:     identName(name),
		Doc:      doc,
		URL:      "https://github.com/agentstation/ago#" + name,
		Requires: []*analysis.Analyzer{ignoresAnalyzer},
	}
	a.Run = func(pass *analysis.Pass) (any, error) {
		ix, _ := pass.ResultOf[ignoresAnalyzer].(*ignoreIndex)
		run(&checkPass{Pass: pass, rule: name, ignores: ix})
		return nil, nil
	}
	return a
}

// A checkPass is the analysis pass a rule sees. It narrows [analysis.Pass] to
// the reporting path that honours //ago:ignore, so that a rule cannot report
// around suppression by accident.
type checkPass struct {
	*analysis.Pass
	rule    string
	ignores *ignoreIndex
}

// reportf records a violation at the position of n unless an //ago:ignore
// directive covers that line for this rule.
func (c *checkPass) reportf(n ast.Node, format string, args ...any) {
	if c.ignores.suppressed(c.Fset, n.Pos(), c.rule) {
		return
	}
	// Reported through the embedded Pass explicitly: this is the single place
	// allowed to bypass reportf, and naming it keeps that visible.
	c.Pass.Report(analysis.Diagnostic{ //nolint:staticcheck // QF1008: the qualifier is deliberate
		Pos:      n.Pos(),
		End:      n.End(),
		Category: c.rule,
		Message:  sprintf(format, args...),
	})
}

// forEachFuncDecl calls fn for every top-level function and method
// declaration in the package under analysis.
func (c *checkPass) forEachFuncDecl(fn func(*ast.FuncDecl)) {
	for _, f := range c.Files {
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok {
				fn(fd)
			}
		}
	}
}

// forEachNode calls fn for every node in the package under analysis. Returning
// false from fn skips the node's children, matching [ast.Inspect].
func (c *checkPass) forEachNode(fn func(ast.Node) bool) {
	for _, f := range c.Files {
		ast.Inspect(f, fn)
	}
}

// namedResults reports whether a signature declares named result parameters,
// which is the precondition for a naked return.
func namedResults(t *ast.FuncType) bool {
	if t.Results == nil {
		return false
	}
	for _, f := range t.Results.List {
		if len(f.Names) > 0 {
			return true
		}
	}
	return false
}

// hasTypeParams reports whether a type parameter list declares anything.
func hasTypeParams(tp *ast.FieldList) bool {
	return tp != nil && len(tp.List) > 0
}
