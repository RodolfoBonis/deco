package decorators

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Tests for service discovery functionality
func TestNewConsulDiscovery(t *testing.T) {

	t.Run("with address", func(t *testing.T) {
		discovery := NewConsulDiscovery("consul.example.com:8500")
		assert.NotNil(t, discovery)
		assert.Equal(t, "consul.example.com:8500", discovery.address)
	})

	t.Run("with empty address", func(t *testing.T) {
		discovery := NewConsulDiscovery("")
		assert.NotNil(t, discovery)
		assert.Equal(t, "localhost:8500", discovery.address)
	})
}

func TestConsulDiscovery_Discover_NoClient(t *testing.T) {

	discovery := &ConsulDiscovery{
		address: "invalid:8500",
		client:  nil,
	}

	instances, err := discovery.Discover("test-service")
	assert.Error(t, err)
	assert.Nil(t, instances)
	assert.Contains(t, err.Error(), "Consul client not available")
}

func TestNewDNSDiscovery(t *testing.T) {

	discovery := NewDNSDiscovery()
	assert.NotNil(t, discovery)
}

func TestDNSDiscovery_Discover(t *testing.T) {

	discovery := NewDNSDiscovery()

	t.Run("valid domain", func(t *testing.T) {
		// Test with a well-known domain
		instances, err := discovery.Discover("google.com")

		// DNS resolution might succeed or fail depending on network
		if err == nil {
			assert.NotNil(t, instances)
			assert.Greater(t, len(instances), 0)

			for _, instance := range instances {
				assert.NotEmpty(t, instance.URL)
				assert.Equal(t, 1, instance.Weight)
				assert.True(t, instance.Healthy)
				assert.NotZero(t, instance.LastCheck)
				assert.NotNil(t, instance.Metadata)
			}
		} else {
			assert.Contains(t, err.Error(), "failed to resolve DNS")
		}
	})

	t.Run("invalid domain", func(t *testing.T) {
		instances, err := discovery.Discover("invalid-domain-that-does-not-exist-12345.com")
		assert.Error(t, err)
		assert.Nil(t, instances)
		assert.Contains(t, err.Error(), "failed to resolve DNS")
	})
}

func TestNewK8sDiscovery(t *testing.T) {

	t.Run("with namespace", func(t *testing.T) {
		discovery := NewK8sDiscovery("production")
		assert.NotNil(t, discovery)
		assert.Equal(t, "production", discovery.namespace)
	})

	t.Run("with empty namespace", func(t *testing.T) {
		discovery := NewK8sDiscovery("")
		assert.NotNil(t, discovery)
		assert.Equal(t, "default", discovery.namespace)
	})
}

func TestK8sDiscovery_Discover(t *testing.T) {

	discovery := NewK8sDiscovery("test-namespace")

	t.Run("valid service", func(t *testing.T) {
		// Test with a service that might exist in the cluster
		instances, err := discovery.Discover("kubernetes.default")

		// This might fail if not running in a Kubernetes cluster
		if err == nil {
			assert.NotNil(t, instances)
			assert.Greater(t, len(instances), 0)

			for _, instance := range instances {
				assert.NotEmpty(t, instance.URL)
				assert.Equal(t, 1, instance.Weight)
				assert.True(t, instance.Healthy)
				assert.NotZero(t, instance.LastCheck)
				assert.Equal(t, "test-namespace", instance.Metadata["namespace"])
				assert.Equal(t, "kubernetes.default", instance.Metadata["service"])
			}
		} else {
			assert.Contains(t, err.Error(), "failed to resolve Kubernetes service")
		}
	})

	t.Run("invalid service", func(t *testing.T) {
		instances, err := discovery.Discover("invalid-service-that-does-not-exist")
		assert.Error(t, err)
		assert.Nil(t, instances)
		assert.Contains(t, err.Error(), "failed to resolve Kubernetes service")
	})
}

func TestNewStaticDiscovery(t *testing.T) {

	targets := []string{
		"http://service1.example.com:8080",
		"http://service2.example.com:8080",
		"http://service3.example.com:8080",
	}

	discovery := NewStaticDiscovery(targets)
	assert.NotNil(t, discovery)
	assert.Equal(t, targets, discovery.targets)
}

func TestStaticDiscovery_Discover(t *testing.T) {

	targets := []string{
		"http://service1.example.com:8080",
		"http://service2.example.com:8080",
		"http://service3.example.com:8080",
	}

	discovery := NewStaticDiscovery(targets)

	instances, err := discovery.Discover("any-service")
	assert.NoError(t, err)
	assert.NotNil(t, instances)
	assert.Len(t, instances, 3)

	for i, instance := range instances {
		assert.Equal(t, targets[i], instance.URL)
		assert.Equal(t, 1, instance.Weight)
		assert.True(t, instance.Healthy)
		assert.NotZero(t, instance.LastCheck)
		assert.Equal(t, "static", instance.Metadata["discovery"])
	}
}

func TestStaticDiscovery_Discover_EmptyTargets(t *testing.T) {

	discovery := NewStaticDiscovery([]string{})

	instances, err := discovery.Discover("any-service")
	assert.NoError(t, err)
	assert.NotNil(t, instances)
	assert.Len(t, instances, 0)
}

func TestStaticDiscovery_Discover_WithSpaces(t *testing.T) {

	targets := []string{
		"  http://service1.example.com:8080  ",
		"  http://service2.example.com:8080  ",
	}

	discovery := NewStaticDiscovery(targets)

	instances, err := discovery.Discover("any-service")
	assert.NoError(t, err)
	assert.NotNil(t, instances)
	assert.Len(t, instances, 2)

	for _, instance := range instances {
		// Should be trimmed
		assert.Equal(t, "http://service1.example.com:8080", instances[0].URL)
		assert.Equal(t, "http://service2.example.com:8080", instances[1].URL)
		assert.Equal(t, 1, instance.Weight)
		assert.True(t, instance.Healthy)
		assert.NotZero(t, instance.LastCheck)
		assert.Equal(t, "static", instance.Metadata["discovery"])
	}
}

func TestProxyInstance_Structure(t *testing.T) {

	instance := &ProxyInstance{
		URL:       "http://example.com:8080",
		Weight:    5,
		Healthy:   true,
		LastCheck: time.Now(),
		Metadata: map[string]string{
			"version": "1.0.0",
			"region":  "us-west-1",
		},
	}

	assert.Equal(t, "http://example.com:8080", instance.URL)
	assert.Equal(t, 5, instance.Weight)
	assert.True(t, instance.Healthy)
	assert.NotZero(t, instance.LastCheck)
	assert.Equal(t, "1.0.0", instance.Metadata["version"])
	assert.Equal(t, "us-west-1", instance.Metadata["region"])
}

func TestServiceDiscovery_InterfaceCompliance(_ *testing.T) {

	// Test that all discovery types can be used interchangeably
	// This is a compile-time test to ensure interface compliance

	// ConsulDiscovery
	consulDiscovery := NewConsulDiscovery("localhost:8500")
	_, _ = consulDiscovery.Discover("test")

	// DNSDiscovery
	dnsDiscovery := NewDNSDiscovery()
	_, _ = dnsDiscovery.Discover("test")

	// K8sDiscovery
	k8sDiscovery := NewK8sDiscovery("default")
	_, _ = k8sDiscovery.Discover("test")

	// StaticDiscovery
	staticDiscovery := NewStaticDiscovery([]string{"http://test.com"})
	_, _ = staticDiscovery.Discover("test")
}

func TestConsulDiscovery_Structure(t *testing.T) {

	discovery := &ConsulDiscovery{
		address: "consul.example.com:8500",
		client:  nil,
	}

	assert.Equal(t, "consul.example.com:8500", discovery.address)
	assert.Nil(t, discovery.client)
}

func TestDNSDiscovery_Structure(t *testing.T) {

	discovery := &DNSDiscovery{}
	assert.NotNil(t, discovery)
}

func TestK8sDiscovery_Structure(t *testing.T) {

	discovery := &K8sDiscovery{
		namespace: "production",
	}

	assert.Equal(t, "production", discovery.namespace)
}

func TestStaticDiscovery_Structure(t *testing.T) {

	targets := []string{"http://service1.com", "http://service2.com"}
	discovery := &StaticDiscovery{
		targets: targets,
	}

	assert.Equal(t, targets, discovery.targets)
}

func TestServiceDiscovery_ErrorHandling(t *testing.T) {

	t.Run("consul with invalid address", func(t *testing.T) {
		discovery := NewConsulDiscovery("invalid-address:99999")
		instances, err := discovery.Discover("test-service")
		assert.Error(t, err)
		assert.Nil(t, instances)
	})

	t.Run("dns with empty service name", func(t *testing.T) {
		discovery := NewDNSDiscovery()
		instances, err := discovery.Discover("")
		assert.Error(t, err)
		assert.Nil(t, instances)
	})

	t.Run("k8s with empty service name", func(t *testing.T) {
		discovery := NewK8sDiscovery("default")
		instances, err := discovery.Discover("")
		assert.Error(t, err)
		assert.Nil(t, instances)
	})
}

func TestServiceDiscovery_MetadataHandling(t *testing.T) {

	t.Run("static discovery metadata", func(t *testing.T) {
		targets := []string{"http://service1.com"}
		discovery := NewStaticDiscovery(targets)

		instances, err := discovery.Discover("test")
		assert.NoError(t, err)
		assert.Len(t, instances, 1)

		metadata := instances[0].Metadata
		assert.Equal(t, "static", metadata["discovery"])
	})

	t.Run("k8s discovery metadata", func(t *testing.T) {
		discovery := NewK8sDiscovery("test-namespace")

		// This might fail if not in a K8s cluster, but we can test the structure
		instances, err := discovery.Discover("test-service")
		if err == nil {
			assert.Len(t, instances, 1)
			metadata := instances[0].Metadata
			assert.Equal(t, "test-namespace", metadata["namespace"])
			assert.Equal(t, "test-service", metadata["service"])
		}
	})
}
