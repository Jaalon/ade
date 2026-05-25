package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupTestEnv(t *testing.T) {
	t.Setenv("PLUGIN_NAME", "test-plugin")
	t.Setenv("PLUGIN_VERSION", "1.0.0")
	t.Setenv("PLUGIN_ORCHESTRATOR_URL", "http://localhost:8082")
}

func TestListTemplates_AtLeast6(t *testing.T) {
	templates := listTemplates()
	if len(templates) < 6 {
		t.Fatalf("expected at least 6 templates, got %d", len(templates))
	}
}

func TestListTemplates_HasGitignore(t *testing.T) {
	templates := listTemplates()
	found := false
	for _, tmpl := range templates {
		if tmpl.Name == "gitignore" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'gitignore' in template list")
	}
}

func TestRender_Gitignore(t *testing.T) {
	result, err := renderTemplate("gitignore", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "*.exe") {
		t.Errorf("output should contain '*.exe', got: %s", result)
	}
}

func TestRender_NotFound(t *testing.T) {
	_, err := renderTemplate("inexistant", nil)
	if err == nil {
		t.Fatal("expected error for nonexistent template")
	}
}

func TestRender_WorkflowLang(t *testing.T) {
	data := TemplateData{Lang: "fr"}
	result, err := renderTemplate("workflow", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "tasks-fr") {
		t.Errorf("output should contain 'tasks-fr' when Lang=fr, got: %s", result)
	}
}

func TestRender_AllTemplatesValid(t *testing.T) {
	templates := listTemplates()
	data := TemplateData{
		ProjectName: "test-project",
		GoVersion:   "1.26",
		ModulePath:  "github.com/user/test",
		Lang:        "fr",
	}
	for _, tmpl := range templates {
		t.Run(tmpl.Name, func(t *testing.T) {
			result, err := renderTemplate(tmpl.Name, data)
			if err != nil {
				t.Fatalf("Render(%q) failed: %v", tmpl.Name, err)
			}
			if result == "" {
				t.Errorf("Render(%q) returned empty", tmpl.Name)
			}
		})
	}
}

func TestRenderFS(t *testing.T) {
	data := TemplateData{ProjectName: "test"}
	result, err := renderFromFS(templatesFS, "embed/gitignore.tmpl", data)
	if err != nil {
		t.Fatalf("renderFromFS failed: %v", err)
	}
	if !strings.Contains(result, "*.exe") {
		t.Errorf("output should contain '*.exe'")
	}
}

func TestRenderFS_NotFound(t *testing.T) {
	_, err := renderFromFS(templatesFS, "embed/nonexistent.tmpl", nil)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestAPI_ListTemplates(t *testing.T) {
	setupTestEnv(t)
	p, err := NewTemplatePlugin()
	if err != nil {
		t.Fatalf("NewTemplatePlugin failed: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.ListTemplates(w, r)
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/templates")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result map[string][]TemplateInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}
	templates := result["templates"]
	if len(templates) < 6 {
		t.Errorf("len(templates) = %d, want >=6", len(templates))
	}
}

func TestAPI_ListTemplates_OnlyGet(t *testing.T) {
	setupTestEnv(t)
	p, err := NewTemplatePlugin()
	if err != nil {
		t.Fatalf("NewTemplatePlugin failed: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.ListTemplates(w, r)
	}))
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/v1/templates", "application/json", bytes.NewReader([]byte("{}")))
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestAPI_RenderTemplate(t *testing.T) {
	setupTestEnv(t)
	p, err := NewTemplatePlugin()
	if err != nil {
		t.Fatalf("NewTemplatePlugin failed: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.RenderTemplate(w, r)
	}))
	defer ts.Close()

	req := RenderRequest{
		TemplateName: "gitignore",
		Variables:    map[string]string{},
	}
	body, _ := json.Marshal(req)
	resp, err := http.Post(ts.URL+"/api/v1/templates/render", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result RenderResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}
	if result.Files == nil {
		t.Fatal("expected Files map")
	}
	content := result.Files["gitignore"]
	if !strings.Contains(content, "*.exe") {
		t.Errorf("output should contain '*.exe', got: %s", content)
	}
}

func TestAPI_RenderTemplate_OnlyPost(t *testing.T) {
	setupTestEnv(t)
	p, err := NewTemplatePlugin()
	if err != nil {
		t.Fatalf("NewTemplatePlugin failed: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.RenderTemplate(w, r)
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/templates/render")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestAPI_RenderTemplate_Unknown(t *testing.T) {
	setupTestEnv(t)
	p, err := NewTemplatePlugin()
	if err != nil {
		t.Fatalf("NewTemplatePlugin failed: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.RenderTemplate(w, r)
	}))
	defer ts.Close()

	req := RenderRequest{
		TemplateName: "nonexistent",
		Variables:    map[string]string{},
		Strict:       true,
	}
	body, _ := json.Marshal(req)
	resp, err := http.Post(ts.URL+"/api/v1/templates/render", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestAPI_RenderTemplate_BadJSON(t *testing.T) {
	setupTestEnv(t)
	p, err := NewTemplatePlugin()
	if err != nil {
		t.Fatalf("NewTemplatePlugin failed: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.RenderTemplate(w, r)
	}))
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/v1/templates/render", "application/json", bytes.NewReader([]byte(`{invalid}`)))
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestAPI_RenderTemplate_EmptyName(t *testing.T) {
	setupTestEnv(t)
	p, err := NewTemplatePlugin()
	if err != nil {
		t.Fatalf("NewTemplatePlugin failed: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.RenderTemplate(w, r)
	}))
	defer ts.Close()

	req := RenderRequest{
		TemplateName: "",
		Variables:    map[string]string{},
	}
	body, _ := json.Marshal(req)
	resp, err := http.Post(ts.URL+"/api/v1/templates/render", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}
