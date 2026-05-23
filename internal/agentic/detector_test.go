package agentic

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"automated_dev_environment/internal/config"
)

func TestFindTool_FoundInPath(t *testing.T) {
	savedLookPath := execLookPath
	defer func() { execLookPath = savedLookPath }()
	execLookPath = func(name string) (string, error) {
		if name == "opencode" {
			return "opencode", nil
		}
		return "", fmt.Errorf("not found")
	}

	savedStat := osStat
	defer func() { osStat = savedStat }()
	osStat = func(name string) (os.FileInfo, error) {
		return nil, fmt.Errorf("not found")
	}

	info := FindTool(config.ToolOpenCode, &config.AgenticConfig{})
	assert.True(t, info.Found)
	assert.Equal(t, "opencode", info.Path)
	assert.Equal(t, "OpenCode", info.Name)
}

func TestFindTool_FoundInDefaultPath(t *testing.T) {
	savedLookPath := execLookPath
	defer func() { execLookPath = savedLookPath }()
	execLookPath = func(name string) (string, error) {
		return "", fmt.Errorf("not found")
	}

	savedStat := osStat
	defer func() { osStat = savedStat }()

	expectedPath := filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "opencode", "opencode.exe")
	osStat = func(name string) (os.FileInfo, error) {
		if name == expectedPath {
			return nil, nil
		}
		return nil, fmt.Errorf("not found")
	}

	info := FindTool(config.ToolOpenCode, &config.AgenticConfig{})
	assert.True(t, info.Found)
	assert.Equal(t, expectedPath, info.Path)
}

func TestFindTool_YamlOverride(t *testing.T) {
	savedLookPath := execLookPath
	defer func() { execLookPath = savedLookPath }()
	execLookPath = func(name string) (string, error) {
		return "opencode", nil
	}

	savedStat := osStat
	defer func() { osStat = savedStat }()
	osStat = func(name string) (os.FileInfo, error) {
		if name == "C:\\custom\\opencode.exe" {
			return nil, nil
		}
		return nil, fmt.Errorf("not found")
	}

	cfg := &config.AgenticConfig{
		Tools: map[config.ToolType]config.ToolConfig{
			config.ToolOpenCode: {Path: "C:\\custom\\opencode.exe"},
		},
	}

	info := FindTool(config.ToolOpenCode, cfg)
	assert.True(t, info.Found)
	assert.Equal(t, "C:\\custom\\opencode.exe", info.Path)
}

func TestFindTool_YamlOverrideNotFound(t *testing.T) {
	savedLookPath := execLookPath
	defer func() { execLookPath = savedLookPath }()
	execLookPath = func(name string) (string, error) {
		return "", fmt.Errorf("not found")
	}

	savedStat := osStat
	defer func() { osStat = savedStat }()
	osStat = func(name string) (os.FileInfo, error) {
		return nil, fmt.Errorf("not found")
	}

	cfg := &config.AgenticConfig{
		Tools: map[config.ToolType]config.ToolConfig{
			config.ToolOpenCode: {Path: "C:\\missing\\opencode.exe"},
		},
	}

	info := FindTool(config.ToolOpenCode, cfg)
	assert.False(t, info.Found)
	assert.Empty(t, info.Path)
}

func TestFindTool_NotFound(t *testing.T) {
	savedLookPath := execLookPath
	defer func() { execLookPath = savedLookPath }()
	execLookPath = func(name string) (string, error) {
		return "", fmt.Errorf("not found")
	}

	savedStat := osStat
	defer func() { osStat = savedStat }()
	osStat = func(name string) (os.FileInfo, error) {
		return nil, fmt.Errorf("not found")
	}

	info := FindTool(config.ToolOpenCode, &config.AgenticConfig{})
	assert.False(t, info.Found)
	assert.Empty(t, info.Path)
}

func TestInstallInstructions_OpenCode(t *testing.T) {
	info := InstallInstructions(config.ToolOpenCode)
	assert.Equal(t, "https://opencode.ai", info.URL)
	assert.NotEmpty(t, info.Message)
	assert.Contains(t, info.Message, "OpenCode")
	assert.Contains(t, info.Message, "https://opencode.ai")
}

func TestInstallInstructions_Cursor(t *testing.T) {
	info := InstallInstructions(config.ToolCursor)
	assert.Equal(t, "https://cursor.com", info.URL)
	assert.NotEmpty(t, info.Message)
	assert.Contains(t, info.Message, "Cursor")
	assert.Contains(t, info.Message, "https://cursor.com")
}

func TestDetectTools_BothFound(t *testing.T) {
	savedLookPath := execLookPath
	defer func() { execLookPath = savedLookPath }()
	execLookPath = func(name string) (string, error) {
		if name == "opencode" || name == "cursor" {
			return name, nil
		}
		return "", fmt.Errorf("not found")
	}

	savedStat := osStat
	defer func() { osStat = savedStat }()
	osStat = func(name string) (os.FileInfo, error) {
		return nil, fmt.Errorf("not found")
	}

	result, err := DetectTools(context.Background(), "")
	assert.NoError(t, err)
	assert.Len(t, result.Tools, 2)
	assert.True(t, result.Tools[0].Found)
	assert.True(t, result.Tools[1].Found)
}

func TestDetectTools_NoneFound(t *testing.T) {
	savedLookPath := execLookPath
	defer func() { execLookPath = savedLookPath }()
	execLookPath = func(name string) (string, error) {
		return "", fmt.Errorf("not found")
	}

	savedStat := osStat
	defer func() { osStat = savedStat }()
	osStat = func(name string) (os.FileInfo, error) {
		return nil, fmt.Errorf("not found")
	}

	result, err := DetectTools(context.Background(), "")
	assert.NoError(t, err)
	assert.Len(t, result.Tools, 2)
	assert.False(t, result.Tools[0].Found)
	assert.False(t, result.Tools[1].Found)
}

func TestDetectTools_WithValidConfig(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(yamlPath, []byte("tools:\n  opencode:\n    path: custom.exe\n"), 0644)
	assert.NoError(t, err)

	savedLookPath := execLookPath
	defer func() { execLookPath = savedLookPath }()
	execLookPath = func(name string) (string, error) {
		return "", errors.New("not found")
	}

	savedStat := osStat
	defer func() { osStat = savedStat }()
	osStat = func(name string) (os.FileInfo, error) {
		if name == "custom.exe" {
			return nil, nil
		}
		return nil, fmt.Errorf("not found")
	}

	result, err := DetectTools(context.Background(), yamlPath)
	assert.NoError(t, err)
	assert.NotNil(t, result.Config)
}

func TestDefaultToolPaths_NotEmpty(t *testing.T) {
	paths := defaultToolPaths(config.ToolOpenCode)
	assert.NotEmpty(t, paths)

	paths = defaultToolPaths(config.ToolCursor)
	assert.NotEmpty(t, paths)
}

func TestToolDisplayName(t *testing.T) {
	assert.Equal(t, "OpenCode", toolDisplayName(config.ToolOpenCode))
	assert.Equal(t, "Cursor", toolDisplayName(config.ToolCursor))
}
