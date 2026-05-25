package validation

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func init() {
	Register(NewGoValidator())
}

type GoValidator struct {
	goCmd   func() (string, error)
	execCmd func(ctx context.Context, name string, args ...string) *exec.Cmd
}

func NewGoValidator() *GoValidator {
	return &GoValidator{
		goCmd: func() (string, error) {
			return exec.LookPath("go")
		},
		execCmd: func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, name, args...)
		},
	}
}

func (v *GoValidator) Name() string {
	return "golang"
}

func (v *GoValidator) Description() string {
	return "Validation de l'environnement Go"
}

func (v *GoValidator) Detect(ctx context.Context) (bool, error) {
	_, err := os.Stat("go.mod")
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, nil
	}
	return true, nil
}

func (v *GoValidator) Validate(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error) {
	startedAt := time.Now()
	result := &ModuleResult{
		ModuleName: v.Name(),
		StartedAt:  startedAt,
	}

	checks := cfg.Checks
	if len(checks) == 0 {
		checks = []string{"version", "build", "test", "vet"}
	}

	checkSet := make(map[string]bool)
	for _, c := range checks {
		checkSet[c] = true
	}

	timeoutSec := 300
	if t, ok := cfg.Options["timeout"]; ok {
		if sec, err := strconv.Atoi(t); err == nil && sec > 0 {
			timeoutSec = sec
		}
	}

	checkCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	if checkSet["version"] {
		result.Checks = append(result.Checks, v.checkGoVersion(checkCtx))
	}
	if checkSet["build"] {
		result.Checks = append(result.Checks, v.checkGoBuild(checkCtx))
	}
	if checkSet["test"] {
		result.Checks = append(result.Checks, v.checkGoTest(checkCtx))
	}
	if checkSet["vet"] {
		result.Checks = append(result.Checks, v.checkGoVet(checkCtx))
	}

	result.Status = StatusPassed
	for _, c := range result.Checks {
		if c.Status == StatusFailed || c.Status == StatusError {
			result.Status = StatusFailed
			break
		}
		if c.Status == StatusWarning && result.Status == StatusPassed {
			result.Status = StatusWarning
		}
	}

	result.Duration = time.Since(startedAt)
	return result, nil
}

func (v *GoValidator) runCommand(ctx context.Context, name string, args ...string) (string, error) {
	cmd := v.execCmd(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (v *GoValidator) checkGoVersion(ctx context.Context) CheckResult {
	now := time.Now()
	start := now

	goPath, err := v.goCmd()
	if err != nil {
		return CheckResult{
			Name:     "go-version",
			Status:   StatusFailed,
			Message:  "Go n'est pas installé. Téléchargez-le depuis https://go.dev/dl/",
			Duration: time.Since(start),
			Err:      err,
		}
	}

	output, err := v.runCommand(ctx, goPath, "version")
	duration := time.Since(start)
	if err != nil {
		return CheckResult{
			Name:     "go-version",
			Status:   StatusFailed,
			Message:  fmt.Sprintf("Impossible d'exécuter 'go version': %v", err),
			Duration: duration,
			Err:      err,
		}
	}

	_ = now
	return CheckResult{
		Name:     "go-version",
		Status:   StatusPassed,
		Message:  fmt.Sprintf("Go %s trouvé", parseGoVersion(output)),
		Duration: duration,
	}
}

func (v *GoValidator) checkGoBuild(ctx context.Context) CheckResult {
	return v.runGoCheck(ctx, "go-build", "go build ./...", "build", "Build réussi")
}

func (v *GoValidator) checkGoTest(ctx context.Context) CheckResult {
	start := time.Now()

	goPath, err := v.goCmd()
	if err != nil {
		return CheckResult{
			Name:     "go-test",
			Status:   StatusFailed,
			Message:  "Go n'est pas installé",
			Duration: time.Since(start),
			Err:      err,
		}
	}

	output, err := v.runCommand(ctx, goPath, "test", "./...")
	duration := time.Since(start)

	if err != nil {
		if strings.Contains(output, "no test files") {
			return CheckResult{
				Name:     "go-test",
				Status:   StatusWarning,
				Message:  "Aucun fichier de test trouvé",
				Duration: duration,
				Details:  output,
			}
		}
		return CheckResult{
			Name:     "go-test",
			Status:   StatusFailed,
			Message:  "Tests échoués",
			Duration: duration,
			Details:  truncateOutput(output, 2048),
			Err:      err,
		}
	}

	if strings.Contains(output, "no test files") {
		return CheckResult{
			Name:     "go-test",
			Status:   StatusWarning,
			Message:  "Aucun fichier de test trouvé",
			Duration: duration,
			Details:  output,
		}
	}

	failCount := countTestFailures(output)
	if failCount > 0 {
		return CheckResult{
			Name:     "go-test",
			Status:   StatusFailed,
			Message:  fmt.Sprintf("%d test(s) échoué(s)", failCount),
			Duration: duration,
			Details:  truncateOutput(output, 2048),
		}
	}

	return CheckResult{
		Name:     "go-test",
		Status:   StatusPassed,
		Message:  "Tests réussis",
		Duration: duration,
	}
}

func (v *GoValidator) checkGoVet(ctx context.Context) CheckResult {
	return v.runGoCheck(ctx, "go-vet", "go vet ./...", "vet", "Vet réussi")
}

func (v *GoValidator) runGoCheck(ctx context.Context, checkName string, cmdStr string, action string, successMsg string) CheckResult {
	start := time.Now()

	goPath, err := v.goCmd()
	if err != nil {
		return CheckResult{
			Name:     checkName,
			Status:   StatusFailed,
			Message:  "Go n'est pas installé",
			Duration: time.Since(start),
			Err:      err,
		}
	}

	args := strings.Fields(cmdStr)
	if len(args) > 0 {
		args = args[1:]
	}

	output, err := v.runCommand(ctx, goPath, args...)
	duration := time.Since(start)

	if err != nil {
		return CheckResult{
			Name:     checkName,
			Status:   StatusFailed,
			Message:  fmt.Sprintf("go %s a échoué", action),
			Duration: duration,
			Details:  truncateOutput(output, 2048),
			Err:      err,
		}
	}

	return CheckResult{
		Name:     checkName,
		Status:   StatusPassed,
		Message:  successMsg,
		Duration: duration,
	}
}

func (v *GoValidator) minGoVersion(cfg ModuleConfig) string {
	if v, ok := cfg.Options["go_version"]; ok && v != "" {
		return v
	}
	return "1.21"
}

func parseGoVersion(output string) string {
	output = strings.TrimSpace(output)
	if !strings.HasPrefix(output, "go version go") {
		return "inconnue"
	}
	rest := strings.TrimPrefix(output, "go version go")
	parts := strings.Fields(rest)
	if len(parts) == 0 {
		return "inconnue"
	}
	return parts[0]
}

func countTestFailures(output string) int {
	count := 0
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "--- FAIL:") {
			count++
		}
	}
	return count
}

func truncateOutput(output string, maxLen int) string {
	if len(output) <= maxLen {
		return output
	}
	return output[:maxLen] + "\n... (tronqué)"
}

func getwd() (string, error) {
	return os.Getwd()
}

var osStat = os.Stat
var osGetwd = os.Getwd
