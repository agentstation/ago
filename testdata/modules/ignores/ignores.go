package ignores

// A well-formed directive is not reported and suppresses its rule.
func Good() {
	//ago:ignore no-goto -- a reason that explains the exception
	goto end
end:
}

// A directive with no reason suppresses nothing and is itself reported.
func NoReason() {
	//ago:ignore no-goto
	goto end
end:
}

// A directive that names no rule.
func NoRule() {
	//ago:ignore
	goto end
end:
}

// A directive naming a rule that does not exist.
func UnknownRule() {
	//ago:ignore no-gotoo -- misspelled rule name
	goto end
end:
}

// A directive whose reason is empty after the separator.
func EmptyReason() {
	//ago:ignore no-goto --
	goto end
end:
}

// Prose that merely starts with the prefix is not a directive.
//
//ago:ignorecase is not a directive
func Prose() {}

// A well-formed directive that suppresses nothing is stale.
//
//ago:ignore no-dot-import -- nothing here dot-imports anything
func Stale() {}
