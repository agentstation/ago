package ago

import (
	"go/ast"
	"go/token"
)

// RuleNoInvalidIgnore keeps suppression directives honest.
var RuleNoInvalidIgnore = register(Rule{
	Name:     "no-invalid-ignore",
	Summary:  "every //ago:ignore must name a known rule and give a reason",
	Default:  true,
	Severity: Error,
	Rationale: `A suppression that names no rule silences everything on the line, and a
suppression that names a misspelled rule silences nothing at all. Both fail
quietly, which is how a lint configuration rots.

This rule requires the full form:

	//ago:ignore no-goto -- hand-written state machine, see docs/parser.md

The reason is not decoration. It is the only record of why ago granted the
exception. A reviewer or a coding agent reads it before deciding whether the
exception still applies.

Turning this rule off is possible but self-defeating. It is the rule that
makes every other rule's escape hatch auditable.`,
	Analyzer: newAnalyzer("no-invalid-ignore",
		"reject //ago:ignore directives that name no known rule or give no reason",
		checkInvalidIgnore),
})

func checkInvalidIgnore(c *checkPass) {
	if c.ignores == nil {
		return
	}
	for _, d := range c.ignores.all {
		if d.Problem != "" {
			c.reportf(directivePos{d.Pos}, "//ago:ignore %s", d.Problem)
		}
	}
}

// directivePos adapts a comment position to the [ast.Node] interface, so
// reportf can report a directive through the same path as a syntax node. The
// helper reports a directive at a single point rather than over a range.
type directivePos struct{ pos token.Pos }

func (d directivePos) Pos() token.Pos { return d.pos }
func (d directivePos) End() token.Pos { return d.pos }

var _ ast.Node = directivePos{}
