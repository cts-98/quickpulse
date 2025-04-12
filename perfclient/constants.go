// constants.go - Configuration constants for the performance client.
//
// This file defines default addresses, message counts, test durations, and payloads
// used by the performance testing client for both gRPC and WebSocket modes.

package perfclient

const (
	gRPCAddress     = "localhost:50051"              // Default gRPC server address
	WSAddress       = "ws://localhost:8081/ws/publish" // Default WebSocket publish endpoint
	totalMessages   = 500000                         // Default total number of messages to send in a test
	testDurationSec = 60                             // Default test duration in seconds
	payloadBase64   = "eyJrIjoidiJ9"                 // Default base64-encoded payload
)
