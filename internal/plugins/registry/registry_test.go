package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"automated_dev_environment/internal/plugins/contract"
)

func TestStore_RegisterAndGet(t *testing.T) {
	s := NewStore()
	p := &PluginInstance{
		Descriptor: &contract.PluginDescriptor{
			Name:       "test-plugin",
			Version:    "1.0.0",
			ApiVersion: "v1",
		},
		HttpAddress: "localhost:8081",
	}

	s.Register(p)

	got, ok := s.Get("test-plugin")
	if !ok {
		t.Fatal("expected plugin to be found")
	}
	if got.Descriptor.Name != "test-plugin" {
		t.Errorf("Name = %q, want %q", got.Descriptor.Name, "test-plugin")
	}
	if got.Descriptor.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", got.Descriptor.Version, "1.0.0")
	}
	if got.Status != contract.HealthHealthy {
		t.Errorf("Status = %d, want %d", got.Status, contract.HealthHealthy)
	}
	if got.FailedChecks != 0 {
		t.Errorf("FailedChecks = %d, want 0", got.FailedChecks)
	}
}

func TestStore_RegisterReplace(t *testing.T) {
	s := NewStore()

	s.Register(&PluginInstance{
		Descriptor:  &contract.PluginDescriptor{Name: "dup", Version: "1.0.0"},
		HttpAddress: "first:8081",
	})
	s.Register(&PluginInstance{
		Descriptor:  &contract.PluginDescriptor{Name: "dup", Version: "2.0.0"},
		HttpAddress: "second:8081",
	})

	got, ok := s.Get("dup")
	if !ok {
		t.Fatal("expected plugin to be found")
	}
	if got.Descriptor.Version != "2.0.0" {
		t.Errorf("Version = %q, want %q", got.Descriptor.Version, "2.0.0")
	}
	if got.HttpAddress != "second:8081" {
		t.Errorf("HttpAddress = %q, want %q", got.HttpAddress, "second:8081")
	}
}

func TestStore_Unregister(t *testing.T) {
	s := NewStore()
	s.Register(&PluginInstance{
		Descriptor: &contract.PluginDescriptor{Name: "to-remove", Version: "1.0.0"},
	})

	if err := s.Unregister("to-remove"); err != nil {
		t.Fatalf("Unregister failed: %v", err)
	}

	if _, ok := s.Get("to-remove"); ok {
		t.Error("expected plugin to be removed")
	}
}

func TestStore_UnregisterUnknown(t *testing.T) {
	s := NewStore()
	err := s.Unregister("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown plugin")
	}
}

func TestStore_List(t *testing.T) {
	s := NewStore()
	s.Register(&PluginInstance{Descriptor: &contract.PluginDescriptor{Name: "a", Version: "1.0.0"}})
	s.Register(&PluginInstance{Descriptor: &contract.PluginDescriptor{Name: "b", Version: "1.0.0"}})
	s.Register(&PluginInstance{Descriptor: &contract.PluginDescriptor{Name: "c", Version: "1.0.0"}})
	s.Register(&PluginInstance{Descriptor: &contract.PluginDescriptor{Name: "d", Version: "1.0.0"}})

	plugins := s.List()
	if len(plugins) != 4 {
		t.Fatalf("len = %d, want 4", len(plugins))
	}

	names := make(map[string]bool)
	for _, p := range plugins {
		names[p.Descriptor.Name] = true
	}
	if !names["a"] || !names["b"] || !names["c"] || !names["d"] {
		t.Errorf("missing some plugins, got %v", names)
	}
}

func TestStore_List_Empty(t *testing.T) {
	s := NewStore()
	plugins := s.List()
	if len(plugins) != 0 {
		t.Errorf("len = %d, want 0", len(plugins))
	}
}

func TestStore_UpdateHealth(t *testing.T) {
	s := NewStore()
	s.Register(&PluginInstance{Descriptor: &contract.PluginDescriptor{Name: "health-test", Version: "1.0.0"}})

	s.UpdateHealth("health-test", contract.HealthHealthy)
	p, _ := s.Get("health-test")
	if p.FailedChecks != 0 {
		t.Errorf("FailedChecks = %d, want 0 after healthy", p.FailedChecks)
	}

	s.UpdateHealth("health-test", contract.HealthUnhealthy)
	p, _ = s.Get("health-test")
	if p.FailedChecks != 1 {
		t.Errorf("FailedChecks = %d, want 1 after unhealthy", p.FailedChecks)
	}

	s.UpdateHealth("health-test", contract.HealthUnhealthy)
	p, _ = s.Get("health-test")
	if p.FailedChecks != 2 {
		t.Errorf("FailedChecks = %d, want 2 after 2 failures", p.FailedChecks)
	}

	s.UpdateHealth("health-test", contract.HealthHealthy)
	p, _ = s.Get("health-test")
	if p.FailedChecks != 0 {
		t.Errorf("FailedChecks = %d, want 0 after recovery", p.FailedChecks)
	}
}

func TestStore_UpdateHealthUnknown(t *testing.T) {
	s := NewStore()
	s.UpdateHealth("nonexistent", contract.HealthHealthy)

	if s.Count() != 0 {
		t.Errorf("Count = %d, want 0", s.Count())
	}
}

func TestStore_Count(t *testing.T) {
	s := NewStore()
	if s.Count() != 0 {
		t.Errorf("Count = %d, want 0", s.Count())
	}

	s.Register(&PluginInstance{Descriptor: &contract.PluginDescriptor{Name: "x", Version: "1.0.0"}})
	if s.Count() != 1 {
		t.Errorf("Count = %d, want 1", s.Count())
	}
}

func TestHealthChecker_RemovesAfterMaxFails(t *testing.T) {
	store := NewStore()

	store.Register(&PluginInstance{
		Descriptor:  &contract.PluginDescriptor{Name: "failing", Version: "1.0.0"},
		HttpAddress: "localhost:1",
	})

	checker := NewHealthChecker(store, 10*time.Millisecond, 3)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	checker.Start(ctx)

	time.Sleep(100 * time.Millisecond)

	_, ok := store.Get("failing")
	if ok {
		t.Error("expected plugin to be removed after max failures")
	}
}

func TestHealthChecker_HealthyPlugin(t *testing.T) {
	healthSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(contract.HealthStatus{
			Status:    contract.HealthHealthy,
			Message:   "ok",
			Timestamp: time.Now().Unix(),
		})
	}))
	defer healthSrv.Close()

	store := NewStore()
	httpAddr := healthSrv.Listener.Addr().String()
	store.Register(&PluginInstance{
		Descriptor:  &contract.PluginDescriptor{Name: "healthy", Version: "1.0.0"},
		HttpAddress: httpAddr,
	})

	checker := NewHealthChecker(store, 50*time.Millisecond, 3)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	checker.Start(ctx)

	time.Sleep(120 * time.Millisecond)

	p, ok := store.Get("healthy")
	if !ok {
		t.Fatal("expected plugin to still be registered")
	}
	if p.Status != contract.HealthHealthy {
		t.Errorf("Status = %d, want %d", p.Status, contract.HealthHealthy)
	}
	if p.FailedChecks != 0 {
		t.Errorf("FailedChecks = %d, want 0", p.FailedChecks)
	}
}

func TestAPI_ListHandler(t *testing.T) {
	store := NewStore()
	api := NewAPI(store)

	store.Register(&PluginInstance{
		Descriptor: &contract.PluginDescriptor{Name: "p1", Version: "1.0.0"},
	})
	store.Register(&PluginInstance{
		Descriptor: &contract.PluginDescriptor{Name: "p2", Version: "2.0.0"},
	})

	ts := httptest.NewServer(http.HandlerFunc(api.ServeHTTP))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/plugins")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result map[string][]*PluginInstance
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}

	plugins, ok := result["plugins"]
	if !ok {
		t.Fatal("missing 'plugins' key")
	}
	if len(plugins) != 2 {
		t.Fatalf("len = %d, want 2", len(plugins))
	}
}

func TestAPI_ListHandler_Empty(t *testing.T) {
	store := NewStore()
	api := NewAPI(store)

	ts := httptest.NewServer(http.HandlerFunc(api.ServeHTTP))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/plugins")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result map[string][]*PluginInstance
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}

	plugins := result["plugins"]
	if len(plugins) != 0 {
		t.Errorf("len = %d, want 0", len(plugins))
	}
}

func TestAPI_GetHandler(t *testing.T) {
	store := NewStore()
	api := NewAPI(store)

	store.Register(&PluginInstance{
		Descriptor: &contract.PluginDescriptor{
			Name:       "known",
			Version:    "1.0.0",
			ApiVersion: "v1",
		},
		HttpAddress: "localhost:8081",
	})

	ts := httptest.NewServer(http.HandlerFunc(api.ServeHTTP))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/plugins/known")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var got PluginInstance
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}
	if got.Descriptor.Name != "known" {
		t.Errorf("Name = %q, want %q", got.Descriptor.Name, "known")
	}
}

func TestAPI_GetHandler_NotFound(t *testing.T) {
	store := NewStore()
	api := NewAPI(store)

	ts := httptest.NewServer(http.HandlerFunc(api.ServeHTTP))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/plugins/unknown")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestAPI_RegisterHandler(t *testing.T) {
	store := NewStore()
	api := NewAPI(store)

	ts := httptest.NewServer(http.HandlerFunc(api.ServeHTTP))
	defer ts.Close()

	req := &contract.RegisterRequest{
		Name:        "new-plugin",
		Version:     "1.0.0",
		ApiVersion:  "v1",
		Description: "A new plugin",
		HttpAddress: "new:8081",
		GrpcAddress: "new:50051",
		Capabilities: []*contract.Capability{
			{Name: "feature", Description: "A feature", Version: "1.0.0"},
		},
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(ts.URL+"/api/v1/plugins/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	inst, ok := store.Get("new-plugin")
	if !ok {
		t.Fatal("expected plugin to be in store")
	}
	if !inst.PushRegistered {
		t.Error("expected plugin to be marked as push registered")
	}
	if len(inst.Descriptor.Capabilities) != 1 {
		t.Errorf("len(Capabilities) = %d, want 1", len(inst.Descriptor.Capabilities))
	}
}

func TestAPI_RegisterHandler_MissingName(t *testing.T) {
	store := NewStore()
	api := NewAPI(store)

	ts := httptest.NewServer(http.HandlerFunc(api.ServeHTTP))
	defer ts.Close()

	req := &contract.RegisterRequest{Version: "1.0.0"}
	body, _ := json.Marshal(req)
	resp, err := http.Post(ts.URL+"/api/v1/plugins/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestAPI_DeleteHandler(t *testing.T) {
	store := NewStore()
	api := NewAPI(store)

	store.Register(&PluginInstance{
		Descriptor: &contract.PluginDescriptor{Name: "to-delete", Version: "1.0.0"},
	})

	ts := httptest.NewServer(http.HandlerFunc(api.ServeHTTP))
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/plugins/to-delete", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if _, ok := store.Get("to-delete"); ok {
		t.Error("expected plugin to be deleted")
	}
}

func TestAPI_DeleteHandler_NotFound(t *testing.T) {
	store := NewStore()
	api := NewAPI(store)

	ts := httptest.NewServer(http.HandlerFunc(api.ServeHTTP))
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/plugins/nonexistent", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestSidecarClient_ListContainers(t *testing.T) {
	sidecarSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("label") != "ade.plugin" {
			t.Errorf("label param = %q, want %q", r.URL.Query().Get("label"), "ade.plugin")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"containers": []map[string]interface{}{
				{
					"id":    "abc123",
					"name":  "templates-plugin",
					"state": "running",
					"labels": map[string]string{
						"ade.plugin.name":      "templates",
						"ade.plugin.version":   "1.0.0",
						"ade.plugin.http-port": "8081",
						"ade.plugin.grpc-port": "50051",
					},
				},
			},
		})
	}))
	defer sidecarSrv.Close()

	client := NewSidecarClient(sidecarSrv.URL)
	containers, err := client.ListContainers(context.Background(), "ade.plugin")
	if err != nil {
		t.Fatalf("ListContainers failed: %v", err)
	}

	if len(containers) != 1 {
		t.Fatalf("len = %d, want 1", len(containers))
	}
	if containers[0].Name != "templates-plugin" {
		t.Errorf("Name = %q, want %q", containers[0].Name, "templates-plugin")
	}
	if containers[0].Labels["ade.plugin.name"] != "templates" {
		t.Errorf("label = %q, want %q", containers[0].Labels["ade.plugin.name"], "templates")
	}
}

func TestDockerDiscoverer_Discover(t *testing.T) {
	sidecarSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"containers": []map[string]interface{}{
				{
					"id":    "d1",
					"name":  "plugin-one",
					"state": "running",
					"labels": map[string]string{
						"ade.plugin.name":      "one",
						"ade.plugin.version":   "1.0.0",
						"ade.plugin.http-port": "8081",
					},
				},
				{
					"id":    "d2",
					"name":  "plugin-two",
					"state": "running",
					"labels": map[string]string{
						"ade.plugin.name":      "two",
						"ade.plugin.http-port": "8082",
					},
				},
			},
		})
	}))
	defer sidecarSrv.Close()

	store := NewStore()
	client := NewSidecarClient(sidecarSrv.URL)
	discoverer := NewDockerDiscoverer(client, store, 1*time.Hour)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = discoverer.discover(ctx)

	plugins := store.List()
	if len(plugins) != 2 {
		t.Fatalf("len = %d, want 2", len(plugins))
	}

	p1, ok := store.Get("one")
	if !ok {
		t.Fatal("expected plugin 'one'")
	}
	if p1.Descriptor.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", p1.Descriptor.Version, "1.0.0")
	}

	p2, ok := store.Get("two")
	if !ok {
		t.Fatal("expected plugin 'two'")
	}
	if p2.Descriptor.Version != "" {
		t.Errorf("Version = %q, want empty", p2.Descriptor.Version)
	}
}

func TestDockerDiscoverer_SkipsPushRegistered(t *testing.T) {
	sidecarSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"containers": []map[string]interface{}{
				{
					"id":    "d1",
					"name":  "existing-plugin",
					"state": "running",
					"labels": map[string]string{
						"ade.plugin.name": "existing",
					},
				},
			},
		})
	}))
	defer sidecarSrv.Close()

	store := NewStore()
	store.Register(&PluginInstance{
		Descriptor: &contract.PluginDescriptor{
			Name:       "existing",
			Version:    "2.0.0",
			ApiVersion: "v1",
			Capabilities: []*contract.Capability{
				{Name: "custom", Version: "1.0.0"},
			},
		},
		HttpAddress:    "existing:8081",
		PushRegistered: true,
	})

	client := NewSidecarClient(sidecarSrv.URL)
	discoverer := NewDockerDiscoverer(client, store, 1*time.Hour)

	_ = discoverer.discover(context.Background())

	p, _ := store.Get("existing")
	if p.Descriptor.Version != "2.0.0" {
		t.Errorf("Version = %q, want '2.0.0' (should not be overwritten)", p.Descriptor.Version)
	}
	if len(p.Descriptor.Capabilities) != 1 {
		t.Errorf("len(Capabilities) = %d, want 1 (should not be overwritten)", len(p.Descriptor.Capabilities))
	}
}

func TestServer_MuxWiring(t *testing.T) {
	store := NewStore()
	api := NewAPI(store)

	mux := http.NewServeMux()
	mux.Handle("/api/v1/plugins/", api)
	mux.Handle("/api/v1/plugins", api)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	store.Register(&PluginInstance{
		Descriptor: &contract.PluginDescriptor{Name: "mux-test", Version: "1.0.0"},
	})

	resp, err := http.Get(ts.URL + "/api/v1/plugins")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("list Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result map[string][]*PluginInstance
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}

	plugins := result["plugins"]
	if len(plugins) != 1 {
		t.Fatalf("len(plugins) = %d, want 1", len(plugins))
	}
	if plugins[0].Descriptor.Name != "mux-test" {
		t.Errorf("Name = %q, want %q", plugins[0].Descriptor.Name, "mux-test")
	}
}
