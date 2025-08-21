package mqtt

import (
	"github.com/grafana/sobek"
	"go.k6.io/k6/js/promises"
)

type endOptions struct {
	Tags map[string]string
}

func (c *client) end(opts *endOptions) error {
	if !c.isConnected() {
		return errNotConnected
	}

	if opts == nil {
		opts = new(endOptions)
	}

	c.log.Debug("Disconnecting from MQTT broker")

	c.fire("end")

	c.addCallMetrics("end", opts.Tags)

	c.disconnect()

	return nil
}

func (c *client) endAsync(opts *endOptions) (*sobek.Promise, error) {
	if !c.isConnected() {
		return nil, errNotConnected
	}

	promise, resolve, reject := promises.New(c.vu)

	go func() {
		if err := c.end(opts); err != nil {
			reject(err)

			return
		}

		resolve(nil)
	}()

	return promise, nil
}
