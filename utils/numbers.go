package utils

import "golang.org/x/exp/constraints"

// Define the Max function that works for any type T that is comparable.
func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}
