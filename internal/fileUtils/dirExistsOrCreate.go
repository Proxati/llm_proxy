package fileUtils

import (
	"fmt"
	"os"
)

// DirExistsOrCreate checks if a directory exists, and creates it if it doesn't
func DirExistsOrCreate(dir string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}
