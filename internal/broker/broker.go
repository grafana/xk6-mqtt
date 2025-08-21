// Package broker provides an embedded MQTT broker for testing purposes.
package broker

import (
	"errors"
	"log"
	"net"
	"os"
	"strings"
	"time"

	mochi "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
)

// EnvBrokerAddress is the environment variable used to set the MQTT broker address.
// It is used by tests to connect to the embedded broker.
const EnvBrokerAddress = "MQTT_BROKER_ADDRESS"

// Setup initializes the embedded MQTT broker for testing purposes.
func Setup() *mochi.Server {
	log.Print("Setting up embedded MQTT broker for tests")

	broker := mochi.New(&mochi.Options{InlineClient: true})

	if err := broker.AddHook(new(auth.AllowHook), nil); err != nil {
		log.Fatal("Failed to add auth hook:", err)
	}

	const brokerHost = "127.0.0.1"

	tcpListener := listeners.NewTCP(listeners.Config{ID: "tcp", Address: brokerHost + ":0"})
	if err := broker.AddListener(tcpListener); err != nil {
		log.Fatal("Failed to add TCP listener:", err)
	}

	go func() {
		log.Print("Starting embedded MQTT broker...")

		err := broker.Serve()
		if err != nil && !errors.Is(err, mochi.ErrConnectionClosed) {
			log.Fatal("Embedded broker server error:", err)
		}
	}()

	// Ensure the listeners are indeed listening before continuing
	var tcpPort string

	for range 10 { // Retry a few times
		if _, port, err := net.SplitHostPort(tcpListener.Address()); err == nil && port != "0" {
			tcpPort = port

			break
		}

		// Wait a bit before retrying
		time.Sleep(100 * time.Millisecond) //nolint:mnd
	}

	if tcpPort == "" {
		log.Fatal("Failed to get assigned port for embedded broker")
	}

	address := "mqtt://" + net.JoinHostPort(brokerHost, tcpPort)

	must(os.Setenv(EnvBrokerAddress, address), "Failed to set environment variable for MQTT broker address")
	log.Println("MQTT broker address set to", address)

	must(broker.Subscribe("test/#", 0, echoHandler(broker)), "Failed to subscribe to echo topic")

	return broker
}

func must(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}

func echoHandler(server *mochi.Server) func(cl *mochi.Client, sub packets.Subscription, pk packets.Packet) {
	const suffix = "/echo"

	return func(_ *mochi.Client, _ packets.Subscription, pk packets.Packet) {
		if !strings.HasSuffix(pk.TopicName, suffix) {
			_ = server.Publish(pk.TopicName+suffix, pk.Payload, false, 0)
		}
	}
}
