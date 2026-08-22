package ago

import "go/ast"

// RuleNoDotImport forbids dot imports.
var RuleNoDotImport = register(Rule{
	Name:     "no-dot-import",
	Summary:  `import . "pkg" is forbidden`,
	Default:  true,
	Severity: Error,
	Rationale: `A dot import makes every identifier in the file ambiguous as to origin, which
defeats both the reader and grep.

The standard library uses dot imports in production code in exactly one
pattern: the two type checkers dot-import a package that holds nothing but
error-code constants, referenced densely enough that qualifying every use
would be noise. That is the narrow case where this rule is wrong. Outside of
it, the standard library's dot imports are confined to tests and to assembly
generators excluded from the build.`,
	Analyzer: newAnalyzer("no-dot-import", `reject import . "pkg"`, checkDotImport),
})

func checkDotImport(c *checkPass) {
	for _, f := range c.Files {
		for _, imp := range f.Imports {
			if imp.Name != nil && imp.Name.Name == "." {
				c.reportf(imp, "dot import of %s; qualify the identifiers instead", imp.Path.Value)
			}
		}
	}
}

// RuleNoBlankImportOutsideMain confines side-effect imports to package main.
var RuleNoBlankImportOutsideMain = register(Rule{
	Name:     "no-blank-import-outside-main",
	Summary:  `import _ "pkg" only allowed in package main`,
	Default:  false,
	Severity: Error,
	Rationale: `A blank import registers a side effect through the import graph, which makes
program behaviour depend on which packages happen to be linked in. Confining
blank imports to package main keeps that dependency explicit and puts it
where a reader looks for wiring.

Off by default because some driver-registration patterns legitimately need a
blank import from a library package.

Files with a _test.go suffix are exempt: a test that blank-imports a driver
is registering it for that binary only.`,
	Analyzer: newAnalyzer("no-blank-import-outside-main",
		`reject import _ "pkg" outside package main`, checkBlankImport),
})

func checkBlankImport(c *checkPass) {
	for _, f := range c.Files {
		if f.Name.Name == "main" {
			continue
		}
		if isTestFile(c, f) {
			continue
		}
		for _, imp := range f.Imports {
			if imp.Name != nil && imp.Name.Name == "_" {
				c.reportf(imp, "blank import of %s outside package main", imp.Path.Value)
			}
		}
	}
}

// isTestFile reports whether a file's name ends in _test.go.
func isTestFile(c *checkPass, f *ast.File) bool {
	name := c.Fset.Position(f.Pos()).Filename
	const suffix = "_test.go"
	return len(name) >= len(suffix) && name[len(name)-len(suffix):] == suffix
}
