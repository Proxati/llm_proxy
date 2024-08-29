package addons

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	px "github.com/proxati/mitmproxy/proxy"

	"github.com/proxati/llm_proxy/v2/config"
	md "github.com/proxati/llm_proxy/v2/proxy/addons/megadumper"
	"github.com/proxati/llm_proxy/v2/schema"
	"github.com/proxati/llm_proxy/v2/schema/proxyadapters/mitm"
)

type MegaTrafficDumper struct {
	px.BaseAddon
	logSources            config.LogSourceConfig
	logDestinationConfigs []md.LogDestination
	filterReqHeaders      *config.HeaderFilterGroup
	filterRespHeaders     *config.HeaderFilterGroup
	wg                    sync.WaitGroup
	closed                atomic.Bool
	logger                *slog.Logger
}

// Requestheaders is a callback that will receive a "flow" from the proxy, will create a
// NewLogDumpContainer and will use the embedded writers to finally write the log.
func (d *MegaTrafficDumper) Requestheaders(f *px.Flow) {
	logger := configLoggerFieldsWithFlow(d.logger, f)

	if d.closed.Load() {
		logger.Warn("MegaDumpAddon is being closed, not logging a request")
		return
	}

	// store a copy of the request in a FlowAdapter right away
	fa := &mitm.FlowAdapter{}
	fa.SetRequest(f.Request)

	d.wg.Add(1) // for blocking this addon during shutdown in .Close()
	go func() {
		logger.Debug("Request starting...")
		defer d.wg.Done()
		start := time.Now()
		<-f.Done() // block this goroutine until the entire flow is done
		doneAt := time.Since(start).Milliseconds()
		logger := configLoggerFieldsWithFlow(d.logger, f)

		// save the other fields in the FlowAdapter
		fa.SetFlow(f)

		// load the selected fields into a container object
		dumpContainer := d.convertFlowToLogDump(logger, fa, doneAt)

		// write the formatted log data to... somewhere
		d.sendToLogDestinations(logger, f.Id.String(), dumpContainer)

		logger.Debug("Request completed")
	}()
}

func (d *MegaTrafficDumper) String() string {
	return "MegaTrafficDumper"
}

func (d *MegaTrafficDumper) Close() error {
	if !d.closed.Swap(true) {
		d.logger.Debug("Closing...")
		d.wg.Wait()
	}

	return nil
}

// convertFlowToLogDump creates a LogDumpContainer from a px.Flow object
func (d *MegaTrafficDumper) convertFlowToLogDump(logger *slog.Logger, flowAdapter *mitm.FlowAdapter, doneAt int64) *schema.LogDumpContainer {
	// load the selected fields into a container object
	dumpContainer, err := schema.NewLogDumpContainer(flowAdapter, d.logSources, doneAt, d.filterReqHeaders, d.filterRespHeaders)
	if err != nil {
		logger.Error("Could not create LogDumpContainer", "error", err)
		return nil
	}
	return dumpContainer
}

// sendToLogDestinations writes the log data to the configured log destinations
func (d *MegaTrafficDumper) sendToLogDestinations(logger *slog.Logger, id string, dumpContainer *schema.LogDumpContainer) {
	for _, ldc := range d.logDestinationConfigs {
		wLogger := logger.With("logDestinationConfig", ldc.String())

		bytesWritten, err := ldc.Write(id, dumpContainer)
		if err != nil {
			wLogger.Error("Could not write log", "error", err)
			continue
		}
		wLogger.Info("Wrote log", "bytesWritten", bytesWritten)
	}
}

// NewMegaTrafficDumperAddon creates a new dumper that creates a new log file for each request
func NewMegaTrafficDumperAddon(
	logger *slog.Logger, // the DI'd logger
	logTarget string, // output directory
	logFormatConfig config.LogFormat, // what file format to write the traffic logs
	logSources config.LogSourceConfig, // which fields from the transaction to log
	filterReqHeaders *config.HeaderFilterGroup, // which headers to filter out from the request before logging
	filterRespHeaders *config.HeaderFilterGroup, // which headers to filter out from the response before logging
) (*MegaTrafficDumper, error) {
	logger = logger.WithGroup("addons.MegaTrafficDumper")
	logger.Debug("Set log output", "logTarget", logTarget)

	logDestinationConfigs, err := md.NewLogDestinations(logger, logTarget, logFormatConfig)
	if err != nil {
		return nil, fmt.Errorf("log destination validation error: %v", err)
	}
	for _, dest := range logDestinationConfigs {
		logger.Debug("Configured log destination", "destination", dest.String())
	}

	mTD := &MegaTrafficDumper{
		logSources:            logSources,
		logDestinationConfigs: logDestinationConfigs,
		filterReqHeaders:      filterReqHeaders,
		filterRespHeaders:     filterRespHeaders,
		logger:                logger,
	}

	mTD.closed.Store(false) // initialize the atomic bool with closed = false
	return mTD, nil
}
