package mqtt

import (
	"os"
	"testing"

	"github.com/grafana/sobek"
	"github.com/grafana/xk6-mqtt/internal/broker"
	"github.com/stretchr/testify/require"
)

func TestClientConnect(t *testing.T) {
	t.Parallel()

	runtime := newTestRuntime(t)
	mm := newMqttMetrics(runtime.VU)
	logger := runtime.VU.InitEnv().Logger

	runtime.MoveToVUContext(newTestVUState(t)) // runtime.VU.InitEnv() will return nil after this

	client := newTestClient(t, logger, runtime.VU, mm)

	toValue := runtime.VU.Runtime().ToValue

	handlerCalled := false

	client.on("connect", func(_ sobek.Value, _ ...sobek.Value) (sobek.Value, error) {
		require.NoError(t, client.end(nil))

		handlerCalled = true

		return sobek.Undefined(), nil
	})

	err := runtime.EventLoop.Start(func() error {
		require.NoError(t, client.connect(toValue(os.Getenv(broker.EnvBrokerAddress)), nil))

		return nil
	})

	require.NoError(t, err)

	runtime.EventLoop.WaitOnRegistered()

	require.True(t, handlerCalled)
}

func TestClientConnectAuthenticated(t *testing.T) {
	t.Parallel()

	server := broker.New(false)

	t.Cleanup(func() {
		require.NoError(t, server.Close())
	})

	tcpListener, ok := server.Listeners.Get("tcp")

	require.True(t, ok)

	addr := "tcp://" + tcpListener.Address()

	runtime := newTestRuntime(t)
	mm := newMqttMetrics(runtime.VU)
	logger := runtime.VU.InitEnv().Logger

	runtime.MoveToVUContext(newTestVUState(t)) // runtime.VU.InitEnv() will return nil after this

	toValue := runtime.VU.Runtime().ToValue

	client := newTestClient(t, logger, runtime.VU, mm)
	client.clientOpts.Username = toValue("test-user")
	client.clientOpts.Password = toValue("test-password")

	handlerCalled := false

	client.on("connect", func(_ sobek.Value, _ ...sobek.Value) (sobek.Value, error) {
		require.NoError(t, client.end(nil))

		handlerCalled = true

		return sobek.Undefined(), nil
	})

	err := runtime.EventLoop.Start(func() error {
		require.NoError(t, client.connect(toValue(addr), nil))

		return nil
	})

	require.NoError(t, err)

	runtime.EventLoop.WaitOnRegistered()

	require.True(t, handlerCalled)
}
