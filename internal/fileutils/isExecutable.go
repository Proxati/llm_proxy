package fileutils

import (
	"fmt"
	"os"
)

// IsExecutable checks if the command is executable by reading the file permissions
func IsExecutable(command string) error {
	fileInfo, err := os.Stat(command)
	if err != nil {
		return fmt.Errorf("unable to stat file: %w", err)
	}

	// Check if the file is executable
	if fileInfo.Mode()&0111 == 0 {
		return fmt.Errorf("command is not executable")
	}

	return nil
}
