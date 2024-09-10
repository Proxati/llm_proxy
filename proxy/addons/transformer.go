package addons

import (
	"log/slog"
	"sync"
	"sync/atomic"

	px "github.com/proxati/mitmproxy/proxy"
)

type TransformerAddon struct {
	px.BaseAddon
	wg     sync.WaitGroup
	closed atomic.Bool
	logger *slog.Logger
}

func (a *TransformerAddon) Request(f *px.Flow) {
	logger := configLoggerFieldsWithFlow(a.logger, f).WithGroup("Request")
	a.wg.Add(1)
	defer a.wg.Done()

	// TODO request body editing here
	logger.Debug("Done editing request body")
	panic("implement me")
}

func (a *TransformerAddon) Response(f *px.Flow) {
	logger := configLoggerFieldsWithFlow(a.logger, f).WithGroup("Response")
	a.wg.Add(1)
	defer a.wg.Done()

	// TODO: response body editing here
	logger.Debug("Done editing response body")
	panic("implement me")
}

func (d *TransformerAddon) String() string {
	return "BodyEditorAddon"
}

func (a *TransformerAddon) Close() error {
	if !a.closed.Swap(true) {
		a.logger.Debug("Closing...")
		a.wg.Wait()
	}

	return nil
}

func NewCBodyEditorAddon(
	logger *slog.Logger,
	storageEngineName string, // name of the storage engine to use
	cacheDir string, // output & cache storage directory
	filterReqHeaders, filterRespHeaders []string, // which headers to filter out
) (*TransformerAddon, error) {
	logger = logger.WithGroup("addons.BodyEditorAddon")
	return &TransformerAddon{
		logger: logger,
	}, nil
}
