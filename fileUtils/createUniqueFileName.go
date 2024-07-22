package fileUtils

import (
	"fmt"
	"log/slog"
	"path/filepath"
)

// CreateUniqueFileName creates a new unused filename at the target directory.
// targetDir: the directory to create the file in.
// identifier: the base name of the file.
// fileExtension: the file extension suffix.
// attempt: the current attempt number, to prevent infinite loops.
func CreateUniqueFileName(
	targetDir, identifier, fileExtension string,
	attempt int,
) string {
	// check if the file already exists
	var fileName string
	if attempt < 1 {
		fileName = filepath.Join(targetDir, fmt.Sprintf("%s.%s", identifier, fileExtension))
	} else if attempt > 100000 {
		panic("Too many attempts to create a unique filename, crashing to prevent infinite loop.")
	} else {
		fileName = filepath.Join(targetDir, fmt.Sprintf("%s-%v.%s", identifier, attempt, fileExtension))
	}

	if FileExists(fileName) {
		slog.Warn("File already exists, trying again...", "fileName", fileName)
		return CreateUniqueFileName(targetDir, identifier, fileExtension, attempt+1)
	}

	return fileName
}
