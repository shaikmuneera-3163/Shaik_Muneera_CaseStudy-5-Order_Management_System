package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	ordercontroller "order-management/services/orders/controller"

	"github.com/gin-gonic/gin"
)

// newOrderTestRouter exercises request validation without connecting to PostgreSQL
// or calling the Product Service. Invalid requests must be rejected before the service is used.
func newOrderTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	ctrl := ordercontroller.NewOrderController(nil)
	router.POST("/api/v1/orders", ctrl.Create)
	router.GET("/api/v1/orders/:id", ctrl.GetByID)
	router.PUT("/api/v1/orders/:id/status", ctrl.UpdateStatus)
	return router
}

func TestCreateOrderRejectsInvalidRequest(t *testing.T) {
	router := newOrderTestRouter()
	cases := []struct {
		name string
		body string
	}{
		{"invalid JSON", `{"customerId":`},
		{"missing customer ID", `{"productId":501,"quantity":2}`},
		{"zero product ID", `{"customerId":101,"productId":0,"quantity":2}`},
		{"zero quantity", `{"customerId":101,"productId":501,"quantity":0}`},
		{"negative quantity", `{"customerId":101,"productId":501,"quantity":-1}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			if resp.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body=%s", resp.Code, http.StatusBadRequest, resp.Body.String())
			}
		})
	}
}

func TestGetOrderRejectsInvalidID(t *testing.T) {
	router := newOrderTestRouter()
	for _, id := range []string{"abc", "0", "-1"} {
		t.Run(id, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/"+id, nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			if resp.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body=%s", resp.Code, http.StatusBadRequest, resp.Body.String())
			}
		})
	}
}

func TestUpdateStatusRejectsInvalidOrderID(t *testing.T) {
	router := newOrderTestRouter()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/orders/abc/status", bytes.NewBufferString(`{"status":"CONFIRMED"}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", resp.Code, http.StatusBadRequest, resp.Body.String())
	}
}
