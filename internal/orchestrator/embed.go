package orchestrator

import "embed"

//go:embed frontend/dist/*
var FrontendFS embed.FS
