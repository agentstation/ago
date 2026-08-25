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

// ConfigName is the file ago looks for when the command line names no
// configuration. ConfigNameAlt is an equivalent spelling.
const (
	ConfigName    = ".ago.yml"
	ConfigNameAlt = ".ago.yaml"

	// maxConfigBytes is the largest config file ago will load. A larger file
	// is an attack or a mistake, not a rule policy.
	maxConfigBytes = 1 << 20
	// maxExcludePatterns caps Config.Exclude.
	maxExcludePatterns = 1024
	// maxExcludePatternLen caps one exclude pattern.
	maxExcludePatternLen = 4096
	// maxRuleNames caps Enable and Disable together.
	maxRuleNames = 4096
)

// A Config is the on-disk rule policy for a repository. A committed policy
// means every developer, every CI job, and every coding agent enforces the
// same subset without extra instruction.
type Config struct {
	// Enable lists rules to turn on. The special value "default" expands to
	// the default rule set and "all" expands to every rule. An empty list
	// means the default set.
	Enable []string `yaml:"enable"`
	// Disable lists rules to turn off after ago applies Enable.
	Disable []string `yaml:"disable"`
	// Tests reports whether ago checks _test.go files.
	Tests bool `yaml:"tests"`
	// Exclude lists path patterns. ago matches them with [path/filepath.Match]
	// against each path element and skips those packages.
	Exclude []string `yaml:"exclude"`

	// path records where ago loaded the config, for error messages.
	path string
}

// Path returns the file ago loaded the config from, or "" for a default
// config.
func (c *Config) Path() string { return c.path }

// Meta rule-set names accepted in Config.Enable.
const (
	setDefault = "default"
	setAll     = "all"
)

// LoadConfig reads a config file. A path of "" searches dir and each parent
// directory for .ago.yml, stopping at the filesystem root. When the search
// finds no file it returns the default config and a nil error.
func LoadConfig(dir, path string) (*Config, error) {
	if path == "" {
		found, ok := findConfig(dir)
		if !ok {
			return &Config{}, nil
		}
		path = found
	}
	b, err := readConfigBytes(path)
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

// readConfigBytes reads a config file. It fails when the file is larger than
// maxConfigBytes.
func readConfigBytes(path string) ([]byte, error) {
	f, err := os.Open(path) //nolint:gosec // reading the config path the user named is the point
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, maxConfigBytes+1))
	if err != nil {
		return nil, err
	}
	if len(b) > maxConfigBytes {
		return nil, fmt.Errorf("%s: config exceeds %d bytes", path, maxConfigBytes)
	}
	return b, nil
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
// Enable only. Errors carry no path prefix. A caller that read the config from
// a file should prefix them with its path.
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
	if err := check("disable", c.Disable); err != nil {
		return err
	}
	if n := len(c.Enable) + len(c.Disable); n > maxRuleNames {
		return fmt.Errorf("enable/disable: %d names; want at most %d", n, maxRuleNames)
	}
	if len(c.Exclude) > maxExcludePatterns {
		return fmt.Errorf("exclude: %d patterns; want at most %d", len(c.Exclude), maxExcludePatterns)
	}
	for _, p := range c.Exclude {
		if p == "" {
			return fmt.Errorf("exclude: empty pattern")
		}
		if len(p) > maxExcludePatternLen {
			return fmt.Errorf("exclude: pattern exceeds %d bytes", maxExcludePatternLen)
		}
		if strings.ContainsRune(p, 0) {
			return fmt.Errorf("exclude: pattern contains NUL")
		}
	}
	return nil
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
// ago matches a pattern three ways, because each way catches a different
// surprise. It matches the whole slash-separated path, so "*.pb.go" works.
// It matches each path element, so "generated" excludes any directory with
// that name at any depth. It matches each leading path prefix, so
// "third_party/*" excludes the whole subtree rather than only the files
// directly inside it.
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
// -init". Rules that are on by default stay uncommented.
func ExampleConfig() string {
	var b strings.Builder
	b.WriteString("# ago rule policy. See https://github.com/agentstation/ago#rules\n")
	b.WriteString("# Run \"ago -list\" for the rule set this binary supports.\n\n")
	b.WriteString("enable:\n")
	for i, r := range registry {
		if i > 0 {
			b.WriteByte('\n')
		}
		prefix := "  # "
		if r.Default {
			prefix = "  - "
		}
		fmt.Fprintf(&b, "%s%-32s # %s.\n", prefix, r.Name, commentSummary(r.Summary))
	}
	b.WriteString("\ndisable: []\n")
	b.WriteString("\n# Check _test.go files as well as production files.\ntests: false\n")
	b.WriteString("\n# Path patterns to skip. The tool always skips vendor and testdata.\nexclude: []\n")
	return b.String()
}

// commentSummary capitalizes summary sentences so YAML comments split.
func commentSummary(s string) string {
	s = strings.TrimSpace(strings.TrimSuffix(s, "."))
	if s == "" {
		return s
	}
	var b strings.Builder
	capNext := true
	for _, r := range s {
		if capNext && r >= 'a' && r <= 'z' {
			r -= 'a' - 'A'
			capNext = false
		} else if r != ' ' {
			capNext = r == '.'
		}
		b.WriteRune(r)
	}
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
