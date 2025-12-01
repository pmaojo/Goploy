// nolint:revive
package util

import (
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/text/language"
)

const (
	// mgmtSecretLen defines the length of the randomly generated management secret.
	mgmtSecretLen = 16
)

var (
	// mgmtSecret holds the cached management secret.
	mgmtSecret string
	// mgmtSecretOnce ensures the management secret is generated only once.
	mgmtSecretOnce sync.Once
)

// GetEnv retrieves the value of the environment variable named by the key.
// If the variable is not present, it returns the default value.
//
// Parameters:
//   - key: The name of the environment variable.
//   - defaultVal: The value to return if the environment variable is not set.
//
// Returns:
//   - string: The value of the environment variable or the default value.
func GetEnv(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return defaultVal
}

// GetEnvEnum retrieves an environment variable and ensures it matches one of the allowed values.
// If the value is invalid or missing, it falls back to the default value.
// It panics if the default value itself is not in the allowed values list.
//
// Parameters:
//   - key: The environment variable name.
//   - defaultVal: The default value to use if missing or invalid.
//   - allowedValues: A slice of valid string values.
//
// Returns:
//   - string: The validated environment variable value.
func GetEnvEnum(key string, defaultVal string, allowedValues []string) string {
	if !slices.Contains(allowedValues, defaultVal) {
		log.Panic().Str("key", key).Str("value", defaultVal).Msg("Default value is not in the allowed values list.")
	}

	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}

	if !slices.Contains(allowedValues, val) {
		log.Error().Str("key", key).Str("value", val).Msg("Value is not allowed. Fallback to default value.")
		return defaultVal
	}

	return val
}

// GetEnvAsInt retrieves an environment variable as an integer.
// Returns the default value if the variable is missing or cannot be parsed.
//
// Parameters:
//   - key: The environment variable name.
//   - defaultVal: The default integer value.
//
// Returns:
//   - int: The parsed integer value.
func GetEnvAsInt(key string, defaultVal int) int {
	strVal := GetEnv(key, "")

	if val, err := strconv.Atoi(strVal); err == nil {
		return val
	}

	return defaultVal
}

// GetEnvAsUint32 retrieves an environment variable as a uint32.
// Returns the default value if the variable is missing or cannot be parsed.
//
// Parameters:
//   - key: The environment variable name.
//   - defaultVal: The default uint32 value.
//
// Returns:
//   - uint32: The parsed uint32 value.
func GetEnvAsUint32(key string, defaultVal uint32) uint32 {
	strVal := GetEnv(key, "")

	if val, err := strconv.ParseUint(strVal, 10, 32); err == nil {
		return uint32(val)
	}

	return defaultVal
}

// GetEnvAsUint8 retrieves an environment variable as a uint8.
// Returns the default value if the variable is missing or cannot be parsed.
//
// Parameters:
//   - key: The environment variable name.
//   - defaultVal: The default uint8 value.
//
// Returns:
//   - uint8: The parsed uint8 value.
func GetEnvAsUint8(key string, defaultVal uint8) uint8 {
	strVal := GetEnv(key, "")

	if val, err := strconv.ParseUint(strVal, 10, 8); err == nil {
		return uint8(val)
	}

	return defaultVal
}

// GetEnvAsBool retrieves an environment variable as a boolean.
// Returns the default value if the variable is missing or cannot be parsed.
//
// Parameters:
//   - key: The environment variable name.
//   - defaultVal: The default boolean value.
//
// Returns:
//   - bool: The parsed boolean value.
func GetEnvAsBool(key string, defaultVal bool) bool {
	strVal := GetEnv(key, "")

	if val, err := strconv.ParseBool(strVal); err == nil {
		return val
	}

	return defaultVal
}

// GetEnvAsStringArr retrieves an environment variable as a slice of strings, split by a separator.
// The default separator is a comma if not specified.
//
// Parameters:
//   - key: The environment variable name.
//   - defaultVal: The default string slice.
//   - separator: Optional custom separator (defaults to ",").
//
// Returns:
//   - []string: The parsed string slice.
func GetEnvAsStringArr(key string, defaultVal []string, separator ...string) []string {
	strVal := GetEnv(key, "")

	if len(strVal) == 0 {
		return defaultVal
	}

	sep := ","
	if len(separator) >= 1 {
		sep = separator[0]
	}

	return strings.Split(strVal, sep)
}

// GetEnvAsStringArrTrimmed retrieves an environment variable as a slice of strings, split by a separator and trimmed of whitespace.
//
// Parameters:
//   - key: The environment variable name.
//   - defaultVal: The default string slice.
//   - separator: Optional custom separator (defaults to ",").
//
// Returns:
//   - []string: The parsed and trimmed string slice.
func GetEnvAsStringArrTrimmed(key string, defaultVal []string, separator ...string) []string {
	slc := GetEnvAsStringArr(key, defaultVal, separator...)

	for i := range slc {
		slc[i] = strings.TrimSpace(slc[i])
	}

	return slc
}

// GetEnvAsURL retrieves an environment variable as a URL.
// Panics if the value (or default) cannot be parsed as a valid URL.
//
// Parameters:
//   - key: The environment variable name.
//   - defaultVal: The default URL string.
//
// Returns:
//   - *url.URL: The parsed URL object.
func GetEnvAsURL(key string, defaultVal string) *url.URL {
	strVal := GetEnv(key, "")

	if len(strVal) == 0 {
		u, err := url.Parse(defaultVal)
		if err != nil {
			log.Panic().Str("key", key).Str("defaultVal", defaultVal).Err(err).Msg("Failed to parse default value for env variable as URL")
		}

		return u
	}

	u, err := url.Parse(strVal)
	if err != nil {
		log.Panic().Str("key", key).Str("strVal", strVal).Err(err).Msg("Failed to parse env variable as URL")
	}

	return u
}

// GetEnvAsLanguageTag retrieves an environment variable as a language.Tag.
// Panics if the value cannot be parsed.
//
// Parameters:
//   - key: The environment variable name.
//   - defaultVal: The default language tag.
//
// Returns:
//   - language.Tag: The parsed language tag.
func GetEnvAsLanguageTag(key string, defaultVal language.Tag) language.Tag {
	strVal := GetEnv(key, "")

	if len(strVal) == 0 {
		return defaultVal
	}

	tag, err := language.Parse(strVal)
	if err != nil {
		log.Panic().Str("key", key).Str("strVal", strVal).Err(err).Msg("Failed to parse env variable as language.Tag")
	}

	return tag
}

// GetEnvAsLanguageTagArr retrieves an environment variable as a slice of language.Tags.
// Panics if any value cannot be parsed.
//
// Parameters:
//   - key: The environment variable name.
//   - defaultVal: The default slice of language tags.
//   - separator: Optional custom separator (defaults to ",").
//
// Returns:
//   - []language.Tag: The parsed slice of language tags.
func GetEnvAsLanguageTagArr(key string, defaultVal []language.Tag, separator ...string) []language.Tag {
	strVal := GetEnv(key, "")

	if len(strVal) == 0 {
		return defaultVal
	}

	sep := ","
	if len(separator) >= 1 {
		sep = separator[0]
	}

	splitString := strings.Split(strVal, sep)
	res := []language.Tag{}
	for _, s := range splitString {
		tag, err := language.Parse(s)
		if err != nil {
			log.Panic().Str("key", key).Str("itemVal", s).Err(err).Msg("Failed to parse item value from env variable as language.Tag")
		}
		res = append(res, tag)
	}

	return res
}

// GetMgmtSecret returns the management secret for the app server.
// It retrieves the value from the environment or generates a random secure string if not set.
// The generated secret is cached and returned on subsequent calls.
//
// Parameters:
//   - envKey: The environment variable key to check first.
//
// Returns:
//   - string: The management secret.
func GetMgmtSecret(envKey string) string {
	val := GetEnv(envKey, "")

	if len(val) > 0 {
		return val
	}

	mgmtSecretOnce.Do(func() {
		var err error
		mgmtSecret, err = GenerateRandomHexString(mgmtSecretLen)
		if err != nil {
			log.Panic().Err(err).Msg("Failed to generate random management secret")
		}

		log.Warn().Str("envKey", envKey).Str("mgmtSecret", mgmtSecret).Msg("Could not retrieve management secret from env key, using randomly generated one")
	})

	return mgmtSecret
}

// GetEnvAsLocation retrieves an environment variable as a time.Location.
// Panics if the location cannot be loaded.
//
// Parameters:
//   - key: The environment variable name.
//   - defaultVal: The default timezone string (e.g., "UTC").
//
// Returns:
//   - *time.Location: The loaded time location.
func GetEnvAsLocation(key string, defaultVal string) *time.Location {
	strVal := GetEnv(key, "")

	if len(strVal) == 0 {
		l, err := time.LoadLocation(defaultVal)
		if err != nil {
			log.Panic().Str("key", key).Str("defaultVal", defaultVal).Err(err).Msg("Failed to parse default value for env variable as location")
		}

		return l
	}

	l, err := time.LoadLocation(strVal)
	if err != nil {
		log.Panic().Str("key", key).Str("strVal", strVal).Err(err).Msg("Failed to parse env variable as location")
	}

	return l
}
