// nolint:revive
package util

// ContainsAllString checks if a string slice contains all the specified substring elements.
//
// Parameters:
//   - slice: The slice to check.
//   - sub: The list of strings to look for in the slice.
//
// Returns:
//   - bool: True if all elements in 'sub' are present in 'slice'.
func ContainsAllString(slice []string, sub ...string) bool {
	contains := make(map[string]bool)
	for _, v := range sub {
		contains[v] = false
	}

	for _, v := range slice {
		if _, ok := contains[v]; ok {
			contains[v] = true
		}
	}

	for _, v := range contains {
		if !v {
			return false
		}
	}

	return true
}

// UniqueString returns a new slice containing only unique strings from the input slice.
// Order is preserved based on the first occurrence of each string.
//
// Parameters:
//   - slice: The input string slice.
//
// Returns:
//   - []string: A new slice with duplicates removed.
func UniqueString(slice []string) []string {
	seen := make(map[string]struct{})
	res := make([]string, 0)

	for _, s := range slice {
		if _, ok := seen[s]; !ok {
			res = append(res, s)
			seen[s] = struct{}{}
		}
	}

	return res
}
