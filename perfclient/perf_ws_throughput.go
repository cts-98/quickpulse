// perf_ws_throughput.go - WebSocket performance test client.
//
// This file implements a performance test for WebSocket message publishing,
// sending messages over a single connection and tracking sent messages, errors,
// and throughput.

package perfclient

import (
	"encoding/base64" // For decoding payloads
	"fmt"             // For formatted output
	"log"             // For logging errors
	"runtime"         // For setting GOMAXPROCS
	"time"            // For timing and test duration

	"github.com/gorilla/websocket" // WebSocket client
)

// RunWSPerfTest runs the WebSocket performance test with the given concurrency and inflight settings.
//
// Parameters:
//   - _concurrency: (unused, for interface compatibility)
//   - _inflight: (unused, for interface compatibility)
func RunWSPerfTest(_concurrency, _inflight int) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Printf("Starting WebSocket perf test: %d messages, single connection, %d seconds max\n", totalMessages, testDurationSec)
	payload, _ := base64.StdEncoding.DecodeString(payloadBase64)

	var sent int64   // Total messages sent
	var errors int64 // Total errors encountered
	start := time.Now()
	endTime := start.Add(time.Duration(testDurationSec) * time.Second)

	// Establish a single WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(WSAddress, nil)
	if err != nil {
		log.Printf("Failed to connect: %v", err)
		return
	}
	defer conn.Close()

	for {
		// Stop if time is up or enough messages sent
		if time.Now().After(endTime) || sent >= int64(totalMessages) {
			break
		}
		// Send a binary message (the payload)
		err := conn.WriteMessage(websocket.BinaryMessage, payload)
		if err != nil {
			errors++
			continue
		}
		// Read the response from the server
		_, resp, err := conn.ReadMessage()
		if err != nil || string(resp) != "ok" {
			errors++
		}
		sent++
		if sent%50000 == 0 {
			fmt.Printf("Progress: %d messages sent (%s)\n", sent, time.Now().Format(time.RFC3339))
		}
	}

	elapsed := time.Since(start)

	fmt.Println("Test complete!")
	fmt.Printf("Total messages sent: %d\n", sent)
	fmt.Printf("Total errors: %d\n", errors)
	fmt.Printf("Elapsed time: %.2f seconds\n", elapsed.Seconds())
	fmt.Printf("Throughput: %.2f messages/sec\n", float64(sent)/elapsed.Seconds())
	fmt.Println("Tip: This test now uses a single WebSocket connection for all messages.")
}