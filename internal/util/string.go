// nolint:revive
package util

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-openapi/swag"
)

var (
	// StringSpaceReplacer is a compiled regex for matching one or more whitespace characters.
	StringSpaceReplacer = regexp.MustCompile(`\s+`)
)

// GenerateRandomBytes returns n random bytes securely generated using the system's default CSPRNG.
//
// An error will be returned if reading from the secure random number generator fails, at which point
// the returned result should be discarded and not used any further.
//
// Parameters:
//   - n: The number of random bytes to generate.
//
// Returns:
//   - []byte: A slice of random bytes.
//   - error: An error if generation fails.
func GenerateRandomBytes(n int) ([]byte, error) {
	result := make([]byte, n)

	_, err := rand.Read(result)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return result, nil
}

// GenerateRandomBase64String returns a string with n random bytes securely generated using the system's
// default CSPRNG in base64 encoding. The resulting string might not be of length n as the encoding for
// the raw bytes generated may vary.
//
// Parameters:
//   - n: The number of random bytes to generate before encoding.
//
// Returns:
//   - string: The base64 encoded string.
//   - error: An error if generation fails.
func GenerateRandomBase64String(n int) (string, error) {
	b, err := GenerateRandomBytes(n)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

// GenerateRandomHexString returns a string with n random bytes securely generated using the system's
// default CSPRNG in hexadecimal encoding. The resulting string might not be of length n as the encoding
// for the raw bytes generated may vary.
//
// Parameters:
//   - n: The number of random bytes to generate before encoding.
//
// Returns:
//   - string: The hex encoded string.
//   - error: An error if generation fails.
func GenerateRandomHexString(n int) (string, error) {
	b, err := GenerateRandomBytes(n)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

// CharRange represents a specific set of characters allowed in a random string.
type CharRange int

const (
	// CharRangeNumeric allows digits '0'-'9'.
	CharRangeNumeric CharRange = iota
	// CharRangeAlphaLowerCase allows lowercase letters 'a'-'z'.
	CharRangeAlphaLowerCase
	// CharRangeAlphaUpperCase allows uppercase letters 'A'-'Z'.
	CharRangeAlphaUpperCase
)

// GenerateRandomString returns a string with n random bytes securely generated using the system's
// default CSPRNG. The characters within the generated string will either be part of one or more supplied
// ranges of characters, or based on characters in the extra string supplied.
//
// Parameters:
//   - n: The length of the string to generate.
//   - ranges: A list of allowed character ranges.
//   - extra: Additional characters allowed in the string.
//
// Returns:
//   - string: The generated random string.
//   - error: An error if generation fails or no character set is provided.
func GenerateRandomString(n int, ranges []CharRange, extra string) (string, error) {
	var str strings.Builder

	if len(ranges) == 0 && len(extra) == 0 {
		return "", errors.New("random string can only be created if set of characters or extra string characters supplied")
	}

	validateFn := func(elem byte) bool {
		// IndexByte(string, byte) is basically Contains(string, string) without casting
		if strings.IndexByte(extra, elem) >= 0 {
			return true
		}

		for _, r := range ranges {
			switch r {
			case CharRangeNumeric:
				if elem >= '0' && elem <= '9' {
					return true
				}
			case CharRangeAlphaLowerCase:
				if elem >= 'a' && elem <= 'z' {
					return true
				}
			case CharRangeAlphaUpperCase:
				if elem >= 'A' && elem <= 'Z' {
					return true
				}
			}
		}

		return false
	}

	for str.Len() < n {
		buf, err := GenerateRandomBytes(n)
		if err != nil {
			return "", err
		}

		for _, b := range buf {
			if validateFn(b) {
				str.WriteByte(b)
			}
			if str.Len() >= n {
				break
			}
		}
	}

	return str.String(), nil
}

// ToUsernameFormat standardizes a string by lowercasing it and trimming whitespace.
//
// Parameters:
//   - s: The input string.
//
// Returns:
//   - string: The formatted string.
func ToUsernameFormat(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

// NonEmptyOrNil returns a pointer to passed string if it is not empty.
// Passing empty strings returns nil instead.
//
// Parameters:
//   - s: The string to check.
//
// Returns:
//   - *string: A pointer to the string or nil.
func NonEmptyOrNil(s string) *string {
	if len(s) > 0 {
		return swag.String(s)
	}

	return nil
}

// EmptyIfNil returns an empty string if the passed pointer is nil.
// Passing a pointer to a string will return the value of the string.
//
// Parameters:
//   - s: A pointer to a string.
//
// Returns:
//   - string: The string value or empty string.
func EmptyIfNil(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

// ContainsAll returns true if a string (str) contains all substrings (subs).
// The search handles overlapping matches logic specifically (e.g., characters are consumed).
//
// Parameters:
//   - str: The string to search in.
//   - subs: The substrings to look for.
//
// Returns:
//   - bool: True if all substrings are found.
func ContainsAll(str string, subs ...string) bool {
	subLen := len(subs)
	contains := make([]bool, subLen)
	indices := make([]int, subLen)
	substrings := make([][]rune, subLen)
	for i, substring := range subs {
		substrings[i] = []rune(substring)
	}

	for _, marked := range str {
		for i, sub := range substrings {
			if len(sub) == 0 {
				contains[i] = true
			}
			if !contains[i] && marked == sub[indices[i]] {
				indices[i]++
				if indices[i] >= len(sub) {
					contains[i] = true
				}
			}
		}
	}

	for _, c := range contains {
		if !c {
			return false
		}
	}

	return true
}
