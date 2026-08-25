package ago

import (
	"path/filepath"
	"strings"
	"testing"
)

// findingSet renders a report's findings as comparable strings.
func findingSet(t *testing.T, r *Report) []string {
	t.Helper()
	out := make([]string, 0, len(r.Findings))
	for _, f := range r.Findings {
		out = append(out, f.Rule+"@"+f.Position())
	}
	return out
}

func checkModule(t *testing.T, name string, opts Options) *Report {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("testdata", "modules", name))
	if err != nil {
		t.Fatal(err)
	}
	opts.Dir = dir
	if len(opts.Patterns) == 0 {
		opts.Patterns = []string{"./..."}
	}
	report, err := Check(opts)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	return report
}

// TestIgnoreDirectives exercises suppression end to end through the real
// driver. The ignore index and the rules interact only once both analyzers
// run together.
func TestIgnoreDirectives(t *testing.T) {
	gotoRule, _ := Lookup("no-goto")
	invalid, _ := Lookup("no-invalid-ignore")
	report := checkModule(t, "ignores", Options{
		Rules:              []Rule{gotoRule, invalid},
		ReportStaleIgnores: true,
	})

	got := strings.Join(findingSet(t, report), "\n")
	// The rule suppresses Good(). The four malformed directives suppress
	// nothing. Each yields both an invalid-ignore finding and the goto it
	// failed to suppress.
	want := []string{
		"no-invalid-ignore@ignores.go:12:2",
		"no-goto@ignores.go:13:2",
		"no-invalid-ignore@ignores.go:19:2",
		"no-goto@ignores.go:20:2",
		"no-invalid-ignore@ignores.go:26:2",
		"no-goto@ignores.go:27:2",
		"no-invalid-ignore@ignores.go:33:2",
		"no-goto@ignores.go:34:2",
	}
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("missing finding %s\ngot:\n%s", w, got)
		}
	}
	// The suppressed goto in Good() must not appear.
	if strings.Contains(got, "no-goto@ignores.go:6:") {
		t.Errorf("directive with a reason failed to suppress\ngot:\n%s", got)
	}
	// Prose beginning with the prefix is not a directive.
	if strings.Contains(got, "ignores.go:39") {
		t.Errorf("prose treated as a directive\ngot:\n%s", got)
	}

	if len(report.StaleIgnores) != 1 {
		t.Fatalf("stale ignores = %d, want 1: %+v", len(report.StaleIgnores), report.StaleIgnores)
	}
	stale := report.StaleIgnores[0]
	if stale.Rules[0] != "no-dot-import" {
		t.Errorf("stale rule = %q, want no-dot-import", stale.Rules[0])
	}
	if stale.Reason == "" {
		t.Error("stale ignore lost its reason")
	}
}

// TestStaleIgnoresOffByDefault confirms that stale reporting is opt-in, so a
// run that does not ask for it cannot fail on a suppression.
func TestStaleIgnoresOffByDefault(t *testing.T) {
	gotoRule, _ := Lookup("no-goto")
	report := checkModule(t, "ignores", Options{Rules: []Rule{gotoRule}})
	if len(report.StaleIgnores) != 0 {
		t.Errorf("stale ignores reported without opting in: %+v", report.StaleIgnores)
	}
}

// TestDeterministicOrder checks that two runs over the same tree produce
// byte-identical output, which CI diffing and agent parsing both depend on.
func TestDeterministicOrder(t *testing.T) {
	gotoRule, _ := Lookup("no-goto")
	invalid, _ := Lookup("no-invalid-ignore")
	opts := Options{Rules: []Rule{gotoRule, invalid}}
	first := strings.Join(findingSet(t, checkModule(t, "ignores", opts)), "\n")
	for i := range 3 {
		got := strings.Join(findingSet(t, checkModule(t, "ignores", opts)), "\n")
		if got != first {
			t.Fatalf("run %d differed:\nfirst:\n%s\ngot:\n%s", i, first, got)
		}
	}
}

// TestDuplicatePatterns confirms that naming the same package twice does not
// double-report, which the pre-analysis driver did.
func TestDuplicatePatterns(t *testing.T) {
	gotoRule, _ := Lookup("no-goto")
	once := checkModule(t, "ignores", Options{
		Rules:    []Rule{gotoRule},
		Patterns: []string{"./..."},
	})
	twice := checkModule(t, "ignores", Options{
		Rules:    []Rule{gotoRule},
		Patterns: []string{"./...", ".", "./..."},
	})
	if len(once.Findings) != len(twice.Findings) {
		t.Errorf("findings: once=%d twice=%d; duplicate patterns must dedupe",
			len(once.Findings), len(twice.Findings))
	}
}

// TestBrokenFileDoesNotHideFindings is the regression test for a driver that
// exited on the first parse error and discarded everything it had already
// found.
func TestBrokenFileDoesNotHideFindings(t *testing.T) {
	gotoRule, _ := Lookup("no-goto")
	report := checkModule(t, "broken", Options{Rules: []Rule{gotoRule}})
	if len(report.Findings) == 0 {
		t.Error("a broken file hid the findings in every other file")
	}
	if len(report.Errors) == 0 {
		t.Error("the parse error was not reported")
	}
	for _, f := range report.Findings {
		if strings.Contains(f.File, "broken") {
			t.Errorf("reported a finding inside the unparseable file: %s", f.Position())
		}
	}
}
