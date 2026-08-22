package noembedded

import "sync"

type Base struct{ n int }

type Embeds struct {
	Base       // want `embedded field; give it a name`
	sync.Mutex // want `embedded field; give it a name`
	Named      Base
}

// Naming the field is the fix.
type Names struct {
	base Base
	mu   sync.Mutex
}

// Interface embedding has no field to name and is not reported.
type Reader interface{ Read() }

type ReadCloser interface {
	Reader
	Close()
}
