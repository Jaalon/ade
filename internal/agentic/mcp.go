package agentic

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"

	"automated_dev_environment/internal/config"
)

type MCPConfig struct {
	MCPServers []config.MCPServer `json:"mcp_servers"`
	SkillsPath string             `json:"skills_path"`
}

type MCPOptions struct {
	OutputDir  string
	ConfigPath string
}

type MCPResult struct {
	ServerName string
	Action     string
	Error      error
}

type MCPReport struct {
	Results   []MCPResult
	Added     int
	Updated   int
	Unchanged int
	Errors    int
}

func ConfigureMCPServers(ctx context.Context, opts MCPOptions) (*MCPReport, error) {
	report := &MCPReport{}

	cfg, err := config.LoadConfig(opts.ConfigPath)
	if err != nil {
		return report, err
	}

	if len(cfg.MCPServers) == 0 {
		return report, nil
	}

	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = "."
	}

	existing, _ := loadOpenCodeConfig(outputDir)

	existingByName := make(map[string]bool)
	for _, s := range existing.MCPServers {
		existingByName[strings.ToLower(s.Name)] = true
	}

	for _, s := range cfg.MCPServers {
		lower := strings.ToLower(s.Name)
		if existingByName[lower] {
			report.Results = append(report.Results, MCPResult{ServerName: s.Name, Action: "updated"})
			report.Updated++
		} else {
			report.Results = append(report.Results, MCPResult{ServerName: s.Name, Action: "added"})
			report.Added++
		}
	}

	incomingByName := make(map[string]bool)
	for _, s := range cfg.MCPServers {
		incomingByName[strings.ToLower(s.Name)] = true
	}
	for _, s := range existing.MCPServers {
		if !incomingByName[strings.ToLower(s.Name)] {
			report.Results = append(report.Results, MCPResult{ServerName: s.Name, Action: "unchanged"})
			report.Unchanged++
		}
	}

	merged := mergeMCPServers(existing.MCPServers, cfg.MCPServers)
	mcpCfg := &MCPConfig{
		MCPServers: merged,
		SkillsPath: ".opencode/skills",
	}

	if err := osMkdirAll(filepath.Join(outputDir, ".opencode"), 0755); err != nil {
		report.Errors++
		return report, err
	}

	if err := saveOpenCodeConfig(outputDir, mcpCfg); err != nil {
		report.Errors++
		return report, err
	}

	return report, nil
}

func loadOpenCodeConfig(projectDir string) (*MCPConfig, error) {
	path := filepath.Join(projectDir, ".opencode", "config.json")
	data, err := osReadFile(path)
	if err != nil {
		return &MCPConfig{}, err
	}

	var cfg MCPConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &MCPConfig{}, err
	}
	return &cfg, nil
}

func saveOpenCodeConfig(projectDir string, cfg *MCPConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(projectDir, ".opencode", "config.json")
	return osWriteFile(path, data, 0644)
}

func mergeMCPServers(existing, incoming []config.MCPServer) []config.MCPServer {
	merged := make(map[string]config.MCPServer)
	for _, s := range existing {
		merged[strings.ToLower(s.Name)] = s
	}
	for _, s := range incoming {
		merged[strings.ToLower(s.Name)] = s
	}

	result := make([]config.MCPServer, 0, len(merged))
	for _, s := range merged {
		result = append(result, s)
	}
	return result
}
