//go:build !scripts

// nolint:revive
package util

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog/log"
)

var (
	// projectRootDir holds the cached project root directory path.
	projectRootDir string
	// dirOnce ensures the project root directory is determined only once.
	dirOnce sync.Once
)

// GetProjectRootDir returns the absolute path to the project root directory for the running application.
//
// It first attempts to determine the directory of the executable.
// If the "PROJECT_ROOT_DIR" environment variable is set, it takes precedence.
//
// This function is excluded from "scripts" build tags.
//
// Returns:
//   - string: The path to the project root directory.
func GetProjectRootDir() string {
	dirOnce.Do(func() {
		ex, err := os.Executable()
		if err != nil {
			log.Panic().Err(err).Msg("Failed to get executable path while retrieving project root directory")
		}

		projectRootDir = filepath.Dir(ex)
	})

	if envRoot := os.Getenv("PROJECT_ROOT_DIR"); envRoot != "" {
		projectRootDir = envRoot
		return envRoot
	}

	return projectRootDir
}
