package templates

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestListTemplates_AtLeast6(t *testing.T) {
	templates := ListTemplates()
	if len(templates) < 6 {
		t.Fatalf("expected at least 6 templates, got %d", len(templates))
	}
}

func TestRender_Gitignore(t *testing.T) {
	out, err := Render("gitignore", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "*.exe") {
		t.Errorf("output should contain '*.exe', got: %s", out)
	}
	if !strings.Contains(out, "*.test") {
		t.Errorf("output should contain '*.test', got: %s", out)
	}
}

func TestRender_McpConfig(t *testing.T) {
	out, err := Render("mcp-config", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\ncontent: %s", err, out)
	}
	if _, ok := parsed["mcp_servers"]; !ok {
		t.Errorf("JSON should contain 'mcp_servers' key")
	}
}

func TestRender_NotFound(t *testing.T) {
	_, err := Render("inexistant", nil)
	if !errors.Is(err, ErrTemplateNotFound) {
		t.Fatalf("expected ErrTemplateNotFound, got %v", err)
	}
}

func TestRender_WorkflowLang(t *testing.T) {
	data := TemplateData{Lang: "fr"}
	out, err := Render("workflow", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "tasks-fr") {
		t.Errorf("workflow output should contain 'tasks-fr' when Lang=fr, got: %s", out)
	}
}

func TestRender_AllTemplatesValid(t *testing.T) {
	templates := ListTemplates()
	defaultData := TemplateData{
		ProjectName: "test-project",
		GoVersion:   "1.26",
		ModulePath:  "github.com/user/test",
		Lang:        "fr",
	}
	for _, tmpl := range templates {
		t.Run(tmpl.Name, func(t *testing.T) {
			out, err := Render(tmpl.Name, defaultData)
			if err != nil {
				t.Fatalf("Render(%q) failed: %v", tmpl.Name, err)
			}
			if out == "" {
				t.Errorf("Render(%q) returned empty output", tmpl.Name)
			}
		})
	}
}

func TestRenderFS(t *testing.T) {
	data := TemplateData{ProjectName: "test"}
	out, err := RenderFS(templateFS, "embed/gitignore.tmpl", data)
	if err != nil {
		t.Fatalf("RenderFS failed: %v", err)
	}
	if !strings.Contains(out, "*.exe") {
		t.Errorf("RenderFS output should contain '*.exe'")
	}
}

func TestRenderFS_NotFound(t *testing.T) {
	_, err := RenderFS(templateFS, "embed/nonexistent.tmpl", nil)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestRender_DockerCompose(t *testing.T) {
	data := TemplateData{
		ProjectName: "mon-projet",
		Compose: ComposeConfig{
			ConfigPort: "9090",
			Network:    "mon-network",
		},
	}
	out, err := Render("docker-compose", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "nginx:alpine") {
		t.Errorf("output should contain 'nginx:alpine', got: %s", out)
	}
	if !strings.Contains(out, "9090:80") {
		t.Errorf("output should contain '9090:80', got: %s", out)
	}
	if !strings.Contains(out, "mon-network") {
		t.Errorf("output should contain 'mon-network', got: %s", out)
	}
}

func TestRender_Env(t *testing.T) {
	data := TemplateData{
		ProjectName: "mon-projet",
		Compose: ComposeConfig{
			ConfigPort: "9090",
			Network:    "ade-net",
		},
	}
	out, err := Render("env", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "ADE_PROJECT_NAME=mon-projet") {
		t.Errorf("output should contain 'ADE_PROJECT_NAME=mon-projet', got: %s", out)
	}
	if !strings.Contains(out, "ADE_CONFIG_PORT=9090") {
		t.Errorf("output should contain 'ADE_CONFIG_PORT=9090', got: %s", out)
	}
	if !strings.Contains(out, "ADE_COMPOSE_NETWORK=ade-net") {
		t.Errorf("output should contain 'ADE_COMPOSE_NETWORK=ade-net', got: %s", out)
	}
}

func TestListTemplates_IncludesNew(t *testing.T) {
	templates := ListTemplates()
	names := make(map[string]bool)
	for _, tmpl := range templates {
		names[tmpl.Name] = true
	}
	if !names["docker-compose"] {
		t.Error("expected 'docker-compose' in template list")
	}
	if !names["env"] {
		t.Error("expected 'env' in template list")
	}
}

func TestRender_DockerComposeMinimalData(t *testing.T) {
	data := TemplateData{ProjectName: "test"}
	out, err := Render("docker-compose", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "nginx:alpine") {
		t.Errorf("output should contain 'nginx:alpine' with minimal data, got: %s", out)
	}
}
