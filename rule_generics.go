package ago

import (
	"go/ast"
	"go/token"
)

// RuleNoGenericMethods forbids the Go 1.27 generic-method feature.
var RuleNoGenericMethods = register(Rule{
	Name:     "no-generic-methods",
	Summary:  "methods may not declare their own type parameters",
	Reverts:  "1.27",
	Default:  true,
	Severity: Error,
	Rationale: `Go 1.27 lets a method declare type parameters of its own. Rewrite any such
method as a package-level function that takes the receiver as its first
argument. The feature buys call-site chaining and costs a second place to
look for a type's operations.

A method whose receiver carries type parameters, such as
func (b *Box[T]) Get() T, is not affected. The rule reports only a method that
introduces new type parameters.

On a toolchain older than Go 1.27 this rule is dormant. The parser rejects the
construct before any analyzer runs.`,
	Analyzer: newAnalyzer("no-generic-methods",
		"reject methods that declare their own type parameters (reverts Go 1.27)",
		checkGenericMethods),
})

func checkGenericMethods(c *checkPass) {
	c.forEachFuncDecl(func(fd *ast.FuncDecl) {
		if fd.Recv == nil || !hasTypeParams(fd.Type.TypeParams) {
			return
		}
		c.reportf(fd.Type.TypeParams,
			"method %s declares its own type parameters; move it to a package-level function",
			fd.Name.Name)
	})
}

// RuleNoGenericDecls forbids type parameters anywhere.
var RuleNoGenericDecls = register(Rule{
	Name:     "no-generic-decls",
	Summary:  "no type parameters on any func or type declaration",
	Reverts:  "1.18",
	Default:  false,
	Severity: Error,
	Rationale: `Reverts generics entirely. This is the strictest rule ago offers. It is the
least likely to be right for a given codebase: it forbids type parameters on
any func or type declaration.

Reasonable for a codebase with no container libraries of its own. Hostile
otherwise. It does not stop you calling generic standard library functions
such as slices.Sort. Recognising a generic call requires resolving the
callee to its declaration in another package.`,
	Analyzer: newAnalyzer("no-generic-decls",
		"reject type parameters on any func or type declaration (reverts Go 1.18)",
		checkGenericDecls),
})

func checkGenericDecls(c *checkPass) {
	for _, f := range c.Files {
		for _, d := range f.Decls {
			switch d := d.(type) {
			case *ast.FuncDecl:
				if hasTypeParams(d.Type.TypeParams) {
					c.reportf(d.Type.TypeParams, "func %s is generic", d.Name.Name)
				}
			case *ast.GenDecl:
				if d.Tok != token.TYPE {
					continue
				}
				for _, s := range d.Specs {
					ts, ok := s.(*ast.TypeSpec)
					if ok && hasTypeParams(ts.TypeParams) {
						c.reportf(ts.TypeParams, "type %s is generic", ts.Name.Name)
					}
				}
			}
		}
	}
}
