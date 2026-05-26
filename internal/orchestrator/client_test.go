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

func TestGetConfig(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/config", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ConfigResponse{
			ProjectName:         "mon-projet",
			OrchestratorVersion: "1.0.0",
			RESTPort:            8080,
			GRPCPort:            9090,
		})
	}))
	defer srv.Close()

	client := NewClientWithURL(srv.URL)
	cfg, err := client.GetConfig(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "mon-projet", cfg.ProjectName)
	assert.Equal(t, "1.0.0", cfg.OrchestratorVersion)
	assert.Equal(t, 8080, cfg.RESTPort)
	assert.Equal(t, 9090, cfg.GRPCPort)
}

func TestGetConfig_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("erreur interne"))
	}))
	defer srv.Close()

	client := NewClientWithURL(srv.URL)
	_, err := client.GetConfig(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erreur interne")
}

func TestGetDashboardStats(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/dashboard/stats", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(DashboardStats{
			Projects:      5,
			Plugins:       3,
			Workflows:     12,
			Reports:       7,
			ActivePlugins: 2,
		})
	}))
	defer srv.Close()

	client := NewClientWithURL(srv.URL)
	stats, err := client.GetDashboardStats(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 5, stats.Projects)
	assert.Equal(t, 3, stats.Plugins)
	assert.Equal(t, 2, stats.ActivePlugins)
}

func TestListProjects(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/projects", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"projects": []map[string]interface{}{
				{"name": "projet-a"},
				{"name": "projet-b"},
			},
		})
	}))
	defer srv.Close()

	client := NewClientWithURL(srv.URL)
	projects, err := client.ListProjects(context.Background())
	assert.NoError(t, err)
	assert.Len(t, projects, 2)
	assert.Equal(t, "projet-a", projects[0].Name)
	assert.Equal(t, "projet-b", projects[1].Name)
}

func TestListProjects_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"projects": []interface{}{},
		})
	}))
	defer srv.Close()

	client := NewClientWithURL(srv.URL)
	projects, err := client.ListProjects(context.Background())
	assert.NoError(t, err)
	assert.Len(t, projects, 0)
}

func TestCreateProject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/projects", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	client := NewClientWithURL(srv.URL)
	err := client.CreateProject(context.Background(), Project{Name: "new-project"})
	assert.NoError(t, err)
}

func TestCreateProject_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"error":"projet déjà existant"}`))
	}))
	defer srv.Close()

	client := NewClientWithURL(srv.URL)
	err := client.CreateProject(context.Background(), Project{Name: "existing"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "déjà existant")
}

func TestDeleteProject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/projects/mon-projet", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := NewClientWithURL(srv.URL)
	err := client.DeleteProject(context.Background(), "mon-projet")
	assert.NoError(t, err)
}

func TestDeleteProject_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := NewClientWithURL(srv.URL)
	err := client.DeleteProject(context.Background(), "inconnu")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "introuvable")
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
