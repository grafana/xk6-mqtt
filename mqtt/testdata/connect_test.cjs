const mqtt = require("k6/x/mqtt")
const assert = require("k6/x/assert")

var connectHandlerCalled = false
var endHandlerCalled = false

module.exports = () => {
  const client = new mqtt.Client()

  client.on("connect", () => {
    assert.true(client.connected, "Client should be connected after connect event")

    connectHandlerCalled = true

    client.end()
  })

  client.on("end", () => {
    assert.false(client.connected, "Client should not be connected after end event")

    endHandlerCalled = true
  })

  assert.false(client.connected, "Client should not be connected initially")

  client.connect(__ENV.MQTT_BROKER_ADDRESS)

  assert.true(client.connected, "Client should be connected after connect call")
}

module.exports.teardown = () => {
  assert.true(connectHandlerCalled, "Connect handler was not called")
  assert.true(endHandlerCalled, "End handler was not called")
}
