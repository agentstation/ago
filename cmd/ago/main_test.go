package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/agentstation/ago"
)

func TestHelpWritesOneCompleteDocumentToStdout(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if status := run([]string{"-h"}, &stdout, &stderr); status != exitClean {
		t.Fatalf("status = %d, want %d", status, exitClean)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	help := stdout.String()
	if got := strings.Count(help, "Usage:"); got != 1 {
		t.Errorf("Usage count = %d, want 1\n%s", got, help)
	}
	for _, want := range []string{"-all", "-config", "Exit status:", "one Go dialect"} {
		if !strings.Contains(help, want) {
			t.Errorf("help does not contain %q\n%s", want, help)
		}
	}
}

func TestInitWritesMinimalPolicyAtModuleRoot(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "go.mod"), "module example.com/project\n\ngo 1.25\n")
	nested := filepath.Join(root, "internal", "example")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if status := writeInitConfig(nested, &stdout, &stderr); status != exitClean {
		t.Fatalf("status = %d, stderr = %q", status, stderr.String())
	}
	path := filepath.Join(root, ago.ConfigName)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != ago.ExampleConfig() {
		t.Errorf("config differs from ExampleConfig:\n%s", b)
	}
	if _, err := os.Stat(filepath.Join(nested, ago.ConfigName)); !os.IsNotExist(err) {
		t.Errorf("nested config exists or stat failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "at go.mod root") {
		t.Errorf("stdout = %q, want go.mod root", stdout.String())
	}
}

func TestInitRejectsAnExistingParentPolicy(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "go.mod"), "module example.com/project\n\ngo 1.25\n")
	policy := filepath.Join(root, ago.ConfigName)
	writeTestFile(t, policy, "enable: [default]\n")
	nested := filepath.Join(root, "internal", "example")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if status := writeInitConfig(nested, &stdout, &stderr); status != exitError {
		t.Fatalf("status = %d, want %d", status, exitError)
	}
	if !strings.Contains(stderr.String(), policy) {
		t.Errorf("stderr = %q, want policy path %q", stderr.String(), policy)
	}
}

func TestInitWritesAtWorkspaceRoot(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "go.work"), "go 1.25\n")
	var stdout, stderr bytes.Buffer
	if status := writeInitConfig(root, &stdout, &stderr); status != exitClean {
		t.Fatalf("status = %d, stderr = %q", status, stderr.String())
	}
	if !strings.Contains(stdout.String(), "at go.work root") {
		t.Errorf("stdout = %q, want go.work root", stdout.String())
	}
}

func TestInitRequiresModuleOrWorkspace(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if status := writeInitConfig(t.TempDir(), &stdout, &stderr); status != exitError {
		t.Fatalf("status = %d, want %d", status, exitError)
	}
	if !strings.Contains(stderr.String(), "cannot find a go.mod or go.work") {
		t.Errorf("stderr = %q", stderr.String())
	}
}

func TestListJSONReportsResolvedPolicy(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, ago.ConfigName)
	writeTestFile(t, configPath, "version: 1\nenable: [default]\ntests: false\nexclude: [generated]\n")

	tests := []struct {
		name         string
		args         []string
		wantSource   string
		wantPath     string
		wantVersion  int
		wantDisabled bool
		wantTests    bool
		wantExclude  []string
	}{
		{name: "built-in", args: []string{"-no-config", "-list", "-format", "json"}, wantSource: "built-in", wantDisabled: true, wantExclude: []string{}},
		{name: "config", args: []string{"-config", configPath, "-list", "-format", "json"}, wantSource: "config", wantPath: configPath, wantVersion: 1, wantExclude: []string{"generated"}},
		{name: "flags", args: []string{"-config", configPath, "-rules", "no-goto", "-tests", "-list", "-format", "json"}, wantSource: "flags", wantPath: configPath, wantVersion: 1, wantTests: true, wantExclude: []string{"generated"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if status := run(tt.args, &stdout, &stderr); status != exitClean {
				t.Fatalf("status = %d, stderr = %q", status, stderr.String())
			}
			var got struct {
				SchemaVersion int        `json:"schemaVersion"`
				Policy        policyJSON `json:"policy"`
				Rules         []ruleJSON `json:"rules"`
			}
			if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
				t.Fatalf("decode JSON: %v\n%s", err, stdout.String())
			}
			if got.SchemaVersion != 1 {
				t.Errorf("schemaVersion = %d, want 1", got.SchemaVersion)
			}
			if got.Policy.RuleSource != tt.wantSource || got.Policy.ConfigPath != tt.wantPath || got.Policy.ConfigVersion != tt.wantVersion || got.Policy.ConfigDisabled != tt.wantDisabled || got.Policy.Tests != tt.wantTests || !slices.Equal(got.Policy.Exclude, tt.wantExclude) {
				t.Errorf("policy = %+v, want source=%q path=%q version=%d disabled=%t tests=%t exclude=%v", got.Policy, tt.wantSource, tt.wantPath, tt.wantVersion, tt.wantDisabled, tt.wantTests, tt.wantExclude)
			}
			if len(got.Rules) != len(ago.Rules()) {
				t.Errorf("rules = %d, want %d", len(got.Rules), len(ago.Rules()))
			}
		})
	}
}

func writeTestFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
