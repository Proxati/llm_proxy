package writers

import (
	"log/slog"

	"github.com/proxati/llm_proxy/v2/internal/fileutils"
	"github.com/proxati/llm_proxy/v2/proxy/addons/megadumper/formatters"
)

type ToDir struct {
	targetDir     string
	fileExtension string
	logger        *slog.Logger
}

// Write writes the bytes to a new file in the target directory
// Every time this method is called, it creates a new filename and writes the bytes to it
func (t *ToDir) Write(identifier string, bytes []byte) (int, error) {
	fileName := fileutils.CreateUniqueFileName(t.targetDir, identifier, t.fileExtension, 0)
	fileObj, err := fileutils.CreateNewFileFromFilename(fileName)
	if err != nil {
		return 0, err
	}
	defer fileObj.Close()
	return fileObj.Write(bytes)
}

// String returns the the name of this writer, and the target directory
func (t *ToDir) String() string {
	return "ToDir: " + t.targetDir
}

func newToDir(logger *slog.Logger, target string, formatter formatters.MegaDumpFormatter) (*ToDir, error) {
	err := fileutils.DirExistsOrCreate(target)
	if err != nil {
		return nil, err
	}

	return &ToDir{
		targetDir:     target,
		fileExtension: formatter.GetFileExtension(),
		logger:        logger,
	}, nil
}

func NewToDir(logger *slog.Logger, target string, formatter formatters.MegaDumpFormatter) (MegaDumpWriter, error) {
	return newToDir(logger, target, formatter)
}
