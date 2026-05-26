package orchestrator

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, 8080, cfg.RESTPort)
	assert.Equal(t, 9090, cfg.GRPCPort)
	assert.Equal(t, "/data", cfg.DataDir)
	assert.Equal(t, 30*time.Second, cfg.DiscoveryInterval)
	assert.Equal(t, 15*time.Second, cfg.HealthInterval)
	assert.Equal(t, 3, cfg.MaxHealthFails)
	assert.Equal(t, "info", cfg.LogLevel)
}

func TestConfigFromEnv(t *testing.T) {
	os.Setenv("ADE_CONFIG_REST_PORT", "9090")
	os.Setenv("ADE_CONFIG_GRPC_PORT", "9999")
	os.Setenv("ADE_CONFIG_DATA_DIR", "/custom/data")
	os.Setenv("ADE_CONFIG_DISCOVERY_INTERVAL", "10s")
	os.Setenv("ADE_CONFIG_HEALTH_INTERVAL", "5s")
	os.Setenv("ADE_CONFIG_MAX_HEALTH_FAILS", "5")
	os.Setenv("ADE_LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("ADE_CONFIG_REST_PORT")
		os.Unsetenv("ADE_CONFIG_GRPC_PORT")
		os.Unsetenv("ADE_CONFIG_DATA_DIR")
		os.Unsetenv("ADE_CONFIG_DISCOVERY_INTERVAL")
		os.Unsetenv("ADE_CONFIG_HEALTH_INTERVAL")
		os.Unsetenv("ADE_CONFIG_MAX_HEALTH_FAILS")
		os.Unsetenv("ADE_LOG_LEVEL")
	}()

	cfg := ConfigFromEnv()
	assert.Equal(t, 9090, cfg.RESTPort)
	assert.Equal(t, 9999, cfg.GRPCPort)
	assert.Equal(t, "/custom/data", cfg.DataDir)
	assert.Equal(t, 10*time.Second, cfg.DiscoveryInterval)
	assert.Equal(t, 5*time.Second, cfg.HealthInterval)
	assert.Equal(t, 5, cfg.MaxHealthFails)
	assert.Equal(t, "debug", cfg.LogLevel)
}

func TestConfigFromEnv_Defaults(t *testing.T) {
	cfg := ConfigFromEnv()
	assert.Equal(t, 8080, cfg.RESTPort)
	assert.Equal(t, 9090, cfg.GRPCPort)
}

func TestNewServer(t *testing.T) {
	cfg := DefaultConfig()
	srv := NewServer(cfg)
	assert.NotNil(t, srv)
	assert.Equal(t, cfg, srv.cfg)
}

func TestHealthHandler(t *testing.T) {
	cfg := DefaultConfig()
	srv := NewServer(cfg)
	srv.setupRoutes()

	server := httptest.NewServer(srv.mux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	assert.Equal(t, "healthy", body["status"])
	assert.Equal(t, "dev", body["version"])
}

func TestHealthHandler_MethodNotAllowed(t *testing.T) {
	cfg := DefaultConfig()
	srv := NewServer(cfg)
	srv.setupRoutes()

	server := httptest.NewServer(srv.mux)
	defer server.Close()

	resp, err := http.Post(server.URL+"/health", "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestServer_Lifecycle(t *testing.T) {
	cfg := DefaultConfig()
	cfg.RESTPort = 0

	srv := NewServer(cfg)

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	cancel()

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("server did not shut down in time")
	}
}

func TestServer_Version(t *testing.T) {
	assert.Equal(t, "dev", Version)
}

func TestProjectStore_Create(t *testing.T) {
	store := NewProjectStore()
	p := &Project{Name: "mon-projet", Description: "test"}
	err := store.Create(p)
	assert.NoError(t, err)
	assert.NotZero(t, p.CreatedAt)
	assert.NotZero(t, p.UpdatedAt)
}

func TestProjectStore_CreateDuplicate(t *testing.T) {
	store := NewProjectStore()
	err := store.Create(&Project{Name: "projet-a"})
	assert.NoError(t, err)
	err = store.Create(&Project{Name: "projet-a"})
	assert.ErrorIs(t, err, ErrProjectAlreadyExists)
}

func TestProjectStore_Get(t *testing.T) {
	store := NewProjectStore()
	store.Create(&Project{Name: "projet-a"})

	p, ok := store.Get("projet-a")
	assert.True(t, ok)
	assert.Equal(t, "projet-a", p.Name)

	_, ok = store.Get("inconnu")
	assert.False(t, ok)
}

func TestProjectStore_List(t *testing.T) {
	store := NewProjectStore()
	assert.Empty(t, store.List())

	store.Create(&Project{Name: "a"})
	store.Create(&Project{Name: "b"})
	assert.Len(t, store.List(), 2)
}

func TestProjectStore_Update(t *testing.T) {
	store := NewProjectStore()
	store.Create(&Project{Name: "projet-a", Description: "original"})

	err := store.Update(&Project{Name: "projet-a", Description: "mise à jour"})
	assert.NoError(t, err)

	p, _ := store.Get("projet-a")
	assert.Equal(t, "mise à jour", p.Description)

	err = store.Update(&Project{Name: "inconnu"})
	assert.ErrorIs(t, err, ErrProjectNotFound)
}

func TestProjectStore_Delete(t *testing.T) {
	store := NewProjectStore()
	store.Create(&Project{Name: "projet-a"})

	err := store.Delete("projet-a")
	assert.NoError(t, err)
	assert.Empty(t, store.List())

	err = store.Delete("inconnu")
	assert.ErrorIs(t, err, ErrProjectNotFound)
}

func TestProjectStore_Count(t *testing.T) {
	store := NewProjectStore()
	assert.Equal(t, 0, store.Count())
	store.Create(&Project{Name: "a"})
	assert.Equal(t, 1, store.Count())
}

func TestWorkflowStore_Create(t *testing.T) {
	store := NewWorkflowStore()
	wf := &Workflow{Name: "test-wf", Status: "pending"}
	err := store.Create(wf)
	assert.NoError(t, err)
	assert.NotEmpty(t, wf.ID)
	assert.NotNil(t, wf.StartedAt)
}

func TestWorkflowStore_Get(t *testing.T) {
	store := NewWorkflowStore()
	wf := &Workflow{Name: "test-wf"}
	store.Create(wf)

	got, ok := store.Get(wf.ID)
	assert.True(t, ok)
	assert.Equal(t, "test-wf", got.Name)
}

func TestWorkflowStore_List(t *testing.T) {
	store := NewWorkflowStore()
	store.Create(&Workflow{Name: "wf-1"})
	store.Create(&Workflow{Name: "wf-2"})
	assert.Len(t, store.List(), 2)
}

func TestWorkflowStore_ListByProject(t *testing.T) {
	store := NewWorkflowStore()
	store.Create(&Workflow{Name: "wf-1", ProjectName: "proj-a"})
	store.Create(&Workflow{Name: "wf-2", ProjectName: "proj-b"})
	store.Create(&Workflow{Name: "wf-3", ProjectName: "proj-a"})

	assert.Len(t, store.ListByProject("proj-a"), 2)
	assert.Len(t, store.ListByProject("proj-b"), 1)
	assert.Len(t, store.ListByProject("inconnu"), 0)
}

func TestWorkflowStore_Count(t *testing.T) {
	store := NewWorkflowStore()
	assert.Equal(t, 0, store.Count())
	store.Create(&Workflow{Name: "a"})
	assert.Equal(t, 1, store.Count())
}

func TestProjectHTTP_API(t *testing.T) {
	cfg := DefaultConfig()
	srv := NewServer(cfg)
	srv.setupRoutes()
	server := httptest.NewServer(srv.mux)
	defer server.Close()

	t.Run("create project", func(t *testing.T) {
		resp, err := http.Post(server.URL+"/api/v1/projects", "application/json",
			jsonBody(t, map[string]string{"name": "mon-app", "description": "app test"}))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var p Project
		json.NewDecoder(resp.Body).Decode(&p)
		assert.Equal(t, "mon-app", p.Name)
	})

	t.Run("create duplicate", func(t *testing.T) {
		resp, err := http.Post(server.URL+"/api/v1/projects", "application/json",
			jsonBody(t, map[string]string{"name": "mon-app"}))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("list projects", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/projects")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		projects := body["projects"].([]interface{})
		assert.GreaterOrEqual(t, len(projects), 1)
	})

	t.Run("get project", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/projects/mon-app")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var p Project
		json.NewDecoder(resp.Body).Decode(&p)
		assert.Equal(t, "mon-app", p.Name)
	})

	t.Run("get project not found", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/projects/inconnu")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("update project", func(t *testing.T) {
		resp, err := httpPut(server.URL+"/api/v1/projects/mon-app",
			jsonBody(t, map[string]string{"description": "nouvelle desc"}))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var p Project
		json.NewDecoder(resp.Body).Decode(&p)
		assert.Equal(t, "nouvelle desc", p.Description)
	})

	t.Run("delete project", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, server.URL+"/api/v1/projects/mon-app", nil)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("get deleted project", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/projects/mon-app")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestWorkflowHTTP_API(t *testing.T) {
	cfg := DefaultConfig()
	srv := NewServer(cfg)
	srv.setupRoutes()
	server := httptest.NewServer(srv.mux)
	defer server.Close()

	t.Run("create workflow", func(t *testing.T) {
		resp, err := http.Post(server.URL+"/api/v1/workflows", "application/json",
			jsonBody(t, map[string]string{"name": "build-deploy", "project_name": "mon-app"}))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var wf Workflow
		json.NewDecoder(resp.Body).Decode(&wf)
		assert.Equal(t, "build-deploy", wf.Name)
		assert.NotEmpty(t, wf.ID)
	})

	t.Run("create workflow missing name", func(t *testing.T) {
		resp, err := http.Post(server.URL+"/api/v1/workflows", "application/json",
			jsonBody(t, map[string]string{}))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("list workflows", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/workflows")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		workflows := body["workflows"].([]interface{})
		assert.GreaterOrEqual(t, len(workflows), 1)
	})

	t.Run("get workflow", func(t *testing.T) {
		createResp, _ := http.Post(server.URL+"/api/v1/workflows", "application/json",
			jsonBody(t, map[string]string{"name": "wf-test"}))
		var created Workflow
		json.NewDecoder(createResp.Body).Decode(&created)
		createResp.Body.Close()

		resp, err := http.Get(server.URL + "/api/v1/workflows/" + created.ID)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var wf Workflow
		json.NewDecoder(resp.Body).Decode(&wf)
		assert.Equal(t, "wf-test", wf.Name)
	})

	t.Run("get workflow not found", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/workflows/inconnu")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestConfigHTTP_API(t *testing.T) {
	cfg := DefaultConfig()
	srv := NewServer(cfg)
	srv.setupRoutes()
	server := httptest.NewServer(srv.mux)
	defer server.Close()

	t.Run("get config", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/config")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var cfgResp ConfigResponse
		json.NewDecoder(resp.Body).Decode(&cfgResp)
		assert.Equal(t, 8080, cfgResp.RESTPort)
		assert.Equal(t, 9090, cfgResp.GRPCPort)
		assert.Equal(t, "dev", cfgResp.OrchestratorVersion)
	})

	t.Run("update config", func(t *testing.T) {
		resp, err := httpPut(server.URL+"/api/v1/config",
			jsonBody(t, map[string]interface{}{
				"settings": map[string]string{"theme": "dark"},
			}))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("get config after update", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/config")
		require.NoError(t, err)
		defer resp.Body.Close()

		var cfgResp ConfigResponse
		json.NewDecoder(resp.Body).Decode(&cfgResp)
		assert.Equal(t, "dark", cfgResp.Settings["theme"])
	})
}

func TestCORSMiddleware(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("CORS headers on GET", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	})

	t.Run("CORS preflight", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodOptions, server.URL+"/", nil)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
}

func TestRecoveryMiddleware(t *testing.T) {
	handler := recoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL + "/")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func jsonBody(t *testing.T, v interface{}) *jsonReader {
	data, err := json.Marshal(v)
	require.NoError(t, err)
	return &jsonReader{data: data}
}

type jsonReader struct {
	data []byte
}

func (r *jsonReader) Read(p []byte) (int, error) {
	return copy(p, r.data), nil
}

func (r *jsonReader) Close() error {
	return nil
}

func httpPut(url string, body *jsonReader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}
