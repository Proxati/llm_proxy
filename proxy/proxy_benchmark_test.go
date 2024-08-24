package proxy

import (
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/stretchr/testify/require"
)

func BenchmarkProxySimple(b *testing.B) {
	// create a proxy with a test config
	proxyPort, err := getFreePort(b)
	require.NoError(b, err)
	tmpDir := b.TempDir()
	proxyShutdown, err := runProxy(b, proxyPort, tmpDir, config.ProxyRunMode, 0)
	require.NoError(b, err)

	// Start a basic web server on another port
	hitCounter := new(atomic.Int32)
	testServerPort, err := getFreePort(b)
	require.NoError(b, err)
	srv, srvShutdown := runWebServer(b, hitCounter, testServerPort)
	require.NotNil(b, srv)
	require.NotNil(b, srvShutdown)

	// Create an http client that will use the proxy to connect to the web server
	client, err := httpClient(b, "http://"+proxyPort)
	require.NoError(b, err)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hitCounter.Store(0) // reset the counter
		// make a request using that client, through the proxy
		b.StartTimer()
		resp, err := client.Post("http://"+testServerPort, "text/plain", strings.NewReader("hello"))
		b.StopTimer()
		require.NoError(b, err)
		require.Equal(b, 200, resp.StatusCode)
	}
	b.Cleanup(func() {
		srvShutdown()
		proxyShutdown()
	})
}

func BenchmarkProxyCacheMemory(b *testing.B) {
	// create a proxy with a test config
	proxyPort, err := getFreePort(b)
	require.NoError(b, err)
	tmpDir := b.TempDir()
	proxyShutdown, err := runProxy(b, proxyPort, tmpDir, config.CacheMode, config.CacheEngineMemory)
	require.NoError(b, err)

	// Start a basic web server on another port
	hitCounter := new(atomic.Int32)
	testServerPort, err := getFreePort(b)
	require.NoError(b, err)
	srv, srvShutdown := runWebServer(b, hitCounter, testServerPort)
	require.NotNil(b, srv)
	require.NotNil(b, srvShutdown)

	// Create an http client that will use the proxy to connect to the web server
	client, err := httpClient(b, "http://"+proxyPort)
	require.NoError(b, err)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hitCounter.Store(0) // reset the counter
		// make a request using that client, through the proxy
		b.StartTimer()
		resp, err := client.Post("http://"+testServerPort, "text/plain", strings.NewReader("hello"))
		b.StopTimer()
		require.NoError(b, err)
		require.Equal(b, 200, resp.StatusCode)
	}
	b.Cleanup(func() {
		srvShutdown()
		proxyShutdown()
	})
}

func BenchmarkProxyCacheBolt(b *testing.B) {
	// create a proxy with a test config
	proxyPort, err := getFreePort(b)
	require.NoError(b, err)
	tmpDir := b.TempDir()
	proxyShutdown, err := runProxy(b, proxyPort, tmpDir, config.CacheMode, config.CacheEngineBolt)
	require.NoError(b, err)

	// Start a basic web server on another port
	hitCounter := new(atomic.Int32)
	testServerPort, err := getFreePort(b)
	require.NoError(b, err)
	srv, srvShutdown := runWebServer(b, hitCounter, testServerPort)
	require.NotNil(b, srv)
	require.NotNil(b, srvShutdown)

	// Create an http client that will use the proxy to connect to the web server
	client, err := httpClient(b, "http://"+proxyPort)
	require.NoError(b, err)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hitCounter.Store(0) // reset the counter
		// make a request using that client, through the proxy
		b.StartTimer()
		resp, err := client.Post("http://"+testServerPort, "text/plain", strings.NewReader("hello"))
		b.StopTimer()
		require.NoError(b, err)
		require.Equal(b, 200, resp.StatusCode)
	}
	b.Cleanup(func() {
		srvShutdown()
		proxyShutdown()
	})
}

func BenchmarkParallelProxyNoCache(b *testing.B) {
	// create a proxy with a test config
	proxyPort, err := getFreePort(b)
	require.NoError(b, err)
	tmpDir := b.TempDir()
	proxyShutdown, err := runProxy(b, proxyPort, tmpDir, config.ProxyRunMode, 0)
	require.NoError(b, err)

	// Start a basic web server on another port
	hitCounter := new(atomic.Int32)
	testServerPort, err := getFreePort(b)
	require.NoError(b, err)
	srv, srvShutdown := runWebServer(b, hitCounter, testServerPort)
	require.NotNil(b, srv)
	require.NotNil(b, srvShutdown)

	// Create an http client that will use the proxy to connect to the web server
	client, err := httpClient(b, "http://"+proxyPort)
	require.NoError(b, err)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// make a request using that client, through the proxy
			resp, err := client.Post("http://"+testServerPort, "text/plain", strings.NewReader("hello"))
			require.NoError(b, err)
			require.Equal(b, 200, resp.StatusCode)
		}
	})
	b.StopTimer()

	b.Cleanup(func() {
		hitCounter.Store(0)
		srvShutdown()
		proxyShutdown()
	})
}

func BenchmarkParallelProxyCacheMemory(b *testing.B) {
	// create a proxy with a test config
	proxyPort, err := getFreePort(b)
	require.NoError(b, err)
	tmpDir := b.TempDir()
	proxyShutdown, err := runProxy(b, proxyPort, tmpDir, config.CacheMode, config.CacheEngineMemory)
	require.NoError(b, err)

	// Start a basic web server on another port
	hitCounter := new(atomic.Int32)
	testServerPort, err := getFreePort(b)
	require.NoError(b, err)
	srv, srvShutdown := runWebServer(b, hitCounter, testServerPort)
	require.NotNil(b, srv)
	require.NotNil(b, srvShutdown)

	// Create an http client that will use the proxy to connect to the web server
	client, err := httpClient(b, "http://"+proxyPort)
	require.NoError(b, err)

	// make one request to prime the cache
	resp, err := client.Post("http://"+testServerPort, "text/plain", strings.NewReader("hello"))
	require.NoError(b, err)
	require.Equal(b, 200, resp.StatusCode)
	time.Sleep(defaultSleepTime) // wait for cache to write

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// make a request using that client, through the proxy
			resp, err := client.Post("http://"+testServerPort, "text/plain", strings.NewReader("hello"))
			require.NoError(b, err)
			require.Equal(b, 200, resp.StatusCode)
		}
	})
	b.StopTimer()

	b.Cleanup(func() {
		hitCounter.Store(0)
		srvShutdown()
		proxyShutdown()
	})
}

func BenchmarkParallelProxyCacheBolt(b *testing.B) {
	// create a proxy with a test config
	proxyPort, err := getFreePort(b)
	require.NoError(b, err)
	tmpDir := b.TempDir()
	proxyShutdown, err := runProxy(b, proxyPort, tmpDir, config.CacheMode, config.CacheEngineBolt)
	require.NoError(b, err)

	// Start a basic web server on another port
	hitCounter := new(atomic.Int32)
	testServerPort, err := getFreePort(b)
	require.NoError(b, err)
	srv, srvShutdown := runWebServer(b, hitCounter, testServerPort)
	require.NotNil(b, srv)
	require.NotNil(b, srvShutdown)

	// Create an http client that will use the proxy to connect to the web server
	client, err := httpClient(b, "http://"+proxyPort)
	require.NoError(b, err)

	// make one request to prime the cache
	resp, err := client.Post("http://"+testServerPort, "text/plain", strings.NewReader("hello"))
	require.NoError(b, err)
	require.Equal(b, 200, resp.StatusCode)
	time.Sleep(defaultSleepTime) // wait for cache to write

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// make a request using that client, through the proxy
			resp, err := client.Post("http://"+testServerPort, "text/plain", strings.NewReader("hello"))
			require.NoError(b, err)
			require.Equal(b, 200, resp.StatusCode)
		}
	})
	b.StopTimer()

	b.Cleanup(func() {
		hitCounter.Store(0)
		srvShutdown()
		proxyShutdown()
	})
}
