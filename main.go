// main.go - Entry point for the QuickPulse server application.
//
// This server can operate in one of three modes, controlled by environment variables:
//   - WebSocket mode (WS_MODE=1): Starts a WebSocket server for publishing and consuming messages.
//   - gRPC unary mode (RPC_MODE=1): Starts a gRPC server supporting unary RPCs.
//   - gRPC streaming mode (RPC_STREAM_MODE=1): Starts a gRPC server supporting streaming RPCs.
//
// The server also exposes Prometheus metrics on :8080/metrics for monitoring.
// Only one mode can be active at a time.

package main

import (
	"log"   // Logging for server events and errors
	"net"   // Networking primitives for TCP listeners
	"net/http" // HTTP server for Prometheus metrics and WebSocket endpoints
	"os"    // For reading environment variables and exiting
	"strconv" // For converting environment variables to integers

	"quickpulse/mq"         // Message queue implementation
	"quickpulse/mqmetrics"  // Instrumented queue and Prometheus metrics
	"quickpulse/proto"      // gRPC protobuf definitions (used for server registration)
	"quickpulse/server"     // WebSocket and gRPC server implementations

	"github.com/prometheus/client_golang/prometheus/promhttp" // Prometheus HTTP handler
	"google.golang.org/grpc"          // gRPC server
	"google.golang.org/grpc/reflection" // gRPC server reflection for debugging
)

// gRPC tuning constants for server configuration
const (
	MaxConcurrentStreams  = uint32(1000000) // Maximum concurrent gRPC streams
	MaxReceiveMessageSize = 1024            // Maximum size of received gRPC messages (1KB)
	WriteBufferSize       = 32 * 1024       // gRPC write buffer size (32KB)
	ReadBufferSize        = 32 * 1024       // gRPC read buffer size (32KB)
)

func main() {
	// Parse mode flags from environment variables (default to 0 if not set or invalid)
	wsMode, _ := strconv.Atoi(os.Getenv("WS_MODE"))
	rpcMode, _ := strconv.Atoi(os.Getenv("RPC_MODE"))
	rpcStreamMode, _ := strconv.Atoi(os.Getenv("RPC_STREAM_MODE"))

	// Ensure exactly one mode is enabled
	modeCount := 0
	if wsMode == 1 {
		modeCount++
	}
	if rpcMode == 1 {
		modeCount++
	}
	if rpcStreamMode == 1 {
		modeCount++
	}
	if modeCount != 1 {
		log.Fatal("Exactly one of WS_MODE, RPC_MODE, or RPC_STREAM_MODE must be set to 1.")
	}

	// Initialize Prometheus metrics and instrumented message queue
	metrics := mqmetrics.NewPrometheusMetrics()
	queue := mq.NewMessageQueue(1000000)
	instrumentedQueue := mqmetrics.NewInstrumentedQueue(queue, metrics)

	// Start Prometheus metrics HTTP server in a separate goroutine
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("Prometheus metrics server listening on :8080/metrics")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("metrics server error: %v", err)
		}
	}()

	// WebSocket server mode
	if wsMode == 1 {
		// Create a new WebSocket server with the instrumented queue
		wsServer := server.NewWsServer(instrumentedQueue)
		// Register HTTP handlers for publish and consume endpoints
		http.HandleFunc("/ws/publish", wsServer.PublishHandler)
		http.HandleFunc("/ws/consume", wsServer.ConsumeHandler)
		log.Println("WebSocket server listening on :8081 (endpoints: /ws/publish, /ws/consume)")
		// Start the HTTP server for WebSocket endpoints
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Fatalf("WebSocket server error: %v", err)
		}
		return
	}

	// gRPC unary server mode
	if rpcMode == 1 {
		// Listen on TCP port 50051 for gRPC connections
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		// Configure gRPC server options for performance tuning
		var serverOpts []grpc.ServerOption
		serverOpts = append(serverOpts, grpc.MaxConcurrentStreams(MaxConcurrentStreams))
		serverOpts = append(serverOpts, grpc.MaxRecvMsgSize(MaxReceiveMessageSize))
		if WriteBufferSize > 0 {
			serverOpts = append(serverOpts, grpc.WriteBufferSize(WriteBufferSize))
		}
		if ReadBufferSize > 0 {
			serverOpts = append(serverOpts, grpc.ReadBufferSize(ReadBufferSize))
		}
		// Create the gRPC server with the configured options
		grpcSrv := grpc.NewServer(serverOpts...)
		// Register the MessageQueue service with a unary handler
		proto.RegisterMessageQueueServer(grpcSrv, server.NewGrpcUnaryServer(instrumentedQueue))
		// Enable server reflection for debugging with tools like grpcurl
		reflection.Register(grpcSrv)

		log.Println("gRPC server (unary) listening on :50051")
		// Start serving gRPC requests
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}

	// gRPC streaming server mode
	if rpcStreamMode == 1 {
		// Listen on TCP port 50051 for gRPC connections
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		// Configure gRPC server options for performance tuning
		var serverOpts []grpc.ServerOption
		serverOpts = append(serverOpts, grpc.MaxConcurrentStreams(MaxConcurrentStreams))
		serverOpts = append(serverOpts, grpc.MaxRecvMsgSize(MaxReceiveMessageSize))
		if WriteBufferSize > 0 {
			serverOpts = append(serverOpts, grpc.WriteBufferSize(WriteBufferSize))
		}
		if ReadBufferSize > 0 {
			serverOpts = append(serverOpts, grpc.ReadBufferSize(ReadBufferSize))
		}
		// Create the gRPC server with the configured options
		grpcSrv := grpc.NewServer(serverOpts...)
		// Register the MessageQueue service with a streaming handler
		proto.RegisterMessageQueueServer(grpcSrv, server.NewGrpcStreamServer(instrumentedQueue))
		// Enable server reflection for debugging with tools like grpcurl
		reflection.Register(grpcSrv)

		log.Println("gRPC server (streaming) listening on :50051")
		// Start serving gRPC requests
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}
}
