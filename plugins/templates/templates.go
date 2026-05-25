package main

import (
	"fmt"
	"io/fs"
	"path"
	"strings"
	"sync"
	"text/template"
)

type TemplateInfo struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Language     string `json:"language"`
	TargetPath   string `json:"target_path"`
	EmbeddedPath string `json:"-"`
}

type TemplateData struct {
	ProjectName string
	GoVersion   string
	ModulePath  string
	Lang        string
	Compose     struct {
		ConfigPort string
		Network    string
	}
}

type RenderRequest struct {
	TemplateName string            `json:"template_name"`
	Variables    map[string]string `json:"variables"`
	Strict       bool              `json:"strict"`
}

type RenderResponse struct {
	Files map[string]string `json:"files"`
}

type namedTemplate struct {
	info TemplateInfo
	tmpl *template.Template
}

var (
	tmplRegistry []namedTemplate
	tmplOnce     sync.Once
)

func registerTemplate(name, desc, lang, targetPath, embeddedPath string) {
	tmplRegistry = append(tmplRegistry, namedTemplate{
		info: TemplateInfo{
			Name:         name,
			Description:  desc,
			Language:     lang,
			TargetPath:   targetPath,
			EmbeddedPath: embeddedPath,
		},
	})
}

func init() {
	registerTemplate("gitignore", "Fichier .gitignore pour un projet Go", "go", ".gitignore", "embed/gitignore.tmpl")
	registerTemplate("mcp-config", "Configuration des serveurs MCP (.opencode/config.json)", "go", ".opencode/config.json", "embed/opencode/config.json.tmpl")
	registerTemplate("workflow", "Workflow de développement (.opencode/workflow.yaml)", "go", ".opencode/workflow.yaml", "embed/opencode/workflow.yaml.tmpl")
	registerTemplate("skill-specification-en", "Skill spécification en anglais", "markdown", ".opencode/skills/specification-en/SKILL.md", "embed/opencode/skills/specification-en/SKILL.md")
	registerTemplate("skill-specification-fr", "Skill spécification en français", "markdown", ".opencode/skills/specification-fr/SKILL.md", "embed/opencode/skills/specification-fr/SKILL.md")
	registerTemplate("skill-story-en", "Skill stories en anglais", "markdown", ".opencode/skills/story-en/SKILL.md", "embed/opencode/skills/story-en/SKILL.md")
	registerTemplate("skill-story-fr", "Skill stories en français", "markdown", ".opencode/skills/story-fr/SKILL.md", "embed/opencode/skills/story-fr/SKILL.md")
	registerTemplate("skill-tasks-en", "Skill tasks en anglais", "markdown", ".opencode/skills/tasks-en/SKILL.md", "embed/opencode/skills/tasks-en/SKILL.md")
	registerTemplate("skill-tasks-fr", "Skill tâches en français", "markdown", ".opencode/skills/tasks-fr/SKILL.md", "embed/opencode/skills/tasks-fr/SKILL.md")
	registerTemplate("skill-feedback-fr", "Skill feedback en français", "markdown", ".opencode/skills/feedback-fr/SKILL.md", "embed/opencode/skills/feedback-fr/SKILL.md")
	registerTemplate("docker-compose", "Environnement de préproduction (docker-compose.yml)", "yaml", "docker-compose.yml", "embed/docker-compose.yml.tmpl")
	registerTemplate("env", "Variables d'environnement (.env)", "env", ".env", "embed/env.tmpl")
}

func buildTemplateRegistry() {
	tmplOnce.Do(func() {
		for i, nt := range tmplRegistry {
			ext := path.Ext(nt.info.EmbeddedPath)
			if ext == ".tmpl" {
				raw, err := templatesFS.ReadFile(nt.info.EmbeddedPath)
				if err != nil {
					tmplRegistry[i].tmpl = template.Must(template.New(nt.info.Name).Parse(
						fmt.Sprintf("<<ERROR reading %s: %v>>", nt.info.EmbeddedPath, err)))
					continue
				}
				t, err := template.New(nt.info.Name).Parse(string(raw))
				if err != nil {
					tmplRegistry[i].tmpl = template.Must(template.New(nt.info.Name).Parse(
						fmt.Sprintf("<<ERROR parsing %s: %v>>", nt.info.EmbeddedPath, err)))
					continue
				}
				tmplRegistry[i].tmpl = t
			}
		}
	})
}

func listTemplates() []TemplateInfo {
	buildTemplateRegistry()
	out := make([]TemplateInfo, len(tmplRegistry))
	for i, nt := range tmplRegistry {
		out[i] = nt.info
	}
	return out
}

func renderTemplate(name string, data interface{}) (string, error) {
	buildTemplateRegistry()

	for _, nt := range tmplRegistry {
		if nt.info.Name == name {
			ext := path.Ext(nt.info.EmbeddedPath)
			if ext != ".tmpl" {
				raw, err := templatesFS.ReadFile(nt.info.EmbeddedPath)
				if err != nil {
					return "", fmt.Errorf("template %q not found: %v", name, err)
				}
				return string(raw), nil
			}
			if nt.tmpl == nil {
				return "", fmt.Errorf("template %q not parsed", name)
			}
			var buf strings.Builder
			if err := nt.tmpl.Execute(&buf, data); err != nil {
				return "", fmt.Errorf("template %q render error: %v", name, err)
			}
			return buf.String(), nil
		}
	}
	return "", fmt.Errorf("template %q not found", name)
}

func renderFromFS(fsys fs.FS, tmplPath string, data interface{}) (string, error) {
	raw, err := fs.ReadFile(fsys, tmplPath)
	if err != nil {
		return "", fmt.Errorf("reading %s: %v", tmplPath, err)
	}
	t, err := template.New(tmplPath).Parse(string(raw))
	if err != nil {
		return "", fmt.Errorf("parsing %s: %v", tmplPath, err)
	}
	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing %s: %v", tmplPath, err)
	}
	return buf.String(), nil
}
