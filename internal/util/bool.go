// nolint:revive
package util

// FalseIfNil returns the value of the boolean pointer if it is not nil.
// If the pointer is nil, it returns false.
//
// Parameters:
//   - b: A pointer to a boolean value.
//
// Returns:
//   - bool: The value of the boolean, or false if nil.
func FalseIfNil(b *bool) bool {
	if b == nil {
		return false
	}

	return *b
}
