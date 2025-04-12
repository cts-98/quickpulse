// perf_grpc_throughput.go - gRPC unary performance test client.
//
// This file implements a concurrent performance test for gRPC unary requests,
// sending messages over multiple workers with configurable concurrency and in-flight
// request limits. It tracks sent messages, errors, and reports throughput.

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

// RunGRPCPerfTest runs the gRPC performance test with the given concurrency and inflight settings.
//
// Parameters:
//   - concurrency: number of parallel workers
//   - inflight: number of in-flight requests per worker
func RunGRPCPerfTest(concurrency, inflight int) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Printf("Starting gRPC perf test: %d messages, %d workers, %d in-flight per worker, %d seconds max\n", totalMessages, concurrency, inflight, testDurationSec)
	payload, _ := base64.StdEncoding.DecodeString(payloadBase64)
	request := &pb.ProduceRequest{Payload: payload}

	var sent int64   // Total messages sent
	var errors int64 // Total errors encountered
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(testDurationSec)*time.Second)
	defer cancel()

	start := time.Now()

	// Create a pool of gRPC connections and clients
	conns := make([]*grpc.ClientConn, concurrency)
	clients := make([]pb.MessageQueueClient, concurrency)
	for i := 0; i < concurrency; i++ {
		conn, err := grpc.Dial(gRPCAddress, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		conns[i] = conn
		clients[i] = pb.NewMessageQueueClient(conn)
	}

	// Worker function with async in-flight requests
	worker := func(id int) {
		defer wg.Done()
		client := clients[id]
		var localWg sync.WaitGroup
		sem := make(chan struct{}, inflight) // Semaphore to limit in-flight requests
		for {
			if ctx.Err() != nil || atomic.LoadInt64(&sent) >= totalMessages {
				break
			}
			sem <- struct{}{}
			localWg.Add(1)
			go func() {
				defer func() {
					<-sem
					localWg.Done()
				}()
				// Use a context per request
				reqCtx, cancelReq := context.WithTimeout(ctx, 5*time.Second)
				defer cancelReq()
				_, err := client.Produce(reqCtx, request)
				if err != nil {
					atomic.AddInt64(&errors, 1)
					fmt.Printf("gRPC error: %v\n", err)
				}
				cur := atomic.AddInt64(&sent, 1)
				if cur%50000 == 0 && id == 0 {
					fmt.Printf("Progress: %d messages sent (%s)\n", cur, time.Now().Format(time.RFC3339))
				}
			}()
			// Stop if enough messages sent
			if atomic.LoadInt64(&sent) >= totalMessages {
				break
			}
		}
		localWg.Wait()
	}

	// Start workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go worker(i)
	}

	wg.Wait()
	elapsed := time.Since(start)
	for _, conn := range conns {
		conn.Close()
	}

	fmt.Println("Test complete!")
	fmt.Printf("Total messages sent: %d\n", sent)
	fmt.Printf("Total errors: %d\n", errors)
	fmt.Printf("Elapsed time: %.2f seconds\n", elapsed.Seconds())
	fmt.Printf("Throughput: %.2f messages/sec\n", float64(sent)/elapsed.Seconds())
	fmt.Println("Tip: You can tune concurrency and in-flight requests with the -concurrency and -inflight flags for best results.")
}