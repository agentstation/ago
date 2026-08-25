package ago

import (
	"bytes"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func FuzzLoadConfig(f *testing.F) {
	f.Add("")
	f.Add("enable: [no-goto]\n")
	f.Add("enable: [all]\ndisable: [no-goto]\n")
	f.Add("exclude: [generated, '*.pb.go', third_party/*]\n")
	f.Add("enable: [no-gotoo]\n")
	f.Add("enabel: [no-goto]\n")
	f.Add("&a [*a, *a, *a]\n")
	f.Fuzz(func(t *testing.T, body string) {
		if len(body) > maxConfigBytes+1024 {
			body = body[:maxConfigBytes+1024]
		}
		dir := t.TempDir()
		path := filepath.Join(dir, ConfigName)
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		_, _ = LoadConfig(dir, path)
	})
}

func FuzzGitHubReport(f *testing.F) {
	f.Add("x.go", "no-goto", "goto; use a loop", 1, 1)
	f.Add("a.go,line=1::x", "no-goto", "::error::forged\nnext", 2, 3)
	f.Add("dir/\nfile.go", "ago stale ignore", "%0A::warning::", 0, 0)
	f.Fuzz(func(t *testing.T, file, rule, msg string, line, col int) {
		r := &Report{
			Findings: []Finding{{
				Rule:      rule,
				Severity:  Error,
				Message:   msg,
				File:      file,
				Line:      line,
				Column:    col,
				EndLine:   line,
				EndColumn: col,
			}},
		}
		var buf bytes.Buffer
		if err := r.Write(&buf, FormatGitHub); err != nil {
			t.Fatal(err)
		}
		out := buf.String()
		if out == "" {
			t.Fatal("empty github report")
		}
		lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
		if len(lines) != 1 {
			t.Fatalf("finding produced %d lines, want 1: %q", len(lines), out)
		}
		if !strings.HasPrefix(lines[0], "::") {
			t.Fatalf("missing workflow command prefix: %q", lines[0])
		}
	})
}

func FuzzSkip(f *testing.F) {
	f.Add("generated", "internal/generated/x.go")
	f.Add("*.pb.go", "api/service.pb.go")
	f.Add("third_party/*", "third_party/lib/a.go")
	f.Add("*", "anything.go")
	f.Fuzz(func(t *testing.T, pattern, path string) {
		cfg := &Config{Exclude: []string{pattern}}
		_ = cfg.Skip(path)
	})
}

func FuzzParseDirective(f *testing.F) {
	f.Add("//ago:ignore no-goto -- reason")
	f.Add("//ago:ignore-file * -- generated")
	f.Add("//ago:ignorecase not a directive")
	f.Add("//ago:ignore")
	f.Add("/* ago:ignore no-goto -- no */")
	f.Fuzz(func(t *testing.T, text string) {
		if len(text) > 65536 {
			t.Skip()
		}
		src := []byte(text)
		if len(src) == 0 {
			src = []byte(" ")
		}
		fset := token.NewFileSet()
		tf := fset.AddFile("x.go", -1, len(src))
		tf.SetLinesForContent(src)
		c := &ast.Comment{Slash: tf.Pos(0), Text: text}
		_ = parseDirective(fset, c, src)
	})
}
