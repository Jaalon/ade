package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"automated_dev_environment/internal/plugins/contract"
	"automated_dev_environment/internal/plugins/registry"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	assert.NotNil(t, client)
	assert.Contains(t, client.restURL, "localhost:8080")
}

func TestNewClientWithURL(t *testing.T) {
	client := NewClientWithURL("http://test:9090")
	assert.Equal(t, "http://test:9090", client.restURL)
}

func TestListPlugins(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/plugins", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"plugins": []map[string]interface{}{
				{
					"descriptor": map[string]interface{}{
						"name":    "test-plugin",
						"version": "1.0.0",
					},
					"status":       1,
					"http_address": "localhost:8081",
				},
			},
		})
	}))
	defer srv.Close()

	client := NewClientWithURL(srv.URL)
	plugins, err := client.ListPlugins(context.Background())
	assert.NoError(t, err)
	assert.Len(t, plugins, 1)
	assert.Equal(t, "test-plugin", plugins[0].Descriptor.Name)
}

func TestListPlugins_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"plugins": []interface{}{},
		})
	}))
	defer srv.Close()

	client := NewClientWithURL(srv.URL)
	plugins, err := client.ListPlugins(context.Background())
	assert.NoError(t, err)
	assert.Len(t, plugins, 0)
}

func TestListPlugins_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	client := NewClientWithURL(srv.URL)
	_, err := client.ListPlugins(context.Background())
	assert.Error(t, err)
}

func TestListPlugins_Unreachable(t *testing.T) {
	client := NewClientWithURL("http://localhost:1")
	_, err := client.ListPlugins(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "orchestrateur non disponible")
}

func TestGetPlugin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/plugins/test-plugin", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"descriptor": map[string]interface{}{
				"name":    "test-plugin",
				"version": "1.0.0",
			},
			"status":       1,
			"http_address": "localhost:8081",
		})
	}))
	defer srv.Close()

	client := NewClientWithURL(srv.URL)
	plugin, err := client.GetPlugin(context.Background(), "test-plugin")
	assert.NoError(t, err)
	assert.Equal(t, "test-plugin", plugin.Descriptor.Name)
}

func TestGetPlugin_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"plugin not found"}`))
	}))
	defer srv.Close()

	client := NewClientWithURL(srv.URL)
	_, err := client.GetPlugin(context.Background(), "unknown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "introuvable")
}

func TestHealth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/health", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClientWithURL(srv.URL)
	err := client.Health(context.Background())
	assert.NoError(t, err)
}

func TestHealth_NotOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	client := NewClientWithURL(srv.URL)
	err := client.Health(context.Background())
	assert.Error(t, err)
}

func TestFormatPluginList_NotEmpty(t *testing.T) {
	plugins := []*registry.PluginInstance{
		{
			Descriptor: &contract.PluginDescriptor{
				Name:    "test",
				Version: "1.0.0",
				Capabilities: []*contract.Capability{
					{Name: "feature-a", Version: "1.0.0"},
				},
			},
			Status:      contract.HealthHealthy,
			HttpAddress: "localhost:8081",
		},
	}
	result := FormatPluginList(plugins)
	assert.Contains(t, result, "test")
	assert.Contains(t, result, "1.0.0")
	assert.Contains(t, result, "HEALTHY")
	assert.Contains(t, result, "feature-a")
}

func TestFormatPluginList_Empty(t *testing.T) {
	result := FormatPluginList(nil)
	assert.Equal(t, "Aucun plugin enregistré.", result)
}

func TestFormatPluginList_EmptySlice(t *testing.T) {
	result := FormatPluginList([]*registry.PluginInstance{})
	assert.Equal(t, "Aucun plugin enregistré.", result)
}

func TestFormatPluginInfo_Nil(t *testing.T) {
	result := FormatPluginInfo(nil)
	assert.Equal(t, "", result)
}

func TestFormatPluginInfo_WithData(t *testing.T) {
	p := &registry.PluginInstance{
		Descriptor: &contract.PluginDescriptor{
			Name:        "test-plugin",
			Version:     "1.0.0",
			Description: "A test plugin",
			ApiVersion:  "v1",
			Capabilities: []*contract.Capability{
				{Name: "feature-a", Description: "Feature A", Version: "1.0.0"},
			},
			Endpoints: map[string]string{
				"custom": "/api/v1/custom",
			},
		},
		Status:      contract.HealthHealthy,
		HttpAddress: "localhost:8081",
		GrpcAddress: "localhost:50051",
	}
	result := FormatPluginInfo(p)
	assert.Contains(t, result, "test-plugin")
	assert.Contains(t, result, "HEALTHY")
	assert.Contains(t, result, "feature-a")
	assert.Contains(t, result, "custom")
}

func TestIsNotFound(t *testing.T) {
	assert.True(t, IsNotFound(fmt.Errorf("plugin \"xyz\" introuvable")))
	assert.True(t, IsNotFound(fmt.Errorf("plugin not found")))
	assert.False(t, IsNotFound(nil))
	assert.False(t, IsNotFound(fmt.Errorf("autre erreur")))
}
