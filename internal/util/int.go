// nolint:revive
package util

import (
	"math"

	"github.com/go-openapi/swag"
)

// IntPtrToInt64Ptr converts a pointer to an int to a pointer to an int64.
//
// Parameters:
//   - num: A pointer to an int.
//
// Returns:
//   - *int64: A pointer to an int64, or nil if the input is nil.
func IntPtrToInt64Ptr(num *int) *int64 {
	if num == nil {
		return nil
	}

	return swag.Int64(int64(*num))
}

// Int64PtrToIntPtr converts a pointer to an int64 to a pointer to an int.
//
// Parameters:
//   - num: A pointer to an int64.
//
// Returns:
//   - *int: A pointer to an int, or nil if the input is nil.
func Int64PtrToIntPtr(num *int64) *int {
	if num == nil {
		return nil
	}

	return swag.Int(int(*num))
}

// IntToInt32Ptr converts an int to a pointer to an int32.
// It returns nil if the value overflows or underflows the int32 range.
//
// Parameters:
//   - num: The int value to convert.
//
// Returns:
//   - *int32: A pointer to the converted int32 value, or nil if the value is out of bounds.
func IntToInt32Ptr(num int) *int32 {
	if num > math.MaxInt32 || num < math.MinInt32 {
		return nil
	}

	return swag.Int32(int32(num))
}
