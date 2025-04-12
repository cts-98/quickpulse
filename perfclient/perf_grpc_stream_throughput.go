// perf_grpc_stream_throughput.go - gRPC streaming performance test client.
//
// This file implements a concurrent performance test for gRPC streaming, sending and
// receiving messages over multiple streams with configurable concurrency and in-flight
// message limits. It tracks sent, received, and error counts, and reports throughput.

package perfclient

import (
	"context"      // For context and cancellation
	"encoding/base64" // For decoding payloads
	"fmt"          // For formatted output
	"log"          // For logging errors
	"runtime"      // For setting GOMAXPROCS
	"sync"         // For WaitGroup and concurrency
	"sync/atomic"  // For atomic counters
	"time"         // For timing and test duration

	"google.golang.org/grpc" // gRPC client
	pb "quickpulse/proto"     // gRPC protobuf definitions
)

// RunGRPCStreamPerfTest runs a concurrent gRPC streaming performance test.
//
// Parameters:
//   - concurrency: number of parallel streams
//   - inflight: number of in-flight messages per stream
//   - totalMessages: total number of messages to send
//   - testDurationSec: maximum test duration in seconds
//   - payloadBase64: base64-encoded payload to send
//   - grpcAddress: gRPC server address
func RunGRPCStreamPerfTest(concurrency, inflight int, totalMessages int64, testDurationSec int, payloadBase64, grpcAddress string) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Printf("Starting gRPC streaming perf test: %d messages, %d streams, %d in-flight per stream, %d seconds max\n",
		totalMessages, concurrency, inflight, testDurationSec)
	payload, _ := base64.StdEncoding.DecodeString(payloadBase64)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(testDurationSec)*time.Second)
	defer cancel()

	var sent int64     // Total messages sent
	var received int64 // Total messages received
	var errors int64   // Total errors encountered
	var wg sync.WaitGroup

	start := time.Now()

	// Worker function for each stream
	worker := func(id int) {
		defer wg.Done()
		conn, err := grpc.Dial(grpcAddress, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()
		client := pb.NewMessageQueueClient(conn)
		stream, err := client.StreamMessages(ctx)
		if err != nil {
			log.Fatalf("Failed to open stream: %v", err)
		}

		var localWg sync.WaitGroup
		sem := make(chan struct{}, inflight) // Semaphore to limit in-flight messages

		// Sender goroutine: sends messages as long as the test is running and quota remains
		localWg.Add(1)
		go func() {
			defer localWg.Done()
			for {
				if ctx.Err() != nil || atomic.LoadInt64(&sent) >= totalMessages {
					return
				}
				sem <- struct{}{}
				go func() {
					defer func() { <-sem }()
					msg := &pb.StreamMessage{Payload: payload}
					if err := stream.Send(msg); err != nil {
						atomic.AddInt64(&errors, 1)
					} else {
						atomic.AddInt64(&sent, 1)
					}
				}()
			}
		}()

		// Receiver goroutine: receives responses from the server
		localWg.Add(1)
		go func() {
			defer localWg.Done()
			for {
				if ctx.Err() != nil || atomic.LoadInt64(&received) >= totalMessages {
					return
				}
				resp, err := stream.Recv()
				if err != nil {
					atomic.AddInt64(&errors, 1)
					return
				}
				if resp.Error != "" {
					atomic.AddInt64(&errors, 1)
				} else {
					atomic.AddInt64(&received, 1)
				}
			}
		}()

		localWg.Wait()
	}

	// Start all workers (streams)
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go worker(i)
	}
	wg.Wait()

	elapsed := time.Since(start).Seconds()
	fmt.Printf("Test complete: sent=%d, received=%d, errors=%d, elapsed=%.2fs, throughput=%.0f msg/sec\n",
		sent, received, errors, elapsed, float64(received)/elapsed)
}