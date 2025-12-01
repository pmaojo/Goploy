package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/subosito/gotenv"
)

// envSetter defines the function signature for setting environment variables.
type envSetter = func(key string, value string) error

// DotEnvTryLoad attempts to load and apply environment variables from a specified .env file.
//
// If the file does not exist, it silently returns.
// If the file exists and is loaded successfully, it logs a warning indicating that ENV variables are being overridden.
// If an error occurs during loading (other than file not found), it panics.
//
// This function is intended for local development to inject secrets.
//
// Parameters:
//   - absolutePathToEnvFile: The absolute path to the .env file.
//   - setEnvFn: A function to set environment variables (e.g., os.Setenv).
func DotEnvTryLoad(absolutePathToEnvFile string, setEnvFn envSetter) {
	err := DotEnvLoad(absolutePathToEnvFile, setEnvFn)

	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Panic().Err(err).Str("envFile", absolutePathToEnvFile).Msg(".env parse error!")
		}
	} else {
		log.Warn().Str("envFile", absolutePathToEnvFile).Msg(".env overrides ENV variables!")
	}
}

// DotEnvLoad forces the loading of environment variables from a specified .env file, overriding existing ones.
//
// This is useful for injecting configuration locally or in tests.
//
// Parameters:
//   - absolutePathToEnvFile: The absolute path to the .env file.
//   - setEnvFn: A function to set environment variables (e.g., os.Setenv or t.Setenv).
//
// Returns:
//   - error: An error if the file cannot be opened, parsed, or if setting a variable fails.
func DotEnvLoad(absolutePathToEnvFile string, setEnvFn envSetter) error {
	file, err := os.Open(absolutePathToEnvFile)
	if err != nil {
		return fmt.Errorf("failed to open .env file: %w", err)
	}

	defer file.Close()

	envs, err := gotenv.StrictParse(file)
	if err != nil {
		return fmt.Errorf("failed to parse .env file: %w", err)
	}

	for key, value := range envs {
		if err := setEnvFn(key, value); err != nil {
			return fmt.Errorf("failed to set environment variable: %w", err)
		}
	}

	return nil
}
