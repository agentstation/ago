package nonakedreturn

import "errors"

// Named results plus a bare return.
func Named() (n int, err error) {
	n = 1
	return // want `naked return in Named`
}

// Unnamed results cannot return nakedly, so nothing is reported.
func Unnamed() (int, error) {
	return 1, nil
}

// An explicit return with named results is the fix and is not reported.
func Explicit() (n int, err error) {
	n = 1
	return n, nil
}

type T struct{}

func (t T) Method() (n int) {
	n = 1
	return // want `naked return in Method`
}

// A closure with its own named results owns its bare return.
func Closure() error {
	fn := func() (err error) {
		err = errors.New("x")
		return // want `naked return in the function literal`
	}
	return fn()
}

// A closure with no results, inside a function that has named results. The
// bare return belongs to the closure, so it is not reported.
func MixedScopes() (n int) {
	f := func() { return }
	f()
	n = 1
	return n
}
