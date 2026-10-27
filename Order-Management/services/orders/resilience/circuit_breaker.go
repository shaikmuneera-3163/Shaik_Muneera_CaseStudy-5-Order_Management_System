package resilience

import (
	"log"
	"time"

	"github.com/sony/gobreaker"
)

func NewCircuitBreaker(name string) *gobreaker.CircuitBreaker {
	return gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        name,
		MaxRequests: 1,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Printf(
				"[CIRCUIT BREAKER] %s state changed: %s -> %s",
				name, from, to,
			)
		},
	})
}
