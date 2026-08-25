package ago

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Severity classifies how strongly ago objects to a construct. Every rule
// currently reports at [Error]. The field exists so that report formats with
// a severity axis, such as SARIF, carry an honest value rather than a
// hardcoded one.
type Severity string

// Severity levels.
const (
	// Error marks a construct the rule forbids outright.
	Error Severity = "error"
	// Warning marks a construct the rule discourages.
	Warning Severity = "warning"
)

// A Rule is one restriction, paired with the analyzer that enforces it.
//
// Name is the canonical kebab-case name used by the ago command, by .ago.yml,
// and by //ago:ignore directives. The Analyzer.Name field is the same name
// with the hyphens removed, because go/analysis requires analyzer names to be
// valid Go identifiers. The ago command accepts either spelling.
type Rule struct {
	// Name is the canonical kebab-case rule name, such as "no-goto".
	Name string
	// Summary is a single line shown by "ago -list".
	Summary string
	// Rationale explains why the rule exists and when to turn it off. It is
	// the analyzer's Doc body and the text an agent reads to decide whether a
	// violation is worth fixing or worth ignoring.
	Rationale string
	// Default reports whether ago enables the rule when no configuration
	// selects a rule set explicitly.
	Default bool
	// Reverts names the Go release that introduced the construct this rule
	// forbids, or "" when the construct is not tied to one release.
	Reverts string
	// Severity is the level reported for violations of this rule.
	Severity Severity
	// Analyzer enforces the rule.
	Analyzer *analysis.Analyzer
}

// DocURL returns the anchor in the README that documents the rule.
func (r Rule) DocURL() string {
	return "https://github.com/agentstation/ago#" + r.Name
}

// registry holds every rule in registration order.
var registry []Rule

// register adds a rule to the registry and wires up its analyzer metadata.
// Each rule file calls it at package level. It panics on a duplicate name
// because that is a programming error in ago itself, not a condition a caller
// can handle.
func register(r Rule) Rule {
	for _, existing := range registry {
		if existing.Name == r.Name {
			panic(fmt.Sprintf("ago: duplicate rule %q", r.Name))
		}
	}
	registry = append(registry, r)
	return r
}

// identName converts a canonical kebab-case rule name into the valid Go
// identifier that go/analysis requires for an analyzer name.
func identName(name string) string {
	return strings.ReplaceAll(name, "-", "")
}

// Rules returns every rule ago knows about, in registration order.
func Rules() []Rule {
	out := make([]Rule, len(registry))
	copy(out, registry)
	return out
}

// Lookup finds a rule by its canonical kebab-case name or by its analyzer
// identifier spelling.
func Lookup(name string) (Rule, bool) {
	for _, r := range registry {
		if r.Name == name || r.Analyzer.Name == name {
			return r, true
		}
	}
	return Rule{}, false
}

// Analyzers returns the analyzer for every rule, for use with multichecker,
// unitchecker, or a golangci-lint module plugin. Callers that want only the
// default set should filter with [Rule.Default].
func Analyzers() []*analysis.Analyzer {
	out := make([]*analysis.Analyzer, 0, len(registry))
	for _, r := range registry {
		out = append(out, r.Analyzer)
	}
	return out
}

// DefaultNames returns the canonical names of the rules that are on when no
// configuration selects a rule set, sorted.
func DefaultNames() []string {
	var out []string
	for _, r := range registry {
		if r.Default {
			out = append(out, r.Name)
		}
	}
	sort.Strings(out)
	return out
}

// Names returns the canonical names of every rule, sorted.
func Names() []string {
	out := make([]string, 0, len(registry))
	for _, r := range registry {
		out = append(out, r.Name)
	}
	sort.Strings(out)
	return out
}
