package fileUtils

import (
	"os"
)

// FileExists returns true if the file exists
func FileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	return !os.IsNotExist(err)
}
