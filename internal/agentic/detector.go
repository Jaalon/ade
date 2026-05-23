package agentic

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"automated_dev_environment/internal/config"
)

var execLookPath = exec.LookPath
var osStat = os.Stat
var osMkdirAll = os.MkdirAll
var osWriteFile = os.WriteFile
var osReadFile = os.ReadFile

func DetectTools(ctx context.Context, configPath string) (*DetectionResult, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	tools := []config.ToolType{config.ToolOpenCode, config.ToolCursor}
	result := &DetectionResult{
		Config:     cfg,
		ConfigPath: configPath,
	}

	for _, t := range tools {
		info := FindTool(t, cfg)
		result.Tools = append(result.Tools, info)
	}

	return result, nil
}

func FindTool(tool config.ToolType, cfg *config.AgenticConfig) ToolInfo {
	if cfg != nil && cfg.Tools != nil {
		if tc, ok := cfg.Tools[tool]; ok && tc.Path != "" {
			if _, err := osStat(tc.Path); err == nil {
				return ToolInfo{
					Type:  tool,
					Name:  toolDisplayName(tool),
					Path:  tc.Path,
					Found: true,
				}
			}
		}
	}

	if path, err := execLookPath(string(tool)); err == nil {
		return ToolInfo{
			Type:  tool,
			Name:  toolDisplayName(tool),
			Path:  path,
			Found: true,
		}
	}

	for _, p := range defaultToolPaths(tool) {
		if _, err := osStat(p); err == nil {
			return ToolInfo{
				Type:  tool,
				Name:  toolDisplayName(tool),
				Path:  p,
				Found: true,
			}
		}
	}

	return ToolInfo{
		Type:  tool,
		Name:  toolDisplayName(tool),
		Found: false,
	}
}

func toolDisplayName(tool config.ToolType) string {
	switch tool {
	case config.ToolOpenCode:
		return "OpenCode"
	case config.ToolCursor:
		return "Cursor"
	default:
		return string(tool)
	}
}

func defaultToolPaths(tool config.ToolType) []string {
	localAppData := os.Getenv("LOCALAPPDATA")
	userProfile := os.Getenv("USERPROFILE")
	programFiles := os.Getenv("PROGRAMFILES")
	programFilesX86 := os.Getenv("PROGRAMFILES(X86)")

	switch tool {
	case config.ToolOpenCode:
		var paths []string
		if localAppData != "" {
			paths = append(paths, filepath.Join(localAppData, "Programs", "opencode", "opencode.exe"))
		}
		if userProfile != "" {
			paths = append(paths, filepath.Join(userProfile, ".opencode", "opencode.exe"))
		}
		if programFiles != "" {
			paths = append(paths, filepath.Join(programFiles, "opencode", "opencode.exe"))
		}
		if programFilesX86 != "" {
			paths = append(paths, filepath.Join(programFilesX86, "opencode", "opencode.exe"))
		}
		return paths

	case config.ToolCursor:
		var paths []string
		if localAppData != "" {
			paths = append(paths, filepath.Join(localAppData, "Programs", "cursor", "cursor.exe"))
		}
		if userProfile != "" {
			paths = append(paths, filepath.Join(userProfile, "AppData", "Local", "cursor", "cursor.exe"))
		}
		if programFiles != "" {
			paths = append(paths, filepath.Join(programFiles, "cursor", "cursor.exe"))
		}
		return paths

	default:
		return nil
	}
}

func InstallInstructions(tool config.ToolType) InstallInfo {
	switch tool {
	case config.ToolOpenCode:
		return InstallInfo{
			Tool:    tool,
			URL:     "https://opencode.ai",
			Message: "OpenCode n'est pas installé. Téléchargez-le depuis https://opencode.ai et assurez-vous qu'il est dans votre PATH.",
		}
	case config.ToolCursor:
		return InstallInfo{
			Tool:    tool,
			URL:     "https://cursor.com",
			Message: "Cursor n'est pas installé. Téléchargez-le depuis https://cursor.com et assurez-vous qu'il est dans votre PATH.",
		}
	default:
		return InstallInfo{
			Tool:    tool,
			URL:     "",
			Message: "Outil inconnu.",
		}
	}
}
