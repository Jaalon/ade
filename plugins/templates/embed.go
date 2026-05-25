package main

import "embed"

//go:embed embed/gitignore.tmpl embed/docker-compose.yml.tmpl embed/env.tmpl embed/opencode/config.json.tmpl embed/opencode/workflow.yaml.tmpl embed/opencode/skills/specification-en/SKILL.md embed/opencode/skills/specification-fr/SKILL.md embed/opencode/skills/story-en/SKILL.md embed/opencode/skills/story-fr/SKILL.md embed/opencode/skills/tasks-fr/SKILL.md embed/opencode/skills/tasks-en/SKILL.md embed/opencode/skills/feedback-fr/SKILL.md
var templatesFS embed.FS
