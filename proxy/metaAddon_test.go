package proxy

import (
	"io"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/proxati/llm_proxy/v2/config"
	px "github.com/proxati/mitmproxy/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAddon implements the px.Addon interface for testing purposes.
type mockAddon struct {
	px.Addon
	clientConnectedCalled        atomic.Bool
	clientDisconnectedCalled     atomic.Bool
	serverConnectedCalled        atomic.Bool
	serverDisconnectedCalled     atomic.Bool
	tlsEstablishedServerCalled   atomic.Bool
	requestHeadersCalled         atomic.Bool
	requestCalled                atomic.Bool
	responseHeadersCalled        atomic.Bool
	responseCalled               atomic.Bool
	streamRequestModifierCalled  atomic.Bool
	streamResponseModifierCalled atomic.Bool
}

func (m *mockAddon) ClientConnected(client *px.ClientConn) {
	m.clientConnectedCalled.Store(true)
}
func (m *mockAddon) ClientDisconnected(client *px.ClientConn) {
	m.clientDisconnectedCalled.Store(true)
}
func (m *mockAddon) ServerConnected(ctx *px.ConnContext) {
	m.serverConnectedCalled.Store(true)
}
func (m *mockAddon) ServerDisconnected(ctx *px.ConnContext) {
	m.serverDisconnectedCalled.Store(true)
}
func (m *mockAddon) TlsEstablishedServer(ctx *px.ConnContext) {
	m.tlsEstablishedServerCalled.Store(true)
}
func (m *mockAddon) Requestheaders(f *px.Flow) {
	m.requestHeadersCalled.Store(true)
}
func (m *mockAddon) Request(f *px.Flow) {
	m.requestCalled.Store(true)
}
func (m *mockAddon) Responseheaders(f *px.Flow) {
	m.responseHeadersCalled.Store(true)
}
func (m *mockAddon) Response(f *px.Flow) {
	m.responseCalled.Store(true)
}
func (m *mockAddon) StreamRequestModifier(f *px.Flow, in io.Reader) io.Reader {
	m.streamRequestModifierCalled.Store(true)
	return in
}
func (m *mockAddon) StreamResponseModifier(f *px.Flow, in io.Reader) io.Reader {
	m.streamResponseModifierCalled.Store(true)
	return in
}

func TestNewMetaAddon(t *testing.T) {
	cfg := &config.Config{}
	addons := []px.Addon{&mockAddon{}, &mockAddon{}}

	meta := newMetaAddon(cfg, addons...)
	assert.Equal(t, cfg, meta.cfg)
	assert.Equal(t, len(addons), len(meta.mitmAddons))
}

func TestAddAddon(t *testing.T) {
	meta := newMetaAddon(&config.Config{})
	mock := &mockAddon{}
	meta.addAddon(mock)

	assert.Contains(t, meta.mitmAddons, mock)
}

func TestAllMethods(t *testing.T) {
	// create a proxy with a test config
	proxyPort, err := getFreePort()
	require.NoError(t, err)
	tmpDir := t.TempDir()

	// setup the metaAddon
	meta := newMetaAddon(&config.Config{})
	mock := &mockAddon{}
	meta.addAddon(mock)

	proxyShutdown, err := runProxy(proxyPort, tmpDir, config.ProxyRunMode, meta)
	require.NoError(t, err)

	// Start a basic web server on another port
	hitCounter := new(atomic.Int32)
	testServerPort, err := getFreePort()
	require.NoError(t, err)
	srv, srvShutdown := runWebServer(hitCounter, testServerPort)
	require.NotNil(t, srv)
	require.NotNil(t, srvShutdown)

	// Create an http client that will use the proxy to connect to the web server
	client, err := httpClient("http://" + proxyPort)
	require.NoError(t, err)

	t.Run("TestAddonMethodsCalled", func(t *testing.T) {
		resp, err := client.Post("http://"+testServerPort, "text/plain", strings.NewReader("hello"))
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		assert.True(t, mock.clientConnectedCalled.Load())
		assert.False(t, mock.clientDisconnectedCalled.Load())
		assert.True(t, mock.serverConnectedCalled.Load())
		assert.False(t, mock.serverDisconnectedCalled.Load())
		assert.False(t, mock.tlsEstablishedServerCalled.Load())
		assert.True(t, mock.requestHeadersCalled.Load())
		assert.True(t, mock.requestCalled.Load())
		assert.True(t, mock.responseHeadersCalled.Load())
		assert.True(t, mock.responseCalled.Load())
		assert.True(t, mock.streamRequestModifierCalled.Load())
		assert.True(t, mock.streamResponseModifierCalled.Load())
	})

	// done with tests, send shutdown signals
	t.Cleanup(func() {
		srvShutdown()
		proxyShutdown()
	})
}
