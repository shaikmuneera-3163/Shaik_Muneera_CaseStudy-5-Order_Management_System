package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"order-management/services/products/model"
	"order-management/shared/discovery"
)

type ProductClient struct {
	BaseURL   string
	Discovery *discovery.ConsulClient
	Client    *http.Client
}

func NewProductClient(url string) *ProductClient {
	return &ProductClient{BaseURL: url, Client: &http.Client{Timeout: 5 * time.Second}}
}
func NewProductClientWithDiscovery(d *discovery.ConsulClient) *ProductClient {
	return &ProductClient{Discovery: d, Client: &http.Client{Timeout: 5 * time.Second}}
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
func (p *ProductClient) GetAll(ctx context.Context) ([]model.Product, error) {
	base, err := p.baseURL()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/products", nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product service returned %d", resp.StatusCode)
	}
	var result struct {
		Data []model.Product `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Data, nil
}
func (p *ProductClient) GetByID(ctx context.Context, id uint) (*model.Product, error) {
	base, err := p.baseURL()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/products/%d", base, id), nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product service returned %d", resp.StatusCode)
	}
	var result struct {
		Data model.Product `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}
func (p *ProductClient) Create(ctx context.Context, product model.Product) (interface{}, error) {
	base, err := p.baseURL()
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(product)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/products", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("product service returned %d", resp.StatusCode)
	}
	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}
