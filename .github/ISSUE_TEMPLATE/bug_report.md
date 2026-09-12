---
name: Bug report
about: A rule reported the wrong thing, or goago failed to run
title: ''
labels: bug
assignees: ''
---

**What happened**

<!-- Include the exact command and its output. -->

```
$ goago ./...
```

**What you expected**

**Minimal reproduction**

<!-- The smallest Go file that shows the problem. Do not paste
     proprietary source. A synthetic reproduction is always preferable. -->

```go
package p
```

**Which rule**

<!-- If a specific rule applies, name it and paste `goago -explain <rule>`
     if the rationale seems to disagree with the behavior you saw. -->

**Environment**

- `goago -version`:
- `go version`:
- OS and architecture:
- `.goago.yml` (if any):
