package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ade/plugins-sdk"
	"github.com/ade/plugins-sdk/contract"
)

type TemplatePlugin struct {
	templates []TemplateInfo
	sdk       *sdk.PluginServer
}

func NewTemplatePlugin() (*TemplatePlugin, error) {
	server, err := sdk.NewPlugin()
	if err != nil {
		return nil, err
	}

	server.AddCapability(&contract.Capability{
		Name:        "template_provider",
		Description: "Provides project templates",
		Version:     "1.0.0",
	})

	p := &TemplatePlugin{
		templates: listTemplates(),
		sdk:       server,
	}

	server.HandleFunc("/api/v1/templates", p.ListTemplates)
	server.HandleFunc("/api/v1/templates/render", p.RenderTemplate)

	return p, nil
}

func (p *TemplatePlugin) Start(ctx context.Context) error {
	log.Printf("[templates-plugin] %d templates available", len(p.templates))
	return p.sdk.Start(ctx)
}

func (p *TemplatePlugin) Shutdown(ctx context.Context) {
	if p.sdk != nil {
		p.sdk.Shutdown(ctx)
	}
}

func (p *TemplatePlugin) ListTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"templates": p.templates,
	})
}

func (p *TemplatePlugin) RenderTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.TemplateName == "" {
		http.Error(w, "template_name is required", http.StatusBadRequest)
		return
	}

	data := TemplateData{
		ProjectName: req.Variables["ProjectName"],
		GoVersion:   req.Variables["GoVersion"],
		ModulePath:  req.Variables["ModulePath"],
		Lang:        req.Variables["Lang"],
	}

	if req.Variables["ConfigPort"] != "" {
		data.Compose.ConfigPort = req.Variables["ConfigPort"]
	}
	if req.Variables["Network"] != "" {
		data.Compose.Network = req.Variables["Network"]
	}

	result, err := renderTemplate(req.TemplateName, data)
	if err != nil {
		if req.Strict {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		result = fmt.Sprintf("<<error: %v>>", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(RenderResponse{
		Files: map[string]string{
			req.TemplateName: result,
		},
	})
}
