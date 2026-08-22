package ago

import "fmt"

// sprintf exists so that rule messages format through a single call site,
// which keeps the vet-style message punctuation consistent.
func sprintf(format string, args ...any) string {
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args...)
}
