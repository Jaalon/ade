package agentic

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"automated_dev_environment/internal/config"
)

func TestMergeMCPServers_EmptyIncoming(t *testing.T) {
	existing := []config.MCPServer{{Name: "fs", Command: "npx"}}
	merged := mergeMCPServers(existing, nil)
	assert.Len(t, merged, 1)
	assert.Equal(t, "fs", merged[0].Name)
}

func TestMergeMCPServers_NewServer(t *testing.T) {
	existing := []config.MCPServer{{Name: "fs", Command: "npx"}}
	incoming := []config.MCPServer{{Name: "github", Command: "npx"}}
	merged := mergeMCPServers(existing, incoming)
	assert.Len(t, merged, 2)
}

func TestMergeMCPServers_Override(t *testing.T) {
	existing := []config.MCPServer{{Name: "fs", Command: "old"}}
	incoming := []config.MCPServer{{Name: "fs", Command: "new"}}
	merged := mergeMCPServers(existing, incoming)
	assert.Len(t, merged, 1)
	assert.Equal(t, "new", merged[0].Command)
}

func TestMergeMCPServers_CaseInsensitiveOverride(t *testing.T) {
	existing := []config.MCPServer{{Name: "FS", Command: "old"}}
	incoming := []config.MCPServer{{Name: "fs", Command: "new"}}
	merged := mergeMCPServers(existing, incoming)
	assert.Len(t, merged, 1)
	assert.Equal(t, "new", merged[0].Command)
}

func TestConfigureMCPServers_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	yamlContent := `
mcp_servers:
  - name: "filesystem"
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-filesystem", "."]
  - name: "github"
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-github"]
    env:
      GITHUB_TOKEN: "${GITHUB_TOKEN}"
`
	yamlPath := filepath.Join(dir, "ade-config.yaml")
	err := os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	assert.NoError(t, err)

	report, err := ConfigureMCPServers(context.Background(), MCPOptions{
		OutputDir:  dir,
		ConfigPath: yamlPath,
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, report.Added)
	assert.Equal(t, 0, report.Updated)
	assert.Equal(t, 0, report.Unchanged)
	assert.Equal(t, 0, report.Errors)

	configPath := filepath.Join(dir, ".opencode", "config.json")
	data, err := os.ReadFile(configPath)
	assert.NoError(t, err)

	var mcpCfg MCPConfig
	err = json.Unmarshal(data, &mcpCfg)
	assert.NoError(t, err)
	assert.Len(t, mcpCfg.MCPServers, 2)
	assert.Equal(t, ".opencode/skills", mcpCfg.SkillsPath)
}

func TestConfigureMCPServers_ExistingConfigNewServer(t *testing.T) {
	dir := t.TempDir()

	existingCfg := MCPConfig{
		MCPServers: []config.MCPServer{{Name: "fs", Command: "npx"}},
		SkillsPath: ".opencode/skills",
	}
	cfgDir := filepath.Join(dir, ".opencode")
	os.MkdirAll(cfgDir, 0755)
	cfgData, _ := json.Marshal(existingCfg)
	os.WriteFile(filepath.Join(cfgDir, "config.json"), cfgData, 0644)

	yamlContent := `
mcp_servers:
  - name: "github"
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-github"]
`
	yamlPath := filepath.Join(dir, "ade-config.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	report, err := ConfigureMCPServers(context.Background(), MCPOptions{
		OutputDir:  dir,
		ConfigPath: yamlPath,
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, report.Added)
	assert.Equal(t, 0, report.Updated)
	assert.Equal(t, 1, report.Unchanged)
	assert.Equal(t, 0, report.Errors)
}

func TestConfigureMCPServers_OverrideExisting(t *testing.T) {
	dir := t.TempDir()

	existingCfg := MCPConfig{
		MCPServers: []config.MCPServer{{Name: "fs", Command: "old"}},
		SkillsPath: ".opencode/skills",
	}
	cfgDir := filepath.Join(dir, ".opencode")
	os.MkdirAll(cfgDir, 0755)
	cfgData, _ := json.Marshal(existingCfg)
	os.WriteFile(filepath.Join(cfgDir, "config.json"), cfgData, 0644)

	yamlContent := `
mcp_servers:
  - name: "fs"
    command: "npx"
    args: ["-y", "new-args"]
`
	yamlPath := filepath.Join(dir, "ade-config.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	report, err := ConfigureMCPServers(context.Background(), MCPOptions{
		OutputDir:  dir,
		ConfigPath: yamlPath,
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, report.Added)
	assert.Equal(t, 1, report.Updated)
	assert.Equal(t, 0, report.Unchanged)
	assert.Equal(t, 0, report.Errors)
}

func TestConfigureMCPServers_NoYamlConfig(t *testing.T) {
	dir := t.TempDir()

	report, err := ConfigureMCPServers(context.Background(), MCPOptions{
		OutputDir: dir,
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, report.Added)
	assert.Equal(t, 0, report.Updated)
	assert.Equal(t, 0, report.Unchanged)
	assert.Equal(t, 0, report.Errors)
}

func TestLoadOpenCodeConfig_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".opencode")
	os.MkdirAll(cfgDir, 0755)
	os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte("{invalid json}"), 0644)

	_, err := loadOpenCodeConfig(dir)
	assert.Error(t, err)
}

func TestConfigureMCPServers_SkillsPathPreserved(t *testing.T) {
	dir := t.TempDir()

	yamlContent := `
mcp_servers:
  - name: "fs"
    command: "npx"
`
	yamlPath := filepath.Join(dir, "ade-config.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	_, err := ConfigureMCPServers(context.Background(), MCPOptions{
		OutputDir:  dir,
		ConfigPath: yamlPath,
	})
	assert.NoError(t, err)

	configPath := filepath.Join(dir, ".opencode", "config.json")
	data, _ := os.ReadFile(configPath)
	var mcpCfg MCPConfig
	json.Unmarshal(data, &mcpCfg)
	assert.Equal(t, ".opencode/skills", mcpCfg.SkillsPath)
}
