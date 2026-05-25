package sdk

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"automated_dev_environment/internal/plugins/contract"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestLoadConfigFromEnv_AllVars(t *testing.T) {
	t.Setenv("PLUGIN_NAME", "test-plugin")
	t.Setenv("PLUGIN_VERSION", "1.0.0")
	t.Setenv("PLUGIN_DESCRIPTION", "A test plugin")
	t.Setenv("PLUGIN_ORCHESTRATOR_URL", "http://orchestrator:8082")
	t.Setenv("PLUGIN_GRPC_PORT", "50052")
	t.Setenv("PLUGIN_HTTP_PORT", "8082")
	t.Setenv("PLUGIN_REGISTER_INTERVAL", "15s")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadConfigFromEnv failed: %v", err)
	}

	if cfg.Name != "test-plugin" {
		t.Errorf("Name = %q, want %q", cfg.Name, "test-plugin")
	}
	if cfg.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", cfg.Version, "1.0.0")
	}
	if cfg.Description != "A test plugin" {
		t.Errorf("Description = %q, want %q", cfg.Description, "A test plugin")
	}
	if cfg.OrchestratorURL != "http://orchestrator:8082" {
		t.Errorf("OrchestratorURL = %q, want %q", cfg.OrchestratorURL, "http://orchestrator:8082")
	}
	if cfg.GRPCPort != 50052 {
		t.Errorf("GRPCPort = %d, want %d", cfg.GRPCPort, 50052)
	}
	if cfg.HTTPPort != 8082 {
		t.Errorf("HTTPPort = %d, want %d", cfg.HTTPPort, 8082)
	}
	if cfg.RegisterInterval != 15*time.Second {
		t.Errorf("RegisterInterval = %v, want %v", cfg.RegisterInterval, 15*time.Second)
	}
}

func TestLoadConfigFromEnv_MissingName(t *testing.T) {
	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Fatal("expected error for missing PLUGIN_NAME")
	}
}

func TestLoadConfigFromEnv_Defaults(t *testing.T) {
	t.Setenv("PLUGIN_NAME", "defaults-test")
	t.Setenv("PLUGIN_VERSION", "0.1.0")
	t.Setenv("PLUGIN_ORCHESTRATOR_URL", "http://orchestrator:8082")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadConfigFromEnv failed: %v", err)
	}

	if cfg.GRPCPort != 50051 {
		t.Errorf("default GRPCPort = %d, want %d", cfg.GRPCPort, 50051)
	}
	if cfg.HTTPPort != 8081 {
		t.Errorf("default HTTPPort = %d, want %d", cfg.HTTPPort, 8081)
	}
	if cfg.RegisterInterval != 30*time.Second {
		t.Errorf("default RegisterInterval = %v, want %v", cfg.RegisterInterval, 30*time.Second)
	}
}

func TestNewPlugin_WithOptions(t *testing.T) {
	t.Setenv("PLUGIN_NAME", "opts-test")
	t.Setenv("PLUGIN_VERSION", "2.0.0")
	t.Setenv("PLUGIN_ORCHESTRATOR_URL", "http://orchestrator:8082")

	p, err := NewPlugin()
	if err != nil {
		t.Fatalf("NewPlugin failed: %v", err)
	}

	if p.cfg.Name != "opts-test" {
		t.Errorf("Name = %q, want %q", p.cfg.Name, "opts-test")
	}
	if p.cfg.Version != "2.0.0" {
		t.Errorf("Version = %q, want %q", p.cfg.Version, "2.0.0")
	}
}

func TestAddCapability(t *testing.T) {
	t.Setenv("PLUGIN_NAME", "caps-test")
	t.Setenv("PLUGIN_VERSION", "1.0.0")
	t.Setenv("PLUGIN_ORCHESTRATOR_URL", "http://orchestrator:8082")

	p, err := NewPlugin()
	if err != nil {
		t.Fatalf("NewPlugin failed: %v", err)
	}

	p.AddCapability(&contract.Capability{
		Name:        "greeter",
		Description: "Greets the world",
		Version:     "1.0.0",
	})

	caps := p.caps.list()
	if len(caps) != 1 {
		t.Fatalf("len(caps) = %d, want 1", len(caps))
	}
	if caps[0].Name != "greeter" {
		t.Errorf("caps[0].Name = %q, want %q", caps[0].Name, "greeter")
	}
}

func TestAddCapability_Multiple(t *testing.T) {
	t.Setenv("PLUGIN_NAME", "multi-caps")
	t.Setenv("PLUGIN_VERSION", "1.0.0")
	t.Setenv("PLUGIN_ORCHESTRATOR_URL", "http://orchestrator:8082")

	p, err := NewPlugin()
	if err != nil {
		t.Fatalf("NewPlugin failed: %v", err)
	}

	p.AddCapability(&contract.Capability{Name: "cap-a", Version: "1.0.0"})
	p.AddCapability(&contract.Capability{Name: "cap-b", Version: "2.0.0"})
	p.AddCapability(&contract.Capability{Name: "cap-c", Version: "3.0.0"})

	caps := p.caps.list()
	if len(caps) != 3 {
		t.Fatalf("len(caps) = %d, want 3", len(caps))
	}
}

func TestHandleFunc_ServesAtPath(t *testing.T) {
	t.Setenv("PLUGIN_NAME", "handler-test")
	t.Setenv("PLUGIN_VERSION", "1.0.0")
	t.Setenv("PLUGIN_ORCHESTRATOR_URL", "http://orchestrator:8082")

	p, err := NewPlugin()
	if err != nil {
		t.Fatalf("NewPlugin failed: %v", err)
	}

	p.HandleFunc("/api/v1/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok": true}`))
	})

	ts := httptest.NewServer(p.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/test")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]bool
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}
	if !body["ok"] {
		t.Errorf("body[ok] = false, want true")
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	ts := httptest.NewServer(recoveryMiddleware(handler))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestHealthEndpoint(t *testing.T) {
	t.Setenv("PLUGIN_NAME", "health-test")
	t.Setenv("PLUGIN_VERSION", "1.0.0")
	t.Setenv("PLUGIN_ORCHESTRATOR_URL", "http://orchestrator:8082")

	p, err := NewPlugin()
	if err != nil {
		t.Fatalf("NewPlugin failed: %v", err)
	}

	ts := httptest.NewServer(p.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var hs contract.HealthStatus
	if err := json.NewDecoder(resp.Body).Decode(&hs); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}
	if hs.Status != contract.HealthHealthy {
		t.Errorf("Status = %d, want %d", hs.Status, contract.HealthHealthy)
	}
}

func TestCapabilitiesEndpoint(t *testing.T) {
	t.Setenv("PLUGIN_NAME", "caps-http")
	t.Setenv("PLUGIN_VERSION", "1.0.0")
	t.Setenv("PLUGIN_ORCHESTRATOR_URL", "http://orchestrator:8082")

	p, err := NewPlugin()
	if err != nil {
		t.Fatalf("NewPlugin failed: %v", err)
	}

	p.AddCapability(&contract.Capability{Name: "greeter", Description: "Greets", Version: "1.0.0"})
	p.AddCapability(&contract.Capability{Name: "reporter", Description: "Reports", Version: "0.5.0"})

	ts := httptest.NewServer(p.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/capabilities")
	if err != nil {
		t.Fatalf("GET /capabilities failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result map[string][]*contract.Capability
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}

	caps, ok := result["capabilities"]
	if !ok {
		t.Fatal("response missing 'capabilities' key")
	}
	if len(caps) != 2 {
		t.Fatalf("len(capabilities) = %d, want 2", len(caps))
	}
	if caps[0].Name != "greeter" {
		t.Errorf("caps[0].Name = %q, want %q", caps[0].Name, "greeter")
	}
}

func TestGRPCHealthCheck(t *testing.T) {
	t.Setenv("PLUGIN_NAME", "grpc-health")
	t.Setenv("PLUGIN_VERSION", "1.0.0")
	t.Setenv("PLUGIN_ORCHESTRATOR_URL", "http://orchestrator:8082")

	p, err := NewPlugin(WithGRPCOptions(grpc.Creds(insecure.NewCredentials())))
	if err != nil {
		t.Fatalf("NewPlugin failed: %v", err)
	}

	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	port := lis.Addr().(*net.TCPAddr).Port

	go p.grpcSrv.Serve(lis)
	defer p.grpcSrv.GracefulStop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc dial failed: %v", err)
	}
	defer conn.Close()

	client := contract.NewPluginServiceClient(conn)
	resp, err := client.HealthCheck(context.Background(), &contract.Empty{})
	if err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}

	if resp.Status != contract.HealthCheckResponse_HEALTHY {
		t.Errorf("Status = %v, want HEALTHY", resp.Status)
	}
	if resp.Message != "ok" {
		t.Errorf("Message = %q, want %q", resp.Message, "ok")
	}
	_ = port
}

func TestGRPCGetCapabilities(t *testing.T) {
	t.Setenv("PLUGIN_NAME", "grpc-caps")
	t.Setenv("PLUGIN_VERSION", "1.0.0")
	t.Setenv("PLUGIN_ORCHESTRATOR_URL", "http://orchestrator:8082")

	p, err := NewPlugin(WithGRPCOptions(grpc.Creds(insecure.NewCredentials())))
	if err != nil {
		t.Fatalf("NewPlugin failed: %v", err)
	}

	p.AddCapability(&contract.Capability{Name: "grpc_feature", Version: "1.0.0"})

	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	go p.grpcSrv.Serve(lis)
	defer p.grpcSrv.GracefulStop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc dial failed: %v", err)
	}
	defer conn.Close()

	client := contract.NewPluginServiceClient(conn)
	resp, err := client.GetCapabilities(context.Background(), &contract.Empty{})
	if err != nil {
		t.Fatalf("GetCapabilities failed: %v", err)
	}

	if len(resp.Capabilities) != 1 {
		t.Fatalf("len(Capabilities) = %d, want 1", len(resp.Capabilities))
	}
	if resp.Capabilities[0].Name != "grpc_feature" {
		t.Errorf("Capabilities[0].Name = %q, want %q", resp.Capabilities[0].Name, "grpc_feature")
	}
}

func TestRegistration_RegisterNow(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/v1/plugins/register" {
			t.Errorf("Path = %s, want /api/v1/plugins/register", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(contract.RegisterResponse{Accepted: true})
	}))
	defer srv.Close()

	desc := &contract.PluginDescriptor{
		Name:       "reg-test",
		Version:    "1.0.0",
		ApiVersion: "v1",
		Endpoints: map[string]string{
			"grpc": ":50051",
			"http": ":8081",
		},
	}

	reg := NewRegistration(desc, srv.URL, 10*time.Minute)
	err := reg.RegisterNow(context.Background())
	if err != nil {
		t.Fatalf("RegisterNow failed: %v", err)
	}
}

func TestRegistration_RegisterNow_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	desc := &contract.PluginDescriptor{
		Name:       "reg-fail",
		Version:    "1.0.0",
		ApiVersion: "v1",
	}

	reg := NewRegistration(desc, srv.URL, 10*time.Minute)
	err := reg.RegisterNow(context.Background())
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestRegistration_Stop(t *testing.T) {
	desc := &contract.PluginDescriptor{
		Name:       "stop-test",
		Version:    "1.0.0",
		ApiVersion: "v1",
	}

	reg := NewRegistration(desc, "http://localhost:1", 10*time.Minute)
	reg.Start(context.Background())
	reg.Stop()
	reg.Stop()
}

func TestCORSMiddleware(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ts := httptest.NewServer(handler)
	defer ts.Close()

	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodOptions, ts.URL+"/", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("OPTIONS failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("OPTIONS Status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}

	if origin := resp.Header.Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", origin, "*")
	}
}

func TestRegistration_StartAndStop(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	desc := &contract.PluginDescriptor{
		Name:       "start-stop",
		Version:    "1.0.0",
		ApiVersion: "v1",
		Endpoints:  map[string]string{"grpc": ":50051", "http": ":8081"},
	}

	reg := NewRegistration(desc, srv.URL, 50*time.Millisecond)
	reg.Start(context.Background())

	time.Sleep(120 * time.Millisecond)
	reg.Stop()

	if callCount < 2 {
		t.Errorf("Expected at least 2 registration calls (initial + 1 retry), got %d", callCount)
	}
}
