// metrics.go - Provides metrics collection for message queue operations.
//
// This file defines the MetricsCollector interface for queue metrics and
// implements DefaultMetrics, which uses atomic counters to track enqueue/dequeue
// operations, queue depth, and throughput.

package mqmetrics

import (
	"sync/atomic" // For atomic operations on counters
	"time"        // For time-based throughput calculations
)

// MetricsCollector defines the interface for collecting queue metrics.
// Implementations may track enqueue/dequeue counts, queue depth, throughput, and latency.
type MetricsCollector interface {
	IncEnqueue()                              // Increment the enqueue counter
	IncDequeue()                              // Increment the dequeue counter
	GetThroughput() (enqueuePerSec, dequeuePerSec int64) // Get enqueue/dequeue throughput per second
	GetQueueDepth() int64                     // Get the current queue depth
	SetQueueDepth(depth int64)                // Set the current queue depth
	ObserveEnqueueLatency(d time.Duration)    // Observe enqueue latency (optional)
}

// DefaultMetrics implements MetricsCollector with atomic counters for thread safety.
type DefaultMetrics struct {
	enqueueCount   int64 // Total number of enqueues
	dequeueCount   int64 // Total number of dequeues
	lastEnqueue    int64 // Enqueue count at last throughput update
	lastDequeue    int64 // Dequeue count at last throughput update
	queueDepth     int64 // Current queue depth
}

// NewDefaultMetrics creates a new DefaultMetrics instance and starts the throughput updater goroutine.
func NewDefaultMetrics() *DefaultMetrics {
	m := &DefaultMetrics{}
	go m.runThroughputUpdater()
	return m
}

// IncEnqueue atomically increments the enqueue counter.
func (m *DefaultMetrics) IncEnqueue() {
	atomic.AddInt64(&m.enqueueCount, 1)
}

// IncDequeue atomically increments the dequeue counter.
func (m *DefaultMetrics) IncDequeue() {
	atomic.AddInt64(&m.dequeueCount, 1)
}

// SetQueueDepth atomically sets the current queue depth.
func (m *DefaultMetrics) SetQueueDepth(depth int64) {
	atomic.StoreInt64(&m.queueDepth, depth)
}

// GetQueueDepth atomically retrieves the current queue depth.
func (m *DefaultMetrics) GetQueueDepth() int64 {
	return atomic.LoadInt64(&m.queueDepth)
}

// GetThroughput returns the number of enqueues and dequeues per second since the last update.
func (m *DefaultMetrics) GetThroughput() (int64, int64) {
	enqueue := atomic.LoadInt64(&m.enqueueCount)
	dequeue := atomic.LoadInt64(&m.dequeueCount)
	lastEnqueue := atomic.LoadInt64(&m.lastEnqueue)
	lastDequeue := atomic.LoadInt64(&m.lastDequeue)
	return enqueue - lastEnqueue, dequeue - lastDequeue
}

// runThroughputUpdater updates the lastEnqueue/lastDequeue counters every second
// to enable throughput calculation.
func (m *DefaultMetrics) runThroughputUpdater() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		atomic.StoreInt64(&m.lastEnqueue, atomic.LoadInt64(&m.enqueueCount))
		atomic.StoreInt64(&m.lastDequeue, atomic.LoadInt64(&m.dequeueCount))
	}
}

// ObserveEnqueueLatency is a no-op for DefaultMetrics, but can be implemented in other collectors.
func (m *DefaultMetrics) ObserveEnqueueLatency(d time.Duration) {}