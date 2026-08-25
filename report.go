package ago

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Format names an output encoding.
type Format string

// Supported output formats.
const (
	// FormatText is the vet-style file:line:col form that editors link on.
	FormatText Format = "text"
	// FormatJSON is a stable machine-readable document. It is the format to
	// use from a coding agent or any other program.
	FormatJSON Format = "json"
	// FormatSARIF is SARIF 2.1.0, which GitHub code scanning ingests.
	FormatSARIF Format = "sarif"
	// FormatGitHub is the GitHub Actions workflow-command form, which turns
	// findings into inline annotations on a pull request.
	FormatGitHub Format = "github"
)

// Formats lists every supported output format.
func Formats() []Format {
	return []Format{FormatText, FormatJSON, FormatSARIF, FormatGitHub}
}

// ParseFormat validates a format name.
func ParseFormat(s string) (Format, error) {
	for _, f := range Formats() {
		if string(f) == s {
			return f, nil
		}
	}
	names := make([]string, 0, len(Formats()))
	for _, f := range Formats() {
		names = append(names, string(f))
	}
	return "", fmt.Errorf("unknown format %q; want one of %s", s, strings.Join(names, ", "))
}

// A Report is everything one ago run produced. The JSON encoding of this type
// is ago's machine-readable contract. Fields grow over time but existing
// fields keep their name and meaning.
type Report struct {
	// Version is the ago version that produced the report.
	Version string `json:"version"`
	// Rules lists the canonical names of the rules that ran, sorted.
	Rules []string `json:"rules"`
	// Findings holds every violation, ordered by file, line, and column.
	Findings []Finding `json:"findings"`
	// StaleIgnores lists //ago:ignore directives that suppressed nothing.
	StaleIgnores []StaleIgnore `json:"staleIgnores"`
	// Errors holds load or parse failures. A non-empty Errors means the run
	// did not finish and Findings may omit violations.
	Errors []string `json:"errors"`
}

// A StaleIgnore is a suppression directive that matched no finding. The
// command reports it separately from Findings because it is a maintenance
// signal rather than a Go-subset violation.
type StaleIgnore struct {
	// Rules lists the rule names the directive claimed to suppress.
	Rules []string `json:"rules"`
	// Reason is the text the author gave for the exception.
	Reason string `json:"reason"`
	// File, Line, and Column locate the directive.
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

// Position renders the directive location in file:line:col form.
func (s StaleIgnore) Position() string {
	return fmt.Sprintf("%s:%d:%d", s.File, s.Line, s.Column)
}

// Write encodes the report in the requested format.
func (r *Report) Write(w io.Writer, format Format) error {
	switch format {
	case FormatJSON:
		return r.writeJSON(w)
	case FormatSARIF:
		return r.writeSARIF(w)
	case FormatGitHub:
		return r.writeGitHub(w)
	default:
		return r.writeText(w)
	}
}

func (r *Report) writeText(w io.Writer) error {
	for _, f := range r.Findings {
		if _, err := fmt.Fprintf(w, "%s: %s (%s)\n", f.Position(), f.Message, f.Rule); err != nil {
			return err
		}
	}
	for _, s := range r.StaleIgnores {
		if _, err := fmt.Fprintf(w, "%s: //ago:ignore suppressed nothing (%s)\n",
			s.Position(), strings.Join(s.Rules, ",")); err != nil {
			return err
		}
	}
	return nil
}

func (r *Report) writeJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

func (r *Report) writeGitHub(w io.Writer) error {
	for _, f := range r.Findings {
		level := "error"
		if f.Severity == Warning {
			level = "warning"
		}
		if _, err := fmt.Fprintf(w,
			"::%s file=%s,line=%d,col=%d,endLine=%d,endColumn=%d,title=ago %s::%s\n",
			level, f.File, f.Line, f.Column, f.EndLine, f.EndColumn, f.Rule,
			escapeWorkflowData(f.Message)); err != nil {
			return err
		}
	}
	for _, s := range r.StaleIgnores {
		if _, err := fmt.Fprintf(w,
			"::warning file=%s,line=%d,col=%d,title=ago stale ignore::%s\n",
			s.File, s.Line, s.Column,
			escapeWorkflowData("//ago:ignore suppressed nothing: "+strings.Join(s.Rules, ","))); err != nil {
			return err
		}
	}
	return nil
}

// escapeWorkflowData escapes the characters that terminate a GitHub Actions
// workflow command, so a message containing them cannot truncate or forge an
// annotation.
func escapeWorkflowData(s string) string {
	return strings.NewReplacer(
		"%", "%25",
		"\r", "%0D",
		"\n", "%0A",
		":", "%3A",
		",", "%2C",
	).Replace(s)
}
