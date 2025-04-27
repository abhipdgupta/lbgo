package core

type HealthStatus string

const (
	HealthUnknown   HealthStatus = "Unknown"
	HealthHealthy   HealthStatus = "Healthy"
	HealthUnhealthy HealthStatus = "Unhealthy"
	HealthTimeout   HealthStatus = "Timeout"
)
