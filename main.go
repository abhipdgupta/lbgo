package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/abhipdgupta/lbgo/core"
)

func main() {

	instance1 := core.NewInstance("instance-1", "http://localhost:8081", time.Second*2, 5)
	instance2 := core.NewInstance("instance-2", "http://localhost:8082", time.Second*2, 5)
	instance3 := core.NewInstance("instance-3", "http://localhost:8083", time.Second*2, 5)

	// can be replace with other load balancing  algorithms, it is just required to follow Balancer interface
	lb := core.NewRoundRobinBalancer()

	lb.Add(instance1)
	lb.Add(instance2)
	lb.Add(instance3)

	var wg sync.WaitGroup

	var aggWg sync.WaitGroup

	aggWg.Add(1)
	go aggregateMetricsPeriodically(lb, &aggWg)

	for i := 0; i < 1000; i++ {

		time.Sleep(20 * time.Millisecond)

		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			instance, err := lb.Select()
			if err != nil {
				panic(err)
			}

			// Randomly decide if the request should succeed or fail
			statusCode := generateRandomStatusCode()

			// Randomly generate latency (between 100ms and 500ms)
			latency := time.Duration(rand.Intn(400)+100) * time.Millisecond

			// Simulate the request and update metrics
			simulateRequest(instance, statusCode, latency)

			fmt.Printf("Request sent to Instance %s with status code %d and latency %s\n", instance.ID, statusCode, latency)
		}(i)
	}

	wg.Wait()

	aggWg.Wait()
}

func simulateRequest(instance *core.Instance, statusCode int, latency time.Duration) {
	instance.UpdateMetrics(statusCode, latency)
}

func generateRandomStatusCode() int {
	statusCode := rand.Intn(100)

	if statusCode < 80 {
		// 80% chance for a success (2xx status codes)
		return 200 + rand.Intn(100) // Random 2xx code (200-299)
	} else if statusCode < 90 {
		// 10% chance for client error (4xx status codes)
		return 400 + rand.Intn(100) // Random 4xx code (400-499)
	} else {
		// 10% chance for server error (5xx status codes)
		return 500 + rand.Intn(100) // Random 5xx code (500-599)
	}
}

// aggregateMetricsPeriodically aggregates metrics from instances periodically
func aggregateMetricsPeriodically(lb *core.RoundRobinBalancer, aggWg *sync.WaitGroup) {
	defer aggWg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		aggregatedMetrics := lb.AggregateMetrics()
		fmt.Printf("Aggregated Metrics: %+v\n", aggregatedMetrics)
	}
}
