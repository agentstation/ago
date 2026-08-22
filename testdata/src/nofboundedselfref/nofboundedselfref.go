// Package nofboundedselfref covers the interaction between
// no-f-bounded-constraints and no-self-referential-constraints.
//
// It needs Go 1.26 to type-check: a type naming itself in its own type
// parameter list is exactly what that release made legal.
package nofboundedselfref

type Cloneable[T any] interface{ Clone() T }

// Already reported by no-self-referential-constraints, so no-f-bounded stays
// quiet rather than reporting the same declaration twice.
type Adder[A Adder[A]] interface{ Add(A) A }

// An ordinary F-bounded constraint alongside it is still reported, so the
// guard above suppresses one declaration rather than the whole file.
type Node[T Cloneable[T]] struct{ v T } // want `type parameter T of type Node appears in its own constraint`
