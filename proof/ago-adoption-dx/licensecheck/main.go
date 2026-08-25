package main

import (
	"fmt"
	"os"

	"github.com/google/licensecheck"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: licensecheck FILE...")
		os.Exit(2)
	}
	failed := false
	for _, path := range os.Args[1:] {
		body, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			failed = true
			continue
		}
		coverage := licensecheck.Scan(body)
		fmt.Printf("%s: coverage=%.1f", path, coverage.Percent)
		for _, match := range coverage.Match {
			fmt.Printf(" license=%s", match.ID)
			if match.ID != "Apache-2.0" && match.ID != "MIT" {
				failed = true
			}
		}
		fmt.Println()
		if coverage.Percent < 75 || len(coverage.Match) == 0 {
			failed = true
		}
	}
	if failed {
		os.Exit(1)
	}
}
