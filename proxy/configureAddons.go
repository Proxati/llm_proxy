package proxy

import (
	"fmt"
	"log/slog"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/proxy/addons"
)

// configureDumper creates and configure MegaDirDumper addon object, but bypass traffic logs when
// no output target is requested (or when verbose is disabled)
func configureDumper(logger *slog.Logger, cfg *config.Config, logSources config.LogSourceConfig) (*addons.MegaTrafficDumper, error) {
	// create and configure MegaDirDumper addon object, but bypass traffic logs when no output is requested
	if cfg.TrafficLogger.Output == "" && !cfg.IsVerboseOrHigher() {
		// no output dir specified and verbose is disabled
		return nil, nil
	}

	dumperAddon, err := addons.NewMegaTrafficDumperAddon(
		logger,
		cfg.TrafficLogger.Output,
		cfg.TrafficLogger.LogFmt,
		logSources,
		cfg.HeaderFilters.RequestToLogs,
		cfg.HeaderFilters.ResponseToLogs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create traffic log dumper: %v", err)
	}

	return dumperAddon, nil
}

func configureCacheAddon(logger *slog.Logger, cfg *config.Config) (*addons.ResponseCacheAddon, error) {
	cacheConfig, err := cfg.Cache.GetCacheStorageConfig(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to load/create cache config: %w", err)
	}

	cacheAddon, err := addons.NewCacheAddon(
		logger,
		cacheConfig.GetStorageEngine(),
		cacheConfig.GetStoragePath(),
		cfg.HeaderFilters.RequestToLogs,
		cfg.HeaderFilters.ResponseToLogs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache addon: %w", err)
	}
	return cacheAddon, nil
}

func configureTrafficTransformers(logger *slog.Logger, cfg *config.Config) (*addons.TrafficTransformerAddon, error) {
	transformersAddon, err := addons.NewTrafficTransformerAddon(
		logger,
		cfg.TrafficTransformers.Request,
		cfg.TrafficTransformers.Response,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create traffic transformers addon: %w", err)
	}
	return transformersAddon, nil
}
