package ago

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, dir, body string) string {
	t.Helper()
	path := filepath.Join(dir, ConfigName)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr string
		check   func(*testing.T, *Config)
	}{
		{
			name: "empty file means defaults",
			body: "",
			check: func(t *testing.T, c *Config) {
				if got := ruleNames(c.Enabled(nil)); !slices.Equal(got, DefaultNames()) {
					t.Errorf("enabled = %v, want the default set %v", got, DefaultNames())
				}
			},
		},
		{
			name: "explicit rule list",
			body: "enable:\n  - no-goto\n  - no-dot-import\n",
			check: func(t *testing.T, c *Config) {
				want := []string{"no-dot-import", "no-goto"}
				if got := ruleNames(c.Enabled(nil)); !slices.Equal(got, want) {
					t.Errorf("enabled = %v, want %v", got, want)
				}
			},
		},
		{
			name: "all minus a disable",
			body: "enable: [all]\ndisable: [no-goto]\n",
			check: func(t *testing.T, c *Config) {
				got := ruleNames(c.Enabled(nil))
				if slices.Contains(got, "no-goto") {
					t.Error("disable did not remove no-goto")
				}
				if len(got) != len(Rules())-1 {
					t.Errorf("enabled %d rules, want %d", len(got), len(Rules())-1)
				}
			},
		},
		{
			name: "default meta name expands",
			body: "enable: [default, no-goto]\n",
			check: func(t *testing.T, c *Config) {
				if got := ruleNames(c.Enabled(nil)); !slices.Equal(got, DefaultNames()) {
					t.Errorf("enabled = %v, want the default set", got)
				}
			},
		},
		{
			name:    "unknown rule fails loudly",
			body:    "enable: [no-gotoo]\n",
			wantErr: `unknown rule "no-gotoo"`,
		},
		{
			name:    "unknown field fails loudly",
			body:    "enabel: [no-goto]\n",
			wantErr: "field enabel not found",
		},
		{
			name:    "meta name in disable is rejected",
			body:    "disable: [all]\n",
			wantErr: `"all" is not a rule name`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := writeConfig(t, dir, tt.body)
			cfg, err := LoadConfig(dir, path)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want one containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadConfig: %v", err)
			}
			tt.check(t, cfg)
		})
	}
}

// TestLoadConfigSearchesParents checks that a config at the repository root
// governs a run started from a subdirectory. That is how an agent invoked in
// a package directory picks up the repository's policy.
func TestLoadConfigSearchesParents(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, "enable: [no-goto]\n")
	sub := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(sub, "")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Path() == "" {
		t.Fatal("config not found from a subdirectory")
	}
	if got := ruleNames(cfg.Enabled(nil)); !slices.Equal(got, []string{"no-goto"}) {
		t.Errorf("enabled = %v, want [no-goto]", got)
	}
}

// TestLoadConfigMissingIsNotAnError confirms that a repository without a
// config runs the default set rather than failing.
func TestLoadConfigMissingIsNotAnError(t *testing.T) {
	cfg, err := LoadConfig(t.TempDir(), "")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Path() != "" {
		t.Errorf("path = %q, want empty", cfg.Path())
	}
	if got := ruleNames(cfg.Enabled(nil)); !slices.Equal(got, DefaultNames()) {
		t.Errorf("enabled = %v, want the default set", got)
	}
}

// TestOverridesBeatConfig confirms the command line wins over the file.
func TestOverridesBeatConfig(t *testing.T) {
	cfg := &Config{Enable: []string{"no-goto"}}
	got := ruleNames(cfg.Enabled([]string{"no-dot-import"}))
	if !slices.Equal(got, []string{"no-dot-import"}) {
		t.Errorf("enabled = %v, want [no-dot-import]", got)
	}
}

// TestDisableAppliesAfterOverrides documents that disable is subtractive even
// when the rule list came from the command line.
func TestDisableAppliesAfterOverrides(t *testing.T) {
	cfg := &Config{Disable: []string{"no-goto"}}
	got := ruleNames(cfg.Enabled([]string{"no-goto", "no-dot-import"}))
	if !slices.Equal(got, []string{"no-dot-import"}) {
		t.Errorf("enabled = %v, want [no-dot-import]", got)
	}
}

func TestConfigSkip(t *testing.T) {
	cfg := &Config{Exclude: []string{"generated", "*.pb.go", "third_party/*"}}
	tests := []struct {
		path string
		want bool
	}{
		{"internal/generated/x.go", true},
		{"api/service.pb.go", true},
		{"third_party/lib/a.go", true},
		{"internal/service/x.go", false},
		{"generated.go", false},
	}
	for _, tt := range tests {
		if got := cfg.Skip(tt.path); got != tt.want {
			t.Errorf("Skip(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

// TestExampleConfigIsValid keeps "ago -init" honest: the file it writes must
// load without error and select exactly the default rule set.
func TestExampleConfigIsValid(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, ExampleConfig())
	cfg, err := LoadConfig(dir, path)
	if err != nil {
		t.Fatalf("the config written by -init does not load: %v", err)
	}
	if got := ruleNames(cfg.Enabled(nil)); !slices.Equal(got, DefaultNames()) {
		t.Errorf("enabled = %v, want the default set %v", got, DefaultNames())
	}
}

func ruleNames(rules []Rule) []string {
	out := make([]string, 0, len(rules))
	for _, r := range rules {
		out = append(out, r.Name)
	}
	slices.Sort(out)
	return out
}
