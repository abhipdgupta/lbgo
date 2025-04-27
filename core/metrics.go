package core

import (
	"sync"
	"time"
)

type Metrics struct {
	TotalLatency    time.Duration `json:"totalLatency"`    // Total accumulated latency
	TotalRequests   int64         `json:"totalRequests"`   // Total number of requests
	Success2xx      int64         `json:"success2xx"`      // Number of successful (2xx) responses
	ClientErrors4xx int64         `json:"clientErrors4xx"` // Number of client error (4xx) responses
	ServerErrors5xx int64         `json:"serverErrors5xx"` // Number of server error (5xx) responses
	ErrorRate       float64       `json:"errorRate"`       // Percentage of failed requests
	SuccessRate     float64       `json:"successRate"`     // Percentage of successful requests
	AverageLatency  time.Duration `json:"averageLatency"`  // Average latency across all requests
}

type RollingMetrics struct {
	mu             sync.Mutex
	currentWindow  *Metrics
	windowDuration time.Duration
	windowStart    time.Time
	metricsBuffer  []Metrics
	maxBufferSize  int
}

func (m *Metrics) CalculateAndUpdateRates() {
	if m.TotalRequests == 0 {
		m.ErrorRate = 0
		m.SuccessRate = 0
		return
	}
	m.ErrorRate = float64(m.ClientErrors4xx+m.ServerErrors5xx) / float64(m.TotalRequests) * 100
	m.SuccessRate = float64(m.Success2xx) / float64(m.TotalRequests) * 100
}

func (m *Metrics) CalculateAndUpdateAverageLatency() {
	if m.TotalRequests == 0 {
		m.AverageLatency = 0
		return
	}
	m.AverageLatency = m.TotalLatency / time.Duration(m.TotalRequests)
}

func (m *Metrics) Update(statusCode int, latency time.Duration) {
	m.TotalRequests++
	m.TotalLatency += latency

	switch {
	case statusCode >= 200 && statusCode < 300:
		m.Success2xx++
	case statusCode >= 400 && statusCode < 500:
		m.ClientErrors4xx++
	case statusCode >= 500:
		m.ServerErrors5xx++
	}
}

func NewRollingMetrics(windowDuration time.Duration, maxBufferSize int) *RollingMetrics {
	metric := &RollingMetrics{
		windowDuration: windowDuration,
		windowStart:    time.Now(),
		metricsBuffer:  make([]Metrics, 0),
		maxBufferSize:  maxBufferSize,
		currentWindow:  &Metrics{},
	}
	go metric.resetBufferPeriodically()

	return metric
}

func (r *RollingMetrics) UpdateMetrics(statusCode int, latency time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.currentWindow.Update(statusCode, latency)
}

func (r *RollingMetrics) FlushBuffer() []Metrics {
	r.mu.Lock()
	defer r.mu.Unlock()

	metrics := r.metricsBuffer
	r.metricsBuffer = make([]Metrics, 0)
	return metrics
}

func (r *RollingMetrics) resetBufferPeriodically() {
	ticker := time.NewTicker(r.windowDuration)
	defer ticker.Stop()

	for range ticker.C {
		r.mu.Lock()

		// Finalize the current window metrics (if expired) and reset it
		elapsed := time.Since(r.windowStart)
		if elapsed >= r.windowDuration {
			r.currentWindow.CalculateAndUpdateRates()
			r.currentWindow.CalculateAndUpdateAverageLatency()

			r.metricsBuffer = append(r.metricsBuffer, *r.currentWindow)

			if len(r.metricsBuffer) > r.maxBufferSize {
				r.metricsBuffer = r.metricsBuffer[1:]
			}

			// Reset for the next window
			r.currentWindow = &Metrics{}
			r.windowStart = time.Now()
		}

		r.mu.Unlock()
	}
}
