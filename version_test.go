package ago

import (
	"runtime/debug"
	"testing"
)

func TestResolveVersion(t *testing.T) {
	tests := []struct {
		name   string
		linked string
		info   *debug.BuildInfo
		want   string
	}{
		{name: "linker value", linked: "v1.2.3", want: "v1.2.3"},
		{
			name:   "module version",
			linked: "dev",
			info: &debug.BuildInfo{Main: debug.Module{
				Path:    "github.com/agentstation/ago",
				Version: "v0.2.0",
			}},
			want: "v0.2.0",
		},
		{
			name:   "local source",
			linked: "dev",
			info: &debug.BuildInfo{Main: debug.Module{
				Path:    "github.com/agentstation/ago",
				Version: "(devel)",
			}},
			want: "dev",
		},
		{
			name:   "unrelated main module",
			linked: "dev",
			info: &debug.BuildInfo{Main: debug.Module{
				Path:    "example.com/consumer",
				Version: "v1.0.0",
			}},
			want: "dev",
		},
		{name: "missing build info", linked: "dev", want: "dev"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveVersion(tt.linked, tt.info); got != tt.want {
				t.Fatalf("resolveVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}
