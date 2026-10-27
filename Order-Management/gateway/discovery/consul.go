package discovery

import (
	"fmt"
	"os"

	consulapi "github.com/hashicorp/consul/api"
)

type ConsulClient struct {
	Client *consulapi.Client
}

func NewConsulClient() (*ConsulClient, error) {
	address := os.Getenv("CONSUL_HTTP_ADDR")
	if address == "" {
		address = "127.0.0.1:8500"
	}

	cfg := consulapi.DefaultConfig()
	cfg.Address = address

	client, err := consulapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &ConsulClient{Client: client}, nil
}

func (c *ConsulClient) GetServiceURL(name string) (string, error) {
	services, _, err := c.Client.Health().Service(name, "", true, nil)
	if err != nil {
		return "", err
	}

	if len(services) == 0 {
		return "", fmt.Errorf("no healthy service found: %s", name)
	}

	svc := services[0].Service
	host := svc.Address
	if host == "" {
		host = "127.0.0.1"
	}

	return fmt.Sprintf("http://%s:%d", host, svc.Port), nil
}
