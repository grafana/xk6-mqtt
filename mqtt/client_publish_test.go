package mqtt

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/grafana/sobek"
	"github.com/grafana/xk6-mqtt/internal/broker"
	"github.com/stretchr/testify/require"
	"go.k6.io/k6/v2/js/modulestest"
)

const pendingPublishTimeout = 5 * time.Second

func TestClientPublish(t *testing.T) {
	t.Parallel()

	runtime := newTestRuntime(t)
	mm := newMqttMetrics(runtime.VU)
	logger := runtime.VU.InitEnv().Logger

	runtime.MoveToVUContext(newTestVUState(t))

	client := newTestClient(t, logger, runtime.VU, mm)

	toValue := runtime.VU.Runtime().ToValue

	client.on("connect", func(_ sobek.Value, _ ...sobek.Value) (sobek.Value, error) {
		require.NoError(t, client.end(nil))

		return sobek.Undefined(), nil
	})

	err := runtime.EventLoop.Start(func() error {
		require.NoError(t, client.connect(toValue(os.Getenv(broker.EnvBrokerAddress)), nil)) //nolint:forbidigo // test reads the embedded broker address from env

		return nil
	})

	require.NoError(t, err)

	runtime.EventLoop.WaitOnRegistered()
}

func TestClientEndWithPendingPublish(t *testing.T) {
	t.Parallel()

	runtime := newTestRuntime(t)
	client, published := newPendingPublishClient(t, runtime)

	result := make(chan error, 1)

	go func() {
		result <- client.publishExecute("test/pending", []byte("data"), &publishOptions{Qos: 1})
	}()

	waitFor(t, published, "the broker to receive the publish")

	ended := make(chan error, 1)

	go func() { ended <- client.end(nil) }()

	require.NoError(t, waitFor(t, ended, "end() to return while a publish is pending"))
	require.Error(t, waitFor(t, result, "the pending publish to fail after end()"))
}

func TestClientPublishPendingWhenVUContextIsDone(t *testing.T) {
	t.Parallel()

	runtime := newTestRuntime(t)
	client, published := newPendingPublishClient(t, runtime)

	result := make(chan error, 1)

	go func() {
		result <- client.publishExecute("test/pending", []byte("data"), &publishOptions{Qos: 1})
	}()

	waitFor(t, published, "the broker to receive the publish")

	runtime.CancelContext()

	err := waitFor(t, result, "the pending publish to return after the VU context is done")
	require.ErrorContains(t, err, context.Canceled.Error())

	require.NoError(t, client.end(nil))
}

// newPendingPublishClient returns a client connected to a broker that never
// acknowledges a publish, so QoS 1 publish tokens stay pending. The returned
// channel receives a value each time the broker gets a PUBLISH packet.
func newPendingPublishClient(t *testing.T, runtime *modulestest.Runtime) (*client, <-chan struct{}) {
	t.Helper()

	address, published := newSilentBroker(t)

	mm := newMqttMetrics(runtime.VU)
	logger := runtime.VU.InitEnv().Logger

	runtime.MoveToVUContext(newTestVUState(t))

	client := newTestClient(t, logger, runtime.VU, mm)

	require.NoError(t, client.connect(runtime.VU.Runtime().ToValue(address), nil))

	return client, published
}

// newSilentBroker starts a TCP server that accepts MQTT connections and
// replies to CONNECT, but never acknowledges a PUBLISH.
func newSilentBroker(t *testing.T) (string, <-chan struct{}) {
	t.Helper()

	var lc net.ListenConfig

	listener, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	t.Cleanup(func() { _ = listener.Close() })

	published := make(chan struct{}, 1)

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}

		defer func() { _ = conn.Close() }()

		for {
			packet, err := packets.ReadPacket(conn)
			if err != nil {
				return
			}

			switch packet.(type) {
			case *packets.ConnectPacket:
				if err := packets.NewControlPacket(packets.Connack).Write(conn); err != nil {
					return
				}
			case *packets.PublishPacket:
				select {
				case published <- struct{}{}:
				default:
				}
			}
		}
	}()

	return "mqtt://" + listener.Addr().String(), published
}

func waitFor[T any](t *testing.T, ch <-chan T, what string) T {
	t.Helper()

	select {
	case v := <-ch:
		return v
	case <-time.After(pendingPublishTimeout):
		require.FailNowf(t, "timeout", "timed out after %s waiting for %s", pendingPublishTimeout, what)

		var zero T

		return zero
	}
}
