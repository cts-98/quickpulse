// grpc_server.go - gRPC server implementations for the message queue.
//
// This file defines GrpcUnaryServer and GrpcStreamServer, which implement the
// gRPC service for the message queue in unary and streaming modes, respectively.
// The servers use a shared queue and provide methods for producing and consuming
// messages via gRPC.

package server

import (
	"context" // For gRPC context

	"quickpulse/mq"    // Message queue interface
	"quickpulse/proto" // gRPC protobuf definitions

	"google.golang.org/grpc/codes"  // gRPC error codes
	"google.golang.org/grpc/status" // gRPC status errors
)

// GrpcUnaryServer implements the gRPC MessageQueue service in unary mode.
type GrpcUnaryServer struct {
	proto.UnimplementedMessageQueueServer // Embeds unimplemented methods for forward compatibility
	Queue mq.Queue                        // Underlying message queue
}

// GrpcStreamServer implements the gRPC MessageQueue service in streaming mode.
type GrpcStreamServer struct {
	proto.UnimplementedMessageQueueServer // Embeds unimplemented methods for forward compatibility
	Queue mq.Queue                        // Underlying message queue
}

// NewGrpcUnaryServer creates a new GrpcUnaryServer with the given queue.
func NewGrpcUnaryServer(queue mq.Queue) *GrpcUnaryServer {
	return &GrpcUnaryServer{Queue: queue}
}

// NewGrpcStreamServer creates a new GrpcStreamServer with the given queue.
func NewGrpcStreamServer(queue mq.Queue) *GrpcStreamServer {
	return &GrpcStreamServer{Queue: queue}
}

// Produce handles unary gRPC requests to enqueue a message.
func (s *GrpcUnaryServer) Produce(ctx context.Context, req *proto.ProduceRequest) (*proto.ProduceResponse, error) {
	err := s.Queue.Enqueue(req.Payload)
	if err != nil {
		return &proto.ProduceResponse{Success: false, Error: err.Error()}, nil
	}
	return &proto.ProduceResponse{Success: true}, nil
}

// Consume handles unary gRPC requests to dequeue a message.
func (s *GrpcUnaryServer) Consume(ctx context.Context, req *proto.ConsumeRequest) (*proto.ConsumeResponse, error) {
	msg, err := s.Queue.Dequeue()
	if err != nil {
		return &proto.ConsumeResponse{Payload: nil, Error: err.Error()}, nil
	}
	return &proto.ConsumeResponse{Payload: msg}, nil
}

// StreamMessages is not implemented in unary mode and returns an error.
func (s *GrpcUnaryServer) StreamMessages(stream proto.MessageQueue_StreamMessagesServer) error {
	return status.Errorf(codes.Unimplemented, "StreamMessages is not implemented in unary mode")
}

// Produce is not implemented in streaming mode and returns an error.
func (s *GrpcStreamServer) Produce(ctx context.Context, req *proto.ProduceRequest) (*proto.ProduceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "Produce is not implemented in streaming mode")
}

// Consume is not implemented in streaming mode and returns an error.
func (s *GrpcStreamServer) Consume(ctx context.Context, req *proto.ConsumeRequest) (*proto.ConsumeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "Consume is not implemented in streaming mode")
}

// StreamMessages handles bidirectional streaming for producing and consuming messages.
func (s *GrpcStreamServer) StreamMessages(stream proto.MessageQueue_StreamMessagesServer) error {
	for {
		// Receive a message from the client
		in, err := stream.Recv()
		if err != nil {
			// End of stream or error
			return err
		}
		// Enqueue the received payload if present
		if in.Payload != nil {
			_ = s.Queue.Enqueue(in.Payload)
		}
		// Dequeue a message to send back to the client
		msg, err := s.Queue.Dequeue()
		resp := &proto.StreamMessage{}
		if err != nil {
			resp.Error = err.Error()
			resp.Payload = []byte{}
		} else if msg == nil {
			resp.Payload = []byte{}
		} else {
			resp.Payload = msg
		}
		// Send the response to the client
		if err := stream.Send(resp); err != nil {
			return err
		}
	}
}