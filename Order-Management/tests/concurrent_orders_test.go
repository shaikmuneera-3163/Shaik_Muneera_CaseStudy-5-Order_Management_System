package tests

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	ordercontroller "order-management/services/orders/controller"

	"github.com/gin-gonic/gin"
)

// This test checks that the HTTP handler safely processes 100 simultaneous requests
// and rejects invalid payloads. It does NOT replace a full load test of successful order
// creation against Product Service and PostgreSQL; that requires test dependencies.
func TestConcurrentInvalidOrderRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	ctrl := ordercontroller.NewOrderController(nil)
	router.POST("/api/v1/orders", ctrl.Create)

	const requestCount = 100
	var wg sync.WaitGroup
	errs := make(chan error, requestCount)

	for i := 0; i < requestCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			body := []byte(fmt.Sprintf(`{"customerId":%d,"productId":0,"quantity":1}`, i+1))
			req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			if resp.Code != http.StatusBadRequest {
				errs <- fmt.Errorf("request %d: status = %d, want %d", i, resp.Code, http.StatusBadRequest)
			}
		}(i)
	}

	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}
