package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func StoreValidationReport(ctx context.Context, report *ValidationReport, outputDir string) error {
	path := orchestratorReportPath(outputDir)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("cr\u00e9ation du r\u00e9pertoire de rapport: %w", err)
	}
	return writeOrchestratorFormat(report, path)
}

func orchestratorReportPath(outputDir string) string {
	return filepath.Join(outputDir, "orchestrator-report.json")
}

type orchestratorReportData struct {
	Status        string                   `json:"status"`
	Duration      string                   `json:"duration"`
	NumModules    int                      `json:"num_modules"`
	NumChecks     int                      `json:"num_checks"`
	NumPassed     int                      `json:"num_passed"`
	NumFailed     int                      `json:"num_failed"`
	Modules       []orchestratorModuleData `json:"modules"`
	FormatVersion string                   `json:"format_version"`
}

type orchestratorModuleData struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Checks  int    `json:"checks"`
	Passed  int    `json:"passed"`
	Failed  int    `json:"failed"`
	Details string `json:"details,omitempty"`
}

func writeOrchestratorFormat(report *ValidationReport, path string) error {
	data := orchestratorReportData{
		Status:        string(report.Status),
		Duration:      fmt.Sprintf("%.2fs", report.Duration.Seconds()),
		NumModules:    len(report.Modules),
		NumChecks:     report.NumChecks(),
		NumPassed:     report.NumPassed(),
		NumFailed:     report.NumFailed(),
		FormatVersion: "1.0",
	}

	for _, m := range report.Modules {
		md := orchestratorModuleData{
			Name:   m.ModuleName,
			Status: string(m.Status),
			Checks: len(m.Checks),
		}
		for _, c := range m.Checks {
			if c.Status == StatusPassed {
				md.Passed++
			} else if c.Status == StatusFailed || c.Status == StatusError {
				md.Failed++
			}
		}
		if m.Error != nil {
			md.Details = m.Error.Error()
		}
		data.Modules = append(data.Modules, md)
	}

	payload, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, payload, 0644)
}
