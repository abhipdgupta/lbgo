package core

import (
	"sync"
	"time"
)

// Instance represents a backend server tracked by the load balancer
type Instance struct {
	ID          string       `json:"id"`           // Unique identifier for the instance
	URL         string       `json:"url"`          // Backend URL
	Health      HealthStatus `json:"health"`       // Current health status
	LastChecked time.Time    `json:"last_checked"` // Last time health was checked
	Uptime      time.Time    `json:"uptime"`       // Time when instance was registered
	isAlive     bool

	Metrics *RollingMetrics `json:"-"`

	mu *sync.Mutex
}

func NewInstance(id, url string, windowDuration time.Duration, maxBufferSize int) *Instance {
	instance := &Instance{
		ID:          id,
		URL:         url,
		Health:      HealthUnknown,
		LastChecked: time.Now(),
		Uptime:      time.Now(),
		isAlive:     true,
		Metrics:     NewRollingMetrics(windowDuration, maxBufferSize),
		mu:          &sync.Mutex{},
	}

	// TODO: go routien for periodic health checks

	return instance

}

func (i *Instance) UpdateHealth(health HealthStatus) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.Health = health
	i.LastChecked = time.Now()
	i.isAlive = (health == HealthHealthy)
}

func (i *Instance) IsAlive() bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.isAlive
}

func (i *Instance) UpdateMetrics(statusCode int, latency time.Duration) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.Metrics.UpdateMetrics(statusCode, latency)
}

func (i *Instance) SnapshotMetrics() []Metrics {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.Metrics.FlushBuffer()
}
