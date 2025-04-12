// Package main provides a command-line tool for running performance tests
// against WebSocket, gRPC, and gRPC streaming servers. It allows configuration
// of concurrency, in-flight requests, message count, test duration, payload, and
// server address via command-line flags. The actual test logic is implemented in
// the quickpulse/perfclient package.
package main

import (
	"flag"   // For parsing command-line flags
	"fmt"    // For formatted I/O
	"os"     // For OS-level functions like exiting the program

	"quickpulse/perfclient" // Import the perfclient package containing test runners
)

// main is the entry point for the performance client.
// It parses command-line flags to determine the test mode and parameters,
// then dispatches to the appropriate performance test function.
func main() {
	// Define command-line flags for configuring the test
	mode := flag.String("mode", "ws", "Test mode: ws, grpc, or grpc_stream")
	concurrency := flag.Int("concurrency", 500, "Number of parallel workers/connections/streams")
	inflight := flag.Int("inflight", 20, "Number of in-flight requests per worker/stream")
	messages := flag.Int64("messages", 2000000, "Total messages to send (default: 2M)")
	duration := flag.Int("duration", 5, "Test duration in seconds")
	payload := flag.String("payload", "aGVsbG8gd29ybGQ=", "Base64-encoded payload")
	address := flag.String("address", "localhost:50051", "gRPC server address")
	flag.Parse() // Parse the command-line flags

	// Select the test mode and run the corresponding performance test
	switch *mode {
	case "ws":
		// Run WebSocket performance test with specified concurrency and inflight settings
		fmt.Println("Running WebSocket perf test...")
		perfclient.RunWSPerfTest(*concurrency, *inflight)
	case "grpc":
		// Run gRPC performance test with specified concurrency and inflight settings
		fmt.Println("Running gRPC perf test...")
		perfclient.RunGRPCPerfTest(*concurrency, *inflight)
	case "grpc_stream":
		// Run gRPC streaming performance test with all provided parameters
		fmt.Println("Running gRPC streaming perf test...")
		perfclient.RunGRPCStreamPerfTest(*concurrency, *inflight, *messages, *duration, *payload, *address)
	default:
		// Handle unknown mode by printing an error and exiting with a non-zero status
		fmt.Fprintf(os.Stderr, "Unknown mode: %s\n", *mode)
		os.Exit(1)
	}
}