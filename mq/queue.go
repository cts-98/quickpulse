// queue.go - High-performance, lock-free message queue implementation.
//
// This file defines the Queue interface and provides a MessageQueue implementation
// using a fixed-size ring buffer and atomic operations for ultra-low latency and
// minimal locking. The queue is designed for concurrent producers and consumers.

package mq

import (
	"errors"   // For error handling
	"log"      // For logging errors
	"sync/atomic" // For atomic operations on queue pointers
)

// Queue defines the interface for a message queue supporting basic operations.
type Queue interface {
	Enqueue(msg []byte) error   // Add a message to the queue
	Dequeue() ([]byte, error)   // Remove and return the next message
	Len() uint64                // Get the current number of messages in the queue
}

// MessageQueue is a high-performance, ultra low latency queue for binary messages.
// It uses a fixed-size ring buffer and atomic operations for minimal locking.
type MessageQueue struct {
	buffer     [][]byte // The ring buffer holding messages
	capacity   uint64   // Maximum number of messages the queue can hold
	head       uint64   // Next position to read (consumer index)
	tail       uint64   // Next position to write (producer index)
	_          [56]byte // Padding to avoid false sharing (cache line alignment)
}

// NewMessageQueue creates a new MessageQueue with the given capacity.
func NewMessageQueue(capacity uint64) *MessageQueue {
	return &MessageQueue{
		buffer:   make([][]byte, capacity),
		capacity: capacity,
	}
}

// Enqueue adds a binary message to the queue.
// Returns an error if the queue is full.
// Uses atomic operations to ensure thread safety for concurrent producers.
func (q *MessageQueue) Enqueue(msg []byte) error {
	for {
		head := atomic.LoadUint64(&q.head)
		tail := atomic.LoadUint64(&q.tail)
		// Check if the queue is full
		if (tail-head) >= q.capacity {
			log.Println("ERROR: MessageQueue capacity breached. Cannot enqueue new message.")
			return errors.New("queue is full")
		}
		pos := tail % q.capacity
		// Atomically claim the next slot for writing
		if atomic.CompareAndSwapUint64(&q.tail, tail, tail+1) {
			q.buffer[pos] = msg
			return nil
		}
		// If CAS fails, another producer won the race; retry
	}
}

// Dequeue removes and returns the next binary message from the queue.
// Returns nil and an error if the queue is empty.
// Uses atomic operations to ensure thread safety for concurrent consumers.
func (q *MessageQueue) Dequeue() ([]byte, error) {
	for {
		head := atomic.LoadUint64(&q.head)
		tail := atomic.LoadUint64(&q.tail)
		// Check if the queue is empty
		if head == tail {
			return nil, errors.New("queue is empty")
		}
		pos := head % q.capacity
		msg := q.buffer[pos]
		// Atomically claim the next slot for reading
		if atomic.CompareAndSwapUint64(&q.head, head, head+1) {
			q.buffer[pos] = nil // Avoid memory leak by clearing the slot
			return msg, nil
		}
		// If CAS fails, another consumer won the race; retry
	}
}

// Len returns the number of messages currently in the queue.
func (q *MessageQueue) Len() uint64 {
	return atomic.LoadUint64(&q.tail) - atomic.LoadUint64(&q.head)
}