package nogoto

func Jump() {
	goto end // want `goto end; use labelled break or continue`
end:
}

// A suppressed goto, in the own-line form.
func Suppressed() {
	//goago:ignore no-goto -- hand-written state machine
	goto end
end:
}

// A suppressed goto, in the trailing form.
func SuppressedTrailing() {
	goto end //goago:ignore no-goto -- trailing directive
end:
}

// The wildcard suppresses every rule on the line.
func SuppressedWildcard() {
	goto end //goago:ignore * -- suppress everything here
end:
}

// A directive with no reason suppresses nothing.
func NotSuppressed() {
	//goago:ignore no-goto
	goto end // want `goto end`
end:
}

// Labelled break is the replacement and is never reported.
func Labelled() {
	total := 0
outer:
	for range 3 {
		for range 3 {
			total++
			if total > 4 {
				break outer
			}
			continue outer
		}
	}
}
