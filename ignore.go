package ago

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"strings"
	"sync/atomic"

	"golang.org/x/tools/go/analysis"
)

// Directive prefixes recognised in comments.
const (
	ignoreLinePrefix = "//ago:ignore"
	ignoreFilePrefix = "//ago:ignore-file"
	// reasonSep separates the rule list from the mandatory reason.
	reasonSep = "--"
	// wildcard suppresses every rule.
	wildcard = "*"
)

// An ignoreDirective is one parsed //ago:ignore or //ago:ignore-file comment.
type ignoreDirective struct {
	// Pos locates the comment itself.
	Pos token.Pos
	// File is the file that contains the directive.
	File string
	// Line is the source line the directive suppresses. It is 0 for a
	// file-scoped directive.
	Line int
	// Rules lists the canonical rule names suppressed, or holds a single
	// wildcard entry.
	Rules []string
	// Reason is the text after the "--" separator.
	Reason string
	// Problem names the malformation. It stays empty for a well-formed
	// directive.
	Problem string
	// used records whether any diagnostic was actually suppressed by this
	// directive. Drivers use it to report stale suppressions. Rules run
	// concurrently, so it is atomic.
	used atomic.Bool
}

// suppresses reports whether the directive covers the named rule.
func (d *ignoreDirective) suppresses(rule string) bool {
	if d.Problem != "" {
		return false
	}
	for _, r := range d.Rules {
		if r == wildcard || r == rule {
			return true
		}
	}
	return false
}

// An ignoreIndex holds every directive found in a package, indexed for lookup
// by position.
type ignoreIndex struct {
	// byLine maps a filename to the directives that scope to a single line.
	byLine map[string]map[int][]*ignoreDirective
	// byFile maps a filename to the directives that scope to the whole file.
	byFile map[string][]*ignoreDirective
	// all holds every directive in source order, including malformed ones.
	all []*ignoreDirective
}

// suppressed reports whether a directive covers a diagnostic for rule at pos,
// and marks that directive as used.
func (ix *ignoreIndex) suppressed(fset *token.FileSet, pos token.Pos, rule string) bool {
	if ix == nil {
		return false
	}
	p := fset.Position(pos)
	for _, d := range ix.byFile[p.Filename] {
		if d.suppresses(rule) {
			d.used.Store(true)
			return true
		}
	}
	for _, d := range ix.byLine[p.Filename][p.Line] {
		if d.suppresses(rule) {
			d.used.Store(true)
			return true
		}
	}
	return false
}

// ignoresAnalyzer collects //ago:ignore directives. Every rule analyzer
// requires it so that a rule never has to parse comments itself.
var ignoresAnalyzer = &analysis.Analyzer{
	Name:       "agoignores",
	Doc:        "collect //ago:ignore directives for the other ago analyzers",
	URL:        "https://github.com/agentstation/ago#suppressing-a-rule",
	Run:        runIgnores,
	ResultType: reflectTypeOfIgnoreIndex,
}

func runIgnores(pass *analysis.Pass) (any, error) {
	ix := &ignoreIndex{
		byLine: map[string]map[int][]*ignoreDirective{},
		byFile: map[string][]*ignoreDirective{},
	}
	for _, f := range pass.Files {
		src := readSource(pass, pass.Fset.Position(f.Pos()).Filename)
		for _, group := range f.Comments {
			for _, c := range group.List {
				d := parseDirective(pass.Fset, c, src)
				if d == nil {
					continue
				}
				ix.all = append(ix.all, d)
				if d.Line == 0 {
					ix.byFile[d.File] = append(ix.byFile[d.File], d)
					continue
				}
				lines := ix.byLine[d.File]
				if lines == nil {
					lines = map[int][]*ignoreDirective{}
					ix.byLine[d.File] = lines
				}
				lines[d.Line] = append(lines[d.Line], d)
			}
		}
	}
	return ix, nil
}

// parseDirective turns a comment into a directive, or returns nil when the
// comment is not an ago directive at all. A malformed directive comes back
// with Problem set so that no-invalid-ignore can report it.
func parseDirective(fset *token.FileSet, c *ast.Comment, src []byte) *ignoreDirective {
	text := strings.TrimRight(c.Text, " \t")
	fileScoped := strings.HasPrefix(text, ignoreFilePrefix)
	if !fileScoped && !strings.HasPrefix(text, ignoreLinePrefix) {
		return nil
	}
	prefix := ignoreLinePrefix
	if fileScoped {
		prefix = ignoreFilePrefix
	}
	rest := text[len(prefix):]
	// Require a separator so that "//ago:ignorecase" in prose is not a
	// directive.
	if rest != "" && !strings.HasPrefix(rest, " ") && !strings.HasPrefix(rest, "\t") {
		return nil
	}

	pos := fset.Position(c.Slash)
	d := &ignoreDirective{Pos: c.Slash, File: pos.Filename}
	if !fileScoped {
		d.Line = directiveTargetLine(pos, src)
	}

	body, reason, found := strings.Cut(strings.TrimSpace(rest), reasonSep)
	d.Reason = strings.TrimSpace(reason)
	names := splitList(body)

	d.Rules = names
	d.Problem = directiveProblem(prefix, names, found && d.Reason != "")
	return d
}

// directiveProblem describes the first thing wrong with a directive, or
// returns "" when the directive is well formed. A directive with a problem
// suppresses nothing, so that a broken escape hatch fails loudly.
func directiveProblem(prefix string, names []string, hasReason bool) string {
	if len(names) == 0 {
		return "names no rule; write " + prefix + " <rule> " + reasonSep + " <reason>"
	}
	for _, name := range names {
		if name == wildcard {
			continue
		}
		if _, ok := Lookup(name); !ok {
			return fmt.Sprintf("names unknown rule %q; run \"ago -list\" for the rule set", name)
		}
	}
	if !hasReason {
		return "gives no reason; write " + prefix + " " + strings.Join(names, ",") + " " + reasonSep + " <reason>"
	}
	return ""
}

// directiveTargetLine reports which source line a line-scoped directive
// suppresses. A directive trailing code suppresses that same line. A directive
// on a line of its own suppresses the line below it.
func directiveTargetLine(pos token.Position, src []byte) int {
	if isOwnLineComment(pos, src) {
		return pos.Line + 1
	}
	return pos.Line
}

// isOwnLineComment reports whether only whitespace precedes the comment on its
// line. When the source is unavailable it falls back to the column, which is
// correct for unindented comments and conservative otherwise.
func isOwnLineComment(pos token.Position, src []byte) bool {
	if src == nil {
		return pos.Column == 1
	}
	start := pos.Offset - (pos.Column - 1)
	if start < 0 || pos.Offset > len(src) || start > pos.Offset {
		return pos.Column == 1
	}
	return strings.TrimSpace(string(src[start:pos.Offset])) == ""
}

// splitList parses a comma-separated rule list, tolerating spaces.
func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

// readSource returns the contents of a file through the pass, so editor
// overlays apply, and falls back to the filesystem.
func readSource(pass *analysis.Pass, filename string) []byte {
	if pass.ReadFile != nil {
		if b, err := pass.ReadFile(filename); err == nil {
			return b
		}
	}
	b, err := os.ReadFile(filename) //nolint:gosec // reading the source files under analysis is the point
	if err != nil {
		return nil
	}
	return b
}
