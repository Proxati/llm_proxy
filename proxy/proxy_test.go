package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/internal/fileutils"
	"github.com/proxati/llm_proxy/v2/proxy/addons"
	"github.com/proxati/llm_proxy/v2/schema"
	"github.com/proxati/llm_proxy/v2/schema/utils"
	px "github.com/proxati/mitmproxy/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// ugly hack to wait for background async
	defaultSleepTime = 300 * time.Millisecond
	outputSubdir     = "output"
	certSubdir       = "certs"
	cacheSubdir      = "cache"
	debugOutput      = false
)

// randomly finds an available port to bind to
func getFreePort(t testing.TB) (string, error) {
	t.Helper()
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return "", err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "", err
	}
	defer l.Close()
	port := l.Addr().(*net.TCPAddr).Port
	return fmt.Sprintf("localhost:%d", port), nil
}

func httpClient(t testing.TB, proxyAddr string) (*http.Client, error) {
	t.Helper()
	proxyURL, err := url.Parse(proxyAddr)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	return &http.Client{
		Transport: transport,
		Timeout:   10 * http.DefaultClient.Timeout,
	}, nil
}

// respBuilder builds a response body test message from the original request body and the hit counter
func respBuilder(t testing.TB, hits int32, body io.Reader) []byte {
	t.Helper()
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		bodyBytes = []byte("")
	}
	return []byte(fmt.Sprintf("counter: %d request_body: %s", hits, string(bodyBytes)))
}

func runWebServer(t testing.TB, hitCounter *atomic.Int32, listenAddr string) (*http.Server, func()) {
	t.Helper()
	if hitCounter == nil {
		panic("hitCounter must be non-nil")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// increment the counter
		hitCounter.Add(1)

		resp := respBuilder(t, hitCounter.Load(), r.Body)
		encodedResp, encoding, err := utils.EncodeBody(resp, r.Header.Get("Accept-Encoding"))
		if err != nil {
			log.Printf("error encoding response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Encoding", encoding)
		w.WriteHeader(http.StatusOK)
		w.Write(encodedResp)
	})

	srv := &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			panic(err)
		}
	}()

	return srv, func() {
		if err := srv.Close(); err != nil {
			log.Printf("HTTP server Close: %v", err)
		}
	}
}

func runProxy(
	t testing.TB,
	proxyPort, tempDir string,
	proxyAppMode config.AppMode,
	cacheEngine config.CacheEngine,
	addons ...px.Addon,
) (shutdownFunc func(), err error) {
	t.Helper()
	// Create a simple proxy config
	cfg := config.NewDefaultConfig()
	cfg.HTTPBehavior.Listen = proxyPort
	cfg.HTTPBehavior.CertDir = filepath.Join(tempDir, certSubdir)
	cfg.TrafficLogger.Output = filepath.Join(tempDir, outputSubdir)
	cfg.Cache.Dir = filepath.Join(tempDir, cacheSubdir)
	cfg.AppMode = proxyAppMode
	cfg.HTTPBehavior.NoHTTPUpgrader = true // disable TLS because our test server doesn't support it
	cfg.Cache.Engine = cacheEngine

	if debugOutput {
		cfg.EnableOutputDebug()
	}

	// create a proxy with the test config
	p, err := configProxy(slog.Default(), cfg)
	if err != nil {
		return nil, err
	}

	// add any additional addons
	for _, addon := range addons {
		p.AddAddon(addon)
	}

	// setup external control of the proxy
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	// start the proxy in the background
	go func() {
		err = startProxy(slog.Default(), p, shutdownChan)
		if err != nil {
			log.Fatal(err)
		}
	}()
	// sleep while waiting for the proxy to start
	time.Sleep(defaultSleepTime)

	return func() {
		// sleep while waiting for the proxy to settle before shutdown
		time.Sleep(defaultSleepTime)

		// returns a function that can be called to shutdown the proxy goroutine
		shutdownChan <- os.Interrupt
	}, nil
}

func TestProxySimple(t *testing.T) {
	// create a proxy with a test config
	proxyPort, err := getFreePort(t)
	require.NoError(t, err)
	tmpDir := t.TempDir()
	proxyShutdown, err := runProxy(t, proxyPort, tmpDir, config.ProxyRunMode, 0)
	require.NoError(t, err)

	// Start a basic web server on another port
	hitCounter := new(atomic.Int32)
	testServerPort, err := getFreePort(t)
	require.NoError(t, err)
	srv, srvShutdown := runWebServer(t, hitCounter, testServerPort)
	require.NotNil(t, srv)
	require.NotNil(t, srvShutdown)

	// Create an http client that will use the proxy to connect to the web server
	client, err := httpClient(t, "http://"+proxyPort)
	require.NoError(t, err)

	t.Run("TestSimpleProxy", func(t *testing.T) {
		hitCounter.Store(0) // reset the counter
		// make a request using that client, through the proxy
		resp, err := client.Post("http://"+testServerPort, "text/plain", strings.NewReader("hello"))
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		expectedResponse := respBuilder(t, 1, strings.NewReader("hello"))

		// check the response body from req1
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, expectedResponse, body)
		assert.Equal(t, int32(1), hitCounter.Load())
	})

	t.Run("TestSimpleProxy2", func(t *testing.T) {
		hitCounter.Store(5) // reset the counter
		// make another request using that client, through the proxy
		resp, err := client.Post("http://"+testServerPort, "text/plain", strings.NewReader("hello"))
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		expectedResponse := respBuilder(t, 6, strings.NewReader("hello"))

		// check the response body from req2
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, expectedResponse, body)
		assert.Equal(t, int32(6), hitCounter.Load())
	})

	// done with tests, send shutdown signals
	t.Cleanup(func() {
		hitCounter.Store(0)
		srvShutdown()
		proxyShutdown()
	})
}

func TestProxyDirLoggerMode(t *testing.T) {
	logger := slog.Default()

	// create a proxy with a test config
	proxyPort, err := getFreePort(t)
	require.NoError(t, err)
	tmpDir := t.TempDir()
	proxyShutdown, err := runProxy(t, proxyPort, tmpDir, config.ProxyRunMode, 0)
	require.NoError(t, err)

	// Start a basic web server on another port
	hitCounter := new(atomic.Int32)
	testServerPort, err := getFreePort(t)
	require.NoError(t, err)
	srv, srvShutdown := runWebServer(t, hitCounter, testServerPort)
	require.NotNil(t, srv)
	require.NotNil(t, srvShutdown)

	// Create an http client that will use the proxy to connect to the web server
	client, err := httpClient(t, "http://"+proxyPort)
	require.NoError(t, err)

	t.Run("TestDirLoggerNormal", func(t *testing.T) {
		hitCounter.Store(0) // reset the counter

		// setup a new watcher
		watch, err := fileutils.NewFileWatcher(logger, tmpDir+"/"+outputSubdir)
		require.NoError(t, err)

		// make a request using that client, through the proxy
		resp, err := client.Post("http://"+testServerPort, "text/plain", strings.NewReader(t.Name()))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		require.Equal(t, int32(1), hitCounter.Load())

		expectedResponse := respBuilder(t, 1, strings.NewReader(t.Name()))

		// check the response body from req1
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, expectedResponse, body)

		// wait for the proxy to write the log file
		err = fileutils.WaitForFile(logger, watch, defaultSleepTime)
		require.NoError(t, err)

		// check that the log file was created
		logFiles, err := filepath.Glob(filepath.Join(tmpDir, outputSubdir, "*"))
		require.NoError(t, err)
		require.Equal(t, 1, len(logFiles))

		expectedResponse = respBuilder(t, 1, strings.NewReader(t.Name()))

		// read the log file, and check that it contains the expected content
		logFile, err := os.ReadFile(logFiles[0])
		require.NoError(t, err)
		assert.Contains(t, string(logFile), string(expectedResponse))

		// delete that log file, and try again
		err = os.Remove(logFiles[0])
		require.NoError(t, err)

		// setup a new watcher
		watch, err = fileutils.NewFileWatcher(logger, tmpDir+"/"+outputSubdir)
		require.NoError(t, err)

		// make another request using that client, through the proxy
		resp, err = client.Post("http://"+testServerPort, "text/plain", strings.NewReader("hello2"))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		require.Equal(t, int32(2), hitCounter.Load())

		// sleep to allow the proxy to write the log file
		err = fileutils.WaitForFile(logger, watch, defaultSleepTime)
		require.NoError(t, err)

		// check the log
		logFiles, err = filepath.Glob(filepath.Join(tmpDir, outputSubdir, "*"))
		require.NoError(t, err)
		require.Equal(t, 1, len(logFiles))

		expectedResponse = respBuilder(t, 2, strings.NewReader("hello2"))

		// read the log file, and check that it contains the expected content
		logFile, err = os.ReadFile(logFiles[0])
		defer os.Remove(logFiles[0])
		require.NoError(t, err)
		assert.Contains(t, string(logFile), string(expectedResponse))
	})

	t.Run("TestDirLoggerJSON", func(t *testing.T) {
		hitCounter.Store(0) // reset the counter

		watch, err := fileutils.NewFileWatcher(logger, tmpDir+"/"+outputSubdir)
		require.NoError(t, err)

		// make another request using that client, through the proxy
		resp, err := client.Post("http://"+testServerPort, "text/plain", strings.NewReader("hello"))
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		require.Equal(t, int32(1), hitCounter.Load())

		// sleep to allow the proxy to write the log file
		err = fileutils.WaitForFile(logger, watch, defaultSleepTime)
		require.NoError(t, err)

		// check the log
		logFiles, err := filepath.Glob(filepath.Join(tmpDir, outputSubdir, "*"))
		require.NoError(t, err)
		require.Equal(t, 1, len(logFiles))

		// read the log file, and check that it contains the expected content
		logFile, err := os.ReadFile(logFiles[0])
		require.NoError(t, err)

		// marshal the log file to a logDumpContainer
		lDump := schema.LogDumpContainer{}
		err = json.Unmarshal(logFile, &lDump)
		require.NoError(t, err)
		fmt.Println(string(logFile))

		// check the logDumpContainer
		assert.Equal(t, schema.DefaultSchemaVersion, lDump.SchemaVersion)
		assert.NotNil(t, lDump.Timestamp)
		assert.NotNil(t, lDump.ConnectionStats)

		require.NotNil(t, lDump.Request)
		assert.Equal(t, "POST", lDump.Request.Method)

		expectedResponse := respBuilder(t, 1, strings.NewReader("hello"))

		require.NotNil(t, lDump.Response)
		assert.Equal(t, http.StatusOK, lDump.Response.Status)
		assert.Equal(t, string(expectedResponse), lDump.Response.Body)

	})

	// done with tests, send shutdown signals
	t.Cleanup(func() {
		hitCounter.Store(0)
		srvShutdown()
		proxyShutdown()
	})
}

func TestProxyCache(t *testing.T) {
	for _, engine := range []config.CacheEngine{config.CacheEngineMemory, config.CacheEngineBolt} {
		t.Run(engine.String(), func(t *testing.T) {
			logger := slog.Default().With("cacheEngine", engine.String())

			// create a proxy with a test config
			proxyPort, err := getFreePort(t)
			require.NoError(t, err)
			tmpDir := t.TempDir()
			proxyShutdown, err := runProxy(t, proxyPort, tmpDir, config.CacheMode, engine)
			require.NoError(t, err)

			// Start a basic web server on another port
			hitCounter := new(atomic.Int32)
			testServerPort, err := getFreePort(t)
			require.NoError(t, err)
			srv, srvShutdown := runWebServer(t, hitCounter, testServerPort)
			require.NotNil(t, srv)
			require.NotNil(t, srvShutdown)

			// Create a client that will use the proxy
			client, err := httpClient(t, "http://"+proxyPort)
			require.NoError(t, err)

			t.Run("Miss", func(t *testing.T) {
				hitCounter.Store(0) // reset the counter
				// make a request using the client, through the proxy
				resp, err := client.Post("http://"+testServerPort, "text/plain", strings.NewReader(t.Name()))
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)

				// check the response body from this request
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				expectedResponse := respBuilder(t, 1, strings.NewReader(t.Name()))

				assert.Equal(t, expectedResponse, body)
				assert.Equal(t, int32(1), hitCounter.Load())
				assert.Equal(t, addons.CacheStatusMiss, resp.Header.Get(addons.CacheStatusHeader))
			})

			t.Run("Hit", func(t *testing.T) {
				hitCounter.Store(0) // reset the counter

				// make a request using the client, through the proxy
				resp, err := client.Post("http://"+testServerPort, "text/plain", strings.NewReader(t.Name()))
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)

				// check the response body from this request, should be a miss
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				expectedResponse := respBuilder(t, 1, strings.NewReader(t.Name()))

				assert.Equal(t, expectedResponse, body)
				assert.Equal(t, int32(1), hitCounter.Load())
				assert.Equal(t, addons.CacheStatusMiss, resp.Header.Get(addons.CacheStatusHeader))

				// wait for the cache to be written, waiting for a file event is not reliable here
				time.Sleep(defaultSleepTime)

				// now, this should be a cache hit...
				// make another request using the client, through the proxy
				resp, err = client.Post("http://"+testServerPort, "text/plain", strings.NewReader(t.Name()))
				require.NoError(t, err)
				assert.Equal(t, 200, resp.StatusCode)

				// check the response body from this request
				// (should be the cached response with value=1, not the incremented value)
				body, err = io.ReadAll(resp.Body)
				require.NoError(t, err)

				expectedResponse = respBuilder(t, 1, strings.NewReader(t.Name()))

				assert.Equal(t, expectedResponse, body)
				assert.Equal(t, int32(1), hitCounter.Load()) // the counter should not be 6, because we got a cache hit
				assert.Equal(t, addons.CacheStatusHit, resp.Header.Get(addons.CacheStatusHeader))
			})

			t.Run("HitNoGzip", func(t *testing.T) {
				// Make a request with gzip, then make a second request without gzip
				hitCounter.Store(0) // reset the counter to align with results from the previous test
				require.Equal(t, int32(0), hitCounter.Load())

				// create the test request
				req1, err := http.NewRequest("POST", "http://"+testServerPort, strings.NewReader(t.Name()))
				require.NoError(t, err)

				// manually set the headers to simulate a request asking for a gzip response from upstream
				req1.Header.Set("Content-Type", "text/plain")
				req1.Header.Set("Accept-Encoding", "gzip")

				resp1, err := client.Do(req1)
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp1.StatusCode)
				assert.Equal(t, "gzip", resp1.Header.Get("Content-Encoding"))

				// check the response body from this request
				body1, err := io.ReadAll(resp1.Body)
				require.NoError(t, err)

				// decode the response, force gzip decoding because that's what we asked for!
				decodedBody, err := utils.DecodeBody(body1, "gzip")
				require.NoError(t, err)

				// check the response and counter state
				expectedResponseBody := respBuilder(t, 1, strings.NewReader(t.Name()))
				assert.Equal(t, expectedResponseBody, decodedBody)
				assert.Equal(t, int32(1), hitCounter.Load())
				assert.Equal(t, addons.CacheStatusMiss, resp1.Header.Get(addons.CacheStatusHeader))

				// wait for the cache to be written, waiting for a file event is not reliable here
				time.Sleep(defaultSleepTime)

				// send another request without gzip, check that it's a cache hit (no gzip)
				client.Transport.(*http.Transport).DisableCompression = false
				resp2, err := client.Post("http://"+testServerPort, "text/plain", strings.NewReader(t.Name()))
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp2.StatusCode)
				assert.NotEqual(t, "gzip", resp2.Header.Get("Content-Encoding"))
				assert.Equal(t, "", resp2.Header.Get("Content-Encoding"))

				// check the response body from this request
				body2, err := io.ReadAll(resp2.Body)
				require.NoError(t, err)
				expectedResponseBody = respBuilder(t, 1, strings.NewReader(t.Name()))

				// no decoding for this body check, because it should be plain-text (no gzip)
				assert.Equal(t, expectedResponseBody, body2)
				assert.Equal(t, int32(1), hitCounter.Load(), "test server hit counter hasn't incremented, bc cache hit")
				assert.Equal(t, addons.CacheStatusHit, resp2.Header.Get(addons.CacheStatusHeader))

				require.Equal(t, decodedBody, body2, "body1 and body2 are equal, because 1 is live and the 2 is from cache")
			})

			// done with this test run, send shutdown signals
			t.Cleanup(func() {
				logger.Info("Cleaning up")
				srvShutdown()
				proxyShutdown()
			})
		})
	}
}

func TestCacheControlHeaders(t *testing.T) {
	for _, engine := range []config.CacheEngine{config.CacheEngineMemory, config.CacheEngineBolt} {
		for _, testHeaderValue := range []string{"no-store", "no-cache"} {
			testName := fmt.Sprintf("%s_%s", engine, testHeaderValue)
			logger := slog.Default().With("testName", testName)

			t.Run(testName, func(t *testing.T) {
				logger.Info("Starting test")

				// create a proxy with a test config
				proxyPort, err := getFreePort(t)
				require.NoError(t, err)
				tmpDir := t.TempDir()
				proxyShutdown, err := runProxy(t, proxyPort, tmpDir, config.CacheMode, engine)
				require.NoError(t, err)

				// Start a basic web server on another port
				hitCounter := new(atomic.Int32)
				testServerPort, err := getFreePort(t)
				require.NoError(t, err)
				srv, srvShutdown := runWebServer(t, hitCounter, testServerPort)
				require.NotNil(t, srv)
				require.NotNil(t, srvShutdown)

				// Create a client that will use the proxy
				client, err := httpClient(t, "http://"+proxyPort)
				require.NoError(t, err)

				// start tests
				hitCounter.Store(0) // reset the counter

				logger.Debug("Starting First Request")
				// First request with "Cache-Control: no-store" header
				req, err := http.NewRequest("POST", "http://"+testServerPort, strings.NewReader("hello"))
				require.NoError(t, err)
				req.Header.Set("Cache-Control", testHeaderValue)

				// Make the first request using the client
				resp, err := client.Do(req)
				require.NoError(t, err)
				require.Equal(t, 200, resp.StatusCode)

				expectedResponse := respBuilder(t, 1, strings.NewReader("hello"))

				// check the response body
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				require.NoError(t, resp.Body.Close())

				assert.Equal(t, expectedResponse, body)
				require.Equal(t, int32(1), hitCounter.Load(), "first request, fresh response expected")

				logger.Debug("Starting Second Request")
				time.Sleep(defaultSleepTime)
				// Create the second request without "Cache-Control" header
				req, err = http.NewRequest("POST", "http://"+testServerPort, strings.NewReader("hello"))
				require.NoError(t, err)

				// Make the second request using the client
				resp, err = client.Do(req)
				require.NoError(t, err)
				require.Equal(t, 200, resp.StatusCode)

				expectedResponse = respBuilder(t, 2, strings.NewReader("hello"))

				// check the second response body
				body, err = io.ReadAll(resp.Body)
				require.NoError(t, err)
				require.NoError(t, resp.Body.Close())

				assert.Equal(t, expectedResponse, body)

				// The counter should be incremented to 2, indicating the first response was not cached
				require.Equal(t, int32(2), hitCounter.Load(), "second request, cache miss")

				logger.Debug("Starting Third Request")
				time.Sleep(defaultSleepTime)
				// Third request should be a cache hit, because the second response was cached
				req, err = http.NewRequest("POST", "http://"+testServerPort, strings.NewReader("hello"))
				require.NoError(t, err)

				// Make the third request using the client
				resp, err = client.Do(req)
				require.NoError(t, err)
				require.Equal(t, 200, resp.StatusCode)

				expectedResponse = respBuilder(t, 2, strings.NewReader("hello"))

				// check the third response body
				body, err = io.ReadAll(resp.Body)
				require.NoError(t, err)
				require.NoError(t, resp.Body.Close())

				assert.Equal(t, expectedResponse, body)

				// The counter should not have incremented, because the response was cached
				require.Equal(t, int32(2), hitCounter.Load(), "third request, cache hit")

				logger.Debug("Starting Fourth Request")
				time.Sleep(defaultSleepTime)
				// Fourth request with "Cache-Control: no-store" header
				req, err = http.NewRequest("POST", "http://"+testServerPort, strings.NewReader("hello"))
				require.NoError(t, err)
				req.Header.Set("Cache-Control", testHeaderValue)

				// Make the fourth request using the client
				resp, err = client.Do(req)
				require.NoError(t, err)
				require.Equal(t, 200, resp.StatusCode)

				expectedResponse = respBuilder(t, 3, strings.NewReader("hello"))

				// check the fourth response body
				body, err = io.ReadAll(resp.Body)
				require.NoError(t, err)
				require.NoError(t, resp.Body.Close())

				assert.Equal(t, expectedResponse, body)

				// The counter should be incremented to 3, indicating the fourth response was not from the cache
				require.Equal(t, int32(3), hitCounter.Load(), "fourth request,fresh response expected")

				t.Cleanup(func() {
					logger.Info("Cleaning up")
					srvShutdown()
					proxyShutdown()
				})
			})

		}
	}
}

// Testing imperative code is tough
func TestNewProxy(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	ca, err := newCA(slog.Default(), tempDir)
	assert.NoError(t, err)

	p, err := newProxy("localhost:8080", false, ca)
	assert.NoError(t, err)
	assert.NotNil(t, p)
}

func TestNewCA(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	ca, err := newCA(slog.Default(), tempDir)
	assert.NoError(t, err)
	assert.NotNil(t, ca)
}

func TestConfigProxy(t *testing.T) {
	t.Parallel()

	t.Run("TestConfigProxy quiet mode", func(t *testing.T) {
		// Create a mock configuration
		cfg := config.NewDefaultConfig()
		cfg.HTTPBehavior.CertDir = t.TempDir()
		cfg.AppMode = config.ProxyRunMode

		// Call the function with the mock configuration
		p, err := configProxy(slog.Default(), cfg)

		// Assert that no error was returned
		assert.NoError(t, err)

		// Assert that a proxy was returned
		assert.NotNil(t, p)

		assert.Equal(t, 1, len(p.Addons))

		// Assert that the first, and only addon is the MetaAddon
		_, ok := p.Addons[0].(*metaAddon)
		assert.True(t, ok)

		// Assert that the MetaAddon has the correct config
		metaAddon := p.Addons[0].(*metaAddon)
		assert.Equal(t, cfg, metaAddon.cfg)

		// Assert that the MetaAddon has two addons (the traffic logger, and the ID header addon)
		assert.Equal(t, 3, len(metaAddon.mitmAddons))
	})

	t.Run("TestConfigProxy verbose mode", func(t *testing.T) {
		// Create a mock configuration
		cfg := config.NewDefaultConfig()
		cfg.HTTPBehavior.CertDir = t.TempDir()
		cfg.AppMode = config.ProxyRunMode
		cfg.EnableOutputVerbose()

		// Call the function with the mock configuration
		p, err := configProxy(slog.Default(), cfg)

		// Assert that no error was returned
		assert.NoError(t, err)

		// Assert that a proxy was returned
		assert.NotNil(t, p)

		assert.Equal(t, 1, len(p.Addons))

		// Assert that the first, and only addon is the MetaAddon
		_, ok := p.Addons[0].(*metaAddon)
		assert.True(t, ok)

		// Assert that the MetaAddon has the correct config
		metaAddon := p.Addons[0].(*metaAddon)
		assert.Equal(t, cfg, metaAddon.cfg)

		// Assert that the MetaAddon has two addons (the traffic logger, stdout logger, the ID header addon, and the base addon)
		assert.Equal(t, 5, len(metaAddon.mitmAddons))
	})

	t.Run("TestConfigProxy output mode", func(t *testing.T) {
		// Create a mock configuration
		cfg := config.NewDefaultConfig()
		cfg.HTTPBehavior.CertDir = t.TempDir()
		cfg.AppMode = config.ProxyRunMode
		cfg.TrafficLogger.Output = t.TempDir() // Set the output directory, which should enable the logger addon

		// Call the function with the mock configuration
		p, err := configProxy(slog.Default(), cfg)

		// Assert that no error was returned
		assert.NoError(t, err)

		// Assert that a proxy was returned
		assert.NotNil(t, p)

		assert.Equal(t, 1, len(p.Addons))

		// Assert that the first, and only addon is the MetaAddon
		_, ok := p.Addons[0].(*metaAddon)
		assert.True(t, ok)

		// Assert that the MetaAddon has the correct config
		metaAddon := p.Addons[0].(*metaAddon)
		assert.Equal(t, cfg, metaAddon.cfg)

		// Assert that the MetaAddon has two addons, the logger, the ID header addon, and the base addon
		assert.Equal(t, 4, len(metaAddon.mitmAddons))
	})
}
