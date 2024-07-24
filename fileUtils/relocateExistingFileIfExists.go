package fileUtils

import (
	"os"
	"path/filepath"
)

// RelocateExistingFileIfExists checks if a file exists, and if it does, renames it to a unique name
func RelocateExistingFileIfExists(fileName string) error {
	if FileExists(fileName) {
		relocatedFile := CreateUniqueFileName(filepath.Dir(fileName), filepath.Base(fileName), "", 0)
		getLogger().Warn("File already exists, relocating", "oldFileName", fileName, "newFileName", relocatedFile)
		return os.Rename(fileName, relocatedFile)
	}
	return nil
}
