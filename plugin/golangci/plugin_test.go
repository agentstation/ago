package golangci

import (
	"testing"

	"github.com/agentstation/ago"
	"github.com/golangci/plugin-module-register/register"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		conf    any
		want    []string
		wantErr bool
	}{
		{
			name: "no settings runs the default rule set",
			conf: map[string]any{},
			want: ago.DefaultNames(),
		},
		{
			name: "explicit enable selects exactly those rules",
			conf: map[string]any{"enable": []string{"no-goto", "no-embedded-field"}},
			want: []string{"no-embedded-field", "no-goto"},
		},
		{
			name: "disable subtracts from the default set",
			conf: map[string]any{"enable": []string{"default"}, "disable": []string{"no-goto"}},
			want: without(ago.DefaultNames(), "no-goto"),
		},
		{
			name: "all selects every rule",
			conf: map[string]any{"enable": []string{"all"}},
			want: ago.Names(),
		},
		{
			name:    "an unknown rule is an error, not a silent no-op",
			conf:    map[string]any{"enable": []string{"no-such-rule"}},
			wantErr: true,
		},
		{
			name:    "a meta-name in disable is an error",
			conf:    map[string]any{"disable": []string{"all"}},
			wantErr: true,
		},
		{
			name:    "an unknown settings key is an error",
			conf:    map[string]any{"enabel": []string{"no-goto"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New(tt.conf)
			if tt.wantErr {
				if err == nil {
					t.Fatal("New: want error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			analyzers, err := p.BuildAnalyzers()
			if err != nil {
				t.Fatalf("BuildAnalyzers: %v", err)
			}
			got := make([]string, len(analyzers))
			for i, a := range analyzers {
				rule, ok := ago.Lookup(a.Name)
				if !ok {
					t.Fatalf("analyzer %q is not a known rule", a.Name)
				}
				got[i] = rule.Name
			}
			if !sameSet(got, tt.want) {
				t.Errorf("analyzers = %v, want %v", got, tt.want)
			}
		})
	}
}

// The rules need type information. Declaring syntax-only would give every
// type-aware rule a nil TypesInfo.
func TestLoadMode(t *testing.T) {
	p, err := New(map[string]any{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := p.GetLoadMode(); got != register.LoadModeTypesInfo {
		t.Errorf("GetLoadMode() = %q, want %q", got, register.LoadModeTypesInfo)
	}
}

func without(names []string, drop string) []string {
	out := make([]string, 0, len(names))
	for _, n := range names {
		if n != drop {
			out = append(out, n)
		}
	}
	return out
}

func sameSet(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	seen := make(map[string]int, len(got))
	for _, g := range got {
		seen[g]++
	}
	for _, w := range want {
		seen[w]--
	}
	for _, n := range seen {
		if n != 0 {
			return false
		}
	}
	return true
}
