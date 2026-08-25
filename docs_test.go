package ago

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

var headingRE = regexp.MustCompile(`(?m)^#{1,6}\s+(.*)$`)

// Every rule's DocURL points at a rule-reference anchor. GitHub derives that
// anchor from a heading. A rule added without a heading ships a broken link.
// The JSON catalogue gives that link to coding agents.
func TestEveryRuleHasADocAnchor(t *testing.T) {
	doc, err := os.ReadFile("docs/rules.md")
	if err != nil {
		t.Fatalf("read docs/rules.md: %v", err)
	}

	anchors := map[string]bool{}
	for _, m := range headingRE.FindAllStringSubmatch(string(doc), -1) {
		anchors[slugify(m[1])] = true
	}

	for _, rule := range Rules() {
		if !anchors[rule.Name] {
			t.Errorf("rule %q has no documentation heading; DocURL %s would not resolve",
				rule.Name, rule.DocURL())
		}
		if rule.Analyzer.URL != rule.DocURL() {
			t.Errorf("rule %q analyzer URL = %q, want %q",
				rule.Name, rule.Analyzer.URL, rule.DocURL())
		}
	}
}

// slugify mirrors how GitHub derives an anchor from a heading: strip
// formatting, lowercase, drop punctuation, and join words with hyphens.
func slugify(heading string) string {
	s := strings.ToLower(strings.TrimSpace(strings.ReplaceAll(heading, "`", "")))
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('-')
		}
	}
	return b.String()
}

// The rule reference documents each rule's default state. A rule whose default
// flips without a documentation change is a defect. Assert that the sections
// agree with the registry.
func TestDefaultRuleCountMatchesDocs(t *testing.T) {
	doc, err := os.ReadFile("docs/rules.md")
	if err != nil {
		t.Fatalf("read docs/rules.md: %v", err)
	}

	body := string(doc)
	on := section(body, "## On by default", "## Off by default")
	off := section(body, "## Off by default", "")
	if on == "" || off == "" {
		t.Fatal("docs/rules.md is missing the On/Off by default sections")
	}

	for _, rule := range Rules() {
		heading := "### `" + rule.Name + "`"
		inOn := strings.Contains(on, heading)
		inOff := strings.Contains(off, heading)
		switch {
		case inOn && inOff:
			t.Errorf("rule %q is documented in both sections", rule.Name)
		case !inOn && !inOff:
			t.Errorf("rule %q is documented in neither section", rule.Name)
		case rule.Default && !inOn:
			t.Errorf("rule %q is on by default but documented as off", rule.Name)
		case !rule.Default && !inOff:
			t.Errorf("rule %q is off by default but documented as on", rule.Name)
		}
	}
}

func section(body, start, end string) string {
	i := strings.Index(body, start)
	if i < 0 {
		return ""
	}
	rest := body[i+len(start):]
	if end != "" {
		if j := strings.Index(rest, end); j >= 0 {
			return rest[:j]
		}
	}
	return rest
}
