// Package golangci registers ago as a golangci-lint module plugin.
//
// Build a custom golangci-lint binary that includes it with a .custom-gcl.yml:
//
//	version: v2.6.0
//	plugins:
//	  - module: github.com/agentstation/ago
//	    import: github.com/agentstation/ago/plugin/golangci
//	    version: latest
//
// Then run "golangci-lint custom" and enable the "ago" linter in
// .golangci.yml. Settings select the rule set. Without them the plugin runs
// ago's default rules.
//
//	linters:
//	  enable:
//	    - ago
//	  settings:
//	    custom:
//	      ago:
//	        type: module
//	        settings:
//	          enable: [no-goto, no-naked-return]
//	          disable: [no-dot-import]
package golangci

import (
	"fmt"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"github.com/agentstation/ago"
)

func init() {
	register.Plugin("ago", New)
}

// Settings selects which rules the plugin runs. It mirrors the enable and
// disable keys of .ago.yml, including the "default" and "all" meta-names.
// Leaving both empty runs ago's default rule set.
type Settings struct {
	Enable  []string `json:"enable"`
	Disable []string `json:"disable"`
}

// Plugin is the registered golangci-lint linter.
type Plugin struct {
	analyzers []*analysis.Analyzer
}

// New builds the plugin from golangci-lint's raw settings.
func New(conf any) (register.LinterPlugin, error) {
	settings, err := register.DecodeSettings[Settings](conf)
	if err != nil {
		return nil, err
	}

	cfg := &ago.Config{Enable: settings.Enable, Disable: settings.Disable}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("ago: %w", err)
	}

	rules := cfg.Enabled(nil)
	analyzers := make([]*analysis.Analyzer, 0, len(rules))
	for _, rule := range rules {
		analyzers = append(analyzers, rule.Analyzer)
	}
	return &Plugin{analyzers: analyzers}, nil
}

// BuildAnalyzers returns the analyzers for the selected rules.
func (p *Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return p.analyzers, nil
}

// GetLoadMode reports that ago's rules need full type information.
func (p *Plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
