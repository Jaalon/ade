package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ToolType string

const (
	ToolOpenCode ToolType = "opencode"
	ToolCursor   ToolType = "cursor"
)

type AgenticConfig struct {
	Tools      map[ToolType]ToolConfig `yaml:"tools"`
	MCPServers []MCPServer             `yaml:"mcp_servers"`
}

type ToolConfig struct {
	Path string `yaml:"path"`
}

type MCPServer struct {
	Name    string            `yaml:"name"`
	Command string            `yaml:"command"`
	Args    []string          `yaml:"args"`
	Env     map[string]string `yaml:"env"`
}

var osStat = os.Stat
var osReadFile = os.ReadFile

func LoadConfig(path string) (*AgenticConfig, error) {
	if path == "" {
		path = FindConfigPath()
	}
	if path == "" {
		return &AgenticConfig{}, nil
	}

	data, err := osReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &AgenticConfig{}, nil
		}
		return &AgenticConfig{}, nil
	}

	var cfg AgenticConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return &AgenticConfig{}, nil
	}
	if cfg.Tools == nil {
		cfg.Tools = make(map[ToolType]ToolConfig)
	}
	return &cfg, nil
}

func FindConfigPath() string {
	candidates := []string{
		"ade-config.yaml",
		".ade.yaml",
	}
	for _, c := range candidates {
		if _, err := osStat(c); err == nil {
			return c
		}
	}

	userProfile := os.Getenv("USERPROFILE")
	if userProfile != "" {
		globalPath := filepath.Join(userProfile, ".ade", "config.yaml")
		if _, err := osStat(globalPath); err == nil {
			return globalPath
		}
	}
	return ""
}
