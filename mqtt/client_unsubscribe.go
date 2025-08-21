package mqtt

import (
	"fmt"
	"reflect"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/grafana/sobek"
	"go.k6.io/k6/js/promises"
)

type unsubscribeOptions struct {
	Tags map[string]string
}

func (c *client) unsubscribe(topic sobek.Value, opts *unsubscribeOptions) error {
	topics, err := c.unsubscribePrepare(topic)
	if err != nil {
		return err
	}

	return c.unsubscribeExecute(topics, opts)
}

func (c *client) unsubscribeAsync(topic sobek.Value, opts *unsubscribeOptions) (*sobek.Promise, error) {
	topics, err := c.unsubscribePrepare(topic)
	if err != nil {
		return nil, err
	}

	promise, resolve, reject := promises.New(c.vu)

	go func() {
		if err := c.unsubscribeExecute(topics, opts); err != nil {
			reject(err)

			return
		}

		resolve(nil)
	}()

	return promise, nil
}

func (c *client) unsubscribePrepare(topic sobek.Value) ([]string, error) {
	if !c.isConnected() {
		return nil, errNotConnected
	}

	return asUnsubscribeTopics(topic, c.vu.Runtime())
}

func (c *client) unsubscribeExecute(topics []string, opts *unsubscribeOptions) error {
	c.log.Debug("Unsubscribing from MQTT topic(s)")

	tokens := make(map[string]paho.Token)

	for _, topic := range topics {
		c.log.WithField("topic", topic).Debug("Unsubscribing from topic")

		token := c.pahoClient.Unsubscribe(topic)

		tokens[topic] = token
	}

	for t, token := range tokens {
		if token.Wait() && token.Error() != nil {
			c.addErrorMetrics(token.Error(), "unsubscribe", opts.Tags, "topic", t)

			return token.Error()
		}

		c.addCallMetrics("unsubscribe", opts.Tags, "topic", t)
	}

	return nil
}

func asUnsubscribeTopics(value sobek.Value, rt *sobek.Runtime) ([]string, error) {
	var topics []string

	switch value.ExportType() {
	case reflect.TypeFor[string]():
		var topic string
		if err := rt.ExportTo(value, &topic); err != nil {
			return nil, err
		}

		topics = append(topics, topic)

	case reflect.TypeFor[[]string]():
		var names []string

		if err := rt.ExportTo(value, &names); err != nil {
			return nil, err
		}

		topics = append(topics, names...)

	default:
		return nil, fmt.Errorf("%w: String or Array of String expected", errInvalidType)
	}

	return topics, nil
}
