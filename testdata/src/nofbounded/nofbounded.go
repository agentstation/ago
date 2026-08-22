package nofbounded

type Cloneable[T any] interface{ Clone() T }
type Edge[N any] interface{ Ends() (N, N) }

// A type parameter inside its own constraint.
type Node[T Cloneable[T]] struct{ v T } // want `type parameter T of type Node appears in its own constraint`

// The same shape on a function.
func Clone[T Cloneable[T]](x T) T { // want `type parameter T of func Clone appears in its own constraint`
	return x.Clone()
}

// A sibling reference is not the parameter's own constraint.
type Graph[N any, E Edge[N]] struct {
	nodes []N
	edges []E
}

// Already reported by no-self-referential-constraints, so this rule stays
// quiet rather than reporting the same declaration twice.
type Adder[A Adder[A]] interface{ Add(A) A }

type Pair[K comparable, V any] struct {
	k K
	v V
}
