// nolint:revive
package util

import "github.com/go-openapi/swag"

const (
	// centFactor is the conversion factor between major currency units and cents (100).
	centFactor = 100
)

// Int64PtrWithCentsToFloat64Ptr converts a pointer to an int64 (in cents) to a pointer to a float64 (in major units).
//
// Parameters:
//   - c: A pointer to the amount in cents.
//
// Returns:
//   - *float64: A pointer to the amount in major units, or nil if the input is nil.
func Int64PtrWithCentsToFloat64Ptr(c *int64) *float64 {
	if c == nil {
		return nil
	}

	return Int64WithCentsToFloat64Ptr(*c)
}

// Int64WithCentsToFloat64Ptr converts an int64 (in cents) to a pointer to a float64 (in major units).
//
// Parameters:
//   - c: The amount in cents.
//
// Returns:
//   - *float64: A pointer to the amount in major units.
func Int64WithCentsToFloat64Ptr(c int64) *float64 {
	return swag.Float64(float64(c) / centFactor)
}

// IntPtrWithCentsToFloat64Ptr converts a pointer to an int (in cents) to a pointer to a float64 (in major units).
//
// Parameters:
//   - c: A pointer to the amount in cents.
//
// Returns:
//   - *float64: A pointer to the amount in major units, or nil if the input is nil.
func IntPtrWithCentsToFloat64Ptr(c *int) *float64 {
	if c == nil {
		return nil
	}

	return IntWithCentsToFloat64Ptr(*c)
}

// IntWithCentsToFloat64Ptr converts an int (in cents) to a pointer to a float64 (in major units).
//
// Parameters:
//   - c: The amount in cents.
//
// Returns:
//   - *float64: A pointer to the amount in major units.
func IntWithCentsToFloat64Ptr(c int) *float64 {
	return swag.Float64(float64(c) / centFactor)
}

// Float64PtrToInt64PtrWithCents converts a pointer to a float64 (in major units) to a pointer to an int64 (in cents).
//
// Parameters:
//   - f: A pointer to the amount in major units.
//
// Returns:
//   - *int64: A pointer to the amount in cents, or nil if the input is nil.
func Float64PtrToInt64PtrWithCents(f *float64) *int64 {
	if f == nil {
		return nil
	}

	return swag.Int64(Float64PtrToInt64WithCents(f))
}

// Float64PtrToInt64WithCents converts a pointer to a float64 (in major units) to an int64 (in cents).
//
// Parameters:
//   - f: A pointer to the amount in major units.
//
// Returns:
//   - int64: The amount in cents.
func Float64PtrToInt64WithCents(f *float64) int64 {
	return int64(swag.Float64Value(f) * centFactor)
}

// Float64ToInt64WithCents converts a float64 (in major units) to an int64 (in cents).
//
// Parameters:
//   - f: The amount in major units.
//
// Returns:
//   - int64: The amount in cents.
func Float64ToInt64WithCents(f float64) int64 {
	return int64(f * centFactor)
}

// Float64PtrToIntPtrWithCents converts a pointer to a float64 (in major units) to a pointer to an int (in cents).
//
// Parameters:
//   - f: A pointer to the amount in major units.
//
// Returns:
//   - *int: A pointer to the amount in cents, or nil if the input is nil.
func Float64PtrToIntPtrWithCents(f *float64) *int {
	if f == nil {
		return nil
	}

	return swag.Int(Float64PtrToIntWithCents(f))
}

// Float64PtrToIntWithCents converts a pointer to a float64 (in major units) to an int (in cents).
//
// Parameters:
//   - f: A pointer to the amount in major units.
//
// Returns:
//   - int: The amount in cents.
func Float64PtrToIntWithCents(f *float64) int {
	return int(swag.Float64Value(f) * centFactor)
}
