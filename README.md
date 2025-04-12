# Message Queue System (Go, gRPC & WebSocket)
```
***************************************************
*   ____       _      _      ___       _          *
*  /___ \_   _(_) ___| | __ / _ \_   _| |___  ___ *
* //  / / | | | |/ __| |/ // /_)/ | | | / __|/ _ \*
*/ \_/ /| |_| | | (__|   </ ___/| |_| | \__ \  __/*
*\___,_\ \__,_|_|\___|_|\_\/     \__,_|_|___/\___|*
***************************************************
```
⚡⚡⚡ **Blazing fast, in-memory message queue** ⚡⚡⚡


QuickPulse is an opinionated, ephemeral message queue system.


## Architecture Overview

- **Message**: Represents a message with a payload (and optional ID).
- **MessageQueue**: Thread-safe queue implementation with Enqueue and Dequeue operations.
- **InstrumentedQueue**: Wraps MessageQueue to collect metrics on queue operations.
- **MetricsCollector**: Interface for collecting queue metrics, with implementations for default and Prometheus metrics.
- **GrpcUnaryServer / GrpcStreamServer**: Implements the gRPC service, exposing endpoints for producing and consuming messages (unary or streaming).
- **WsServer**: Implements the WebSocket API, exposing endpoints for publishing and consuming messages.
- **PerfClient**: Tools for benchmarking queue performance via gRPC (unary and streaming) and WebSocket.

## Components

- **Message**: Go struct with fields for payload (and optional ID).
- **MessageQueue**: Go struct managing a slice of messages and providing thread-safe Enqueue/Dequeue.
- **InstrumentedQueue**: Go struct that wraps a MessageQueue and a MetricsCollector, providing instrumented Enqueue/Dequeue.
- **MetricsCollector**: Interface for metrics collection. Implemented by:
  - **DefaultMetrics**: Basic in-memory metrics.
  - **PrometheusMetrics**: Exposes metrics in Prometheus format.
- **GrpcUnaryServer / GrpcStreamServer**: Go structs implementing the gRPC service methods (unary or streaming).
- **WsServer**: Go struct implementing WebSocket handlers for publish/consume.
- **PerfClient**: Go tools for running throughput and load tests via gRPC (unary and streaming) and WebSocket.
- **Protobuf Definitions**: Located in `proto/messagequeue.proto`, defining the gRPC service and message formats.

## Mode Selection

The server supports three mutually exclusive modes, controlled by environment variables:

- `WS_MODE=1`: Run in WebSocket mode (HTTP server on port 8081).
- `RPC_MODE=1`: Run in gRPC unary mode (gRPC server on port 50051).
- `RPC_STREAM_MODE=1`: Run in gRPC streaming mode (gRPC server on port 50051, bidirectional streaming enabled).

**Note:** Only one mode can be active at a time. If more than one or none are set, the server will exit with an error.

## gRPC API

The gRPC service is defined as follows:

- **Service**: `MessageQueue`
    - `Produce(ProduceRequest) returns (ProduceResponse)`
    - `Consume(ConsumeRequest) returns (ConsumeResponse)`
    - `StreamMessages(stream StreamMessage) returns (stream StreamMessage)` (bidirectional streaming, enabled in `RPC_STREAM_MODE`)

### Protobuf Messages

- **ProduceRequest**: `{ bytes payload }`
- **ProduceResponse**: `{ bool success, string error }`
- **ConsumeRequest**: `{}`
- **ConsumeResponse**: `{ bytes payload, string error }`

See `proto/messagequeue.proto` for details.

### Streaming Mode

When running in streaming mode (`RPC_STREAM_MODE=1`), the gRPC server exposes the `StreamMessages` RPC, which allows clients to send and receive messages in a bidirectional stream. Each message sent by the client is enqueued, and the server responds with the next available message from the queue (or an error if the queue is empty).

## WebSocket API

When running in WebSocket mode (`WS_MODE=1`), the server exposes two endpoints:

### Prometheus Dashboard

Below is a sample Prometheus dashboard visualizing QuickPulse metrics:

[Prometheus Dashboard](docs/prometheus_dashboard.png)

_Place your Prometheus dashboard screenshot at `docs/prometheus_dashboard.png` to display it here._
- `ws://<host>:8081/ws/publish`:  
  Clients connect and send messages (as binary/text frames) to be enqueued.  
  The server responds with "ok" or "error: ..." for each message.

- `ws://<host>:8081/ws/consume`:  
  Clients connect and send a request (any message, e.g., "next") to receive a message from the queue.  
  The server responds with the next message (as a binary frame), or "error: ..." if the queue is empty.

## Metrics and Monitoring

- **Prometheus metrics** are exposed on `http://<host>:8080/metrics` in all modes.
- Metrics are collected via the `MetricsCollector` interface, with support for both in-memory and Prometheus-compatible metrics.

## Performance Testing

- **PerfClient** tools are provided for benchmarking queue throughput:
  - `perf_grpc_throughput.go`: gRPC unary performance tests.
  - `perf_grpc_stream_throughput.go`: gRPC streaming performance tests.
  - `perf_ws_throughput.go`: WebSocket performance tests.
- These tools allow you to simulate concurrent producers/consumers and measure system throughput.
### Performance Benchmark System Specs

Performance tests were run on the following system:

- CPU: Apple M1 Pro
- Cores: 8
- RAM: 16 GB
- Hardware Model: MacBookPro18,3
- Operating System: macOS 14.2.1
- Default Shell: /bin/zsh

These specs provide context for interpreting the performance results obtained using the perfclient tools.
### Running PerfClient

You can benchmark the message queue system using the provided perfclient tool in different modes. Use the following commands from the project root:

- **gRPC Unary Mode:**
  ```sh
  go run ./cmd/perfclient/main.go -mode grpc
  ```

- **gRPC Streaming Mode:**
  ```sh
  go run ./cmd/perfclient/main.go -mode grpc_stream
  ```

- **WebSocket Mode:**
  ```sh
  go run ./cmd/perfclient/main.go -mode ws
  ```

Each mode will run the corresponding performance test as described above.

## UML Diagram

The design is described in `message_queue.puml` using PlantUML syntax.

### Viewing the UML Diagram

You can render the UML diagram using any of the following methods:

- **PlantUML Online Editor**:  
  1. Go to [PlantUML Online Server](https://www.plantuml.com/plantuml/uml/).
  2. Copy the contents of `message_queue.puml` and paste it into the editor.
  3. The diagram will be rendered automatically.

- **VSCode PlantUML Extension**:  
  1. Install the "PlantUML" extension in VSCode.
  2. Open `message_queue.puml`.
  3. Use the "Preview Current Diagram" command.

- **Command Line**:  
  1. Install PlantUML and Java.
  2. Run:  
     ```
     plantuml message_queue.puml
     ```
  3. This will generate a PNG or SVG image.

## Usage Scenario

1. A client (gRPC or WebSocket) sends a request to enqueue a message.
2. The server enqueues the message and responds with a success or error.
3. A client sends a request to dequeue a message.
4. The server dequeues a message (if available) and responds with the message or an error.
5. Metrics are collected and can be visualized via Prometheus.
6. Performance can be tested using the provided perfclient tools.

This design supports multiple producers and consumers, message acknowledgments, extensible metrics, and can be extended for persistence.

## Running with Docker

You can build and run the message queue server using Docker. The Docker image supports all modes (gRPC, gRPC streaming, WebSocket) and exposes the necessary ports.

### 1. Build the Docker image

```sh
docker build -t quickpulse .
```

### 2. Run the server in different modes

Only one mode can be active at a time. Use environment variables to select the mode.

#### gRPC Unary Mode (default, port 50051)

```sh
docker run --rm -p 50051:50051 -p 8080:8080 quickpulse
```

#### gRPC Streaming Mode (port 50051)

```sh
docker run --rm -e RPC_MODE=0 -e RPC_STREAM_MODE=1 -p 50051:50051 -p 8080:8080 quickpulse
```

#### WebSocket Mode (port 8081)

```sh
docker run --rm -e WS_MODE=1 -e RPC_MODE=0 -p 8081:8081 -p 8080:8080 quickpulse
```

### 3. Access Prometheus Metrics

Metrics are available at [http://localhost:8080/metrics](http://localhost:8080/metrics) in all modes.

### 4. Customizing Environment Variables

You can override the default environment variables using the `-e` flag with `docker run` to select the desired mode.

---