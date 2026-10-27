## Run from the Order-Management project root

```bash
go test ./...
go test -race ./...
```

## What these tests cover

- `order_api_test.go`: verifies that order API handlers reject malformed or invalid input.
- `resilience_test.go`: verifies the existing bounded `resilience.Retry` helper retries and stops at the configured attempt count.
- `concurrent_orders_test.go`: sends 100 simultaneous invalid create-order requests to exercise concurrent HTTP handler validation.

## Important limitation

The concurrent test does **not** prove that 100 successful orders can be created concurrently. A full successful-order concurrency/integration test must run with a dedicated test PostgreSQL database and a mock or test Product Service. The API tests here intentionally avoid database/network dependencies.
