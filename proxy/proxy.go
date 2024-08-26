package proxy

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/proxati/mitmproxy/cert"
	px "github.com/proxati/mitmproxy/proxy"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/proxy/addons"
	"github.com/proxati/llm_proxy/v2/version"
)

func newCA(logger *slog.Logger, certDir string) (*cert.CA, error) {
	if certDir == "" {
		logger.Debug("No cert dir specified, defaulting to ~/.mitmproxy/")
	} else {
		logger.Debug("Loading certs", "certDir", certDir)
	}

	l, err := cert.NewPathLoader(certDir)
	if err != nil {
		return nil, fmt.Errorf("unable to create or load certs from %v: %w", certDir, err)
	}

	ca, err := cert.New(l)
	if err != nil {
		return nil, fmt.Errorf("problem with CA config: %w", err)
	}

	return ca, nil
}

// newProxy returns a new proxy object with some basic configuration
func newProxy(listenOn string, skipVerifyTLS bool, ca *cert.CA) (*px.Proxy, error) {
	opts := &px.Options{
		Addr:                  listenOn,
		InsecureSkipVerifyTLS: skipVerifyTLS,
		CA:                    ca,
		StreamLargeBodies:     1024 * 1024 * 100,                     // responses larger than 100MB will be streamed
		Logger:                slog.Default().WithGroup("mitmproxy"), // don't use the logger from slog.go in this package!
	}

	p, err := px.NewProxy(opts)
	if err != nil {
		return nil, err
	}
	return p, nil
}

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

// configProxy returns a configured proxy object w/ addons. This proxy still needs to be "started"
// with a blocking call to .Start() (which is handled elsewhere)
func configProxy(logger *slog.Logger, cfg *config.Config) (*px.Proxy, error) {
	metaAdd := newMetaAddon(logger, cfg)

	ca, err := newCA(logger, cfg.HttpBehavior.CertDir)
	if err != nil {
		return nil, fmt.Errorf("setupCA error: %w", err)
	}

	p, err := newProxy(cfg.HttpBehavior.Listen, cfg.HttpBehavior.InsecureSkipVerifyTLS, ca)
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy: %w", err)
	}

	// always validate the request and response objects
	metaAdd.addAddon(addons.NewRequestAndResponseValidator(logger))

	if cfg.IsVerboseOrHigher() {
		// add the verbose logger to the proxy
		metaAdd.addAddon(addons.NewStdOutLogger(logger))
	}

	// struct of bools to toggle the various traffic log outputs
	logSources := cfg.TrafficLogger.GetLogSourceConfig()

	// create the mega traffic dumper addon
	dumperAddon, err := configureDumper(logger, cfg, logSources)
	if err != nil {
		return nil, fmt.Errorf("failed to create traffic log dumper: %w", err)
	}
	if dumperAddon != nil {
		logger.Debug(
			"Created "+dumperAddon.String(),
			"outputDir", cfg.TrafficLogger.Output,
			"logFormat", cfg.TrafficLogger.LogFmt,
			"logSources", logSources,
			// "filterReqHeaders", cfg.FilterReqHeaders,
			// "filterRespHeaders", cfg.FilterRespHeaders,
		)

		// add the traffic log dumper to the metaAddon
		metaAdd.addAddon(dumperAddon)
	}

	// Always add the request ID to the response headers
	metaAdd.addAddon(addons.NewAddIDToHeaders())

	logger.Debug("http to https upgrade", "enabled", !cfg.HttpBehavior.NoHttpUpgrader)
	// upgrade the request _after_ it's logged
	if !cfg.HttpBehavior.NoHttpUpgrader {
		// upgrade all http requests to https
		metaAdd.addAddon(addons.NewSchemeUpgrader(logger))
	}

	logger.Debug("Building proxy config", "AppMode", cfg.AppMode.String())
	switch cfg.AppMode {
	case config.CacheMode:
		cacheConfig, err := cfg.Cache.GetCacheStorageConfig(logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create cache config: %w", err)
		}

		cacheAddon, err := addons.NewCacheAddon(
			logger,
			cacheConfig.GetStorageEngine(),
			cacheConfig.GetStoragePath(),
			cfg.HeaderFilters.RequestToLogs,
			cfg.HeaderFilters.ResponseToLogs,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load cache addon: %w", err)
		}
		logger.Debug(
			"Created "+cacheAddon.String(),
			"storageEngine", cacheConfig.GetStorageEngine(),
			"storagePath", cacheConfig.GetStoragePath(),
		)
		metaAdd.addAddon(cacheAddon)
	case config.APIAuditMode:
		metaAdd.addAddon(addons.NewAPIAuditor(logger))
		logger.Debug("APIAuditor mode enabled")
	case config.ProxyRunMode:
		// log.Debugf("No addons enabled for the basic proxy mode")
	default:
		return nil, fmt.Errorf("unknown app mode: %v", cfg.AppMode)
	}

	// add our single metaAddon abstraction to the proxy
	p.AddAddon(metaAdd)

	return p, nil
}

// startProxy receives a pointer to a proxy object, runs it, and handles the shutdown signal
func startProxy(logger *slog.Logger, p *px.Proxy, shutdown chan os.Signal) error {
	go func() {
		<-shutdown
		logger.Info("Received SIGINT, shutting down now...")

		// Then close all of the addon connections
		for _, addon := range p.Addons {
			myAddon, ok := addon.(addons.ClosableAddon)
			if !ok {
				continue
			}
			logger.Debug("Closing addon", "addonName", myAddon)
			if err := myAddon.Close(); err != nil {
				logger.Error("Could not close", "addon", myAddon, "error", err)
			}
		}
		// Close the http client/server connections first
		logger.Debug("Closing proxy server...")

		// Manual sleep to avoid race condition on connection close
		time.Sleep(100 * time.Millisecond)

		// Create a context that will be cancelled after N seconds
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(60*time.Second))
		defer cancel()

		if err := p.Shutdown(ctx); err != nil {
			logger.Error("Unexpected error shutting down proxy server", "error", err)
		}
	}()

	// block here while the proxy is running
	if err := p.Start(); err != http.ErrServerClosed {
		return fmt.Errorf("proxy server error: %v", err)
	}
	return nil
}

// Run is the main entry point for the proxy, configures the proxy and runs it
func Run(cfg *config.Config) error {
	logger := cfg.GetLogger().WithGroup("proxy")
	logger.Info("Starting LLM_Proxy", "version", version.String())

	// setup background signal handler for clean shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	p, err := configProxy(logger, cfg)
	if err != nil {
		return fmt.Errorf("failed to configure proxy: %w", err)
	}

	if err := startProxy(logger, p, shutdown); err != nil {
		return fmt.Errorf("failed to start proxy: %w", err)
	}
	logger.Info("LLM_Proxy shutdown complete")

	return nil
}
