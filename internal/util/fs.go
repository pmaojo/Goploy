// nolint:revive
package util

import (
	"fmt"
	"os"
	"time"
)

// TouchFile mimics the unix `touch` command.
// It creates an empty file if the file doesn’t already exist.
// If the file already exists, it updates the modified time of the file to the current time.
//
// Parameters:
//   - absolutePathToFile: The absolute path to the file.
//
// Returns:
//   - time.Time: The modification time of the created/updated file.
//   - error: An error if the file creation or time update fails.
func TouchFile(absolutePathToFile string) (time.Time, error) {
	_, err := os.Stat(absolutePathToFile)

	if os.IsNotExist(err) {
		file, err := os.Create(absolutePathToFile)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to create file: %w", err)
		}

		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to stat file: %w", err)
		}

		return stat.ModTime(), nil
	}

	currentTime := time.Now().Local()
	err = os.Chtimes(absolutePathToFile, currentTime, currentTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to change file time: %w", err)
	}

	return currentTime, nil
}
