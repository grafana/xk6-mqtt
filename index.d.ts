/**
 * **MQTT protocol support for k6**
 *
 * > **⚠️ Warning**
 * >
 * > This is an **experimental extension** for k6. It is **not officially supported yet**.
 * > We are actively working on it to make it officially supported in the future.
 *
 * **xk6-mqtt** is a k6 extension that adds first-class support for the [MQTT](https://mqtt.org)
 * protocol to your load testing and performance scripts. With this extension, you can connect
 * to MQTT brokers, publish and subscribe to topics, and interact with MQTT systems directly
 * from your k6 tests.
 *
 * The API is intentionally designed to feel familiar to users of [MQTT.js](https://github.com/mqttjs/MQTT.js),
 * the popular JavaScript MQTT client. This means you can leverage event-driven programming,
 * both synchronous and asynchronous operations, and migrate existing MQTT.js-based test logic with minimal changes.
 * The extension aims to provide a modern, ergonomic developer experience for MQTT load testing in JavaScript.
 *
 * @module mqtt
 *
 * @example Hello World
 * Comparing HTTP-based tests to MQTT ones, you’ll find differences in both structure and inner workings.
 * The primary difference is that instead of continuously looping the main function
 * (`export default function() { ... }`) over and over, each VU is now runs an asynchronous event loop.
 *
 * When the MQTT connection is created, the `connect` handler function will be immediately called,
 * all code inside it will be executed (usually code to set up other event handlers),
 * and then blocked until the MQTT connection is closed (by the remote host or by using `client.end()`).
 *
 * ```javascript
 * import { Client } from "k6/x/mqtt";
 *
 * export default function () {
 *   const client = new Client()
 *
 *   client.on("connect", async () => {
 *     console.log("Connected to MQTT broker")
 *     client.subscribe("greeting")
 *     client.publish("greeting", "Hello MQTT!")
 *   })
 *
 *   client.on("message", (topic, message) => {
 *     const str = String.fromCharCode.apply(null, new Uint8Array(message))
 *     console.info("topic:", topic, "message:", str)
 *     client.end()
 *   })
 *
 *   client.on("end", () => {
 *     console.log("Disconnected from MQTT broker")
 *   })
 *
 *   client.connect(__ENV["MQTT_BROKER_ADDRESS"] || "mqtt://broker.emqx.io:1883")
 * }
 * ```
 *
 * @example Async Programming
 * Async and event-based programming is fully supported.
 * You can use [setTimeout()](https://developer.mozilla.org/en-US/docs/Web/API/Window/setTimeout),
 * [setInterval()](https://developer.mozilla.org/en-US/docs/Web/API/Window/setInterval),
 * and other async patterns with xk6-mqtt event handlers.
 * ```javascript
 * import { Client } from "k6/x/mqtt";
 *
 * export default function () {
 *   const client = new Client()
 *
 *   client.on("connect", async () => {
 *     console.log("Connected to MQTT broker")
 *     client.subscribe("probe")
 *
 *     const intervalId = setInterval(() => {
 *       client.publish("probe", "ping MQTT!")
 *     }, 1000)
 *
 *     setTimeout(() => {
 *       clearInterval(intervalId)
 *       client.end()
 *     }, 3100)
 *   })
 *
 *   client.on("message", (topic, message) => {
 *     console.info(String.fromCharCode.apply(null, new Uint8Array(message)))
 *   })
 *
 *   client.on("end", () => {
 *     console.log("Disconnected from MQTT broker")
 *   })
 *
 *   client.connect(__ENV["MQTT_BROKER_ADDRESS"] || "mqtt://broker.emqx.io:1883")
 * }
 * ```
 */
export as namespace mqtt;

/**
 * Quality of Service levels for MQTT message delivery.
 */
export declare enum QoS {
  /** At most once delivery (fire and forget). */
  AtMostOnce = 0,
  /** At least once delivery (guaranteed delivery). */
  AtLeastOnce = 1,
  /** Exactly once delivery (guaranteed and duplicate-free). */
  ExactlyOnce = 2
}

/**
 * Defines the "Last Will and Testament" message for MQTT clients.
 * This message is sent by the broker when the client disconnects unexpectedly.
 */
export declare interface Will {
  /** Topic for the will message. */
  topic: string;
  /** Payload for the will message. */
  payload: string;
  /** QoS level for the will message. */
  qos?: QoS;
  /** Whether the will message should be retained. */
  retain?: boolean;
}

/**
 * Base interface for objects that can have tags.
 * Tags are key-value pairs used for metrics and logging.
 * The keys are names of tags and the values are tag values.
 */
export declare interface HasTags {
  /** Optional tags for metrics and logging. */
  tags?: Record<string, string>;
}

/**
 * Options for creating a new MQTT client.
 */
export declare interface ClientOptions extends HasTags {
  /** Last Will and Testament message. */
  will?: Will;
}

/**
 * Options for connecting to an MQTT broker.
 */
export declare interface ConnectOptions extends HasTags {
  /** Client identifier (must be unique per broker connection) */
  clientId?: string;
  /** Keep-alive interval in seconds (default: 60) */
  keepalive?: number;
  /** Connection timeout in milliseconds (default: 30000) */
  connectTimeout?: number;
  /** Array of broker URLs to connect to (for failover) */
  servers?: string[];
}

/**
 * Options for ending a client connection.
 */
export declare interface EndOptions extends HasTags { }

/**
 * Options for subscribing to MQTT topics.
 */
export declare interface SubscribeOptions extends HasTags {
  /** Quality of Service level for the subscription. */
  qos?: QoS;
}

/**
 * Type alias for accepting either a single string or an array of strings.
 * Used for topics and other string-based parameters.
 */
export declare type StringOrStrings = string | string[];

/**
 * Options for unsubscribing from MQTT topics.
 */
export declare interface UnsubscribeOptions extends HasTags { }

/**
 * Options for publishing MQTT messages.
 */
export declare interface PublishOptions extends HasTags {
  /** Quality of Service level for the message (default: 0) */
  qos?: QoS;
  /** Whether the message should be retained by the broker (default: false) */
  retain?: boolean;
}

/**
 * Type alias for message payloads.
 * Accepts either string or ArrayBuffer for binary data.
 */
export declare type StringOrArrayBuffer = string | ArrayBuffer;

/**
 * MQTT client for connecting to brokers and managing MQTT operations
 *
 * @example
 * ```javascript
 * import { Client } from "k6/x/mqtt";
 *
 * export default function () {
 *   const client = new Client()
 *
 *   client.on("connect", async () => {
 *     console.log("Connected to MQTT broker")
 *     client.subscribe("greeting")
 *
 *     client.publish("greeting", "Hello MQTT!")
 *   })
 *
 *   client.on("message", (topic, message) => {
 *     const str = String.fromCharCode.apply(null, new Uint8Array(message))
 *     console.info("topic:", topic, "message:", str)
 *     client.end()
 *   })
 *
 *   client.on("end", () => {
 *     console.log("Disconnected from MQTT broker")
 *   })
 *
 *   client.connect(__ENV["MQTT_BROKER_ADDRESS"] || "mqtt://broker.emqx.io:1883")
 * }
 * ```
 */
export declare class Client {
  /** Indicates if the client is currently connected. */
  readonly connected: boolean;

  /**
   * Create a new MQTT client.
   * @param options Optional client options.
   */
  constructor(options?: ClientOptions);

  /**
   * Connects to an MQTT broker.
   * @param url Broker URL or connection options.
   * @param options Optional connection options.
   */
  connect(url: string | ConnectOptions, options?: ConnectOptions): void;

  /**
   * Disconnects from the MQTT broker synchronously.
   * @param options - Optional disconnect options.
   */
  end(options?: EndOptions): void;

  /**
   * Disconnects from the MQTT broker asynchronously.
   * @param options - Optional disconnect options.
   * @returns Promise that resolves when disconnection is complete.
   */
  endAsync(options?: EndOptions): Promise<void>;

  /**
   * Attempts to reconnect to the MQTT broker.
   * Uses the same connection parameters as the last successful connection.
   */
  reconnect(): void;

  /**
   * Subscribe to one or more topics synchronously.
   * @param topic Topic(s) or subscription options.
   * @param options Optional subscription options.
   */
  subscribe(topic: StringOrStrings | SubscribeOptions, options?: SubscribeOptions): void;

  /**
   * Subscribe to one or more topics asynchronously.
   * @param topic Topic(s) or subscription options.
   * @param options Optional subscription options.
   * @returns Promise that resolves when subscription is complete.
   */
  subscribeAsync(topic: StringOrStrings | SubscribeOptions, options?: SubscribeOptions): Promise<void>;

  /**
   * Unsubscribe from one or more topics synchronously.
   * @param topics Topic(s) to unsubscribe from.
   * @param options - Optional unsubscribe options.
   */
  unsubscribe(topics: StringOrStrings, options?: UnsubscribeOptions): void;

  /**
   * Unsubscribe from one or more topics asynchronously.
   * @param topics Topic(s) to unsubscribe from.
   * @param options - Optional unsubscribe options.
   * @returns Promise that resolves when unsubscription is complete.
   */
  unsubscribeAsync(topics: StringOrStrings, options?: UnsubscribeOptions): Promise<void>;

  /**
   * Publish a message to a MQTT topic synchronously.
   * @param topic - The topic to publish to.
   * @param payload - The message payload (string or ArrayBuffer).
   * @param options - Optional publish options.
   */
  publish(topic: string, payload: StringOrArrayBuffer, options?: PublishOptions): void;

  /**
   * Publishes a message to an MQTT topic asynchronously.
   * @param topic - The topic to publish to.
   * @param payload - The message payload (string or ArrayBuffer).
   * @param options - Optional publish options.
   * @returns Promise that resolves when publish is complete.
   */
  publishAsync(topic: string, payload: StringOrArrayBuffer, options?: PublishOptions): Promise<void>;

  /**
   * Listen for the `connect` event.
   * @param listener Callback for connect event.
   */
  on(event: "connect", listener: () => void): void;

  /**
   * Listen for the `end` event.
   * @param listener Callback for end event.
   */
  on(event: "end", listener: () => void): void;

  /**
   * Listen for the `reconnect` event.
   * @param listener Callback for reconnect event.
   */
  on(event: "reconnect", listener: () => void): void;

  /**
   * Listen for incoming messages.
   * @param listener Callback for message event.
   */
  on(event: "message", listener: (topic: string, payload: ArrayBuffer) => void): void;

  /**
   * Listen for errors.
   * @param listener Callback for error event.
   */
  on(event: "error", listener: (error: Error) => void): void;
}
