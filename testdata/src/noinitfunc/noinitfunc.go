package noinitfunc

var registry = map[string]int{}

func init() { // want `func init\(\)`
	registry["a"] = 1
}

// A method named init is not a package initializer.
type T struct{}

func (t T) init() {}

// An ordinary function named anything else is fine.
func Setup() { registry["b"] = 2 }
