package writers

import "log/slog"

type ToStdOut struct{}

func (t *ToStdOut) Write(identifier string, bytes []byte) (int, error) {
	slog.Info(string(bytes), "identifier", identifier)
	return len(bytes), nil
}

func newToStdOut() (*ToStdOut, error) {
	return &ToStdOut{}, nil
}

func NewToStdOut() (MegaDumpWriter, error) {
	return newToStdOut()
}
