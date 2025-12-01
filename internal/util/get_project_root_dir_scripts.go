//go:build scripts

// nolint:revive
package util

import "os"

// GetProjectRootDir returns the path as string to the project_root while **scripts generation**.
// Note: This function replaces the original util.GetProjectRootDir when go runs with the "script" build tag.
// https://stackoverflow.com/questions/43215655/building-multiple-binaries-using-different-packages-and-build-tags
// Should be in sync with "scripts/internal/util/get_project_root_dir.go"
//
// Returns:
//   - string: The project root directory path, defaulting to "/app" if PROJECT_ROOT_DIR is not set.
func GetProjectRootDir() string {
	if val, ok := os.LookupEnv("PROJECT_ROOT_DIR"); ok {
		return val
	}

	return "/app"
}
