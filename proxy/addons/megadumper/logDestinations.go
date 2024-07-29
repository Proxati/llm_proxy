package megadumper

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/internal/fileUtils"
	"github.com/proxati/llm_proxy/v2/proxy/addons/megadumper/formatters"
	"github.com/proxati/llm_proxy/v2/proxy/addons/megadumper/writers"
	"github.com/proxati/llm_proxy/v2/schema"
)

type LogDestinationConfig struct {
	target    string
	writer    writers.MegaDumpWriter
	formatter formatters.MegaDumpFormatter
	logger    *slog.Logger
}

func (ld *LogDestinationConfig) String() string {
	return fmt.Sprintf("%s.%s", ld.writer.String(), ld.formatter.GetFileExtension())
}

func (ld *LogDestinationConfig) Write(identifier string, logDumpContainer *schema.LogDumpContainer) (int, error) {
	bytes, err := ld.formatter.Read(logDumpContainer)
	if err != nil {
		return 0, fmt.Errorf("could not format log dump container: %w", err)
	}
	return ld.writer.Write(identifier, bytes)
}

// NewLogDestinationConfig creates a new log destination configuration object to select and store:
//
//	destination type (e.g., stdOut, dir, restAPI)
//	format (how to format the output, e.g., JSON, TXT)
//	target (text to describe the target, such as dir name or rest API URL)

func NewLogDestinationConfigs(
	logger *slog.Logger,
	logTarget string,
	format config.LogFormat,
) ([]LogDestinationConfig, error) {
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
		ldc := LogDestinationConfig{
			target:    target,
			formatter: formatter,
			writer:    writer,
		}
		ldc.logger = logger.With("logDestination", ldc.String())
		return []LogDestinationConfig{ldc}, nil
	}

	targets := strings.Split(logTarget, ",")
	LDCs := make([]LogDestinationConfig, len(targets))

	for i, target := range targets {
		target = strings.TrimPrefix(target, "file://")
		target = strings.TrimSpace(target)
		if target == "" {
			continue
		}

		ld := LogDestinationConfig{
			target:    target,
			formatter: formatter,
		}

		if fileUtils.IsValidFilePathFormat(target) {
			ld.writer, err = writers.NewToDir(logger, target, formatter)
			if err != nil {
				return nil, fmt.Errorf("could not create writer: %w", err)
			}

			ld.logger = logger.With("logDestination", ld.String())
			LDCs[i] = ld
			continue
		}

		/*
			if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
				logDestinations = append(logDestinations, md.WriteToAsyncREST)
				continue
			}
		*/
		return nil, fmt.Errorf("target unhandled by log destination conditionals: %s", target)
	}

	if len(LDCs) == 0 {
		return nil, fmt.Errorf("no valid log destinations found")
	}

	return LDCs, nil
}
