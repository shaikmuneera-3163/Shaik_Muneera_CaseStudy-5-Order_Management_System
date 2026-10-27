package tests

import (
	"errors"
	"testing"

	"order-management/services/orders/resilience"
)

func TestRetrySucceedsAfterTemporaryFailures(t *testing.T) {
	calls := 0
	err := resilience.Retry(3, func() error {
		calls++
		if calls < 3 {
			return errors.New("temporary failure")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Retry() returned unexpected error: %v", err)
	}
	if calls != 3 {
		t.Fatalf("operation called %d times, want 3", calls)
	}
}

func TestRetryStopsAfterConfiguredAttempts(t *testing.T) {
	calls := 0
	wantErr := errors.New("service unavailable")
	err := resilience.Retry(2, func() error {
		calls++
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Retry() error = %v, want %v", err, wantErr)
	}
	if calls != 2 {
		t.Fatalf("operation called %d times, want 2", calls)
	}
}

func TestRetryUsesAtLeastOneAttempt(t *testing.T) {
	calls := 0
	_ = resilience.Retry(0, func() error {
		calls++
		return nil
	})
	if calls != 1 {
		t.Fatalf("operation called %d times, want 1", calls)
	}
}
