package ago

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
)

// TestRules runs every rule against its fixture package. analysistest fails
// the test when an expected diagnostic does not appear. It also fails when an
// unexpected one appears. Each fixture doubles as a false-positive guard.
func TestRules(t *testing.T) {
	tests := []struct {
		rule Rule
		pkg  string
		// minGo is the Go minor version the fixture needs to type-check.
		// A fixture for a rule that reverts a language change cannot compile
		// on a toolchain from before that change. The rule is dormant there
		// anyway: the compiler already enforces it.
		minGo int
	}{
		{rule: RuleNoGoto, pkg: "nogoto"},
		{rule: RuleNoNewExpr, pkg: "nonewexpr", minGo: 26},
		{rule: RuleNoNakedReturn, pkg: "nonakedreturn"},
		{rule: RuleNoSelfReferentialConstraints, pkg: "noselfref", minGo: 26},
		{rule: RuleNoFBoundedConstraints, pkg: "nofbounded"},
		{rule: RuleNoFBoundedConstraints, pkg: "nofboundedselfref", minGo: 26},
		{rule: RuleNoGenericDecls, pkg: "nogenericdecls"},
		{rule: RuleNoDotImport, pkg: "nodotimport"},
		{rule: RuleNoBlankImportOutsideMain, pkg: "noblankimport"},
		{rule: RuleNoEmbeddedField, pkg: "noembedded"},
		{rule: RuleNoInitFunc, pkg: "noinitfunc"},
		{rule: RuleNoRedundantShortDecl, pkg: "noshortdecl"},
	}
	for _, tt := range tests {
		t.Run(tt.pkg, func(t *testing.T) {
			if tt.minGo > 0 && !goAtLeast(tt.minGo) {
				t.Skipf("fixture needs Go 1.%d; %s cannot compile it, so %s is dormant",
					tt.minGo, runtime.Version(), tt.rule.Name)
			}
			analysistest.Run(t, analysistest.TestData(), tt.rule.Analyzer, tt.pkg)
		})
	}
}

// goAtLeast reports whether the toolchain running the tests is at least
// Go 1.<minor>. An unrecognized version string counts as new enough, so a
// development toolchain runs the tests rather than silently skipping them.
func goAtLeast(minor int) bool {
	v := strings.TrimPrefix(runtime.Version(), "go1.")
	if v == runtime.Version() {
		return true
	}
	if i := strings.IndexAny(v, ".rc-+ "); i >= 0 {
		v = v[:i]
	}
	got, err := strconv.Atoi(v)
	if err != nil {
		return true
	}
	return got >= minor
}

// TestNoGenericMethods checks the Go 1.27 rule. The test writes the fixture at
// run time rather than committing it. A toolchain older than Go 1.27 cannot
// parse a generic method and would fail gofmt over a committed file. On such a
// toolchain the language already enforces the rule and the test skips.
func TestNoGenericMethods(t *testing.T) {
	const src = `package nogenericmethods

type Box struct{ v any }

// A method that declares its own type parameters.
func (b *Box) Get[T any]() T { // want ` + "`" + `method Get declares its own type parameters` + "`" + `
	return b.v.(T)
}

type Typed[E any] struct{ v E }

// A method whose receiver carries type parameters declares none of its own.
func (t *Typed[E]) Value() E { return t.v }
`
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "p.go", src, parser.SkipObjectResolution); err != nil {
		t.Skipf("toolchain rejects generic methods, so the rule is dormant: %v", err)
	}

	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "src", "nogenericmethods")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "box.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	analysistest.Run(t, dir, RuleNoGenericMethods.Analyzer, "nogenericmethods")
}

// TestAnalyzersValidate checks the invariants go/analysis itself imposes, so
// that a golangci-lint plugin or a go vet -vettool build cannot fail on
// metadata ago controls.
func TestAnalyzersValidate(t *testing.T) {
	if err := analysis.Validate(append(Analyzers(), ignoresAnalyzer)); err != nil {
		t.Fatal(err)
	}
}

// TestRuleMetadata guards the invariants the command and the config rely on.
func TestRuleMetadata(t *testing.T) {
	seen := map[string]bool{}
	for _, r := range Rules() {
		switch {
		case r.Name == "":
			t.Error("rule with empty name")
		case seen[r.Name]:
			t.Errorf("duplicate rule name %q", r.Name)
		case r.Summary == "":
			t.Errorf("%s: empty summary", r.Name)
		case r.Rationale == "":
			t.Errorf("%s: empty rationale", r.Name)
		case r.Severity == "":
			t.Errorf("%s: empty severity", r.Name)
		case r.Analyzer == nil:
			t.Errorf("%s: nil analyzer", r.Name)
		}
		seen[r.Name] = true
		if r.Analyzer != nil && r.Analyzer.Name != identName(r.Name) {
			t.Errorf("%s: analyzer name %q does not match", r.Name, r.Analyzer.Name)
		}
		if _, ok := Lookup(r.Name); !ok {
			t.Errorf("%s: not findable by canonical name", r.Name)
		}
		if _, ok := Lookup(r.Analyzer.Name); !ok {
			t.Errorf("%s: not findable by analyzer name", r.Name)
		}
	}
}
