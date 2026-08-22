# Security Policy

## Supported versions

The latest tagged release receives security fixes. Development snapshots and
older releases do not receive separate security support.

## Reporting a vulnerability

Email `security@agentstation.ai`. Do not open a public issue for a suspected
vulnerability.

Include the affected version (`ago -version`), the Go toolchain version, the
command you ran, a minimal reproduction, and the security effect. Remove any
proprietary source from your report — a synthetic reproduction is always
preferable.

We will confirm receipt, assess the report, and coordinate remediation and
disclosure with you. Please do not disclose an unresolved report publicly
before that coordination is complete.

## Threat model

`ago` is a static analyzer. It reads Go source and configuration and writes a
report. It executes no code from the packages it analyzes.

In scope:

- A crafted `.ago.yml` or `//ago:ignore` directive that causes `ago` to crash,
  hang, or consume unbounded memory.
- A crafted source file that causes `ago` to silently skip analysis, so a
  violation goes unreported while the exit status stays `0`.
- Output injection — a finding message, file path, or rule name that escapes
  its encoding in `-format sarif` or `-format github` and forges an annotation
  or a code-scanning alert.
- Path traversal through an `exclude` pattern or a package pattern.

Out of scope:

- `ago` invokes the Go toolchain through `golang.org/x/tools/go/packages`,
  which loads and builds package metadata. Running `ago` on untrusted source
  is equivalent to running `go list` on it, and carries the same risk. Do not
  run `ago` on source you would not run `go build` on.
- Vulnerabilities in the Go toolchain or in `golang.org/x/tools`. Report those
  upstream; we will pick up the fix on the next release.
- A rule producing a false positive or a false negative. That is a correctness
  bug — please file it as a normal issue.
