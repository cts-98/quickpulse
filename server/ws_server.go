// ws_server.go - WebSocket server for publishing and consuming messages.
//
// This file defines WsServer, which provides WebSocket endpoints for clients to
// publish messages to and consume messages from a message queue. It uses the
// gorilla/websocket package for WebSocket support.

package server

import (
	"log"      // For logging errors and events
	"net/http" // For HTTP server and handlers

	"github.com/gorilla/websocket" // WebSocket support
	"quickpulse/mq"                // Message queue interface
)

// WsServer provides WebSocket endpoints for publishing and consuming messages.
type WsServer struct {
	Queue mq.Queue // Underlying message queue
}

// NewWsServer creates a new WsServer with the given queue.
func NewWsServer(queue mq.Queue) *WsServer {
	return &WsServer{Queue: queue}
}

// upgrader is used to upgrade HTTP connections to WebSocket connections.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Allow all origins
}

// PublishHandler handles WebSocket connections for publishing messages to the queue.
// Each message received from the client is enqueued, and an "ok" or "error" response is sent back.
func (s *WsServer) PublishHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		// Read a message from the client
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		// Enqueue the message
		err = s.Queue.Enqueue(msg)
		resp := "ok"
		if err != nil {
			resp = "error: " + err.Error()
		}
		// Send response to the client
		if err := conn.WriteMessage(websocket.TextMessage, []byte(resp)); err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}

// ConsumeHandler handles WebSocket connections for consuming messages from the queue.
// The client sends a request (any message) to receive the next message from the queue.
func (s *WsServer) ConsumeHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		// Wait for client to request a message (could be any message, e.g., "next")
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		// Dequeue the next message from the queue
		msg, err := s.Queue.Dequeue()
		if err != nil {
			// Send error response if queue is empty
			if err := conn.WriteMessage(websocket.TextMessage, []byte("error: "+err.Error())); err != nil {
				log.Println("Write error:", err)
				break
			}
			continue
		}
		// Send the message to the client as a binary WebSocket message
		if err := conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}