package mqtt

import (
	"sync"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/grafana/sobek"
	"github.com/sirupsen/logrus"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
)

type will struct {
	Topic   string
	Payload string
	Qos     byte
	Retain  bool
}

type clientOptions struct {
	Will *will
	Tags map[string]string
}

type client struct {
	pahoClient paho.Client

	url string
	log logrus.FieldLogger

	clientOpts *clientOptions
	connOpts   *connectOptions

	handlers map[string]sobek.Callable

	vu       modules.VU
	callChan chan func() error
	stop     chan struct{}

	metrics *mqttMetrics

	mu sync.RWMutex
}

func newClient(log logrus.FieldLogger, vu modules.VU, metrics *mqttMetrics) *client {
	c := new(client)
	c.log = log
	c.handlers = make(map[string]sobek.Callable)
	c.vu = vu
	c.callChan = make(chan func() error)
	c.stop = make(chan struct{})

	c.metrics = metrics

	return c
}

func (m *module) client(call sobek.ConstructorCall) *sobek.Object {
	toValue := m.vu.Runtime().ToValue

	c := newClient(m.log, m.vu, m.metrics)
	this := call.This
	must := func(err error) {
		if err != nil {
			common.Throw(m.vu.Runtime(), err)
		}
	}

	c.clientOpts = new(clientOptions)

	if len(call.Arguments) > 0 {
		must(m.vu.Runtime().ExportTo(call.Arguments[0], &c.clientOpts))
	}

	must(this.Set("connect", toValue(c.connect)))
	must(this.Set("connectAsync", toValue(c.connectAsync)))
	must(this.Set("end", toValue(c.end)))
	must(this.Set("endAsync", toValue(c.endAsync)))
	must(this.Set("reconnect", toValue(c.reconnect)))
	must(this.Set("reconnectAsync", toValue(c.reconnectAsync)))
	must(this.Set("publish", toValue(c.publish)))
	must(this.Set("publishAsync", toValue(c.publishAsync)))
	must(this.Set("subscribe", toValue(c.subscribe)))
	must(this.Set("subscribeAsync", toValue(c.subscribeAsync)))
	must(this.Set("unsubscribe", toValue(c.unsubscribe)))
	must(this.Set("unsubscribeAsync", toValue(c.unsubscribeAsync)))
	must(this.Set("on", toValue(c.on)))

	must(this.DefineAccessorProperty("connected", toValue(c.isConnected), nil, sobek.FLAG_FALSE, sobek.FLAG_FALSE))

	go c.loop()

	return nil
}
