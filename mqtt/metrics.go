package mqtt

import (
	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/metrics"
)

const (
	mqttMessageSent     = "mqtt_message_sent"
	mqttMessageReceived = "mqtt_message_received"
	mqttDataSent        = "mqtt_data_sent"
	mqttDataReceived    = "mqtt_data_received"
	mqttErrors          = "mqtt_errors"
	mqttCalls           = "mqtt_calls"
)

type mqttMetrics struct {
	dataSent            *metrics.Metric
	dataReceived        *metrics.Metric
	mqttDataSent        *metrics.Metric
	mqttDataReceived    *metrics.Metric
	mqttMessageSent     *metrics.Metric
	mqttMessageReceived *metrics.Metric
	mqttErrors          *metrics.Metric
	mqttCalls           *metrics.Metric
}

func newMqttMetrics(vu modules.VU) *mqttMetrics {
	return &mqttMetrics{
		dataSent:            vu.InitEnv().BuiltinMetrics.DataSent,
		dataReceived:        vu.InitEnv().BuiltinMetrics.DataReceived,
		mqttMessageSent:     vu.InitEnv().Registry.MustNewMetric(mqttMessageSent, metrics.Counter),
		mqttMessageReceived: vu.InitEnv().Registry.MustNewMetric(mqttMessageReceived, metrics.Counter),
		mqttDataSent:        vu.InitEnv().Registry.MustNewMetric(mqttDataSent, metrics.Counter),
		mqttDataReceived:    vu.InitEnv().Registry.MustNewMetric(mqttDataReceived, metrics.Counter),
		mqttErrors:          vu.InitEnv().Registry.MustNewMetric(mqttErrors, metrics.Counter),
		mqttCalls:           vu.InitEnv().Registry.MustNewMetric(mqttCalls, metrics.Counter),
	}
}
