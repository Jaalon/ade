package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"automated_dev_environment/internal/templates"
)

type FileStatus string

const (
	StatusCreated     FileStatus = "created"
	StatusSkipped     FileStatus = "skipped"
	StatusError       FileStatus = "error"
	StatusOverwritten FileStatus = "overwritten"
)

type FileResult struct {
	TemplateName string
	TargetPath   string
	Status       FileStatus
	Err          error
}

type Report struct {
	Files   []FileResult
	Success int
	Skipped int
	Errors  int
}

func Generate(ctx context.Context, opts Options) (*Report, error) {
	if opts.Prompter == nil {
		opts.Prompter = &StdPrompter{}
	}
	if opts.OutputDir == "" {
		opts.OutputDir = "."
	}

	info, err := os.Stat(opts.OutputDir)
	if err != nil {
		if os.IsNotExist(err) {
			if mkErr := os.MkdirAll(opts.OutputDir, 0755); mkErr != nil {
				return nil, fmt.Errorf("cannot create output directory %s: %w", opts.OutputDir, mkErr)
			}
		} else {
			return nil, fmt.Errorf("cannot stat output directory %s: %w", opts.OutputDir, err)
		}
	} else if !info.IsDir() {
		return nil, fmt.Errorf("output path %s is not a directory", opts.OutputDir)
	}

	allTemplates := templates.ListTemplates()

	var filtered []templates.TemplateInfo
	if len(opts.TemplatesFilter) > 0 {
		filterMap := make(map[string]bool, len(opts.TemplatesFilter))
		for _, name := range opts.TemplatesFilter {
			filterMap[name] = true
		}
		for _, t := range allTemplates {
			if filterMap[t.Name] {
				filtered = append(filtered, t)
			}
		}
	} else {
		filtered = allTemplates
	}

	report := &Report{}

	for _, tmpl := range filtered {
		select {
		case <-ctx.Done():
			return report, ctx.Err()
		default:
		}

		content, err := templates.Render(tmpl.Name, opts.TemplateData)
		if err != nil {
			report.Files = append(report.Files, FileResult{
				TemplateName: tmpl.Name,
				TargetPath:   tmpl.TargetPath,
				Status:       StatusError,
				Err:          fmt.Errorf("render error: %w", err),
			})
			report.Errors++
			continue
		}

		if content == "" {
			report.Files = append(report.Files, FileResult{
				TemplateName: tmpl.Name,
				TargetPath:   tmpl.TargetPath,
				Status:       StatusSkipped,
				Err:          fmt.Errorf("empty rendered content"),
			})
			report.Skipped++
			continue
		}

		dest := filepath.Join(opts.OutputDir, tmpl.TargetPath)
		destDir := filepath.Dir(dest)

		if mkErr := os.MkdirAll(destDir, 0755); mkErr != nil {
			report.Files = append(report.Files, FileResult{
				TemplateName: tmpl.Name,
				TargetPath:   tmpl.TargetPath,
				Status:       StatusError,
				Err:          fmt.Errorf("cannot create directory %s: %w", destDir, mkErr),
			})
			report.Errors++
			continue
		}

		existed := false
		if _, statErr := os.Stat(dest); statErr == nil {
			existed = true
			if !opts.Force {
				confirmed, promptErr := opts.Prompter.Confirm(
					fmt.Sprintf("Le fichier %s existe déjà. Écraser ?", dest))
				if promptErr != nil {
					report.Files = append(report.Files, FileResult{
						TemplateName: tmpl.Name,
						TargetPath:   tmpl.TargetPath,
						Status:       StatusSkipped,
						Err:          fmt.Errorf("prompt error: %w", promptErr),
					})
					report.Skipped++
					continue
				}
				if !confirmed {
					report.Files = append(report.Files, FileResult{
						TemplateName: tmpl.Name,
						TargetPath:   tmpl.TargetPath,
						Status:       StatusSkipped,
					})
					report.Skipped++
					continue
				}
			}
		}

		if writeErr := os.WriteFile(dest, []byte(content), 0644); writeErr != nil {
			report.Files = append(report.Files, FileResult{
				TemplateName: tmpl.Name,
				TargetPath:   tmpl.TargetPath,
				Status:       StatusError,
				Err:          fmt.Errorf("write error: %w", writeErr),
			})
			report.Errors++
			continue
		}

		status := StatusCreated
		if existed {
			status = StatusOverwritten
		}
		report.Files = append(report.Files, FileResult{
			TemplateName: tmpl.Name,
			TargetPath:   tmpl.TargetPath,
			Status:       status,
		})
		report.Success++
	}

	return report, nil
}
