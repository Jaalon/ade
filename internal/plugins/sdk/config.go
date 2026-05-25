package sdk

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Name             string
	Version          string
	Description      string
	OrchestratorURL  string
	GRPCPort         int
	HTTPPort         int
	RegisterInterval time.Duration
}

func LoadConfigFromEnv() (Config, error) {
	cfg := Config{
		GRPCPort:         50051,
		HTTPPort:         8081,
		RegisterInterval: 30 * time.Second,
	}

	cfg.Name = os.Getenv("PLUGIN_NAME")
	if cfg.Name == "" {
		return cfg, fmt.Errorf("%w: PLUGIN_NAME is required", ErrInvalidConfig)
	}

	cfg.Version = os.Getenv("PLUGIN_VERSION")
	if cfg.Version == "" {
		return cfg, fmt.Errorf("%w: PLUGIN_VERSION is required", ErrInvalidConfig)
	}

	cfg.Description = os.Getenv("PLUGIN_DESCRIPTION")

	cfg.OrchestratorURL = os.Getenv("PLUGIN_ORCHESTRATOR_URL")
	if cfg.OrchestratorURL == "" {
		return cfg, fmt.Errorf("%w: PLUGIN_ORCHESTRATOR_URL is required", ErrInvalidConfig)
	}

	if v := os.Getenv("PLUGIN_GRPC_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return cfg, fmt.Errorf("%w: PLUGIN_GRPC_PORT=%q is not a number", ErrInvalidConfig, v)
		}
		cfg.GRPCPort = p
	}

	if v := os.Getenv("PLUGIN_HTTP_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return cfg, fmt.Errorf("%w: PLUGIN_HTTP_PORT=%q is not a number", ErrInvalidConfig, v)
		}
		cfg.HTTPPort = p
	}

	if v := os.Getenv("PLUGIN_REGISTER_INTERVAL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return cfg, fmt.Errorf("%w: PLUGIN_REGISTER_INTERVAL=%q is not a duration", ErrInvalidConfig, v)
		}
		cfg.RegisterInterval = d
	}

	return cfg, nil
}
