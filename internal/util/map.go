// nolint:revive
package util

// MergeStringMap merges the contents of the 'toMerge' map into the 'base' map.
//
// Keys that already exist in the 'base' map are NOT overwritten.
// This function modifies the 'base' map in place and returns it.
//
// Parameters:
//   - base: The destination map.
//   - toMerge: The source map containing key-value pairs to add.
//
// Returns:
//   - map[string]string: The modified 'base' map.
func MergeStringMap(base map[string]string, toMerge map[string]string) map[string]string {
	for k, v := range toMerge {
		if _, ok := base[k]; !ok {
			base[k] = v
		}
	}

	return base
}
