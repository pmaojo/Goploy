// nolint:revive
package util

import (
	"context"
	"strings"
)

// CacheControlDirective represents bitmask flags for Cache-Control directives.
type CacheControlDirective uint8

const (
	// CacheControlDirectiveNoCache indicates the "no-cache" directive.
	CacheControlDirectiveNoCache CacheControlDirective = 1 << iota
	// CacheControlDirectiveNoStore indicates the "no-store" directive.
	CacheControlDirectiveNoStore
)

// HasDirective checks if a specific directive is set.
//
// Parameters:
//   - dir: The directive to check.
//
// Returns:
//   - bool: True if the directive is set, false otherwise.
func (d *CacheControlDirective) HasDirective(dir CacheControlDirective) bool { return *d&dir != 0 }

// AddDirective adds a specific directive to the bitmask.
//
// Parameters:
//   - dir: The directive to add.
func (d *CacheControlDirective) AddDirective(dir CacheControlDirective) { *d |= dir }

// ClearDirective removes a specific directive from the bitmask.
//
// Parameters:
//   - dir: The directive to remove.
func (d *CacheControlDirective) ClearDirective(dir CacheControlDirective) { *d &= ^dir }

// ToggleDirective toggles the state of a specific directive in the bitmask.
//
// Parameters:
//   - dir: The directive to toggle.
func (d *CacheControlDirective) ToggleDirective(dir CacheControlDirective) { *d ^= dir }

// String returns the string representation of the set directives, separated by pipes.
//
// Returns:
//   - string: A pipe-separated string of directive names (e.g., "no-cache|no-store").
func (d *CacheControlDirective) String() string {
	res := make([]string, 0)

	if d.HasDirective(CacheControlDirectiveNoCache) {
		res = append(res, "no-cache")
	}
	if d.HasDirective(CacheControlDirectiveNoStore) {
		res = append(res, "no-store")
	}

	return strings.Join(res, "|")
}

// ParseCacheControlDirective parses a single Cache-Control directive string.
//
// Parameters:
//   - d: The directive string (e.g., "no-cache").
//
// Returns:
//   - CacheControlDirective: The corresponding bitmask value, or 0 if unknown.
func ParseCacheControlDirective(d string) CacheControlDirective {
	parts := strings.Split(d, "=")
	switch strings.ToLower(parts[0]) {
	case "no-cache":
		return CacheControlDirectiveNoCache
	case "no-store":
		return CacheControlDirectiveNoStore
	default:
		return 0
	}
}

// ParseCacheControlHeader parses a full Cache-Control header value.
//
// Parameters:
//   - val: The Cache-Control header string (e.g., "no-cache, no-store").
//
// Returns:
//   - CacheControlDirective: The combined bitmask of all parsed directives.
func ParseCacheControlHeader(val string) CacheControlDirective {
	res := CacheControlDirective(0)

	directives := strings.Split(val, ",")
	for _, dir := range directives {
		res |= ParseCacheControlDirective(dir)
	}

	return res
}

// CacheControlDirectiveFromContext retrieves the CacheControlDirective from the context.
//
// Parameters:
//   - ctx: The context to retrieve the value from.
//
// Returns:
//   - CacheControlDirective: The value from the context, or 0 if not found or invalid.
func CacheControlDirectiveFromContext(ctx context.Context) CacheControlDirective {
	d := ctx.Value(CTXKeyCacheControl)
	if d == nil {
		return CacheControlDirective(0)
	}

	directive, ok := d.(CacheControlDirective)
	if !ok {
		return CacheControlDirective(0)
	}

	return directive
}
