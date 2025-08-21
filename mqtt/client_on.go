package mqtt

import (
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/grafana/sobek"
	"github.com/sirupsen/logrus"
	"go.k6.io/k6/metrics"
)

var events = map[string]struct{}{ //nolint:gochecknoglobals
	"connect":   {},
	"reconnect": {},
	"end":       {},
	"error":     {},
	"message":   {},
}

func (c *client) on(event string, handler sobek.Callable) {
	if _, ok := events[event]; !ok {
		c.log.WithField("event", event).Warn("Unknown event type")

		return
	}

	if _, ok := c.handlers[event]; ok {
		c.log.WithField("event", event).Warn("Event handler already registered, overriding")
	}

	c.log.WithField("event", event).Debug("Event handler registered")

	c.handlers[event] = handler
}

func (c *client) fire(event string, args ...sobek.Value) {
	fn, ok := c.handlers[event]
	if !ok {
		return
	}

	c.log.WithField("event", event).Debug("Queuing event handler")

	c.callChan <- func() error {
		c.log.WithField("event", event).Debug("Firing event handler")

		_, err := fn(sobek.Undefined(), args...)

		return err
	}
}

func (c *client) messageHandler(_ paho.Client, msg paho.Message) {
	c.log.WithFields(logrus.Fields{
		"topic":     msg.Topic(),
		"messageID": msg.MessageID(),
	}).Debug("Received MQTT message")

	rt := c.vu.Runtime()

	payload := rt.NewArrayBuffer(msg.Payload())

	now := time.Now()
	bytes := float64(len(msg.Payload()))
	tags := c.tags().With("topic", msg.Topic())

	samples := metrics.Samples{
		metrics.Sample{
			TimeSeries: metrics.TimeSeries{
				Metric: c.metrics.mqttCalls,
				Tags:   c.tagsForMethod("message", nil, "topic", msg.Topic()),
			},
			Time:  now,
			Value: float64(1),
		},
		metrics.Sample{
			TimeSeries: metrics.TimeSeries{
				Metric: c.metrics.mqttMessageReceived,
				Tags:   tags,
			},
			Time:  now,
			Value: float64(1),
		},
		metrics.Sample{
			TimeSeries: metrics.TimeSeries{
				Metric: c.metrics.mqttDataReceived,
				Tags:   tags,
			},
			Time:  now,
			Value: bytes,
		},
		metrics.Sample{
			TimeSeries: metrics.TimeSeries{
				Metric: c.metrics.dataReceived,
				Tags:   c.currentTags(),
			},
			Time:  now,
			Value: bytes,
		},
	}

	metrics.PushIfNotDone(c.vu.Context(), c.vu.State().Samples, samples)

	c.fire("message", rt.ToValue(msg.Topic()), rt.ToValue(payload))
}

func (c *client) connectHandler(_ paho.Client) {
	c.log.Debug("Connected to MQTT broker")

	c.fire("connect")
}

func (c *client) reconnectHandler(_ paho.Client, _ *paho.ClientOptions) {
	c.log.Debug("Reconnecting to MQTT broker")

	c.fire("reconnect")
}
