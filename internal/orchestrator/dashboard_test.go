package orchestrator

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardStats_WithData(t *testing.T) {
	srv := NewServer(DefaultConfig())
	projectStore := NewProjectStore()
	projectStore.Create(&Project{Name: "proj-a", Description: "desc-a"})
	projectStore.Create(&Project{Name: "proj-b", Description: "desc-b"})
	projectStore.Create(&Project{Name: "proj-c", Description: "desc-c"})
	srv.projectStore = projectStore

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/stats", nil)
	w := httptest.NewRecorder()
	srv.handleDashboardStats(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var stats DashboardStats
	err := json.Unmarshal(w.Body.Bytes(), &stats)
	require.NoError(t, err)
	assert.Equal(t, 3, stats.Projects)
	assert.Equal(t, 0, stats.Plugins)
	assert.Equal(t, 0, stats.Workflows)
	assert.Equal(t, 0, stats.Reports)
	assert.Equal(t, 0, stats.ActivePlugins)
}

func TestDashboardStats_Empty(t *testing.T) {
	srv := NewServer(DefaultConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/stats", nil)
	w := httptest.NewRecorder()
	srv.handleDashboardStats(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var stats DashboardStats
	err := json.Unmarshal(w.Body.Bytes(), &stats)
	require.NoError(t, err)
	assert.Equal(t, 0, stats.Projects)
	assert.Equal(t, 0, stats.Plugins)
}

func TestDashboardStats_MethodNotAllowed(t *testing.T) {
	srv := NewServer(DefaultConfig())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/dashboard/stats", nil)
	w := httptest.NewRecorder()
	srv.handleDashboardStats(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestDashboardActivity_WithEvents(t *testing.T) {
	srv := NewServer(DefaultConfig())
	srv.activityStore.Add("plugin_register", "plugin-a enregistré")
	srv.activityStore.Add("workflow_start", "build démarré")
	srv.activityStore.Add("project_create", "projet créé")
	srv.activityStore.Add("plugin_register", "plugin-b enregistré")
	srv.activityStore.Add("workflow_complete", "build terminé")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/activity", nil)
	w := httptest.NewRecorder()
	srv.handleDashboardActivity(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Events []ActivityEvent `json:"events"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Events, 5)
	assert.Equal(t, "plugin_register", resp.Events[0].Type)
}

func TestDashboardActivity_Empty(t *testing.T) {
	srv := NewServer(DefaultConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/activity", nil)
	w := httptest.NewRecorder()
	srv.handleDashboardActivity(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Events []ActivityEvent `json:"events"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Empty(t, resp.Events)
}

func TestActivityStore_RingBuffer(t *testing.T) {
	store := NewActivityStore(3)
	for i := 0; i < 10; i++ {
		store.Add("test", "event")
	}
	assert.Equal(t, 3, store.Len())
}

func TestDashboardHTTP_API(t *testing.T) {
	srv := NewServer(DefaultConfig())
	srv.setupRoutes()

	ts := httptest.NewServer(srv.handler())
	defer ts.Close()

	t.Run("stats", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/dashboard/stats")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("activity", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/dashboard/activity")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("ws_not_implemented", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/ws")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)
	})
}

func TestStaticSPA(t *testing.T) {
	srv := NewServer(DefaultConfig())
	srv.setupRoutes()

	ts := httptest.NewServer(srv.handler())
	defer ts.Close()

	t.Run("api_route_not_caught_by_spa", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/plugins")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.NotEqual(t, http.StatusNotFound, resp.StatusCode,
			"API route should not return 404, should be handled by mux")
	})

	t.Run("root_serves_index", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("unknown_route_spa_fallback", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/une-page-frontend")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("unknown_route_deep_spa_fallback", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/dashboard/projects/123")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestActivityStore_Concurrency(t *testing.T) {
	store := NewActivityStore(100)
	done := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			store.Add("test", "concurrent event")
		}
		done <- struct{}{}
	}()

	go func() {
		for i := 0; i < 50; i++ {
			store.Add("test", "concurrent event")
		}
		done <- struct{}{}
	}()

	<-done
	<-done

	store.Add("final", "done")
	events := store.List()
	assert.GreaterOrEqual(t, len(events), 1)
}
