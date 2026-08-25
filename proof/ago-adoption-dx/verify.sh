#!/bin/sh
set -eu

failures=0

check() {
	name=$1
	shift
	if "$@"; then
		printf 'PASS: %s\n' "$name"
	else
		printf 'FAIL: %s\n' "$name"
		failures=$((failures + 1))
	fi
}

check "README documents go get -tool" \
	grep -q 'go get -tool github.com/agentstation/ago/cmd/ago' README.md
check "README library example matches Check signature" \
	sh -c '! grep -q "ago.Check(ctx" README.md'
check "package docs do not claim the shipped command is a vet tool" \
	sh -c '! grep -q "through go vet with -vettool" ago.go'
check "pkg.go.dev license files contain no pointer file" \
	sh -c 'test ! -f LICENSE'
check "pkg.go.dev recognizes each license text" \
	sh -c 'cd proof/ago-adoption-dx/licensecheck && go run . ../../../LICENSE-APACHE ../../../LICENSE-MIT'
check "README contains no duplicated generics sentence" \
	sh -c 'test "$(grep -o "generic containers in the language without giving programmers access" README.md | wc -l | tr -d " ")" -le 1'

printf 'Summary: %d failed\n' "$failures"
test "$failures" -eq 0
