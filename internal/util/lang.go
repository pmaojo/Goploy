// nolint:revive
package util

import (
	"sort"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// SortCollateStringSlice sorts a slice of strings based on the collation rules of a specific language.
//
// This is useful when the sorting order depends on locale-specific character rules (e.g., accents).
//
// Note: The slice passed is sorted in place.
//
// Parameters:
//   - slice: The slice of strings to sort.
//   - lang: The language tag defining the collation rules.
//   - options: Optional collation options (defaults to IgnoreCase and IgnoreWidth).
func SortCollateStringSlice(slice []string, lang language.Tag, options ...collate.Option) {
	if len(options) == 0 {
		options = []collate.Option{collate.IgnoreCase, collate.IgnoreWidth}
	}
	coll := collate.New(lang, options...)

	sort.Slice(slice, func(i int, j int) bool {
		return coll.CompareString(slice[i], slice[j]) < 0
	})
}
