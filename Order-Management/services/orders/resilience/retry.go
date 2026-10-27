package resilience

import (
	"log"
	"time"
)

func Retry(attempts int, operation func() error) error {
	if attempts < 1 {
		attempts = 1
	}

	var err error

	for i := 0; i < attempts; i++ {
		err = operation()
		if err == nil {
			if i > 0 {
				log.Printf("[RETRY] Request succeeded on attempt %d/%d", i+1, attempts)
			}
			return nil
		}

		log.Printf("[RETRY] Attempt %d/%d failed: %v", i+1, attempts, err)

		if i < attempts-1 {
			time.Sleep(time.Duration(i+1) * 200 * time.Millisecond)
		}
	}

	log.Printf("[RETRY] All %d attempts failed: %v", attempts, err)
	return err
}
