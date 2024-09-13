package megadumper

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/internal/fileutils"
	"github.com/proxati/llm_proxy/v2/proxy/addons/megadumper/formatters"
	"github.com/proxati/llm_proxy/v2/proxy/addons/megadumper/writers"
	"github.com/proxati/llm_proxy/v2/schema"
)

// LogDestination is a struct that holds the configuration for a log destination.
// target: the target of the log destination (e.g., file path, rest API URL)
// writer: the writer to use for the log destination (e.g., to a dir, to rest API)
// formatter: the formatter to use for the log destination (e.g., JSON, TXT)
// logger: the logger used to print status to the terminal
type LogDestination struct {
	target    string
	writer    writers.MegaDumpWriter
	formatter formatters.MegaDumpFormatter
	logger    *slog.Logger
}

func (ld *LogDestination) String() string {
	if ld.writer == nil {
		return fmt.Sprintf("LogDestination: %s", ld.target)
	}

	return fmt.Sprintf("LogDestination: %s", ld.writer.String())
}

// Write writes a log dump container object to it's log destination. The
// formatter is responsible for converting the log dump container the correct
// format (json, text, etc) before writing.
func (ld *LogDestination) Write(identifier string, logDumpContainer schema.LogDumpContainer) (int, error) {
	bytes, err := ld.formatter.Read(&logDumpContainer)
	if err != nil {
		return 0, fmt.Errorf("could not format log dump container: %w", err)
	}
	return ld.writer.Write(identifier, bytes)
}

// NewLogDestinationConfig creates a new log destination configuration object to select and store:
// logger: the logger used to print status to the terminal
// logTarget: the target of the log destination as a comma-delimited string (e.g., file path, rest API URL)
// format: the format of the log destination (e.g., JSON, TXT)
func NewLogDestinations(
	logger *slog.Logger,
	logTarget string,
	format config.LogFormat,
) ([]LogDestination, error) {
	formatter, err := formatters.NewMegaDumpFormatter(format)
	if err != nil {
		return nil, fmt.Errorf("could not load the formatter: %w", err)
	}

	if logTarget == "" {
		// default to stdout if none selected
		writer, err := writers.NewToStdOut(logger, "", formatter)
		if err != nil {
			return nil, fmt.Errorf("could not create stdout writer: %w", err)
		}

		target := "stdout"
		ldc := LogDestination{
			target:    target,
			formatter: formatter,
			writer:    writer,
		}
		ldc.logger = logger.With("logDestination", ldc.String())
		return []LogDestination{ldc}, nil
	}

	targets := strings.Split(logTarget, ",")
	LDCs := make([]LogDestination, len(targets))

	for i, target := range targets {
		target = strings.TrimPrefix(target, "file://")
		target = strings.TrimSpace(target)
		if target == "" {
			continue
		}

		ld := LogDestination{
			target:    target,
			formatter: formatter,
		}

		if fileutils.IsValidFilePathFormat(target) {
			ld.writer, err = writers.NewToDir(logger, target, formatter)
			if err != nil {
				return nil, fmt.Errorf("could not create writer: %w", err)
			}

			ld.logger = logger.With("logDestination", ld.String())
			LDCs[i] = ld
			continue
		}

		if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
			ld.writer, err = writers.NewToAsyncREST(logger, target, formatter)
			if err != nil {
				return nil, fmt.Errorf("could not create writer: %w", err)
			}
			ld.logger = logger.With("logDestination", ld.String())
			LDCs[i] = ld
			continue
		}
		return nil, fmt.Errorf("target unhandled by log destination conditionals: %s", target)
	}

	if len(LDCs) == 0 {
		return nil, fmt.Errorf("no valid log destinations found")
	}

	return LDCs, nil
}
