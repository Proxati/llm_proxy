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
	"github.com/proxati/llm_proxy/v2/proxy/addons/megadumper/formatters"
	"github.com/proxati/llm_proxy/v2/proxy/addons/megadumper/writers"
	"github.com/proxati/llm_proxy/v2/schema"
)

type MegaDumpAddon struct {
	px.BaseAddon
	formatter         formatters.MegaDumpFormatter
	logSources        config.LogSourceConfig
	writers           []writers.MegaDumpWriter
	filterReqHeaders  []string
	filterRespHeaders []string
	wg                sync.WaitGroup
	closed            atomic.Bool
}

// Requestheaders is a callback that will receive a "flow" from the proxy, will create a
// NewLogDumpContainer and will use the embedded writers to finally write the log.
func (d *MegaDumpAddon) Requestheaders(f *px.Flow) {
	logger := slog.With("addon", "MegaDumpAddon.Requestheaders", "URL", f.Request.URL, "StatusCode", "ID", f.Id.String())

	if d.closed.Load() {
		logger.Warn("MegaDumpAddon is being closed, not logging a request")
		return
	}

	start := time.Now()

	d.wg.Add(1) // for blocking this addon during shutdown in .Close()
	go func() {
		defer d.wg.Done()
		<-f.Done() // block this goroutine until the entire flow is done
		doneAt := time.Since(start).Milliseconds()

		// load the selected fields into a container object
		dumpContainer, err := schema.NewLogDumpContainer(f, d.logSources, doneAt, d.filterReqHeaders, d.filterRespHeaders)
		if err != nil {
			logger.Error("could not create LogDumpContainer", "error", err)
			return
		}

		id := f.Id.String() // TODO: is the internal request ID unique enough?

		// format the container object, reformatted into a byte array
		formattedDump, err := d.formatter.Read(dumpContainer)
		if err != nil {
			logger.Error("could not format LogDumpContainer", "error", err)
			return
		}

		// write the formatted log data to... somewhere
		for _, w := range d.writers {
			if w == nil {
				logger.Error("Writer is nil, skipping")
				continue
			}
			_, err := w.Write(id, formattedDump)
			if err != nil {
				logger.Error("could not write log", "error", err)
				continue
			}
		}
	}()
}

func (d *MegaDumpAddon) String() string {
	return "MegaDumpAddon"
}

func (d *MegaDumpAddon) Close() error {
	if !d.closed.Swap(true) {
		slog.Debug("Waiting for MegaDumpAddon shutdown...")
		d.wg.Wait()
	}

	return nil
}

// newLogDestinations parses the logTarget string and returns a slice of log destinations.
// The actual validation of log destinations happens in formatter. No validation here!
func newLogDestinations(logTarget string) ([]md.LogDestination, error) {
	if logTarget == "" {
		slog.Debug("logTarget empty, defaulting to stdout")
		return []md.LogDestination{md.WriteToStdOut}, nil
	}

	var logDestinations []md.LogDestination
	slog.Debug("set log output directory", "logTarget", logTarget)
	logDestinations = append(logDestinations, md.WriteToDir)

	return logDestinations, nil
}

// formatPicker will setup the log formatter object, which converts the config enum to a
// local enum, which is used only inside this package
func formatPicker(format config.TrafficLogFormat) (formatters.MegaDumpFormatter, error) {
	var f formatters.MegaDumpFormatter

	switch format {
	case config.TrafficLog_JSON:
		slog.Debug("Traffic logging format set to JSON")
		f = &formatters.JSON{}
	case config.TrafficLog_TXT:
		slog.Debug("Traffic logging format set to text")
		f = &formatters.PlainText{}
	default:
		return nil, fmt.Errorf("invalid log format: %v", format)
	}

	return f, nil
}

// newWriters creates and configured the writer objects based on the log destinations and other parameters
func newWriters(logDestinations []md.LogDestination, logTarget string, f formatters.MegaDumpFormatter) ([]writers.MegaDumpWriter, error) {
	var w = make([]writers.MegaDumpWriter, 0)
	for _, logDest := range logDestinations {
		switch logDest {
		case md.WriteToDir:
			slog.Debug("Directory logger enabled")
			dirWriter, err := writers.NewToDir(logTarget, f)
			if err != nil {
				return nil, err
			}
			w = append(w, dirWriter)
		case md.WriteToStdOut:
			slog.Debug("Standard out logger enabled")
			stdoutWriter, err := writers.NewToStdOut()
			if err != nil {
				return nil, err
			}
			w = append(w, stdoutWriter)
		default:
			return nil, fmt.Errorf("invalid log destination: %v", logDest)
		}
	}
	return w, nil
}

// NewMegaDumpAddon creates a new dumper that creates a new log file for each request
func NewMegaDumpAddon(
	logTarget string, // output directory
	logFormatConfig config.TrafficLogFormat, // what file format to write the traffic logs
	logSources config.LogSourceConfig, // which fields from the transaction to log
	filterReqHeaders, filterRespHeaders []string, // which headers to filter out
) (*MegaDumpAddon, error) {

	logDestinations, err := newLogDestinations(logTarget)
	if err != nil {
		return nil, fmt.Errorf("log destination validation error: %v", err)
	}

	f, err := formatPicker(logFormatConfig)
	if err != nil {
		return nil, fmt.Errorf("log format validation error: %v", err)
	}

	w, err := newWriters(logDestinations, logTarget, f)
	if err != nil {
		return nil, fmt.Errorf("writer creation error: %v", err)
	}

	mda := &MegaDumpAddon{
		formatter:         f,
		logSources:        logSources,
		writers:           w,
		filterReqHeaders:  filterReqHeaders,
		filterRespHeaders: filterRespHeaders,
	}

	mda.closed.Store(false) // initialize the atomic bool with closed = false
	slog.Debug("Created MegaDirDumper", "sources", logSources.String(), "writers", len(w))
	return mda, nil
}
