package writers

import log "github.com/sirupsen/logrus"

type ToStdOut struct{}

func (t *ToStdOut) Write(identifier string, bytes []byte) (int, error) {
	log.Infof("%v: %v", identifier, string(bytes))
	return len(bytes), nil
}

func newToStdOut() (*ToStdOut, error) {
	return &ToStdOut{}, nil
}

func NewToStdOut() (MegaDumpWriter, error) {
	return newToStdOut()
}
