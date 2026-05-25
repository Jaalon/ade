package validation

import "time"

type CheckStatus string

const (
	StatusPassed  CheckStatus = "passed"
	StatusFailed  CheckStatus = "failed"
	StatusWarning CheckStatus = "warning"
	StatusSkipped CheckStatus = "skipped"
	StatusError   CheckStatus = "error"
)

type CheckResult struct {
	Name     string        `json:"name"`
	Status   CheckStatus   `json:"status"`
	Message  string        `json:"message"`
	Duration time.Duration `json:"duration"`
	Err      error         `json:"-"`
	Details  string        `json:"details,omitempty"`
}

type ModuleResult struct {
	ModuleName string        `json:"module_name"`
	Status     CheckStatus   `json:"status"`
	Checks     []CheckResult `json:"checks"`
	Duration   time.Duration `json:"duration"`
	StartedAt  time.Time     `json:"started_at"`
	Error      error         `json:"-"`
}

type ValidationReport struct {
	Status      CheckStatus      `json:"status"`
	Modules     []ModuleResult   `json:"modules"`
	Duration    time.Duration    `json:"duration"`
	StartedAt   time.Time        `json:"started_at"`
	CompletedAt time.Time        `json:"completed_at"`
	Config      ValidationConfig `json:"-"`
}

func (r *ValidationReport) Passed() bool {
	for _, m := range r.Modules {
		if m.Status == StatusFailed || m.Status == StatusError {
			return false
		}
	}
	return true
}

func (r *ValidationReport) Failed() bool {
	for _, m := range r.Modules {
		if m.Status == StatusFailed || m.Status == StatusError {
			return true
		}
	}
	return false
}

func (r *ValidationReport) HasWarnings() bool {
	for _, m := range r.Modules {
		if m.Status == StatusWarning {
			return true
		}
		for _, c := range m.Checks {
			if c.Status == StatusWarning {
				return true
			}
		}
	}
	return false
}

func (r *ValidationReport) NumChecks() int {
	total := 0
	for _, m := range r.Modules {
		total += len(m.Checks)
	}
	return total
}

func (r *ValidationReport) NumPassed() int {
	total := 0
	for _, m := range r.Modules {
		for _, c := range m.Checks {
			if c.Status == StatusPassed {
				total++
			}
		}
	}
	return total
}

func (r *ValidationReport) NumFailed() int {
	total := 0
	for _, m := range r.Modules {
		for _, c := range m.Checks {
			if c.Status == StatusFailed || c.Status == StatusError {
				total++
			}
		}
	}
	return total
}
