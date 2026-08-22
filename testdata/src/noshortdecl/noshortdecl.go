package noshortdecl

func Plain() int {
	n := 1 // want `short declaration in plain statement position`
	return n
}

func Var() int {
	var n = 1
	return n
}

// := is load-bearing in these positions and is not reported.
func LoadBearing(xs []int, x any) int {
	total := 0 // want `short declaration in plain statement position`
	for i, v := range xs {
		total += i + v
	}
	if n := len(xs); n > 0 {
		total += n
	}
	for i := 0; i < 3; i++ {
		total += i
	}
	switch t := x.(type) {
	case int:
		total += t
	}
	switch n := len(xs); n {
	case 0:
		total++
	}
	ch := make(chan int, 1) // want `short declaration in plain statement position`
	ch <- 1
	select {
	case v := <-ch:
		total += v
	}
	return total
}
