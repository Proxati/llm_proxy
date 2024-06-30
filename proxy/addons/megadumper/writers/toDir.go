package writers

import (
	"github.com/proxati/llm_proxy/fileUtils"
	md "github.com/proxati/llm_proxy/proxy/addons/megadumper"
	log "github.com/sirupsen/logrus"
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
	log.Infof("Writing to file: %v", fileName)
	return fileObj.Write(bytes)
}

func newToDir(target string, logFormat md.LogFormat) (*ToDir, error) {
	err := fileUtils.DirExistsOrCreate(target)
	if err != nil {
		return nil, err
	}

	return &ToDir{
		targetDir:     target,
		fileExtension: logFormat.FileExtension(),
	}, nil
}

func NewToDir(target string, logFormat md.LogFormat) (MegaDumpWriter, error) {
	return newToDir(target, logFormat)
}
