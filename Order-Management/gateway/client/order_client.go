package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	ordermodel "order-management/services/orders/model"
	"order-management/services/orders/resilience"
	"order-management/shared/discovery"
	"order-management/shared/dto"
	"os"
	"strconv"
	"time"

	"github.com/sony/gobreaker"
)

type OrderClient struct {
	BaseURL   string
	Discovery *discovery.ConsulClient
	Client    *http.Client
	Breaker   *gobreaker.CircuitBreaker
	Retries   int
}

func NewOrderClient(url string) *OrderClient {
	timeout, retries := resilienceSettings()

	return &OrderClient{
		BaseURL: url,
		Client:  &http.Client{Timeout: timeout},
		Breaker: resilience.NewCircuitBreaker("order-service"),
		Retries: retries,
	}
}

func NewOrderClientWithDiscovery(d *discovery.ConsulClient) *OrderClient {
	timeout, retries := resilienceSettings()

	return &OrderClient{
		Discovery: d,
		Client:    &http.Client{Timeout: timeout},
		Breaker:   resilience.NewCircuitBreaker("order-service"),
		Retries:   retries,
	}
}

func (o *OrderClient) baseURL() (string, error) {
	if o.Discovery != nil {
		return o.Discovery.GetServiceURL("orders")
	}

	if o.BaseURL == "" {
		return "", fmt.Errorf("order service URL is empty")
	}

	return o.BaseURL, nil
}

// doRequest sends an HTTP request with retry and circuit-breaker protection.
func (o *OrderClient) doRequest(
	ctx context.Context,
	method string,
	url string,
	body []byte,
) (*http.Response, error) {
	attempts := o.Retries
	if attempts < 1 || attempts > 5 {
		attempts = 3
	}

	var response *http.Response

	err := resilience.Retry(attempts, func() error {
		if err := ctx.Err(); err != nil {
			return err
		}

		// Create a new request for every attempt.
		request, err := http.NewRequestWithContext(
			ctx,
			method,
			url,
			bytes.NewReader(body),
		)
		if err != nil {
			return err
		}

		if body != nil {
			request.Header.Set("Content-Type", "application/json")
		}

		operation := func() (interface{}, error) {
			resp, err := o.Client.Do(request)
			if err != nil {
				return nil, err
			}

			// Treat server errors as failures.
			// Do not retry client errors such as 400 or 404.
			if resp.StatusCode >= 500 {
				resp.Body.Close()
				return nil, fmt.Errorf(
					"order service returned status %d",
					resp.StatusCode,
				)
			}

			return resp, nil
		}

		var result interface{}
		var requestErr error

		if o.Breaker != nil {
			result, requestErr = o.Breaker.Execute(operation)
		} else {
			result, requestErr = operation()
		}

		if requestErr != nil {
			log.Printf(
				"[ORDER CLIENT] Request attempt failed: %v",
				requestErr,
			)
			return requestErr
		}

		var ok bool
		response, ok = result.(*http.Response)
		if !ok {
			return fmt.Errorf("unexpected order service response")
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("order service request failed: %w", err)
	}

	return response, nil
}

func (o *OrderClient) Create(
	ctx context.Context,
	req dto.OrderRequest,
) (*ordermodel.Order, error) {
	base, err := o.baseURL()
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := o.doRequest(ctx, http.MethodPost, base+"/orders", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("order service returned %d", resp.StatusCode)
	}

	var envelope struct {
		Data ordermodel.Order `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, err
	}

	return &envelope.Data, nil
}

func (o *OrderClient) GetAll(
	ctx context.Context,
) ([]ordermodel.Order, error) {
	base, err := o.baseURL()
	if err != nil {
		return nil, err
	}

	raw, err := o.doJSON(ctx, http.MethodGet, base+"/orders")
	if err != nil {
		return nil, err
	}

	var envelope struct {
		Data []ordermodel.Order `json:"data"`
	}

	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, err
	}

	return envelope.Data, nil
}

func (o *OrderClient) GetByID(
	ctx context.Context,
	id uint,
) (*ordermodel.Order, error) {
	base, err := o.baseURL()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/orders/%d", base, id)

	raw, err := o.doJSON(ctx, http.MethodGet, url)
	if err != nil {
		return nil, err
	}

	var envelope struct {
		Data ordermodel.Order `json:"data"`
	}

	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, err
	}

	return &envelope.Data, nil
}

func (o *OrderClient) UpdateStatus(
	ctx context.Context,
	id uint,
	status string,
) (*ordermodel.Order, error) {
	base, err := o.baseURL()
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(map[string]string{
		"status": status,
	})
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/orders/%d/status", base, id)

	resp, err := o.doRequest(ctx, http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("order service returned %d", resp.StatusCode)
	}

	var envelope struct {
		Data ordermodel.Order `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, err
	}

	return &envelope.Data, nil
}

func (o *OrderClient) doJSON(
	ctx context.Context,
	method string,
	url string,
) ([]byte, error) {
	resp, err := o.doRequest(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("order service returned %d", resp.StatusCode)
	}

	var result json.RawMessage

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

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
