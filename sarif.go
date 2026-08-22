package ago

import (
	"encoding/json"
	"io"
)

// SARIF 2.1.0 is the interchange format GitHub code scanning ingests. Only the
// subset of the schema ago populates is modelled here; the omitempty tags keep
// the document to what a static analysis tool without fixes needs to emit.
type sarifLog struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name           string      `json:"name"`
	Version        string      `json:"version"`
	InformationURI string      `json:"informationUri"`
	Rules          []sarifRule `json:"rules"`
}

type sarifRule struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	ShortDescription sarifText         `json:"shortDescription"`
	FullDescription  sarifText         `json:"fullDescription"`
	HelpURI          string            `json:"helpUri"`
	Properties       sarifRuleProps    `json:"properties"`
	DefaultConfig    sarifRuleDefaults `json:"defaultConfiguration"`
}

type sarifRuleDefaults struct {
	Level string `json:"level"`
}

type sarifRuleProps struct {
	Tags []string `json:"tags,omitempty"`
}

type sarifText struct {
	Text string `json:"text"`
}

type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   sarifText       `json:"message"`
	Locations []sarifLocation `json:"locations"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysical `json:"physicalLocation"`
}

type sarifPhysical struct {
	ArtifactLocation sarifArtifact `json:"artifactLocation"`
	Region           sarifRegion   `json:"region"`
}

type sarifArtifact struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn"`
	EndLine     int `json:"endLine,omitempty"`
	EndColumn   int `json:"endColumn,omitempty"`
}

// writeSARIF emits a SARIF 2.1.0 document describing the run. Every rule that
// ran is declared, whether or not it fired, so that a code scanning dashboard
// can distinguish "rule passed" from "rule not configured".
func (r *Report) writeSARIF(w io.Writer) error {
	ran := make(map[string]bool, len(r.Rules))
	for _, name := range r.Rules {
		ran[name] = true
	}

	var declared []sarifRule
	for _, rule := range registry {
		if !ran[rule.Name] {
			continue
		}
		tags := []string{"go", "style"}
		if rule.Reverts != "" {
			tags = append(tags, "reverts-go"+rule.Reverts)
		}
		declared = append(declared, sarifRule{
			ID:               rule.Name,
			Name:             rule.Name,
			ShortDescription: sarifText{Text: rule.Summary},
			FullDescription:  sarifText{Text: rule.Rationale},
			HelpURI:          rule.DocURL(),
			Properties:       sarifRuleProps{Tags: tags},
			DefaultConfig:    sarifRuleDefaults{Level: string(rule.Severity)},
		})
	}

	results := make([]sarifResult, 0, len(r.Findings))
	for _, f := range r.Findings {
		results = append(results, sarifResult{
			RuleID:  f.Rule,
			Level:   string(f.Severity),
			Message: sarifText{Text: f.Message},
			Locations: []sarifLocation{{
				PhysicalLocation: sarifPhysical{
					ArtifactLocation: sarifArtifact{URI: f.File},
					Region: sarifRegion{
						StartLine:   f.Line,
						StartColumn: f.Column,
						EndLine:     f.EndLine,
						EndColumn:   f.EndColumn,
					},
				},
			}},
		})
	}

	doc := sarifLog{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []sarifRun{{
			Tool: sarifTool{Driver: sarifDriver{
				Name:           "ago",
				Version:        Version,
				InformationURI: "https://github.com/agentstation/ago",
				Rules:          declared,
			}},
			Results: results,
		}},
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(doc)
}
