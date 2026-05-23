package agentic

import (
	"context"
	"path/filepath"
	"strings"

	"automated_dev_environment/internal/templates"
)

type SkillInfo struct {
	Name        string
	Description string
	Installed   bool
	TargetPath  string
}

type SkillsReport struct {
	Installed    []string
	AlreadyExist []string
	Skipped      []string
	Errors       []SkillError
	Total        int
}

type SkillError struct {
	Name string
	Err  error
}

func ListAvailableSkills() ([]SkillInfo, error) {
	all := templates.ListTemplates()
	var skills []SkillInfo

	for _, t := range all {
		if strings.HasPrefix(t.Name, "skill-") {
			skills = append(skills, SkillInfo{
				Name:        t.Name,
				Description: t.Description,
				TargetPath:  t.TargetPath,
			})
		}
	}
	return skills, nil
}

func ListProjectSkills(outputDir string) ([]SkillInfo, error) {
	available, err := ListAvailableSkills()
	if err != nil {
		return nil, err
	}

	var found []SkillInfo
	for _, s := range available {
		fullPath := filepath.Join(outputDir, s.TargetPath)
		if _, err := osStat(fullPath); err == nil {
			s.Installed = true
			found = append(found, s)
		}
	}
	return found, nil
}

func MissingSkills(outputDir string) ([]SkillInfo, error) {
	available, err := ListAvailableSkills()
	if err != nil {
		return nil, err
	}

	var missing []SkillInfo
	for _, s := range available {
		fullPath := filepath.Join(outputDir, s.TargetPath)
		if _, err := osStat(fullPath); err != nil {
			missing = append(missing, s)
		}
	}
	return missing, nil
}

func EnsureSkills(ctx context.Context, outputDir string, force bool) (*SkillsReport, error) {
	report := &SkillsReport{}

	available, err := ListAvailableSkills()
	if err != nil {
		return report, err
	}
	report.Total = len(available)

	for _, s := range available {
		select {
		case <-ctx.Done():
			return report, ctx.Err()
		default:
		}

		fullPath := filepath.Join(outputDir, s.TargetPath)

		exists := false
		if _, err := osStat(fullPath); err == nil {
			exists = true
		}

		if exists && !force {
			report.AlreadyExist = append(report.AlreadyExist, s.Name)
			continue
		}

		dir := filepath.Dir(fullPath)
		if mkErr := osMkdirAll(dir, 0755); mkErr != nil {
			report.Errors = append(report.Errors, SkillError{Name: s.Name, Err: mkErr})
			continue
		}

		content, rErr := templates.Render(s.Name, nil)
		if rErr != nil {
			report.Errors = append(report.Errors, SkillError{Name: s.Name, Err: rErr})
			continue
		}

		if wErr := osWriteFile(fullPath, []byte(content), 0644); wErr != nil {
			report.Errors = append(report.Errors, SkillError{Name: s.Name, Err: wErr})
			continue
		}

		report.Installed = append(report.Installed, s.Name)
	}

	return report, nil
}
