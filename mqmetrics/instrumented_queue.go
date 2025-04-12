// instrumented_queue.go - Provides InstrumentedQueue, a wrapper for MessageQueue that records metrics.
//
// This file defines InstrumentedQueue, which wraps a MessageQueue and updates
// metrics on each enqueue and dequeue operation. It is used to monitor queue
// activity and performance in real time.

package mqmetrics

import (
	"quickpulse/mq" // MessageQueue implementation
	"time"          // For measuring operation latency
)

// InstrumentedQueue wraps a MessageQueue and updates metrics on each operation.
type InstrumentedQueue struct {
	Queue   *mq.MessageQueue   // Underlying message queue
	Metrics MetricsCollector   // Metrics collector for recording queue stats
}

// NewInstrumentedQueue creates a new InstrumentedQueue with the given queue and metrics collector.
func NewInstrumentedQueue(q *mq.MessageQueue, m MetricsCollector) *InstrumentedQueue {
	return &InstrumentedQueue{
		Queue:   q,
		Metrics: m,
	}
}

// Enqueue adds a message to the queue and updates metrics for enqueue count, queue depth, and latency.
func (iq *InstrumentedQueue) Enqueue(msg []byte) error {
	start := time.Now()
	err := iq.Queue.Enqueue(msg)
	if err == nil {
		iq.Metrics.IncEnqueue() // Increment enqueue counter
		iq.Metrics.SetQueueDepth(int64(iq.Queue.Len())) // Update queue depth metric
		iq.Metrics.ObserveEnqueueLatency(time.Since(start)) // Record enqueue latency
	}
	return err
}

// Dequeue removes a message from the queue and updates metrics for dequeue count and queue depth.
func (iq *InstrumentedQueue) Dequeue() ([]byte, error) {
	msg, err := iq.Queue.Dequeue()
	if err == nil {
		iq.Metrics.IncDequeue() // Increment dequeue counter
		iq.Metrics.SetQueueDepth(int64(iq.Queue.Len())) // Update queue depth metric
	}
	return msg, err
}

// Len returns the current number of messages in the queue.
func (iq *InstrumentedQueue) Len() uint64 {
	return iq.Queue.Len()
}