---
name: Rule proposal
about: Propose a new restriction
title: 'rule: no-'
labels: rule proposal
assignees: ''
---

<!-- ago is restriction-only. A rule must forbid legal Go, and working around
     it must never require adding syntax. See CONTRIBUTING.md. -->

**The construct**

```go
// The code the rule would reject.
```

**The replacement**

```go
// What you would write instead. This must be ordinary Go, not new syntax.
```

**What it costs a reader**

<!-- Say what someone reading this code cold has to do that they would not
     otherwise. A second place to look. A scroll upward. An identifier whose
     origin is not local. "It is confusing" is not a rationale. -->

**On or off by default**

<!-- On by default only if the replacement is direct and mechanical and no
     reasonable codebase would miss the construct. If banning it is a taste
     call, propose it off by default. -->

**Evidence**

<!-- Measure it against the standard library. See docs/stdlib-survey.md.
     A count the standard library uses heavily does not disqualify a rule.
     It does make the rule off by default. -->

- Occurrences in the standard library (non-test):
- Method used:

**What the rule must NOT report**

<!-- The near-misses. Most defects in this tool were a rule catching a
     neighbouring construct that looked similar. -->

**Does it revert a Go release?**

<!-- If yes, which one, and verified how? If no, say so plainly rather than
     framing a preference as a revert. -->
