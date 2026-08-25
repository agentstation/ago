package ago

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/checker"
	"golang.org/x/tools/go/packages"
)

// Options control one [Check] run.
type Options struct {
	// Dir is the directory ago resolves patterns against. An empty Dir means
	// the process working directory.
	Dir string
	// Patterns are go/packages patterns such as "./..." or a list of .go
	// files. An empty Patterns means "./...".
	Patterns []string
	// Rules are the rules to run. An empty Rules means the default set.
	Rules []Rule
	// Tests reports whether ago analyzes _test.go files.
	Tests bool
	// Config supplies exclude patterns. It may be nil.
	Config *Config
	// ReportStaleIgnores reports //ago:ignore directives that suppressed
	// nothing.
	ReportStaleIgnores bool
}

// Check loads the named packages and runs the selected rules over them.
//
// A load or type error does not abort the run. Check collects errors into
// [Report.Errors] and analysis continues over whatever parsed. A single
// unbuildable file cannot hide every finding in the repository.
func Check(opts Options) (*Report, error) {
	rules := opts.Rules
	if len(rules) == 0 {
		rules = (&Config{}).Enabled(nil)
	}
	patterns := opts.Patterns
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}

	report := &Report{
		Version:      Version,
		Rules:        sortedNames(rules),
		Findings:     []Finding{},
		StaleIgnores: []StaleIgnore{},
		Errors:       []string{},
	}

	cfg := &packages.Config{
		// checker.Analyze requires typed syntax for the initial packages and
		// all their dependencies. The mode includes NeedModule so a driver can
		// tell which module a finding belongs to.
		Mode:  packages.LoadAllSyntax | packages.NeedModule,
		Dir:   opts.Dir,
		Tests: opts.Tests,
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, fmt.Errorf("loading packages: %w", err)
	}
	report.Errors = append(report.Errors, loadErrors(pkgs)...)

	loaded := pkgs
	pkgs = keepAnalyzable(pkgs, opts.Config)
	if len(pkgs) == 0 {
		if excludedAll(loaded, opts.Config) {
			report.Errors = append(report.Errors, "all packages excluded by config")
		}
		return report, nil
	}

	analyzers := make([]*analysis.Analyzer, 0, len(rules))
	byAnalyzer := make(map[string]Rule, len(rules))
	for _, r := range rules {
		analyzers = append(analyzers, r.Analyzer)
		byAnalyzer[r.Analyzer.Name] = r
	}
	if opts.ReportStaleIgnores {
		// The driver only retains an action's Result when the action is a
		// root. Request the ignore index explicitly. Do not reach it through
		// the rules that depend on it.
		analyzers = append(analyzers, ignoresAnalyzer)
	}

	graph, err := checker.Analyze(analyzers, pkgs, nil)
	if err != nil {
		return nil, fmt.Errorf("running analyzers: %w", err)
	}

	for act := range graph.All() {
		// An action error in a package that already failed to load is a
		// cascade of the load error. Check reports that load error once above.
		// Reporting it again per analyzer buries the root cause.
		if act.Err != nil && len(act.Package.Errors) == 0 {
			report.Errors = append(report.Errors,
				fmt.Sprintf("%s: %v", act.Analyzer.Name, act.Err))
		}
		rule, ok := byAnalyzer[act.Analyzer.Name]
		if !ok {
			continue
		}
		for _, d := range act.Diagnostics {
			report.Findings = append(report.Findings,
				toFinding(act.Package, rule, d, opts.Dir))
		}
	}

	if opts.ReportStaleIgnores {
		report.StaleIgnores = collectStaleIgnores(graph, opts.Dir)
	}

	sortFindings(report.Findings)
	report.Findings = dedupe(report.Findings)
	sort.Slice(report.StaleIgnores, func(i, j int) bool {
		a, b := report.StaleIgnores[i], report.StaleIgnores[j]
		if a.File != b.File {
			return a.File < b.File
		}
		return a.Line < b.Line
	})
	return report, nil
}

// toFinding converts an analysis diagnostic into ago's flat finding form,
// making the path relative to the run directory so that output is stable
// across machines.
func toFinding(pkg *packages.Package, rule Rule, d analysis.Diagnostic, dir string) Finding {
	start := pkg.Fset.Position(d.Pos)
	end := start
	if d.End.IsValid() {
		end = pkg.Fset.Position(d.End)
	}
	return Finding{
		Rule:      rule.Name,
		Severity:  rule.Severity,
		Message:   d.Message,
		File:      relPath(start.Filename, dir),
		Line:      start.Line,
		Column:    start.Column,
		EndLine:   end.Line,
		EndColumn: end.Column,
		DocURL:    rule.DocURL(),
	}
}

// collectStaleIgnores finds every //ago:ignore directive that suppressed
// nothing during the run.
func collectStaleIgnores(graph *checker.Graph, dir string) []StaleIgnore {
	out := []StaleIgnore{}
	seen := map[*ignoreDirective]bool{}
	for act := range graph.All() {
		if act.Analyzer != ignoresAnalyzer {
			continue
		}
		ix, ok := act.Result.(*ignoreIndex)
		if !ok {
			continue
		}
		for _, d := range ix.all {
			if seen[d] || d.used.Load() || d.Problem != "" {
				continue
			}
			seen[d] = true
			pos := act.Package.Fset.Position(d.Pos)
			out = append(out, StaleIgnore{
				Rules:  d.Rules,
				Reason: d.Reason,
				File:   relPath(pos.Filename, dir),
				Line:   pos.Line,
				Column: pos.Column,
			})
		}
	}
	return out
}

// keepAnalyzable drops packages that have no syntax to analyze and packages
// the config excludes.
func keepAnalyzable(pkgs []*packages.Package, cfg *Config) []*packages.Package {
	out := pkgs[:0]
	for _, p := range pkgs {
		if len(p.Syntax) == 0 {
			continue
		}
		if cfg != nil && len(p.CompiledGoFiles) > 0 && cfg.Skip(p.CompiledGoFiles[0]) {
			continue
		}
		out = append(out, p)
	}
	return out
}

// excludedAll reports whether Config.Skip dropped every package that had
// syntax to analyze. A config that excludes the whole tree would otherwise
// exit 0 with an empty report.
func excludedAll(pkgs []*packages.Package, cfg *Config) bool {
	if cfg == nil {
		return false
	}
	saw := false
	for _, p := range pkgs {
		if len(p.Syntax) == 0 {
			continue
		}
		saw = true
		if len(p.CompiledGoFiles) == 0 || !cfg.Skip(p.CompiledGoFiles[0]) {
			return false
		}
	}
	return saw
}

// loadErrors flattens package load and type errors into readable strings,
// deduplicated because go/packages repeats a dependency's errors in every
// package that imports it.
func loadErrors(pkgs []*packages.Package) []string {
	var out []string
	seen := map[string]bool{}
	packages.Visit(pkgs, nil, func(p *packages.Package) {
		for _, e := range p.Errors {
			msg := e.Error()
			if seen[msg] {
				continue
			}
			seen[msg] = true
			out = append(out, msg)
		}
	})
	sort.Strings(out)
	return out
}

// relPath makes a path relative to dir when that is shorter and still points
// inside the tree, and always returns slash-separated output.
func relPath(path, dir string) string {
	if dir == "" {
		var err error
		dir, err = filepath.Abs(".")
		if err != nil {
			return filepath.ToSlash(path)
		}
	} else if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	rel, err := filepath.Rel(dir, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}
