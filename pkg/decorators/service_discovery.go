package decorators

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
)

// ConsulDiscovery implements service discovery using Consul
type ConsulDiscovery struct {
	address string
	client  *api.Client
}

// NewConsulDiscovery creates a new Consul service discovery
func NewConsulDiscovery(address string) *ConsulDiscovery {
	if address == "" {
		address = "localhost:8500"
	}

	config := api.DefaultConfig()
	config.Address = address

	client, err := api.NewClient(config)
	if err != nil {
		LogVerbose("Failed to create Consul client: %v", err)
		return &ConsulDiscovery{address: address}
	}

	return &ConsulDiscovery{
		address: address,
		client:  client,
	}
}

// Discover discovers service instances using Consul
func (cd *ConsulDiscovery) Discover(service string) ([]*ProxyInstance, error) {
	if cd.client == nil {
		return nil, fmt.Errorf("Consul client not available")
	}

	// Query Consul for service instances
	services, _, err := cd.client.Health().Service(service, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query Consul: %v", err)
	}

	var instances []*ProxyInstance
	for _, service := range services {
		instance := &ProxyInstance{
			URL:       fmt.Sprintf("http://%s:%d", service.Service.Address, service.Service.Port),
			Weight:    1,
			Healthy:   true,
			LastCheck: time.Now(),
			Metadata:  make(map[string]string),
		}

		// Add service metadata
		for key, value := range service.Service.Meta {
			instance.Metadata[key] = value
		}

		instances = append(instances, instance)
	}

	return instances, nil
}

// DNSDiscovery implements service discovery using DNS
type DNSDiscovery struct{}

// NewDNSDiscovery creates a new DNS service discovery
func NewDNSDiscovery() *DNSDiscovery {
	return &DNSDiscovery{}
}

// Discover discovers service instances using DNS
func (dd *DNSDiscovery) Discover(service string) ([]*ProxyInstance, error) {
	// Resolve DNS
	ips, err := net.LookupIP(service)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve DNS for %s: %v", service, err)
	}

	var instances []*ProxyInstance
	for _, ip := range ips {
		// Assume HTTP on port 80 for DNS discovery
		instance := &ProxyInstance{
			URL:       fmt.Sprintf("http://%s:80", ip.String()),
			Weight:    1,
			Healthy:   true,
			LastCheck: time.Now(),
			Metadata:  make(map[string]string),
		}
		instances = append(instances, instance)
	}

	return instances, nil
}

// K8sDiscovery implements service discovery using Kubernetes
type K8sDiscovery struct {
	namespace string
}

// NewK8sDiscovery creates a new Kubernetes service discovery
func NewK8sDiscovery(namespace string) *K8sDiscovery {
	if namespace == "" {
		namespace = "default"
	}

	return &K8sDiscovery{
		namespace: namespace,
	}
}

// Discover discovers service instances using Kubernetes
func (kd *K8sDiscovery) Discover(service string) ([]*ProxyInstance, error) {
	// For now, implement a simple DNS-based approach for Kubernetes
	// In a real implementation, you would use the Kubernetes API
	k8sServiceName := fmt.Sprintf("%s.%s.svc.cluster.local", service, kd.namespace)

	// Resolve the Kubernetes service DNS
	ips, err := net.LookupIP(k8sServiceName)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve Kubernetes service %s: %v", k8sServiceName, err)
	}

	var instances []*ProxyInstance
	for _, ip := range ips {
		// Assume HTTP on port 80 for Kubernetes services
		instance := &ProxyInstance{
			URL:       fmt.Sprintf("http://%s:80", ip.String()),
			Weight:    1,
			Healthy:   true,
			LastCheck: time.Now(),
			Metadata: map[string]string{
				"namespace": kd.namespace,
				"service":   service,
			},
		}
		instances = append(instances, instance)
	}

	return instances, nil
}

// StaticDiscovery implements static service discovery
type StaticDiscovery struct {
	targets []string
}

// NewStaticDiscovery creates a new static service discovery
func NewStaticDiscovery(targets []string) *StaticDiscovery {
	return &StaticDiscovery{
		targets: targets,
	}
}

// Discover returns the static targets as instances
func (sd *StaticDiscovery) Discover(service string) ([]*ProxyInstance, error) {
	var instances []*ProxyInstance

	for _, target := range sd.targets {
		instance := &ProxyInstance{
			URL:       strings.TrimSpace(target),
			Weight:    1,
			Healthy:   true,
			LastCheck: time.Now(),
			Metadata: map[string]string{
				"discovery": "static",
			},
		}
		instances = append(instances, instance)
	}

	return instances, nil
}
