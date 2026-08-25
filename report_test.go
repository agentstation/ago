package ago

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestGitHubFormatEscapesProperties(t *testing.T) {
	r := &Report{
		Findings: []Finding{{
			Rule:      "no-goto",
			Severity:  Error,
			Message:   "goto; use a loop\n::error title=forged::pwned",
			File:      "a.go,line=1::injected",
			Line:      2,
			Column:    3,
			EndLine:   2,
			EndColumn: 4,
		}},
		StaleIgnores: []StaleIgnore{{
			Rules:  []string{"no-goto"},
			File:   "b.go,title=stale",
			Line:   5,
			Column: 1,
		}},
	}
	var buf bytes.Buffer
	if err := r.Write(&buf, FormatGitHub); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2 (one command per record):\n%s", len(lines), out)
	}
	if strings.Contains(out, "file=a.go,line=1") {
		t.Fatalf("file property was not escaped:\n%s", out)
	}
	if !strings.Contains(out, "%2C") {
		t.Fatalf("comma in file was not escaped:\n%s", out)
	}
	if strings.Contains(out, "\n::error title=forged::") || strings.Contains(out, "::error title=forged::") {
		t.Fatalf("message injected a workflow command:\n%s", out)
	}
	if !strings.Contains(out, "%3A%3Aerror") {
		t.Fatalf(":: in message was not escaped:\n%s", out)
	}
}

func TestGitHubFormatOneLinePerFinding(t *testing.T) {
	r := &Report{Findings: []Finding{{
		Rule:     "no-goto",
		Severity: Error,
		Message:  "ok",
		File:     "x.go",
		Line:     1,
		Column:   1,
	}}}
	var buf bytes.Buffer
	if err := r.Write(&buf, FormatGitHub); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	want := "::error file=x.go,line=1,col=1,endLine=0,endColumn=0,title=ago no-goto::ok\n"
	if got != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}

func TestSARIFStripsControlCharactersFromURI(t *testing.T) {
	r := &Report{
		Rules: []string{"no-goto"},
		Findings: []Finding{{
			Rule:      "no-goto",
			Severity:  Error,
			Message:   "goto; use a loop",
			File:      "dir/\nsecret.go",
			Line:      0,
			Column:    0,
			EndLine:   0,
			EndColumn: 0,
		}},
	}
	var buf bytes.Buffer
	if err := r.Write(&buf, FormatSARIF); err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Runs []struct {
			Results []struct {
				Locations []struct {
					PhysicalLocation struct {
						ArtifactLocation struct {
							URI string `json:"uri"`
						} `json:"artifactLocation"`
						Region struct {
							StartLine int `json:"startLine"`
						} `json:"region"`
					} `json:"physicalLocation"`
				} `json:"locations"`
			} `json:"results"`
		} `json:"runs"`
	}
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Runs) != 1 || len(doc.Runs[0].Results) != 1 {
		t.Fatalf("unexpected SARIF shape: %s", buf.String())
	}
	uri := doc.Runs[0].Results[0].Locations[0].PhysicalLocation.ArtifactLocation.URI
	if strings.ContainsAny(uri, "\n\r") {
		t.Fatalf("URI still contains a control character: %q", uri)
	}
	if uri != "dir/secret.go" {
		t.Fatalf("URI = %q, want dir/secret.go", uri)
	}
	if line := doc.Runs[0].Results[0].Locations[0].PhysicalLocation.Region.StartLine; line < 1 {
		t.Fatalf("startLine = %d, want >= 1", line)
	}
}
