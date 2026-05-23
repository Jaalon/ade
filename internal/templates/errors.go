package templates

import "errors"

var (
	ErrTemplateNotFound = errors.New("template not found")
	ErrTemplateParse    = errors.New("template parse error")
	ErrTemplateRender   = errors.New("template render error")
)
