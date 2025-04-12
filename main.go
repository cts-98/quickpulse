package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"google.golang.org/grpc"
	"quickpulse/mq"
	"quickpulse/quickpulse/proto"
	"quickpulse/server"
	"quickpulse/mqmetrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Parse mode flags from environment variables
	wsMode, _ := strconv.Atoi(os.Getenv("WS_MODE"))
	rpcMode, _ := strconv.Atoi(os.Getenv("RPC_MODE"))

	if wsMode == 1 && rpcMode == 1 {
		log.Fatal("Both WS_MODE and RPC_MODE are set to 1. Please set only one mode.")
	}
	if wsMode != 1 && rpcMode != 1 {
		log.Fatal("Neither WS_MODE nor RPC_MODE is set to 1. Please set one mode.")
	}

	metrics := mqmetrics.NewPrometheusMetrics()
	queue := mq.NewMessageQueue(1000000)
	instrumentedQueue := mqmetrics.NewInstrumentedQueue(queue, metrics)

	// Start Prometheus metrics HTTP server in a goroutine
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("Prometheus metrics server listening on :8080/metrics")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("metrics server error: %v", err)
		}
	}()

	if wsMode == 1 {
		wsServer := server.NewWsServer(instrumentedQueue)
		http.HandleFunc("/ws/publish", wsServer.PublishHandler)
		http.HandleFunc("/ws/consume", wsServer.ConsumeHandler)
		log.Println("WebSocket server listening on :8081 (endpoints: /ws/publish, /ws/consume)")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Fatalf("WebSocket server error: %v", err)
		}
		return
	}

	// Default to gRPC mode if rpcMode == 1
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcSrv := grpc.NewServer()
	proto.RegisterMessageQueueServer(grpcSrv, server.NewGrpcServer(instrumentedQueue))
	reflection.Register(grpcSrv)

	log.Println("gRPC server listening on :50051")
	if err := grpcSrv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}