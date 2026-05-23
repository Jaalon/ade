package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_ValidFile(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `
tools:
  opencode:
    path: "C:\\opencode\\opencode.exe"
  cursor:
    path: "C:\\cursor\\cursor.exe"
mcp_servers:
  - name: "filesystem"
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-filesystem", "."]
`
	yamlPath := filepath.Join(dir, "ade-config.yaml")
	err := os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	assert.NoError(t, err)

	cfg, err := LoadConfig(yamlPath)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, `C:\opencode\opencode.exe`, cfg.Tools[ToolOpenCode].Path)
	assert.Equal(t, `C:\cursor\cursor.exe`, cfg.Tools[ToolCursor].Path)
	assert.Len(t, cfg.MCPServers, 1)
	assert.Equal(t, "filesystem", cfg.MCPServers[0].Name)
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	cfg, err := LoadConfig("./nonexistent-file.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.Tools)
}

func TestLoadConfig_EmptyPath(t *testing.T) {
	cfg, err := LoadConfig("")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "ade-config.yaml")
	err := os.WriteFile(yamlPath, []byte("{invalid yaml: [broken}"), 0644)
	assert.NoError(t, err)

	cfg, err := LoadConfig(yamlPath)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.Tools)
}

func TestLoadConfig_PartialConfig(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, ".ade.yaml")
	err := os.WriteFile(yamlPath, []byte("tools:\n  opencode:\n    path: custom.exe\n"), 0644)
	assert.NoError(t, err)

	cfg, err := LoadConfig(yamlPath)
	assert.NoError(t, err)
	assert.Equal(t, "custom.exe", cfg.Tools[ToolOpenCode].Path)
	assert.Nil(t, cfg.MCPServers)
}

func TestFindConfigPath_Found(t *testing.T) {
	dir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origWd)

	yamlPath := filepath.Join(dir, "ade-config.yaml")
	os.WriteFile(yamlPath, []byte("tools: {}"), 0644)

	path := FindConfigPath()
	assert.Equal(t, "ade-config.yaml", path)
}

func TestFindConfigPath_NotFound(t *testing.T) {
	dir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origWd)

	path := FindConfigPath()
	assert.Empty(t, path)
}

func TestToolTypeValues(t *testing.T) {
	assert.Equal(t, ToolType("opencode"), ToolOpenCode)
	assert.Equal(t, ToolType("cursor"), ToolCursor)
}
