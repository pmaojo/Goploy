// nolint:revive
package util

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LogFromContext retrieves a zerolog.Logger instance from the provided context.
//
// If the context contains a logger, it is returned.
// If the logger in the context is disabled (and not explicitly disabled via ShouldDisableLogger),
// it falls back to the global logger.
//
// Parameters:
//   - ctx: The context to retrieve the logger from.
//
// Returns:
//   - *zerolog.Logger: A pointer to the logger instance.
func LogFromContext(ctx context.Context) *zerolog.Logger {
	logger := log.Ctx(ctx)
	if logger.GetLevel() == zerolog.Disabled {
		if ShouldDisableLogger(ctx) {
			return logger
		}
		logger = &log.Logger
	}

	return logger
}

// LogFromEchoContext retrieves a zerolog.Logger instance from the Echo context.
//
// This is a wrapper around LogFromContext using the request's context.
//
// Parameters:
//   - c: The echo Context.
//
// Returns:
//   - *zerolog.Logger: A pointer to the logger instance.
func LogFromEchoContext(c echo.Context) *zerolog.Logger {
	return LogFromContext(c.Request().Context())
}

// LogLevelFromString parses a string into a zerolog.Level.
//
// If parsing fails, it logs an error and defaults to DebugLevel.
//
// Parameters:
//   - s: The string representation of the log level (e.g., "info", "debug", "error").
//
// Returns:
//   - zerolog.Level: The parsed log level.
func LogLevelFromString(s string) zerolog.Level {
	level, err := zerolog.ParseLevel(s)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to parse log level, defaulting to %s", zerolog.DebugLevel)
		return zerolog.DebugLevel
	}

	return level
}
