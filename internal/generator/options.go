package generator

import "automated_dev_environment/internal/templates"

type Options struct {
	OutputDir string

	Force bool

	TemplateData templates.TemplateData

	Prompter Prompter

	TemplatesFilter []string
}
