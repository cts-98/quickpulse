// message.go - Defines the Message type used in the message queue system.
//
// This file provides the Message struct, which encapsulates a message's unique
// identifier and its payload, along with methods for creating and accessing messages.

package mq

// Message represents a message in the queue, consisting of an ID and a payload.
type Message struct {
	id      string // Unique identifier for the message
	payload []byte // Message payload (arbitrary binary data)
}

// NewMessage creates a new Message with the given id and payload.
// Returns a pointer to the created Message.
func NewMessage(id string, payload []byte) *Message {
	return &Message{
		id:      id,
		payload: payload,
	}
}

// GetID returns the unique identifier of the message.
func (m *Message) GetID() string {
	return m.id
}

// GetPayload returns the payload of the message as a byte slice.
func (m *Message) GetPayload() []byte {
	return m.payload
}