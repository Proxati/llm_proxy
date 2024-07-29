package fileUtils

import (
	"fmt"
	"os"
)

// DirExistsOrCreate checks if a directory exists, and creates it if it doesn't
func DirExistsOrCreate(dir string) error {
	getLogger().Debug("Checking if directory exists", "dirName", dir)
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			// If it doesn't exist, create it
			err := os.MkdirAll(dir, 0750)
			if err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			getLogger().Info("Directory created", "dirName", dir)
		} else {
			// If os.Stat failed for another reason, return the error
			return fmt.Errorf("failed to check if directory exists: %w", err)
		}
	}
	return nil
}
