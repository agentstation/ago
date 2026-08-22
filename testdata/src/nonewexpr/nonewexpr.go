package nonewexpr

type Config struct{ n int }

// Default is a variable, not a type. Distinguishing it from the type Config
// requires type information.
var Default Config

func value(n int) int { return n + 1 }

func Types() {
	_ = new(Config)
	_ = new([]byte)
	_ = new(map[string]int)
	_ = new(chan int)
	_ = new(struct{ a int })
	_ = new(interface{ M() })
	_ = new(func(int) int)
	_ = new(*Config)
}

func Expressions() {
	_ = new(value(1)) // want `new\(\) applied to an expression`
	_ = new(Default)  // want `new\(\) applied to an expression`
	_ = new(1 + 2)    // want `new\(\) applied to an expression`
}

// Shadowing the builtin means the call is not the builtin new at all.
func Shadowed() {
	new := func(int) *int { n := 0; return &n }
	_ = new(1)
}
