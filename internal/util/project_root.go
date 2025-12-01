package util

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindProjectRoot walks parent directories from the start path until it finds a go.mod file.
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
