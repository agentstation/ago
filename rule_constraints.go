package ago

import (
	"go/ast"
	"go/token"
)

// RuleNoSelfReferentialConstraints forbids the Go 1.26 self-reference.
var RuleNoSelfReferentialConstraints = register(Rule{
	Name:     "no-self-referential-constraints",
	Summary:  "a generic type may not name itself in its own type parameter list",
	Reverts:  "1.26",
	Default:  true,
	Severity: Error,
	Rationale: `Go 1.26 lifted the restriction that a generic type may not refer to itself in
its type parameter list, so type Adder[A Adder[A]] interface{ Add(A) A } now
compiles. Before Go 1.26 the compiler rejected it as "invalid recursive
type".

This rule reports exactly that construct: the declared type's own name
appearing inside its own type parameter list. It does not report F-bounded
constraints written through a separate interface, such as
type Node[T Cloneable[T]], which has been legal since Go 1.18 and is not
what Go 1.26 changed. Use no-f-bounded-constraints for those.

Detection is syntactic and catches direct self-reference. Mutual recursion
across two declarations is not reported.`,
	Analyzer: newAnalyzer("no-self-referential-constraints",
		"reject a generic type that names itself in its own type parameter list (reverts Go 1.26)",
		checkSelfReferentialConstraints),
})

func checkSelfReferentialConstraints(c *checkPass) {
	eachGenericTypeSpec(c, func(ts *ast.TypeSpec) {
		for _, f := range ts.TypeParams.List {
			ast.Inspect(f.Type, func(n ast.Node) bool {
				id, ok := n.(*ast.Ident)
				if ok && id.Name == ts.Name.Name {
					c.reportf(id,
						"type %s refers to itself in its own type parameter list; break the cycle with a separate constraint interface",
						ts.Name.Name)
				}
				return true
			})
		}
	})
}

// RuleNoFBoundedConstraints forbids F-bounded polymorphism.
var RuleNoFBoundedConstraints = register(Rule{
	Name:     "no-f-bounded-constraints",
	Summary:  "a type parameter may not appear in its own constraint",
	Default:  false,
	Severity: Error,
	Rationale: `Reports F-bounded polymorphism: a type parameter that appears inside its own
constraint, as in type Node[T Cloneable[T]] or func algo[A Adder[A]](x A) A.

This is not a version revert. The construct has been legal since generics
landed in Go 1.18, and the standard library uses it. It is the most
abstraction-dense shape the type system allows, and it is usually reachable
by an ordinary interface plus a concrete type, so some codebases choose to
ban it. That is a house-style decision, which is why the rule is off by
default.

A constraint that refers to a sibling type parameter, such as
type Graph[N any, E Edge[N]], is not reported: the parameter does not appear
in its own constraint.`,
	Analyzer: newAnalyzer("no-f-bounded-constraints",
		"reject a type parameter that appears in its own constraint",
		checkFBoundedConstraints),
})

func checkFBoundedConstraints(c *checkPass) {
	report := func(owner string, ts *ast.TypeSpec, tp *ast.FieldList) {
		// A declaration already reported by no-self-referential-constraints
		// is not reported twice.
		selfNamed := ts != nil && namesIdent(tp, ts.Name.Name)
		for _, f := range tp.List {
			for _, name := range f.Names {
				ast.Inspect(f.Type, func(n ast.Node) bool {
					id, ok := n.(*ast.Ident)
					if !ok || id.Name != name.Name {
						return true
					}
					if selfNamed {
						return true
					}
					c.reportf(id,
						"type parameter %s of %s appears in its own constraint; use a plain interface and a concrete type",
						name.Name, owner)
					return true
				})
			}
		}
	}
	eachGenericTypeSpec(c, func(ts *ast.TypeSpec) {
		report("type "+ts.Name.Name, ts, ts.TypeParams)
	})
	c.forEachFuncDecl(func(fd *ast.FuncDecl) {
		if hasTypeParams(fd.Type.TypeParams) {
			report("func "+fd.Name.Name, nil, fd.Type.TypeParams)
		}
	})
}

// eachGenericTypeSpec calls fn for every type declaration that has a type
// parameter list.
func eachGenericTypeSpec(c *checkPass, fn func(*ast.TypeSpec)) {
	for _, f := range c.Files {
		for _, d := range f.Decls {
			gd, ok := d.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, s := range gd.Specs {
				ts, ok := s.(*ast.TypeSpec)
				if ok && hasTypeParams(ts.TypeParams) {
					fn(ts)
				}
			}
		}
	}
}

// namesIdent reports whether name appears anywhere in a type parameter list's
// constraints.
func namesIdent(tp *ast.FieldList, name string) bool {
	found := false
	for _, f := range tp.List {
		ast.Inspect(f.Type, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && id.Name == name {
				found = true
			}
			return !found
		})
	}
	return found
}
