# xk6-mqtt

> **⚠️ Warning**
>
> This is an **experimental extension** for k6. It is **not officially supported yet**.  
> We are actively working on it to make it officially supported in the future.

**MQTT protocol support for k6**

**xk6-mqtt** is a k6 extension that adds first-class support for the [MQTT](https://mqtt.org) protocol to your load testing and performance scripts. With this extension, you can connect to MQTT brokers, publish and subscribe to topics, and interact with MQTT systems directly from your k6 tests.

The [xk6-mqtt API](https://mqtt.x.k6.io) is intentionally designed to feel familiar to users of [MQTT.js](https://github.com/mqttjs/MQTT.js), the popular JavaScript MQTT client. This means you can leverage event-driven programming, both synchronous and asynchronous operations, and migrate existing MQTT.js-based test logic with minimal changes. The extension aims to provide a modern, ergonomic developer experience for MQTT load testing in JavaScript.

## Example

```javascript file=examples/hello.js
import { Client } from "k6/x/mqtt";

export default function () {
  const client = new Client()

  client.on("connect", async () => {
    console.log("Connected to MQTT broker")
    client.subscribe("greeting")
    client.publish("greeting", "Hello MQTT!")
  })

  client.on("message", (topic, message) => {
    const str = String.fromCharCode.apply(null, new Uint8Array(message))
    console.info("topic:", topic, "message:", str)
    client.end()
  })

  client.on("end", () => {
    console.log("Disconnected from MQTT broker")
  })

  client.connect(__ENV["MQTT_BROKER_ADDRESS"] || "mqtt://broker.emqx.io:1883")
}
```

You can find more examples in the [examples](./examples/) folder.

## Async Programming

**xk6-mqtt** fully supports async and event-based programming. You can use [setTimeout()](https://developer.mozilla.org/en-US/docs/Web/API/Window/setTimeout), [setInterval()](https://developer.mozilla.org/en-US/docs/Web/API/Window/setInterval), and other asynchronous patterns together with xk6-mqtt's event handlers such as `on("connect")`, `on("message")`, and `on("end")`. This allows you to implement complex MQTT workflows, timers, and message handling logic in a style familiar to JavaScript developers.

```javascript file=examples/ping.js
import { Client } from "k6/x/mqtt";

export default function () {
  const client = new Client()

  client.on("connect", async () => {
    console.log("Connected to MQTT broker")
    client.subscribe("probe")

    const intervalId = setInterval(() => {
      client.publish("probe", "ping MQTT!")
    }, 1000)

    setTimeout(() => {
      clearInterval(intervalId)
      client.end()
    }, 3100)
  })

  client.on("message", (topic, message) => {
    console.info(String.fromCharCode.apply(null, new Uint8Array(message)))
  })

  client.on("end", () => {
    console.log("Disconnected from MQTT broker")
  })

  client.connect(__ENV["MQTT_BROKER_ADDRESS"] || "mqtt://broker.emqx.io:1883")
}
```

## Quick Start

1. **Build a custom k6 binary with xk6-mqtt**  
   You need to build k6 with this extension using [xk6](https://github.com/grafana/xk6):

   ```sh
   go install go.k6.io/xk6/cmd/xk6@latest
   xk6 build --with github.com/grafana/xk6-mqtt
   ```

2. **Write your test script**  
   Use the example above or create your own test script using the `Client` API.

3. **Run your test**  
   Use your custom k6 binary to run the script:

   ```sh
   ./k6 run script.js
   ```

## Download

Pre-built binaries for k6 with the xk6-mqtt extension are available on the [Releases page](https://github.com/grafana/xk6-mqtt/releases/).

## Limitations

- The underlying [Eclipse Paho](https://eclipse.dev/paho/) MQTT library supports version v3.1.1 of the MQTT protocol. MQTT v5 is not supported yet.

## Contributing

We welcome contributions! Please see the [Contributing Guidelines](CONTRIBUTING.md) for details on how to get started.

## Status

**xk6-mqtt** is in an early stage of development but is already usable for many MQTT load testing scenarios. Please note that some features are not yet implemented:

- **Authentication**: Neither basic username/password nor TLS-based authentication is currently supported.
- **Error event handling**: The `error` event handler is not yet implemented.

We are actively working to add these features and improve the extension. Feedback and contributions are welcome!
