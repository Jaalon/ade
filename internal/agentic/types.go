package agentic

import "automated_dev_environment/internal/config"

type ToolInfo struct {
	Type    config.ToolType
	Name    string
	Path    string
	Found   bool
	Version string
}

type DetectionResult struct {
	Tools      []ToolInfo
	Config     *config.AgenticConfig
	ConfigPath string
}

type InstallInfo struct {
	Tool    config.ToolType
	URL     string
	Message string
}
