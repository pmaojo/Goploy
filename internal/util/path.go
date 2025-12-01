// nolint:revive
package util

import (
	"path/filepath"
	"strings"
)

// FileNameWithoutExtension extracts the filename from a path, excluding the extension.
//
// Parameters:
//   - path: The file path.
//
// Returns:
//   - string: The filename without its extension. Returns an empty string if the path is invalid.
func FileNameWithoutExtension(path string) string {
	base := filepath.Base(path)
	if base == "." || base == "/" {
		return ""
	}

	return strings.TrimSuffix(base, filepath.Ext(path))
}

// FileNameAndExtension splits a path into its filename (without extension) and the extension itself.
//
// Parameters:
//   - path: The file path.
//
// Returns:
//   - string: The filename without extension.
//   - string: The extension (including the dot).
func FileNameAndExtension(path string) (string, string) {
	base := filepath.Base(path)
	if base == "." || base == "/" {
		return "", ""
	}

	extension := filepath.Ext(path)
	fileName := strings.TrimSuffix(base, extension)

	return fileName, extension
}
