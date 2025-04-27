package core

import (
	"errors"
)

var ErrNoInstancesAvailable = errors.New("no instance available")
var ErrNoHealthyInstancesAvailable = errors.New("no healthy instance available")

type Balancer interface {
	Select() (*Instance, error)
	Update(instances []*Instance)
	Remove(instance *Instance)
	Add(instance *Instance)
	Len() int
	Reset()
	AggregateMetrics() Metrics
}
