package mqtt

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/grafana/sobek"
	"go.k6.io/k6/js/promises"
)

var errNotConnected = errors.New("not connected")

type connectOptions struct {
	ClientId       sobek.Value //nolint:revive
	Keepalive      sobek.Value
	ConnectTimeout sobek.Value
	Servers        []string
	Tags           map[string]string
}

func (c *client) connect(urlOrOpts sobek.Value, optsOrEmpty sobek.Value) error {
	err := c.connectPrepare(urlOrOpts, optsOrEmpty)
	if err != nil {
		return err
	}

	return c.connectExecute()
}

func (c *client) connectAsync(urlOrOpts sobek.Value, optsOrEmpty sobek.Value) (*sobek.Promise, error) {
	err := c.connectPrepare(urlOrOpts, optsOrEmpty)
	if err != nil {
		return nil, err
	}

	promise, resolve, reject := promises.New(c.vu)

	go func() {
		if err := c.connectExecute(); err != nil {
			reject(err)

			return
		}

		resolve(nil)
	}()

	return promise, nil
}

func (c *client) reconnect() error {
	c.disconnect()

	c.log.Debug("Reconnecting to MQTT broker")

	return c.connectExecute()
}

func (c *client) reconnectAsync() (*sobek.Promise, error) {
	promise, resolve, reject := promises.New(c.vu)

	go func() {
		if err := c.reconnect(); err != nil {
			reject(err)

			return
		}

		resolve(nil)
	}()

	return promise, nil
}

func (c *client) isConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.pahoClient == nil {
		return false
	}

	return c.pahoClient.IsConnected()
}

func (c *client) connectPrepare(urlOrOpts sobek.Value, optsOrEmpty sobek.Value) error {
	c.disconnect()

	var (
		url  string
		opts *connectOptions
	)

	switch urlOrOpts.ExportType() {
	case reflect.TypeFor[string]():
		url = urlOrOpts.String()
		urlOrOpts = optsOrEmpty

	case reflect.TypeFor[map[string]any]():

	default:
		return fmt.Errorf("%w: expected string or object", errInvalidType)
	}

	if urlOrOpts != nil && !sobek.IsUndefined(urlOrOpts) && !sobek.IsNull(urlOrOpts) {
		if urlOrOpts.ExportType() != reflect.TypeFor[map[string]any]() {
			return fmt.Errorf("%w: expected object", errInvalidType)
		}

		if err := c.vu.Runtime().ExportTo(urlOrOpts, &opts); err != nil {
			return err
		}
	} else {
		opts = new(connectOptions)
	}

	c.url = url
	c.connOpts = opts

	return nil
}

func (c *client) connectExecute() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.log.Debug("Connecting to MQTT broker")

	c.pahoClient = c.newPahoClient()

	if token := c.pahoClient.Connect(); token.Wait() && token.Error() != nil {
		c.addErrorMetrics(token.Error(), "connect", c.connOpts.Tags, "url", c.url)

		return token.Error()
	}

	c.addCallMetrics("connect", nil)

	return nil
}

func (c *client) disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.pahoClient == nil {
		return
	}

	c.stop <- struct{}{}

	if c.pahoClient.IsConnected() {
		c.pahoClient.Disconnect(0)
	}

	c.pahoClient = nil
}

func (c *client) newPahoClient() paho.Client {
	opts := paho.NewClientOptions()

	if len(c.url) != 0 {
		opts.AddBroker(c.url)
	}

	for _, server := range c.connOpts.Servers {
		opts.AddBroker(server)
	}

	opts.SetDefaultPublishHandler(c.messageHandler)
	opts.SetOnConnectHandler(c.connectHandler)
	opts.SetReconnectingHandler(c.reconnectHandler)

	if sobek.IsString(c.connOpts.ClientId) {
		opts.SetClientID(c.connOpts.ClientId.String())
	}

	if sobek.IsNumber(c.connOpts.Keepalive) {
		opts.SetKeepAlive(time.Second * time.Duration(c.connOpts.Keepalive.ToInteger()))
	}

	if sobek.IsNumber(c.connOpts.ConnectTimeout) && c.connOpts.ConnectTimeout.ToInteger() >= 0 {
		opts.SetConnectTimeout(time.Millisecond * time.Duration(c.connOpts.ConnectTimeout.ToInteger()))
	}

	if c.clientOpts.Will != nil {
		opts.SetWill(c.clientOpts.Will.Topic, c.clientOpts.Will.Payload, c.clientOpts.Will.Qos, c.clientOpts.Will.Retain)
	}

	return paho.NewClient(opts)
}
