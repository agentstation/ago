package ago

import (
	"go/token"
	"sort"
)

// A Finding is one reported violation, in the form the report writers consume.
// It is deliberately a flat value rather than an [analysis.Diagnostic] so that
// the JSON schema is stable across x/tools upgrades.
type Finding struct {
	// Rule is the canonical kebab-case rule name.
	Rule string `json:"rule"`
	// Severity is the rule's severity.
	Severity Severity `json:"severity"`
	// Message states what is wrong and what to write instead.
	Message string `json:"message"`
	// File is the path as the caller passed it, slash separated.
	File string `json:"file"`
	// Line and Column are 1-based, matching go vet and gofmt.
	Line   int `json:"line"`
	Column int `json:"column"`
	// EndLine and EndColumn bound the offending syntax. They equal Line and
	// Column when the rule reports a point rather than a range.
	EndLine   int `json:"endLine"`
	EndColumn int `json:"endColumn"`
	// DocURL points at the rule's section in the README.
	DocURL string `json:"docURL"`
}

// Position renders the finding in the file:line:col form that editors and
// terminals link on.
func (f Finding) Position() string {
	return token.Position{Filename: f.File, Line: f.Line, Column: f.Column}.String()
}

// sortFindings orders findings by file, then line, then column, then rule, so
// that output is byte-identical across runs and diffable in CI.
func sortFindings(fs []Finding) {
	sort.SliceStable(fs, func(i, j int) bool {
		a, b := fs[i], fs[j]
		switch {
		case a.File != b.File:
			return a.File < b.File
		case a.Line != b.Line:
			return a.Line < b.Line
		case a.Column != b.Column:
			return a.Column < b.Column
		default:
			return a.Rule < b.Rule
		}
	})
}

// dedupe removes findings that are identical in rule and position, which
// happens when the command line names the same package twice.
func dedupe(fs []Finding) []Finding {
	if len(fs) < 2 {
		return fs
	}
	type key struct {
		rule string
		file string
		line int
		col  int
	}
	seen := make(map[key]bool, len(fs))
	out := fs[:0]
	for _, f := range fs {
		k := key{f.Rule, f.File, f.Line, f.Column}
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, f)
	}
	return out
}
