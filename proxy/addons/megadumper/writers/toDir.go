package writers

import (
	"github.com/proxati/llm_proxy/v2/fileUtils"
	"github.com/proxati/llm_proxy/v2/proxy/addons/megadumper/formatters"
)

type ToDir struct {
	targetDir     string
	fileExtension string
}

// Write writes the bytes to a new file in the target directory
// Every time this method is called, it creates a new filename and writes the bytes to it
func (t *ToDir) Write(identifier string, bytes []byte) (int, error) {
	fileName := fileUtils.CreateUniqueFileName(t.targetDir, identifier, t.fileExtension, 0)
	fileObj, err := fileUtils.CreateNewFileFromFilename(fileName)
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

func newToDir(target string, formatter formatters.MegaDumpFormatter) (*ToDir, error) {
	err := fileUtils.DirExistsOrCreate(target)
	if err != nil {
		return nil, err
	}

	return &ToDir{
		targetDir:     target,
		fileExtension: formatter.GetFileExtension(),
	}, nil
}

func NewToDir(target string, formatter formatters.MegaDumpFormatter) (MegaDumpWriter, error) {
	return newToDir(target, formatter)
}
