package templates

import (
	"fmt"
	"io/fs"
	"path"
	"strings"
	"sync"
	"text/template"
)

type ComposeConfig struct {
	ConfigPort string
	Network    string
}

type TemplateData struct {
	ProjectName string
	GoVersion   string
	ModulePath  string
	Lang        string
	Compose     ComposeConfig
}

type TemplateInfo struct {
	Name         string
	Description  string
	TargetPath   string
	EmbeddedPath string
}

type namedTemplate struct {
	info TemplateInfo
	tmpl *template.Template
}

var (
	registry []namedTemplate
	once     sync.Once
)

func register(name, desc, targetPath, embeddedPath string) {
	registry = append(registry, namedTemplate{
		info: TemplateInfo{
			Name:         name,
			Description:  desc,
			TargetPath:   targetPath,
			EmbeddedPath: embeddedPath,
		},
	})
}

func init() {
	register("gitignore",
		"Fichier .gitignore pour un projet Go",
		".gitignore",
		"embed/gitignore.tmpl")
	register("mcp-config",
		"Configuration des serveurs MCP (.opencode/config.json)",
		".opencode/config.json",
		"embed/opencode/config.json.tmpl")
	register("workflow",
		"Workflow de développement (.opencode/workflow.yaml)",
		".opencode/workflow.yaml",
		"embed/opencode/workflow.yaml.tmpl")

	register("skill-specification-en",
		"Skill spécification en anglais",
		".opencode/skills/specification-en/SKILL.md",
		"embed/opencode/skills/specification-en/SKILL.md")
	register("skill-specification-fr",
		"Skill spécification en français",
		".opencode/skills/specification-fr/SKILL.md",
		"embed/opencode/skills/specification-fr/SKILL.md")
	register("skill-story-en",
		"Skill stories en anglais",
		".opencode/skills/story-en/SKILL.md",
		"embed/opencode/skills/story-en/SKILL.md")
	register("skill-story-fr",
		"Skill stories en français",
		".opencode/skills/story-fr/SKILL.md",
		"embed/opencode/skills/story-fr/SKILL.md")
	register("skill-tasks-fr",
		"Skill tâches en français",
		".opencode/skills/tasks-fr/SKILL.md",
		"embed/opencode/skills/tasks-fr/SKILL.md")
	register("skill-tasks-en",
		"Skill tasks en anglais",
		".opencode/skills/tasks-en/SKILL.md",
		"embed/opencode/skills/tasks-en/SKILL.md")
	register("skill-feedback-fr",
		"Skill feedback en français",
		".opencode/skills/feedback-fr/SKILL.md",
		"embed/opencode/skills/feedback-fr/SKILL.md")

	register("docker-compose",
		"Environnement de préproduction (docker-compose.yml)",
		"docker-compose.yml",
		"embed/docker-compose.yml.tmpl")
	register("env",
		"Variables d'environnement (.env)",
		".env",
		"embed/env.tmpl")
}

func buildRegistry() {
	once.Do(func() {
		for i, nt := range registry {
			ext := path.Ext(nt.info.EmbeddedPath)
			if ext == ".tmpl" {
				raw, err := templateFS.ReadFile(nt.info.EmbeddedPath)
				if err != nil {
					registry[i].tmpl = template.Must(template.New(nt.info.Name).Parse(
						fmt.Sprintf("<<ERROR reading %s: %v>>", nt.info.EmbeddedPath, err)))
					continue
				}
				t, err := template.New(nt.info.Name).Parse(string(raw))
				if err != nil {
					registry[i].tmpl = template.Must(template.New(nt.info.Name).Parse(
						fmt.Sprintf("<<ERROR parsing %s: %v>>", nt.info.EmbeddedPath, err)))
					continue
				}
				registry[i].tmpl = t
			}
		}
	})
}

func ListTemplates() []TemplateInfo {
	buildRegistry()
	out := make([]TemplateInfo, len(registry))
	for i, nt := range registry {
		out[i] = nt.info
	}
	return out
}

func Render(name string, data interface{}) (string, error) {
	buildRegistry()

	for _, nt := range registry {
		if nt.info.Name == name {
			ext := path.Ext(nt.info.EmbeddedPath)
			if ext != ".tmpl" {
				raw, err := templateFS.ReadFile(nt.info.EmbeddedPath)
				if err != nil {
					return "", fmt.Errorf("%w: %s: %v", ErrTemplateRender, name, err)
				}
				return string(raw), nil
			}
			if nt.tmpl == nil {
				return "", fmt.Errorf("%w: %s: template not parsed", ErrTemplateRender, name)
			}
			var buf strings.Builder
			if err := nt.tmpl.Execute(&buf, data); err != nil {
				return "", fmt.Errorf("%w: %s: %v", ErrTemplateRender, name, err)
			}
			return buf.String(), nil
		}
	}
	return "", fmt.Errorf("%w: %s", ErrTemplateNotFound, name)
}

func RenderFS(fsys fs.FS, tmplPath string, data interface{}) (string, error) {
	raw, err := fs.ReadFile(fsys, tmplPath)
	if err != nil {
		return "", fmt.Errorf("%w: reading %s: %v", ErrTemplateNotFound, tmplPath, err)
	}
	t, err := template.New(tmplPath).Parse(string(raw))
	if err != nil {
		return "", fmt.Errorf("%w: parsing %s: %v", ErrTemplateParse, tmplPath, err)
	}
	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("%w: executing %s: %v", ErrTemplateRender, tmplPath, err)
	}
	return buf.String(), nil
}
