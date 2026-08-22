package noselfref

type Cloneable[T any] interface{ Clone() T }
type Edge[N any] interface{ Ends() (N, N) }

// The Go 1.26 change: the declared type names itself in its own type
// parameter list.
type Adder[A Adder[A]] interface { // want `type Adder refers to itself`
	Add(A) A
}

// F-bounded polymorphism written through a separate interface. Legal since
// Go 1.18, so this rule does not report it.
type Node[T Cloneable[T]] struct{ v T }

// A constraint that refers to a sibling type parameter. Also legal since
// Go 1.18.
type Graph[N any, E Edge[N]] struct {
	nodes []N
	edges []E
}

// Ordinary generics are untouched.
type Pair[K comparable, V any] struct {
	k K
	v V
}
