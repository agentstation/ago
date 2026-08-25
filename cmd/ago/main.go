// Command ago enforces one way to write Go across a codebase.
//
// A project selects the Go constructs that it accepts. Developers and coding
// agents use the same rule policy, and CI enforces it.
//
// ago only ever rejects language constructs. It never adds syntax, never
// rewrites code, and never changes semantics. Code that passes ago is
// ordinary Go that builds with the stock toolchain.
//
// Usage:
//
//	ago [flags] [packages]
//
// The package arguments are go/packages patterns, the same ones go build and
// go vet accept. With no arguments ago checks ./... under the working
// directory.
//
// Flags:
//
//	-list             List the rule set and exit.
//
//	-explain rule     Print a rule's full rationale and exit.
//
//	-rules a,b,c      Run these rules, overriding the config file.
//
//	-all              Run every rule.
//
//	-tests            Also check _test.go files.
//
//	-format f         text, json, sarif, or github (default text).
//
//	-config path      Read this config instead of searching for .ago.yml.
//
//	-no-config        Ignore any .ago.yml on disk.
//
//	-init             Write an optional .ago.yml at the project root.
//
//	-stale-ignores    Report //ago:ignore directives that suppressed nothing.
//
//	-version          Print the version and exit.
//
// Exit status:
//
//	0  No violations.
//	1  At least one violation.
//	2  ago could not complete the run.
//
// A load or type error does not by itself abort the run. The command reports
// what it could analyze and exits 2 only when it produced no usable result.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/agentstation/ago"
)

// Exit statuses. They are part of the command's contract with CI and with
// coding agents, so they have names rather than inline numbers.
const (
	exitClean      = 0
	exitViolations = 1
	exitError      = 2
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("ago", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		listFlag    = fs.Bool("list", false, "list the rule set and exit")
		explainFlag = fs.String("explain", "", "print a rule's full rationale and exit")
		rulesFlag   = fs.String("rules", "", "comma-separated rules to run, overriding the config file")
		allFlag     = fs.Bool("all", false, "run every rule")
		testsFlag   = fs.Bool("tests", false, "also check _test.go files")
		formatFlag  = fs.String("format", "text", "output format: text, json, sarif, or github")
		configFlag  = fs.String("config", "", "read this config instead of searching for .ago.yml")
		noConfig    = fs.Bool("no-config", false, "ignore any .ago.yml on disk")
		initFlag    = fs.Bool("init", false, "write an optional .ago.yml at the project root and exit")
		staleFlag   = fs.Bool("stale-ignores", false, "report //ago:ignore directives that suppressed nothing")
		versionFlag = fs.Bool("version", false, "print the version and exit")
	)
	fs.Usage = func() { usage(stdout, fs) }
	if err := fs.Parse(args); err != nil {
		// -h is a request that succeeded, not a failed run.
		if errors.Is(err, flag.ErrHelp) {
			return exitClean
		}
		return exitError
	}

	switch {
	case *versionFlag:
		fmt.Fprintf(stdout, "ago %s\n", ago.Version)
		return exitClean
	case *initFlag:
		return writeInitConfig(".", stdout, stderr)
	case *explainFlag != "":
		return explain(stdout, stderr, *explainFlag)
	}

	format, err := ago.ParseFormat(*formatFlag)
	if err != nil {
		fmt.Fprintf(stderr, "ago: %v\n", err)
		return exitError
	}

	cfg := &ago.Config{}
	if !*noConfig {
		cfg, err = ago.LoadConfig(".", *configFlag)
		if err != nil {
			fmt.Fprintf(stderr, "ago: %v\n", err)
			return exitError
		}
	}

	overrides, err := ruleOverrides(*rulesFlag, *allFlag)
	if err != nil {
		fmt.Fprintf(stderr, "ago: %v\n", err)
		return exitError
	}
	rules := cfg.Enabled(overrides)
	if len(rules) == 0 {
		fmt.Fprintln(stderr, "ago: no rules enabled; check the rule flags or policy config")
		return exitError
	}

	if *listFlag {
		listRules(stdout, format, rules, resolvedPolicy(cfg, overrides, *noConfig, *testsFlag))
		return exitClean
	}

	report, err := ago.Check(ago.Options{
		Patterns:           fs.Args(),
		Rules:              rules,
		Tests:              *testsFlag || cfg.Tests,
		Config:             cfg,
		ReportStaleIgnores: *staleFlag,
	})
	if err != nil {
		fmt.Fprintf(stderr, "ago: %v\n", err)
		return exitError
	}
	if err := report.Write(stdout, format); err != nil {
		fmt.Fprintf(stderr, "ago: writing report: %v\n", err)
		return exitError
	}

	// Load errors go to stderr so that they never contaminate a machine-read
	// stdout, and are already carried in the JSON and SARIF documents.
	if format == ago.FormatText || format == ago.FormatGitHub {
		for _, e := range report.Errors {
			fmt.Fprintf(stderr, "ago: %s\n", e)
		}
	}

	switch {
	case len(report.Findings) > 0 || len(report.StaleIgnores) > 0:
		if format == ago.FormatText {
			fmt.Fprintf(stderr, "\n%s\n", summary(report))
		}
		return exitViolations
	case len(report.Errors) > 0 && len(report.Rules) > 0 && reportedNothing(report):
		// Nothing was analyzable. Exiting 0 here would report a clean tree
		// that was never actually checked.
		fmt.Fprintln(stderr, "ago: no package could be analyzed")
		return exitError
	default:
		return exitClean
	}
}

// reportedNothing reports whether the run produced no findings at all, which
// combined with load errors means the run was not meaningful.
func reportedNothing(r *ago.Report) bool {
	return len(r.Findings) == 0 && len(r.StaleIgnores) == 0
}

// summary renders the one-line tally printed after text output.
func summary(r *ago.Report) string {
	var parts []string
	if n := len(r.Findings); n > 0 {
		parts = append(parts, fmt.Sprintf("%d violation%s", n, plural(n)))
	}
	if n := len(r.StaleIgnores); n > 0 {
		parts = append(parts, fmt.Sprintf("%d stale ignore%s", n, plural(n)))
	}
	return strings.Join(parts, ", ")
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// ruleOverrides turns the -rules and -all flags into an explicit rule list.
func ruleOverrides(rulesFlag string, all bool) ([]string, error) {
	if all {
		if rulesFlag != "" {
			return nil, fmt.Errorf("-all and -rules are mutually exclusive")
		}
		return []string{"all"}, nil
	}
	if rulesFlag == "" {
		return nil, nil
	}
	var out []string
	for _, name := range strings.Split(rulesFlag, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if name == "all" || name == "default" {
			out = append(out, name)
			continue
		}
		r, ok := ago.Lookup(name)
		if !ok {
			return nil, fmt.Errorf("unknown rule %q; run \"ago -list\" for the rule set", name)
		}
		out = append(out, r.Name)
	}
	return out, nil
}

// listRules prints the rule set. In text form an asterisk marks the enabled
// rules, so "ago -list" doubles as a check on what the current config
// resolves to.
func listRules(stdout io.Writer, format ago.Format, enabled []ago.Rule, policy policyJSON) {
	on := map[string]bool{}
	for _, r := range enabled {
		on[r.Name] = true
	}
	if format == ago.FormatJSON {
		writeRulesJSON(stdout, on, policy)
		return
	}
	source := map[string]string{
		"built-in": "built-in defaults",
		"config":   "config file",
		"flags":    "command-line flags",
	}[policy.RuleSource]
	fmt.Fprintf(stdout, "Policy: %s\n", source)
	switch {
	case policy.ConfigDisabled:
		fmt.Fprintln(stdout, "Config: disabled by -no-config")
	case policy.ConfigPath != "":
		fmt.Fprintf(stdout, "Config: %s\n", policy.ConfigPath)
	default:
		fmt.Fprintln(stdout, "Config: none found")
	}
	fmt.Fprintf(stdout, "Tests:  %t\nExcludes: %d\n\n", policy.Tests, len(policy.Exclude))
	width := 0
	for _, r := range ago.Rules() {
		if len(r.Name) > width {
			width = len(r.Name)
		}
	}
	for _, r := range ago.Rules() {
		mark := " "
		if on[r.Name] {
			mark = "*"
		}
		reverts := ""
		if r.Reverts != "" {
			reverts = "  [reverts Go " + r.Reverts + "]"
		}
		fmt.Fprintf(stdout, "%s %-*s  %s%s\n", mark, width, r.Name, r.Summary, reverts)
	}
	fmt.Fprintf(stdout, "\n* = enabled for this run (%d of %d)\n", len(enabled), len(ago.Rules()))
	fmt.Fprintln(stdout, `Run "ago -explain <rule>" for the full rationale.`)
}

// ruleJSON is the -list -format json schema. It is what a coding agent reads
// to discover the rule set without parsing prose.
type ruleJSON struct {
	Name      string `json:"name"`
	Analyzer  string `json:"analyzer"`
	Summary   string `json:"summary"`
	Rationale string `json:"rationale"`
	Default   bool   `json:"default"`
	Enabled   bool   `json:"enabled"`
	Reverts   string `json:"reverts,omitempty"`
	Severity  string `json:"severity"`
	DocURL    string `json:"docURL"`
}

// policyJSON records the inputs that resolved the active rule set.
type policyJSON struct {
	RuleSource     string   `json:"ruleSource"`
	ConfigPath     string   `json:"configPath,omitempty"`
	ConfigVersion  int      `json:"configVersion,omitempty"`
	ConfigDisabled bool     `json:"configDisabled"`
	Tests          bool     `json:"tests"`
	Exclude        []string `json:"exclude"`
}

func resolvedPolicy(cfg *ago.Config, overrides []string, noConfig, testsFlag bool) policyJSON {
	source := "built-in"
	if cfg.Path() != "" {
		source = "config"
	}
	if len(overrides) > 0 {
		source = "flags"
	}
	configPath := cfg.Path()
	if configPath != "" {
		if abs, err := filepath.Abs(configPath); err == nil {
			configPath = abs
		}
	}
	return policyJSON{
		RuleSource:     source,
		ConfigPath:     configPath,
		ConfigVersion:  cfg.Version,
		ConfigDisabled: noConfig,
		Tests:          testsFlag || cfg.Tests,
		Exclude:        append([]string{}, cfg.Exclude...),
	}
}

func writeRulesJSON(stdout io.Writer, enabled map[string]bool, policy policyJSON) {
	out := struct {
		SchemaVersion int        `json:"schemaVersion"`
		Version       string     `json:"version"`
		Policy        policyJSON `json:"policy"`
		Rules         []ruleJSON `json:"rules"`
	}{SchemaVersion: ruleCatalogueSchemaVersion, Version: ago.Version, Policy: policy}
	for _, r := range ago.Rules() {
		out.Rules = append(out.Rules, ruleJSON{
			Name:      r.Name,
			Analyzer:  r.Analyzer.Name,
			Summary:   r.Summary,
			Rationale: r.Rationale,
			Default:   r.Default,
			Enabled:   enabled[r.Name],
			Reverts:   r.Reverts,
			Severity:  string(r.Severity),
			DocURL:    r.DocURL(),
		})
	}
	encodeJSON(stdout, out)
}

const ruleCatalogueSchemaVersion = 1

// explain prints one rule's full rationale.
func explain(stdout, stderr io.Writer, name string) int {
	r, ok := ago.Lookup(name)
	if !ok {
		fmt.Fprintf(stderr, "ago: unknown rule %q\n", name)
		suggest(stderr, name)
		return exitError
	}
	fmt.Fprintf(stdout, "%s\n\n", r.Name)
	fmt.Fprintf(stdout, "  %s\n", r.Summary)
	if r.Reverts != "" {
		fmt.Fprintf(stdout, "  Reverts a language change introduced in Go %s.\n", r.Reverts)
	}
	state := "off"
	if r.Default {
		state = "on"
	}
	fmt.Fprintf(stdout, "  Default: %s\n  Docs:    %s\n\n%s\n", state, r.DocURL(), r.Rationale)
	return exitClean
}

// suggest prints the rules whose names share a prefix with a misspelling.
func suggest(stderr io.Writer, name string) {
	var near []string
	for _, n := range ago.Names() {
		if strings.Contains(n, name) || strings.Contains(name, n) {
			near = append(near, n)
		}
	}
	sort.Strings(near)
	if len(near) > 0 {
		fmt.Fprintf(stderr, "ago: did you mean %s?\n", strings.Join(near, ", "))
		return
	}
	fmt.Fprintln(stderr, `ago: run "ago -list" for the rule set`)
}

// writeInitConfig writes a minimal policy at the nearest Go module or
// workspace root. It refuses to create a second policy below one that already
// applies.
func writeInitConfig(dir string, stdout, stderr io.Writer) int {
	cfg, err := ago.LoadConfig(dir, "")
	if err != nil {
		fmt.Fprintf(stderr, "ago: %v\n", err)
		return exitError
	}
	if cfg.Path() != "" {
		fmt.Fprintf(stderr, "ago: policy already exists: %s\n", cfg.Path())
		return exitError
	}
	root, marker, err := findProjectRoot(dir)
	if err != nil {
		fmt.Fprintf(stderr, "ago: %v\n", err)
		return exitError
	}
	path := filepath.Join(root, ago.ConfigName)
	// A config the whole team commits and reads is 0644.
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644) //nolint:gosec // team-readable by design
	if err != nil {
		fmt.Fprintf(stderr, "ago: %v\n", err)
		return exitError
	}
	removeOnError := true
	defer func() {
		if removeOnError {
			_ = os.Remove(path)
		}
	}()
	if _, err := io.WriteString(f, ago.ExampleConfig()); err != nil {
		_ = f.Close()
		fmt.Fprintf(stderr, "ago: writing %s: %v\n", path, err)
		return exitError
	}
	if err := f.Close(); err != nil {
		fmt.Fprintf(stderr, "ago: writing %s: %v\n", path, err)
		return exitError
	}
	removeOnError = false
	fmt.Fprintf(stdout, "wrote %s at %s root\n", path, marker)
	return exitClean
}

func findProjectRoot(dir string) (string, string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", "", err
	}
	for {
		for _, marker := range []string{"go.mod", "go.work"} {
			path := filepath.Join(abs, marker)
			if info, err := os.Stat(path); err == nil && !info.IsDir() {
				return abs, marker, nil
			}
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return "", "", fmt.Errorf("cannot find a go.mod or go.work above %s", dir)
		}
		abs = parent
	}
}

func usage(w io.Writer, fs *flag.FlagSet) {
	fmt.Fprint(w, `ago enforces one way to write Go across a codebase.

Usage:
  ago [flags] [packages]

Package arguments are go/packages patterns, the same ones go build accepts.
With no arguments ago checks ./... under the working directory.

Flags:
`)
	fs.SetOutput(w)
	fs.PrintDefaults()
	fmt.Fprint(w, `
Exit status:
  0  no violations
  1  at least one violation
  2  ago could not complete the run

Docs: https://github.com/agentstation/ago
`)
}
