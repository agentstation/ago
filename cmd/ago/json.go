package main

import (
	"encoding/json"
	"io"
)

// encodeJSON writes v as indented JSON. Errors writing to stdout are not
// actionable, so the function drops them rather than reporting into the same
// stream.
func encodeJSON(w io.Writer, v any) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
