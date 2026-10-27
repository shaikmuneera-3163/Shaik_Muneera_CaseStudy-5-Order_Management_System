package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/sony/gobreaker"
	"order-management/services/orders/resilience"
	"order-management/services/products/model"
	"order-management/shared/discovery"
)

type ProductClient struct {
	BaseURL   string
	Discovery *discovery.ConsulClient
	Client    *http.Client
	Breaker   *gobreaker.CircuitBreaker
	Retries   int
}

func NewProductClient(url string) *ProductClient {
	timeout, retries := resilienceSettings()
	return &ProductClient{BaseURL: url, Client: &http.Client{Timeout: timeout}, Breaker: resilience.NewCircuitBreaker("product-service"), Retries: retries}
}
func NewProductClientWithDiscovery(d *discovery.ConsulClient) *ProductClient {
	timeout, retries := resilienceSettings()
	return &ProductClient{Discovery: d, Client: &http.Client{Timeout: timeout}, Breaker: resilience.NewCircuitBreaker("product-service"), Retries: retries}
}

// resilienceSettings reads safe local defaults from environment variables.
func resilienceSettings() (time.Duration, int) {
	timeoutSeconds, err := strconv.Atoi(os.Getenv("HTTP_TIMEOUT_SECONDS"))
	if err != nil || timeoutSeconds < 1 || timeoutSeconds > 30 {
		timeoutSeconds = 5
	}
	retries, err := strconv.Atoi(os.Getenv("RETRY_ATTEMPTS"))
	if err != nil || retries < 1 || retries > 5 {
		retries = 3
	}
	return time.Duration(timeoutSeconds) * time.Second, retries
}
func (p *ProductClient) baseURL() (string, error) {
	if p.Discovery != nil {
		return p.Discovery.GetServiceURL("products")
	}
	if p.BaseURL == "" {
		return "", fmt.Errorf("product service URL is empty")
	}
	return p.BaseURL, nil
}

// GetByID retries the safe GET request a limited number of times.
// The context stops both HTTP calls and backoff waits when the request is cancelled.
func (p *ProductClient) GetByID(ctx context.Context, id uint) (*model.Product, error) {
	var product *model.Product
	var lastErr error

	attempts := p.Retries
	if attempts < 1 || attempts > 5 {
		attempts = 3
	}
	for attempt := 0; attempt < attempts; attempt++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		call := func() (interface{}, error) {
			return p.getByIDOnce(ctx, id)
		}
		var result interface{}
		var err error
		if p.Breaker != nil {
			result, err = p.Breaker.Execute(call)
		} else {
			result, err = call()
		}
		if err == nil {
			product = result.(*model.Product)
			return product, nil
		}
		lastErr = err

		if attempt < attempts-1 {
			// Exponential backoff: 200 ms, then 400 ms, and so on.
			delay := time.Duration(200*(1<<attempt)) * time.Millisecond
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
			}
		}
	}
	return nil, lastErr
}

func (p *ProductClient) getByIDOnce(ctx context.Context, id uint) (*model.Product, error) {
	base, err := p.baseURL()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/products/%d", base, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product service returned status %d", resp.StatusCode)
	}
	var result struct {
		Data model.Product `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// ChangeStock is a single, non-retried POST because repeating a stock reservation
// could subtract the same quantity twice if a network response is lost.
func (p *ProductClient) ChangeStock(ctx context.Context, id uint, quantity int, action string) error {
	// Stock changes are mutations, so they use a circuit breaker but are not retried:
	// retrying a reservation after a lost response could subtract stock twice.
	operation := func() (interface{}, error) {
		return nil, p.changeStockOnce(ctx, id, quantity, action)
	}
	if p.Breaker != nil {
		_, err := p.Breaker.Execute(operation)
		return err
	}
	return p.changeStockOnce(ctx, id, quantity, action)
}

func (p *ProductClient) changeStockOnce(ctx context.Context, id uint, quantity int, action string) error {
	base, err := p.baseURL()
	if err != nil {
		return err
	}
	body, err := json.Marshal(map[string]interface{}{"quantity": quantity, "action": action})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/internal/products/%d/stock", base, id), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var result struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&result)
		if result.Message != "" {
			return fmt.Errorf("%s", result.Message)
		}
		if result.Error != "" {
			return fmt.Errorf("%s", result.Error)
		}
		return fmt.Errorf("product service returned status %d", resp.StatusCode)
	}
	return nil
}
