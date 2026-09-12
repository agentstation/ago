package goago

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestConfigMigration(t *testing.T) {
	for _, name := range []string{".goago.yml", ".goago.yaml", ".ago.yml", ".ago.yaml"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, name)
			if err := os.WriteFile(path, []byte("enable: [no-goto]\ntests: true\nexclude: [generated]\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			sub := filepath.Join(root, "nested")
			if err := os.Mkdir(sub, 0o700); err != nil {
				t.Fatal(err)
			}
			cfg, err := LoadConfig(sub, "")
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Path() != path || !cfg.Tests || !cfg.Skip("generated/x.go") || !slices.Equal(ruleNames(cfg.Enabled(nil)), []string{"no-goto"}) {
				t.Fatalf("policy changed during migration: %+v", cfg)
			}
		})
	}
}

func TestConfigMigrationRejectsCompetingNames(t *testing.T) {
	for _, old := range []string{".ago.yml", ".ago.yaml"} {
		for _, current := range []string{".goago.yml", ".goago.yaml"} {
			t.Run(old+current, func(t *testing.T) {
				root := t.TempDir()
				for _, name := range []string{old, current} {
					if err := os.WriteFile(filepath.Join(root, name), []byte("enable: [no-goto]\n"), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				if _, err := LoadConfig(root, ""); err == nil || !strings.Contains(err.Error(), "multiple policy files") {
					t.Fatalf("error = %v, want competing policy error", err)
				}
				if _, err := LoadConfig(root, filepath.Join(root, current)); err != nil {
					t.Fatalf("explicit config: %v", err)
				}
			})
		}
	}
}

func TestDirectiveMigration(t *testing.T) {
	for _, name := range []string{"ago", "goago"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			files := map[string]string{
				"go.mod": "module example.com/migration\n\ngo 1.25\n",
				"line.go": `package migration
func good() {
//NAME:ignore no-goto -- state transition
goto done
done:
return
}
func invalid() {
//NAME:ignore no-goto
goto done
done:
return
}
//NAME:ignore no-goto -- stale example
func stale() {}
//NAME:ignorecase is ordinary prose
func prose() {}
`,
				"file.go": `//NAME:ignore-file no-goto -- file exception
package migration
func file() { goto done; done: return }
`,
			}
			for path, source := range files {
				if err := os.WriteFile(filepath.Join(root, path), []byte(strings.ReplaceAll(source, "NAME", name)), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			gotoRule, _ := Lookup("no-goto")
			invalidRule, _ := Lookup("no-invalid-ignore")
			r, err := Check(Options{Dir: root, Rules: []Rule{gotoRule, invalidRule}, ReportStaleIgnores: true})
			if err != nil {
				t.Fatal(err)
			}
			if len(r.Errors) != 0 || len(r.Findings) != 2 || len(r.StaleIgnores) != 1 {
				t.Fatalf("unexpected report: %+v", r)
			}
			if r.Findings[0].Rule != "no-invalid-ignore" || r.Findings[1].Rule != "no-goto" || r.Findings[1].Line != 10 {
				t.Fatalf("unexpected findings: %+v", r.Findings)
			}
		})
	}
}
