// prometheus_metrics.go - Prometheus-based metrics collection for the message queue.
//
// This file defines PrometheusMetrics, which implements the MetricsCollector interface
// and exposes queue metrics (enqueue/dequeue counts, queue depth, throughput, latency)
// to Prometheus for monitoring and alerting.

package mqmetrics

import (
	"sync/atomic" // For atomic operations on counters
	"github.com/prometheus/client_golang/prometheus" // Prometheus client library
	"time"        // For time-based throughput calculations
)

// PrometheusMetrics collects and exposes queue metrics to Prometheus.
type PrometheusMetrics struct {
	EnqueueCounter    prometheus.Counter // Total number of enqueued messages
	DequeueCounter    prometheus.Counter // Total number of dequeued messages
	QueueDepth        prometheus.Gauge   // Current queue depth
	EnqueueThroughput prometheus.Gauge   // Enqueue throughput (messages/sec)
	DequeueThroughput prometheus.Gauge   // Dequeue throughput (messages/sec)
	EnqueueLatency    prometheus.Histogram // Histogram of enqueue latencies

	enqueueCount     int64 // Internal counter for enqueues (for throughput)
	dequeueCount     int64 // Internal counter for dequeues (for throughput)
	lastEnqueueCount int64 // Last recorded enqueue count (for throughput)
	lastDequeueCount int64 // Last recorded dequeue count (for throughput)
}

// NewPrometheusMetrics creates and registers Prometheus metrics, and starts the throughput updater goroutine.
func NewPrometheusMetrics() *PrometheusMetrics {
	m := &PrometheusMetrics{
		EnqueueCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "unnamedmq_enqueue_total",
			Help: "Total number of enqueued messages",
		}),
		DequeueCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "unnamedmq_dequeue_total",
			Help: "Total number of dequeued messages",
		}),
		QueueDepth: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "unnamedmq_queue_depth",
			Help: "Current queue depth",
		}),
		EnqueueThroughput: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "unnamedmq_enqueue_throughput",
			Help: "Enqueue throughput (messages per second)",
		}),
		DequeueThroughput: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "unnamedmq_dequeue_throughput",
			Help: "Dequeue throughput (messages per second)",
		}),
		EnqueueLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "unnamedmq_enqueue_latency_seconds",
			Help:    "Histogram of enqueue latencies in seconds",
			Buckets: prometheus.ExponentialBuckets(0.0001, 2, 16), // 100us to ~3s
		}),
	}
	// Register all metrics with Prometheus
	prometheus.MustRegister(
		m.EnqueueCounter, m.DequeueCounter, m.QueueDepth,
		m.EnqueueThroughput, m.DequeueThroughput, m.EnqueueLatency,
	)
	// Start a goroutine to update throughput metrics every second
	go m.runThroughputUpdater()
	return m
}

// runThroughputUpdater updates the enqueue/dequeue throughput metrics every second.
func (m *PrometheusMetrics) runThroughputUpdater() {
	for {
		time.Sleep(time.Second)
		enqueue := atomic.LoadInt64(&m.enqueueCount)
		dequeue := atomic.LoadInt64(&m.dequeueCount)
		m.EnqueueThroughput.Set(float64(enqueue - m.lastEnqueueCount))
		m.DequeueThroughput.Set(float64(dequeue - m.lastDequeueCount))
		m.lastEnqueueCount = enqueue
		m.lastDequeueCount = dequeue
	}
}

// IncEnqueue increments the enqueue counter and updates the internal count.
func (m *PrometheusMetrics) IncEnqueue() {
	m.EnqueueCounter.Inc()
	atomic.AddInt64(&m.enqueueCount, 1)
}

// IncDequeue increments the dequeue counter and updates the internal count.
func (m *PrometheusMetrics) IncDequeue() {
	m.DequeueCounter.Inc()
	atomic.AddInt64(&m.dequeueCount, 1)
}

// SetQueueDepth sets the current queue depth gauge.
func (m *PrometheusMetrics) SetQueueDepth(depth int64) {
	m.QueueDepth.Set(float64(depth))
}

// The following are no-ops for PrometheusMetrics, but required for interface compatibility.
func (m *PrometheusMetrics) GetThroughput() (int64, int64) { return 0, 0 }
/*
GetQueueDepth returns the current queue depth as required by the MetricsCollector interface.
For PrometheusMetrics, this is a no-op because Prometheus scrapes the value directly from the gauge.
*/
func (m *PrometheusMetrics) GetQueueDepth() int64 {
	return 0
}

// ObserveEnqueueLatency records the enqueue latency in seconds in the histogram.
func (m *PrometheusMetrics) ObserveEnqueueLatency(d time.Duration) {
	m.EnqueueLatency.Observe(d.Seconds())
}