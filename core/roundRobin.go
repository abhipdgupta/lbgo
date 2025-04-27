package core

import "sync"

type RoundRobinBalancer struct {
	instances            []*Instance
	lastSelectedInstance int
	mu                   sync.Mutex
}

var _ Balancer = (*RoundRobinBalancer)(nil)

func NewRoundRobinBalancer() *RoundRobinBalancer {
	return &RoundRobinBalancer{}
}

func (b *RoundRobinBalancer) Select() (*Instance, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.Len() == 0 {
		return nil, ErrNoInstancesAvailable
	}

	// Start from the next instance after lastSelectedInstance
	startIndex := (b.lastSelectedInstance + 1) % b.Len()

	// Try to find the next healthy instance starting from the next instance
	for i := startIndex; i != b.lastSelectedInstance; i = (i + 1) % b.Len() {
		instance := b.instances[i]
		if instance.isAlive {
			b.lastSelectedInstance = i
			return instance, nil
		}
	}

	// If no healthy instance is found, fallback to the last selected instance
	lastInstance := b.instances[b.lastSelectedInstance]
	if lastInstance.isAlive {
		return lastInstance, nil
	}

	return nil, ErrNoHealthyInstancesAvailable
}

func (b *RoundRobinBalancer) Update(instances []*Instance) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.instances = instances
}

func (b *RoundRobinBalancer) Remove(instance *Instance) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i, inst := range b.instances {
		if inst == instance {
			b.instances = append(b.instances[:i], b.instances[i+1:]...)
			break
		}
	}
}

func (b *RoundRobinBalancer) Add(instance *Instance) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.instances = append(b.instances, instance)
}

func (b *RoundRobinBalancer) Len() int {
	return len(b.instances)
}

func (b *RoundRobinBalancer) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.instances = nil
}

func (b *RoundRobinBalancer) AggregateMetrics() Metrics {
	b.mu.Lock()
	defer b.mu.Unlock()

	aggreatedMetrics := Metrics{}

	for _, instance := range b.instances {

		instancesMetrics := instance.SnapshotMetrics()

		agrregatedInstanceMetric := Metrics{}

		for _, metric := range instancesMetrics {

			agrregatedInstanceMetric.ClientErrors4xx += metric.ClientErrors4xx
			agrregatedInstanceMetric.ServerErrors5xx += metric.ServerErrors5xx
			agrregatedInstanceMetric.Success2xx += metric.Success2xx
			agrregatedInstanceMetric.TotalLatency += metric.TotalLatency
			agrregatedInstanceMetric.TotalRequests += metric.TotalRequests
		}

		agrregatedInstanceMetric.CalculateAndUpdateAverageLatency()
		agrregatedInstanceMetric.CalculateAndUpdateRates()

		aggreatedMetrics.ClientErrors4xx += agrregatedInstanceMetric.ClientErrors4xx
		aggreatedMetrics.ServerErrors5xx += agrregatedInstanceMetric.ServerErrors5xx
		aggreatedMetrics.Success2xx += agrregatedInstanceMetric.Success2xx
		aggreatedMetrics.TotalLatency += agrregatedInstanceMetric.TotalLatency
		aggreatedMetrics.TotalRequests += agrregatedInstanceMetric.TotalRequests

	}

	aggreatedMetrics.CalculateAndUpdateAverageLatency()
	aggreatedMetrics.CalculateAndUpdateRates()

	return aggreatedMetrics
}
