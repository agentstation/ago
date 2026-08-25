# Surveying the standard library

Every count in the rule reference comes from a measurement, not from memory. This
document holds the method so you can reproduce and update the numbers.

## Population

All `.go` files under `$(go env GOROOT)/src`, excluding any `testdata/`,
`vendor/`, or `_asm/` directory. Most counts further exclude `_test.go` files.
A rule that fires only in tests is a different argument. State which
population a number came from. Mixing them is how the first version of this
rule reference could otherwise quote an incorrect figure.

As of `go1.26.5`: **5,594** files in total, **3,916** of them non-test.

## Text-level counts

Adequate for constructs a regular expression can find without ambiguity.

```sh
python3 - <<'PY'
import os, re, subprocess
src = os.path.join(subprocess.run(["go","env","GOROOT"],
                                  capture_output=True, text=True).stdout.strip(), "src")
files = []
for dp, dn, fn in os.walk(src):
    dn[:] = [d for d in dn if d not in ("testdata", "vendor", "_asm")]
    files += [os.path.join(dp, f) for f in fn
              if f.endswith(".go") and not f.endswith("_test.go")]

def read(p):
    return open(p, encoding="utf-8", errors="replace").read()

print("non-test files:", len(files))
print("containing ':=':", sum(1 for p in files if ":=" in read(p)))
print("goto statements:", sum(len(re.findall(r'^\s*goto\s+\w+', read(p), re.M)) for p in files))
print("with dot imports:", sum(1 for p in files if re.search(r'^\s*\.\s+"', read(p), re.M)))
PY
```

## Declaration-level counts

Anything about declarations needs a parser. A regular expression cannot tell a
generic *declaration* from a generic *instantiation*. That is the error that
produced the "~430 files declare generics" figure in an early draft. The real
number is 115.

Use `go/ast`: walk each file, and count a `*ast.FuncDecl` or `*ast.TypeSpec`
whose `TypeParams` is non-nil. For embedded fields, count `*ast.StructType`
fields whose `Names` is empty. The full program used for the rule-reference figures is
short enough to rewrite from that description. Keep it out of the module so it
does not become a dependency.

## Current figures

Measured against `go1.26.5`, non-test files only.

| Measure | Count |
| --- | --- |
| Files | 3,916 |
| Files containing `:=` | 2,522 |
| Files declaring a generic func or type | 115 |
| `goto` statements | 522 (19 generated) |
| Files with dot imports | 53 |
| Embedded struct fields | 680 across 273 files |
| Embedded `sync.Mutex` / `sync.RWMutex` | 24 |

`goto` concentrates in `syscall` (120), `runtime` (77), and the two type
checkers (71). Dot imports outside tests are almost entirely `go/types` and
`cmd/compile/internal/types2` importing `internal/types/errors` (26 files
each). 224 `_test.go` files use them.
