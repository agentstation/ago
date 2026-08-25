# ADX3 adoption and documentation evidence

## Fail-before findings

The previous README made a global installation the primary path. It also
contained two interface defects:

- The library example called `ago.Check` with a context argument that the API
  does not accept.
- The package documentation said that the shipped command worked with
  `go vet -vettool`. Only a command built with an analysis driver such as
  `multichecker` has that protocol.

A clean module that installed ago through the Go tool directive also reported
`ago dev`. Release linker flags do not apply to `go get -tool` or
`go install` builds.

## Target behavior

The primary adoption path is repository-owned:

```sh
go get -tool github.com/agentstation/ago/cmd/ago@latest
go tool ago -init
go tool ago ./...
```

The README assigns one contract to each repository file: `go.mod` pins the
tool, `.ago.yml` selects policy, `AGENTS.md` instructs coding agents, and CI
enforces the result. Detailed rule and design material lives in
`docs/rules.md` and `docs/design.md`.

## Clean-module verification

The clean module installed the published `v0.1.0` command with `go get -tool`.
Go added the tool directive and its module requirements.
`go tool ago -init` wrote `.ago.yml`, and `go tool ago ./...` completed with
no findings.

The fixture then replaced the module with this worktree and required the next
patch version. The command read the version from Go build information:

```text
$ go tool ago -version
ago v0.1.1

$ go tool ago -list -format json
{
  "version": "v0.1.1",
  ...
}
```

`version_test.go` covers a release linker value, a pinned module version, a
local source build, an unrelated main module, and missing build information.

## Checks

The following checks passed on 2026-08-25:

```text
make check
make release-check
make prose
./proof/ago-adoption-dx/verify.sh
```

`make check` includes `go vet ./...`, `go test -race ./...`, a release-style
build, and ago's self-check. The strict prose gate checked 57 files with no
diagnostics. The proof verifier reported zero failures.
