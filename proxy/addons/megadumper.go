package addons

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	px "github.com/kardianos/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"

	"github.com/proxati/llm_proxy/config"
	md "github.com/proxati/llm_proxy/proxy/addons/megadumper"
	"github.com/proxati/llm_proxy/proxy/addons/megadumper/formatters"
	"github.com/proxati/llm_proxy/proxy/addons/megadumper/writers"
	"github.com/proxati/llm_proxy/schema"
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
	if d.closed.Load() {
		log.Warn("MegaDirDumper is being closed, not logging a request")
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
			log.Error(err)
			return
		}

		id := f.Id.String() // TODO: is the internal request ID unique enough?

		// format the container object, reformatted into a byte array
		formattedDump, err := d.formatter.Read(dumpContainer)
		if err != nil {
			log.Error(err)
			return
		}

		// write the formatted log data to... somewhere
		for _, w := range d.writers {
			if w == nil {
				log.Error("Writer is nil, skipping")
				continue
			}
			_, err := w.Write(id, formattedDump)
			if err != nil {
				log.Error(err)
				continue
			}
		}
	}()
}

func (d *MegaDumpAddon) String() string {
	return "MegaDirDumper"
}

func (d *MegaDumpAddon) Close() error {
	if !d.closed.Swap(true) {
		log.Debug("Waiting for MegaDirDumper shutdown...")
		d.wg.Wait()
	}

	return nil
}

// NewMegaDirDumper creates a new dumper that creates a new log file for each request
func NewMegaDirDumper(
	logTarget string, // output directory
	logFormatConfig config.TrafficLogFormat, // what file format to write the traffic logs
	logSources config.LogSourceConfig, // which fields from the transaction to log
	filterReqHeaders, filterRespHeaders []string, // which headers to filter out
) (*MegaDumpAddon, error) {
	var f formatters.MegaDumpFormatter
	var w = make([]writers.MegaDumpWriter, 0)
	var logDestinations []md.LogDestination
	var logFormat md.LogFormat

	if logTarget == "" {
		log.Debug("No traffic log output directory set, using stdout")
		logDestinations = append(logDestinations, md.WriteToStdOut)
	} else {
		log.Debugf("Traffic log output directory set to: %s", logTarget)
		logDestinations = append(logDestinations, md.WriteToDir)
	}

	// Setup the log format, converts the config enum to a local enum, used only inside this package
	switch logFormatConfig {
	case config.TrafficLog_JSON:
		log.Debug("Traffic logging format set to JSON")
		f = &formatters.JSON{}
		logFormat = md.Format_JSON
	case config.TrafficLog_TXT:
		log.Debug("Traffic logging format set to text")
		f = &formatters.PlainText{}
		logFormat = md.Format_PLAINTEXT
	default:
		return nil, fmt.Errorf("invalid log format: %v", logFormatConfig)
	}

	for _, logDest := range logDestinations {
		switch logDest {
		case md.WriteToDir:
			log.Debug("Directory logger enabled")
			dirWriter, err := writers.NewToDir(logTarget, logFormat)
			if err != nil {
				return nil, err
			}
			w = append(w, dirWriter)

		case md.WriteToFile:
			log.Debug("Single file logger enabled")
			fileWriter, err := writers.NewToFile(logTarget, logFormat)
			if err != nil {
				return nil, err
			}
			w = append(w, fileWriter)
		case md.WriteToStdOut:
			log.Debug("Standard out logger enabled")
			stdoutWriter, err := writers.NewToStdOut()
			if err != nil {
				return nil, err
			}
			w = append(w, stdoutWriter)
		default:
			return nil, fmt.Errorf("invalid log destination: %v", logDest)
		}
	}

	mda := &MegaDumpAddon{
		formatter:         f,
		logSources:        logSources,
		writers:           w,
		filterReqHeaders:  filterReqHeaders,
		filterRespHeaders: filterRespHeaders,
	}
	mda.closed.Store(false) // initialize the atomic bool with closed = false

	log.Debugf("Created MegaDirDumper with %s sources and %v writer(s)", logSources.String(), len(w))

	return mda, nil

}
