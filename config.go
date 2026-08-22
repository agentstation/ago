package ago

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	yaml "go.yaml.in/yaml/v3"
)

// ConfigName is the file ago looks for when no configuration is named on the
// command line. ConfigNameAlt is accepted as an equivalent spelling.
const (
	ConfigName    = ".ago.yml"
	ConfigNameAlt = ".ago.yaml"
)

// A Config is the on-disk rule policy for a repository. Keeping the policy in
// the repository rather than in a command line means every developer, every CI
// job, and every coding agent enforces the same subset without being told.
type Config struct {
	// Enable lists rules to turn on. The special value "default" expands to
	// the default rule set and "all" expands to every rule. An empty list
	// means the default set.
	Enable []string `yaml:"enable"`
	// Disable lists rules to turn off after Enable is applied.
	Disable []string `yaml:"disable"`
	// Tests reports whether _test.go files are checked.
	Tests bool `yaml:"tests"`
	// Exclude lists path patterns, matched with [path/filepath.Match] against
	// each path element, whose packages are skipped.
	Exclude []string `yaml:"exclude"`

	// path records where the config was loaded from, for error messages.
	path string
}

// Path returns the file the config was loaded from, or "" for a default
// config.
func (c *Config) Path() string { return c.path }

// Meta rule-set names accepted in Config.Enable.
const (
	setDefault = "default"
	setAll     = "all"
)

// LoadConfig reads a config file. A path of "" searches dir and each parent
// directory for .ago.yml, stopping at the filesystem root; when no file is
// found it returns the default config and a nil error.
func LoadConfig(dir, path string) (*Config, error) {
	if path == "" {
		found, ok := findConfig(dir)
		if !ok {
			return &Config{}, nil
		}
		path = found
	}
	b, err := os.ReadFile(path) //nolint:gosec // reading the config path the user named is the point
	if err != nil {
		return nil, err
	}
	cfg := &Config{path: path}
	dec := yaml.NewDecoder(bytes.NewReader(b))
	dec.KnownFields(true)
	// A file that is empty or holds only comments decodes to io.EOF. That is a
	// valid config meaning "every default", not an error.
	if err := dec.Decode(cfg); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return cfg, nil
}

// findConfig walks up from dir looking for a config file.
func findConfig(dir string) (string, bool) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", false
	}
	for {
		for _, name := range []string{ConfigName, ConfigNameAlt} {
			p := filepath.Join(abs, name)
			if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
				return p, true
			}
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return "", false
		}
		abs = parent
	}
}

// Validate reports whether every name in Enable and Disable refers to a rule
// this build knows about. The meta-names "default" and "all" are valid in
// Enable only. Errors are unqualified; a caller that read the config from a
// file should prefix them with its path.
func (c *Config) Validate() error {
	check := func(field string, names []string) error {
		for _, n := range names {
			if n == setDefault || n == setAll {
				if field == "disable" {
					return fmt.Errorf("disable: %q is not a rule name", n)
				}
				continue
			}
			if _, ok := Lookup(n); !ok {
				return fmt.Errorf("%s: unknown rule %q", field, n)
			}
		}
		return nil
	}
	if err := check("enable", c.Enable); err != nil {
		return err
	}
	return check("disable", c.Disable)
}

// Enabled resolves the config into the set of rules to run, sorted in
// registration order. overrides, when non-empty, replaces Config.Enable and
// comes from the command line.
func (c *Config) Enabled(overrides []string) []Rule {
	selected := map[string]bool{}
	enable := c.Enable
	if len(overrides) > 0 {
		enable = overrides
	}
	if len(enable) == 0 {
		enable = []string{setDefault}
	}
	for _, name := range enable {
		switch name {
		case setDefault:
			for _, r := range registry {
				if r.Default {
					selected[r.Name] = true
				}
			}
		case setAll:
			for _, r := range registry {
				selected[r.Name] = true
			}
		default:
			if r, ok := Lookup(name); ok {
				selected[r.Name] = true
			}
		}
	}
	for _, name := range c.Disable {
		if r, ok := Lookup(name); ok {
			delete(selected, r.Name)
		}
	}
	var out []Rule
	for _, r := range registry {
		if selected[r.Name] {
			out = append(out, r)
		}
	}
	return out
}

// Skip reports whether a file path matches any exclude pattern.
//
// A pattern is matched three ways, because a single one surprises in a
// different case each time. It is matched against the whole slash-separated
// path, so "*.pb.go" works; against each path element, so "generated" excludes
// any directory with that name at any depth; and against each leading path
// prefix, so "third_party/*" excludes the whole subtree rather than only the
// files directly inside it.
func (c *Config) Skip(path string) bool {
	if len(c.Exclude) == 0 {
		return false
	}
	slashed := filepath.ToSlash(path)
	elems := strings.Split(slashed, "/")
	candidates := make([]string, 0, 2*len(elems))
	for i := range elems {
		candidates = append(candidates, elems[i], strings.Join(elems[:i+1], "/"))
	}
	for _, pattern := range c.Exclude {
		for _, candidate := range candidates {
			if ok, _ := filepath.Match(pattern, candidate); ok {
				return true
			}
		}
	}
	return false
}

// ExampleConfig returns a commented config listing every rule, for "ago
// -init". Rules that are on by default are left uncommented.
func ExampleConfig() string {
	var b strings.Builder
	b.WriteString("# ago rule policy. See https://github.com/agentstation/ago#rules\n")
	b.WriteString("# Run \"ago -list\" for the rule set this binary supports.\n\n")
	b.WriteString("enable:\n")
	for _, r := range registry {
		prefix := "  # - "
		if r.Default {
			prefix = "  - "
		}
		fmt.Fprintf(&b, "%s%-32s # %s\n", prefix, r.Name, r.Summary)
	}
	b.WriteString("\ndisable: []\n")
	b.WriteString("\n# Check _test.go files as well as production files.\ntests: false\n")
	b.WriteString("\n# Path patterns to skip. vendor and testdata are always skipped.\nexclude: []\n")
	return b.String()
}

// sortedNames is a small helper used by the command's error messages.
func sortedNames(rules []Rule) []string {
	out := make([]string, 0, len(rules))
	for _, r := range rules {
		out = append(out, r.Name)
	}
	sort.Strings(out)
	return out
}
