# LBGO (Load Balancer in Golang)

Prototyping a simple load balancer in Go, with a metrics aggregator and the flexibility to use various load balancing algorithms. Currently, it supports only the **Round Robin** algorithm.

## Round Robin Algorithm
1. The algorithm selects an instance in a circular manner.
2. After the first instance is selected, the second instance is chosen, and so on. If the last selected instance was the first one, the algorithm will select the second one next.
