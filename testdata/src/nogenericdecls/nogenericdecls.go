package nogenericdecls

type Box[T any] struct{ v T } // want `type Box is generic`

func Map[F, T any](s []F, f func(F) T) []T { // want `func Map is generic`
	out := make([]T, 0, len(s))
	for _, v := range s {
		out = append(out, f(v))
	}
	return out
}

// A method on a generic type does not itself declare type parameters.
func (b *Box[T]) Get() T { return b.v }

type Plain struct{ n int }

func Ordinary(n int) int { return n + 1 }
