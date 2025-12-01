package util

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindProjectRoot recursively searches up the directory tree to find the project root.
// It identifies the project root by the presence of a "go.mod" file.
//
// Parameters:
//   - start: The directory path to start the search from.
//
// Returns:
//   - string: The absolute path to the project root directory.
//   - error: An error if the project root cannot be found or if there is a file system error.
func FindProjectRoot(start string) (string, error) {
	dir := filepath.Clean(start)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir || parent == "." {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("could not find project root from %s", start)
}
