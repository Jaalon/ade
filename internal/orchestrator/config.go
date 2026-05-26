package orchestrator

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Config struct {
	RESTPort          int
	GRPCPort          int
	DataDir           string
	DiscoveryInterval time.Duration
	HealthInterval    time.Duration
	MaxHealthFails    int
	LogLevel          string
	DevMode           bool
	FrontendDir       string
}

func DefaultConfig() Config {
	return Config{
		RESTPort:          8080,
		GRPCPort:          9090,
		DataDir:           "/data",
		DiscoveryInterval: 30 * time.Second,
		HealthInterval:    15 * time.Second,
		MaxHealthFails:    3,
		LogLevel:          "info",
		FrontendDir:       "internal/orchestrator/frontend/dist",
	}
}

func ConfigFromEnv() Config {
	cfg := DefaultConfig()

	if v := os.Getenv("ADE_CONFIG_REST_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.RESTPort = p
		}
	}

	if v := os.Getenv("ADE_CONFIG_GRPC_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.GRPCPort = p
		}
	}

	if v := os.Getenv("ADE_CONFIG_DATA_DIR"); v != "" {
		cfg.DataDir = v
	}

	if v := os.Getenv("ADE_CONFIG_DISCOVERY_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.DiscoveryInterval = d
		}
	}

	if v := os.Getenv("ADE_CONFIG_HEALTH_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.HealthInterval = d
		}
	}

	if v := os.Getenv("ADE_CONFIG_MAX_HEALTH_FAILS"); v != "" {
		if f, err := strconv.Atoi(v); err == nil {
			cfg.MaxHealthFails = f
		}
	}

	if v := os.Getenv("ADE_LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}

	cfg.DevMode = os.Getenv("ADE_DEV_MODE") == "true"

	if v := os.Getenv("ADE_FRONTEND_DIR"); v != "" {
		cfg.FrontendDir = v
	}

	return cfg
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	s.settingsMu.RLock()
	settings := make(map[string]string)
	for k, v := range s.runtimeSettings {
		settings[k] = v
	}
	s.settingsMu.RUnlock()

	cfg := ConfigResponse{
		ProjectName:         os.Getenv("ADE_PROJECT_NAME"),
		OrchestratorVersion: Version,
		RESTPort:            s.cfg.RESTPort,
		GRPCPort:            s.cfg.GRPCPort,
		Settings:            settings,
	}
	if cfg.ProjectName == "" {
		cfg.ProjectName = "default"
	}

	writeJSON(w, http.StatusOK, cfg)
}

func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Settings map[string]string `json:"settings"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "corps de requête invalide")
		return
	}

	s.settingsMu.Lock()
	if body.Settings != nil {
		for k, v := range body.Settings {
			s.runtimeSettings[k] = v
		}
	}
	s.settingsMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]string{"message": "configuration mise à jour"})
}
