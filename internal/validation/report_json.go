package validation

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type JSONReporter struct{}

func NewJSONReporter() *JSONReporter {
	return &JSONReporter{}
}

func (r *JSONReporter) Format() string {
	return "json"
}

func (r *JSONReporter) Write(report *ValidationReport, w io.Writer) error {
	jr := convertToJSON(report)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(jr)
}

func convertToJSON(report *ValidationReport) JSONReport {
	jr := JSONReport{
		Status:      string(report.Status),
		Duration:    fmtDuration(report.Duration),
		StartedAt:   report.StartedAt.Format(time.RFC3339),
		CompletedAt: report.CompletedAt.Format(time.RFC3339),
		NumChecks:   report.NumChecks(),
		NumPassed:   report.NumPassed(),
		NumFailed:   report.NumFailed(),
	}

	for _, m := range report.Modules {
		jm := JSONModuleResult{
			ModuleName: m.ModuleName,
			Status:     string(m.Status),
			Duration:   fmtDuration(m.Duration),
		}
		for _, c := range m.Checks {
			jm.Checks = append(jm.Checks, JSONCheckResult{
				Name:     c.Name,
				Status:   string(c.Status),
				Message:  c.Message,
				Duration: fmtDuration(c.Duration),
				Details:  c.Details,
			})
		}
		jr.Modules = append(jr.Modules, jm)
	}

	return jr
}

type JSONReport struct {
	Status      string             `json:"status"`
	Duration    string             `json:"duration"`
	StartedAt   string             `json:"started_at"`
	CompletedAt string             `json:"completed_at"`
	NumChecks   int                `json:"num_checks"`
	NumPassed   int                `json:"num_passed"`
	NumFailed   int                `json:"num_failed"`
	Modules     []JSONModuleResult `json:"modules"`
}

type JSONModuleResult struct {
	ModuleName string            `json:"module_name"`
	Status     string            `json:"status"`
	Duration   string            `json:"duration"`
	Checks     []JSONCheckResult `json:"checks"`
}

type JSONCheckResult struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Message  string `json:"message"`
	Duration string `json:"duration"`
	Details  string `json:"details,omitempty"`
}

func fmtDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		sec := d.Seconds()
		return fmt.Sprintf("%.2fs", sec)
	}
	return fmt.Sprintf("%.0fs", d.Seconds())
}
