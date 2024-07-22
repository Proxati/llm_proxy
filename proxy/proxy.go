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

func newCA(certDir string) (*cert.CA, error) {
	if certDir == "" {
		slog.Debug("No cert dir specified, defaulting to ~/.mitmproxy/")
	} else {
		slog.Debug("Loading certs", "certDir", certDir)
	}

	l, err := cert.NewPathLoader(certDir)
	if err != nil {
		return nil, fmt.Errorf("unable to create or load certs from %v: %v", certDir, err)
	}

	ca, err := cert.New(l)
	if err != nil {
		return nil, fmt.Errorf("problem with CA config: %v", err)
	}

	return ca, nil
}

// newProxy returns a new proxy object with some basic configuration
func newProxy(debugLevel int, listenOn string, skipVerifyTLS bool, ca *cert.CA) (*px.Proxy, error) {
	opts := &px.Options{
		Debug:                 debugLevel,
		Addr:                  listenOn,
		InsecureSkipVerifyTLS: skipVerifyTLS,
		CA:                    ca,
		StreamLargeBodies:     1024 * 1024 * 100, // responses larger than 100MB will be streamed
	}

	p, err := px.NewProxy(opts)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// configProxy returns a configured proxy object w/ addons. This proxy still needs to be "started"
// with a blocking call to .Start() (which is handled elsewhere)
func configProxy(cfg *config.Config) (*px.Proxy, error) {
	metaAdd := newMetaAddon(cfg)

	ca, err := newCA(cfg.CertDir)
	if err != nil {
		return nil, fmt.Errorf("setupCA error: %v", err)
	}

	p, err := newProxy(cfg.IsDebugEnabled(), cfg.Listen, cfg.InsecureSkipVerifyTLS, ca)
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy: %v", err)
	}

	if !cfg.NoHttpUpgrader {
		// upgrade all http requests to https
		slog.Debug("NoHttpUpgrader is false, enabling http to https upgrade")
		metaAdd.addAddon(&addons.SchemeUpgrader{})
	}

	// struct of bools to toggle the various traffic log outputs
	logSources := config.NewLogSourceConfig(cfg)

	// create and configure MegaDirDumper addon object, but bypass traffic logs when no output is requeste
	if (cfg.OutputDir == "" && cfg.IsVerboseOrHigher()) || cfg.OutputDir != "" {
		dumperAddon, err := addons.NewMegaDumpAddon(
			cfg.OutputDir,
			cfg.LogFormat,
			logSources,
			cfg.FilterReqHeaders, cfg.FilterRespHeaders,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create traffic log dumper: %v", err)
		}

		// add the traffic log dumper to the proxy
		metaAdd.addAddon(dumperAddon)
	}

	slog.Debug("Starting proxy config for AppMode", "AppMode", cfg.AppMode)
	switch cfg.AppMode {
	case config.CacheMode:
		cacheConfig, err := config.NewCacheStorageConfig(cfg.Cache.Dir)
		if err != nil {
			return nil, fmt.Errorf("failed to create cache config: %v", err)
		}

		cacheAddon, err := addons.NewCacheAddon(
			cacheConfig.StorageEngine,
			cacheConfig.StoragePath,
			cfg.FilterReqHeaders, // filters from logging, bc we want to filter cache same as the logs
			cfg.FilterRespHeaders,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load cache addon: %v", err)
		}
		metaAdd.addAddon(cacheAddon)
	case config.APIAuditMode:
		slog.Debug("Enabling API Auditor addon")
		metaAdd.addAddon(addons.NewAPIAuditor())
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
func startProxy(p *px.Proxy, shutdown chan os.Signal) error {
	go func() {
		<-shutdown
		slog.Info("Received SIGINT, shutting down now...")

		// Then close all of the addon connections
		for _, addon := range p.Addons {
			myAddon, ok := addon.(addons.LLM_Addon)
			if !ok {
				continue
			}
			slog.Debug("Closing", "addon", myAddon)
			if err := myAddon.Close(); err != nil {
				slog.Error("Could not close", "addon", myAddon, "error", err)
			}
		}
		// Close the http client/server connections first
		slog.Debug("Closing proxy server...")

		// Manual sleep to avoid race condition on connection close
		time.Sleep(100 * time.Millisecond)

		// Create a context that will be cancelled after N seconds
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(60*time.Second))
		defer cancel()

		if err := p.Shutdown(ctx); err != nil {
			slog.Error("Unexpected error shutting down proxy server", "error", err)
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
	slog.Debug("Starting LLM_Proxy", "version", version.String())

	// setup background signal handler for clean shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	p, err := configProxy(cfg)
	if err != nil {
		return fmt.Errorf("failed to configure proxy: %v", err)
	}

	if err := startProxy(p, shutdown); err != nil {
		return fmt.Errorf("failed to start proxy: %v", err)
	}

	return nil
}
