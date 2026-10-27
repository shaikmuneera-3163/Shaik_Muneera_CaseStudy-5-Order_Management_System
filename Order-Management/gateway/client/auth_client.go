package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"order-management/shared/discovery"
	"order-management/shared/dto"
)

type AuthClient struct {
	BaseURL   string
	Discovery *discovery.ConsulClient
	Client    *http.Client
}

func NewAuthClient(url string) *AuthClient {
	return &AuthClient{BaseURL: url, Client: &http.Client{Timeout: 5 * time.Second}}
}
func NewAuthClientWithDiscovery(d *discovery.ConsulClient) *AuthClient {
	return &AuthClient{Discovery: d, Client: &http.Client{Timeout: 5 * time.Second}}
}
func (a *AuthClient) baseURL() (string, error) {
	if a.Discovery != nil {
		return a.Discovery.GetServiceURL("auth")
	}
	if a.BaseURL == "" {
		return "", fmt.Errorf("auth service URL is empty")
	}
	return a.BaseURL, nil
}
func (a *AuthClient) Login(ctx context.Context, req dto.LoginRequest) (interface{}, error) {
	base, err := a.baseURL()
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/auth/login", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	resp, err := a.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth service returned %d", resp.StatusCode)
	}
	return result, nil
}
