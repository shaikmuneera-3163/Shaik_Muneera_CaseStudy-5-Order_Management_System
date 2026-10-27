package discovery

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	consulapi "github.com/hashicorp/consul/api"
)

type ConsulClient struct{ Client *consulapi.Client }

func NewConsulClient() (*ConsulClient, error) {
	address := strings.TrimSpace(os.Getenv("CONSUL_HTTP_ADDR"))
	if address == "" {
		address = "127.0.0.1:8500"
	}
	if parsed, err := url.Parse(address); err == nil && parsed.Host != "" {
		address = parsed.Host
	}
	cfg := consulapi.DefaultConfig()
	cfg.Address = address
	c, err := consulapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &ConsulClient{Client: c}, nil
}

func (c *ConsulClient) RegisterService(name, port string) error {
	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port for %s: %w", name, err)
	}
	host := strings.TrimSpace(os.Getenv("SERVICE_HOST"))
	if host == "" {
		host = "127.0.0.1"
	}
	id := fmt.Sprintf("%s-%s-%d", name, host, p)
	registration := &consulapi.AgentServiceRegistration{
		ID: id, Name: name, Address: host, Port: p,
		Check: &consulapi.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d/health", host, p),
			Interval: "10s", Timeout: "3s", DeregisterCriticalServiceAfter: "1m",
		},
	}
	if err := c.Client.Agent().ServiceRegister(registration); err != nil {
		return fmt.Errorf("register %s in Consul: %w", name, err)
	}
	return nil
}

func (c *ConsulClient) DeregisterService(name, port string) error {
	p, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	host := strings.TrimSpace(os.Getenv("SERVICE_HOST"))
	if host == "" {
		host = "127.0.0.1"
	}
	id := fmt.Sprintf("%s-%s-%d", name, host, p)
	return c.Client.Agent().ServiceDeregister(id)
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
