# ADX2 license repair evidence

Date: 2026-08-24

## Change

`COPYRIGHT` now holds the dual-license grant and contribution terms. The file
name is outside the set that pkg.go.dev scans as license text.

`LICENSE-APACHE` and `LICENSE-MIT` retain the exact license texts. Release
archives include all three files.

## Local classifier result

The proof command uses `github.com/google/licensecheck` v0.3.1, which pkg.go.dev
uses for license detection.

```text
../../../LICENSE-APACHE: coverage=100.0 license=Apache-2.0
../../../LICENSE-MIT: coverage=98.8 license=MIT
```

Both files exceed pkg.go.dev's 75-percent coverage threshold. The module has
no recognized license filename with unknown contents.

## Publication gate

The live verification stays blocked until the adoption pull request merges
and the patch release reaches the Go proxy. At that point, verify that:

- the Go proxy lists the patch version.
- pkg.go.dev lists Apache-2.0 and MIT without UNKNOWN.
- pkg.go.dev renders the package overview and API.
- the GitHub license API reports a recognized repository license.
