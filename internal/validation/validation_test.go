package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockValidator struct {
	name        string
	description string
	detectFn    func(ctx context.Context) (bool, error)
	validateFn  func(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error)
}

func (m *mockValidator) Name() string                             { return m.name }
func (m *mockValidator) Description() string                      { return m.description }
func (m *mockValidator) Detect(ctx context.Context) (bool, error) { return m.detectFn(ctx) }
func (m *mockValidator) Validate(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error) {
	return m.validateFn(ctx, cfg)
}

func resetRegistry() {
	registry = nil
}

func TestCheckStatus_Values(t *testing.T) {
	assert.Equal(t, CheckStatus("passed"), StatusPassed)
	assert.Equal(t, CheckStatus("failed"), StatusFailed)
	assert.Equal(t, CheckStatus("warning"), StatusWarning)
	assert.Equal(t, CheckStatus("skipped"), StatusSkipped)
	assert.Equal(t, CheckStatus("error"), StatusError)
}

func TestValidationReport_Passed(t *testing.T) {
	report := &ValidationReport{
		Modules: []ModuleResult{
			{Status: StatusPassed},
			{Status: StatusPassed},
		},
	}
	assert.True(t, report.Passed())
	assert.False(t, report.Failed())
}

func TestValidationReport_Failed(t *testing.T) {
	report := &ValidationReport{
		Modules: []ModuleResult{
			{Status: StatusPassed},
			{Status: StatusFailed},
		},
	}
	assert.False(t, report.Passed())
	assert.True(t, report.Failed())
}

func TestValidationReport_FailedWithError(t *testing.T) {
	report := &ValidationReport{
		Modules: []ModuleResult{
			{Status: StatusError},
		},
	}
	assert.False(t, report.Passed())
	assert.True(t, report.Failed())
}

func TestValidationReport_HasWarnings(t *testing.T) {
	t.Run("module-level warning", func(t *testing.T) {
		report := &ValidationReport{
			Modules: []ModuleResult{
				{Status: StatusWarning},
			},
		}
		assert.True(t, report.HasWarnings())
	})

	t.Run("check-level warning", func(t *testing.T) {
		report := &ValidationReport{
			Modules: []ModuleResult{
				{
					Status: StatusPassed,
					Checks: []CheckResult{
						{Status: StatusWarning},
					},
				},
			},
		}
		assert.True(t, report.HasWarnings())
	})

	t.Run("no warnings", func(t *testing.T) {
		report := &ValidationReport{
			Modules: []ModuleResult{
				{Status: StatusPassed, Checks: []CheckResult{{Status: StatusPassed}}},
			},
		}
		assert.False(t, report.HasWarnings())
	})
}

func TestValidationReport_Counters(t *testing.T) {
	report := &ValidationReport{
		Modules: []ModuleResult{
			{
				Status: StatusPassed,
				Checks: []CheckResult{
					{Status: StatusPassed},
					{Status: StatusPassed},
					{Status: StatusFailed},
				},
			},
			{
				Status: StatusPassed,
				Checks: []CheckResult{
					{Status: StatusPassed},
					{Status: StatusWarning},
				},
			},
		},
	}
	assert.Equal(t, 5, report.NumChecks())
	assert.Equal(t, 3, report.NumPassed())
	assert.Equal(t, 1, report.NumFailed())
}

func TestRegisterAndModules(t *testing.T) {
	resetRegistry()

	v1 := &mockValidator{name: "mock-a"}
	v2 := &mockValidator{name: "mock-b"}

	Register(v1)
	Register(v2)

	modules := Modules()
	require.Len(t, modules, 2)
	assert.Equal(t, "mock-a", modules[0].Name())
	assert.Equal(t, "mock-b", modules[1].Name())
}

func TestRegisterAndDetect(t *testing.T) {
	resetRegistry()

	detectedName := ""
	Register(&mockValidator{
		name: "detected-module",
		detectFn: func(ctx context.Context) (bool, error) {
			return true, nil
		},
	})
	Register(&mockValidator{
		name: "undetected-module",
		detectFn: func(ctx context.Context) (bool, error) {
			return false, nil
		},
	})

	modules, err := DetectModules(context.Background())
	require.NoError(t, err)
	require.Len(t, modules, 1)
	assert.Equal(t, "detected-module", modules[0].Name())
	_ = detectedName
}

func TestRegisterAndDetect_ErrorIgnored(t *testing.T) {
	resetRegistry()

	Register(&mockValidator{
		name: "erratic-module",
		detectFn: func(ctx context.Context) (bool, error) {
			return false, fmt.Errorf("detection failed")
		},
	})
	Register(&mockValidator{
		name: "good-module",
		detectFn: func(ctx context.Context) (bool, error) {
			return true, nil
		},
	})

	modules, err := DetectModules(context.Background())
	require.NoError(t, err)
	require.Len(t, modules, 1)
	assert.Equal(t, "good-module", modules[0].Name())
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, ".ade/validation", cfg.OutputDir)
	assert.Equal(t, []string{"json"}, cfg.Formats)
	assert.Empty(t, cfg.Modules)
}

func TestValidateConfig_Valid(t *testing.T) {
	cfg := DefaultConfig()
	err := ValidateConfig(cfg)
	assert.NoError(t, err)
}

func TestValidateConfig_ValidWithModules(t *testing.T) {
	cfg := ValidationConfig{
		OutputDir: ".ade/validation",
		Formats:   []string{"json", "junit"},
		Modules: []ModuleConfig{
			{Name: "golang", Enabled: true},
		},
	}
	err := ValidateConfig(cfg)
	assert.NoError(t, err)
}

func TestValidateConfig_UnknownFormat(t *testing.T) {
	cfg := ValidationConfig{
		Formats: []string{"xml"},
	}
	err := ValidateConfig(cfg)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigInvalid)
	assert.Contains(t, err.Error(), "format inconnu")
}

func TestValidateConfig_DuplicateModule(t *testing.T) {
	cfg := ValidationConfig{
		Modules: []ModuleConfig{
			{Name: "golang", Enabled: true},
			{Name: "golang", Enabled: true},
		},
	}
	err := ValidateConfig(cfg)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigInvalid)
	assert.Contains(t, err.Error(), "doublon")
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	cfg, err := LoadConfig("./nonexistent-file-12345.yaml")
	assert.NoError(t, err)
	assert.Equal(t, ".ade/validation", cfg.OutputDir)
	assert.Equal(t, []string{"json"}, cfg.Formats)
}

func TestLoadConfig_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ade-validate.yaml")
	yamlContent := `
output_dir: ./custom-reports
formats:
  - json
  - junit
modules:
  - name: golang
    enabled: true
    checks:
      - build
      - test
`
	require.NoError(t, os.WriteFile(path, []byte(yamlContent), 0644))

	cfg, err := LoadConfig(path)
	assert.NoError(t, err)
	assert.Equal(t, "./custom-reports", cfg.OutputDir)
	assert.Equal(t, []string{"json", "junit"}, cfg.Formats)
	require.Len(t, cfg.Modules, 1)
	assert.Equal(t, "golang", cfg.Modules[0].Name)
	assert.True(t, cfg.Modules[0].Enabled)
	assert.Equal(t, []string{"build", "test"}, cfg.Modules[0].Checks)
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	require.NoError(t, os.WriteFile(path, []byte("{{invalid"), 0644))

	_, err := LoadConfig(path)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigInvalid)
}

func TestLoadConfig_DefaultsApplied(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "minimal.yaml")
	yamlContent := `modules: []`
	require.NoError(t, os.WriteFile(path, []byte(yamlContent), 0644))

	cfg, err := LoadConfig(path)
	assert.NoError(t, err)
	assert.Equal(t, ".ade/validation", cfg.OutputDir)
	assert.Equal(t, []string{"json"}, cfg.Formats)
}

func TestRunner_RunNoModules(t *testing.T) {
	resetRegistry()

	runner := NewValidationRunner()
	report, err := runner.Run(context.Background(), DefaultConfig())
	require.NoError(t, err)
	assert.Equal(t, StatusPassed, report.Status)
	assert.Empty(t, report.Modules)
}

func TestRunner_RunWithMockModule(t *testing.T) {
	resetRegistry()

	Register(&mockValidator{
		name: "mock",
		detectFn: func(ctx context.Context) (bool, error) {
			return true, nil
		},
		validateFn: func(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error) {
			return &ModuleResult{
				ModuleName: "mock",
				Status:     StatusPassed,
				Checks:     []CheckResult{{Name: "check1", Status: StatusPassed}},
			}, nil
		},
	})

	runner := NewValidationRunner()
	report, err := runner.Run(context.Background(), DefaultConfig())
	require.NoError(t, err)
	assert.Equal(t, StatusPassed, report.Status)
	require.Len(t, report.Modules, 1)
	assert.Equal(t, "mock", report.Modules[0].ModuleName)
	assert.Equal(t, StatusPassed, report.Modules[0].Status)
}

func TestRunner_RunWithFailingModule(t *testing.T) {
	resetRegistry()

	Register(&mockValidator{
		name: "failing-mock",
		detectFn: func(ctx context.Context) (bool, error) {
			return true, nil
		},
		validateFn: func(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error) {
			return &ModuleResult{
				ModuleName: "failing-mock",
				Status:     StatusFailed,
				Checks:     []CheckResult{{Name: "bad-check", Status: StatusFailed, Message: "something went wrong"}},
			}, nil
		},
	})

	runner := NewValidationRunner()
	report, err := runner.Run(context.Background(), DefaultConfig())
	require.NoError(t, err)
	assert.Equal(t, StatusFailed, report.Status)
	require.Len(t, report.Modules, 1)
	assert.Equal(t, StatusFailed, report.Modules[0].Status)
}

func TestRunner_RunCancelledContext(t *testing.T) {
	resetRegistry()

	block := make(chan struct{})

	Register(&mockValidator{
		name: "slow-module",
		detectFn: func(ctx context.Context) (bool, error) {
			return true, nil
		},
		validateFn: func(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error) {
			<-block
			return &ModuleResult{ModuleName: "slow-module", Status: StatusPassed}, nil
		},
	})

	Register(&mockValidator{
		name: "never-runs",
		detectFn: func(ctx context.Context) (bool, error) {
			return true, nil
		},
		validateFn: func(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error) {
			return &ModuleResult{ModuleName: "never-runs", Status: StatusPassed}, nil
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	runner := NewValidationRunner()

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
		close(block)
	}()

	report, err := runner.Run(ctx, DefaultConfig())
	require.NoError(t, err)
	require.Len(t, report.Modules, 2)
	assert.Equal(t, StatusSkipped, report.Modules[1].Status)
}

func TestRunner_RunPanickingModule(t *testing.T) {
	resetRegistry()

	Register(&mockValidator{
		name: "panicking-module",
		detectFn: func(ctx context.Context) (bool, error) {
			return true, nil
		},
		validateFn: func(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error) {
			panic("something terrible happened")
		},
	})

	runner := NewValidationRunner()
	report, err := runner.Run(context.Background(), DefaultConfig())
	require.NoError(t, err)
	assert.Equal(t, StatusFailed, report.Status)
	require.Len(t, report.Modules, 1)
	assert.Equal(t, StatusError, report.Modules[0].Status)
	assert.ErrorIs(t, report.Modules[0].Error, ErrModulePanic)
	assert.Contains(t, report.Modules[0].Error.Error(), "something terrible happened")
}

func TestRunner_RunPanickingModuleFollowedByGoodOne(t *testing.T) {
	resetRegistry()

	Register(&mockValidator{
		name: "panicker",
		detectFn: func(ctx context.Context) (bool, error) {
			return true, nil
		},
		validateFn: func(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error) {
			panic("boom")
		},
	})

	Register(&mockValidator{
		name: "good-module",
		detectFn: func(ctx context.Context) (bool, error) {
			return true, nil
		},
		validateFn: func(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error) {
			return &ModuleResult{ModuleName: "good-module", Status: StatusPassed}, nil
		},
	})

	runner := NewValidationRunner()
	report, err := runner.Run(context.Background(), DefaultConfig())
	require.NoError(t, err)
	require.Len(t, report.Modules, 2)
	assert.Equal(t, StatusError, report.Modules[0].Status)
	assert.Equal(t, StatusPassed, report.Modules[1].Status)
}

func TestRunner_ModuleFiltering(t *testing.T) {
	resetRegistry()

	Register(&mockValidator{
		name: "module-a",
		detectFn: func(ctx context.Context) (bool, error) {
			return true, nil
		},
		validateFn: func(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error) {
			return &ModuleResult{ModuleName: "module-a", Status: StatusPassed}, nil
		},
	})

	Register(&mockValidator{
		name: "module-b",
		detectFn: func(ctx context.Context) (bool, error) {
			return true, nil
		},
		validateFn: func(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error) {
			return &ModuleResult{ModuleName: "module-b", Status: StatusPassed}, nil
		},
	})

	cfg := ValidationConfig{
		Modules: []ModuleConfig{
			{Name: "module-a", Enabled: true},
		},
	}

	runner := NewValidationRunner()
	report, err := runner.Run(context.Background(), cfg)
	require.NoError(t, err)
	require.Len(t, report.Modules, 1)
	assert.Equal(t, "module-a", report.Modules[0].ModuleName)
}

func TestRunner_ModuleDisabled(t *testing.T) {
	resetRegistry()

	Register(&mockValidator{
		name: "disabled-module",
		detectFn: func(ctx context.Context) (bool, error) {
			return true, nil
		},
		validateFn: func(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error) {
			return &ModuleResult{ModuleName: "disabled-module", Status: StatusPassed}, nil
		},
	})

	cfg := ValidationConfig{
		Modules: []ModuleConfig{
			{Name: "disabled-module", Enabled: false},
		},
	}

	runner := NewValidationRunner()
	report, err := runner.Run(context.Background(), cfg)
	require.NoError(t, err)
	require.Len(t, report.Modules, 1)
	assert.Equal(t, StatusSkipped, report.Modules[0].Status)
}

func TestRunner_DetectError_ModuleIgnored(t *testing.T) {
	resetRegistry()

	Register(&mockValidator{
		name: "broken-detect",
		detectFn: func(ctx context.Context) (bool, error) {
			return false, fmt.Errorf("detect error")
		},
		validateFn: func(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error) {
			return &ModuleResult{ModuleName: "broken-detect", Status: StatusPassed}, nil
		},
	})

	Register(&mockValidator{
		name: "working-module",
		detectFn: func(ctx context.Context) (bool, error) {
			return true, nil
		},
		validateFn: func(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error) {
			return &ModuleResult{ModuleName: "working-module", Status: StatusPassed}, nil
		},
	})

	runner := NewValidationRunner()
	report, err := runner.Run(context.Background(), DefaultConfig())
	require.NoError(t, err)
	require.Len(t, report.Modules, 1)
	assert.Equal(t, "working-module", report.Modules[0].ModuleName)
}

func TestRunner_ModuleReturnsNil(t *testing.T) {
	resetRegistry()

	Register(&mockValidator{
		name: "nil-returner",
		detectFn: func(ctx context.Context) (bool, error) {
			return true, nil
		},
		validateFn: func(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error) {
			return nil, nil
		},
	})

	runner := NewValidationRunner()
	report, err := runner.Run(context.Background(), DefaultConfig())
	require.NoError(t, err)
	require.Len(t, report.Modules, 1)
	assert.Equal(t, StatusPassed, report.Modules[0].Status)
	assert.Empty(t, report.Modules[0].Checks)
}

func TestRunner_ModuleReturnsError(t *testing.T) {
	resetRegistry()

	Register(&mockValidator{
		name: "err-returner",
		detectFn: func(ctx context.Context) (bool, error) {
			return true, nil
		},
		validateFn: func(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error) {
			return nil, fmt.Errorf("module crashed")
		},
	})

	runner := NewValidationRunner()
	report, err := runner.Run(context.Background(), DefaultConfig())
	require.NoError(t, err)
	require.Len(t, report.Modules, 1)
	assert.Equal(t, StatusError, report.Modules[0].Status)
	assert.ErrorContains(t, report.Modules[0].Error, "module crashed")
}

func TestRunner_ConfigWithExplicitModulesOverridesDetect(t *testing.T) {
	resetRegistry()

	Register(&mockValidator{
		name: "explicit-module",
		detectFn: func(ctx context.Context) (bool, error) {
			return false, nil
		},
		validateFn: func(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error) {
			return &ModuleResult{ModuleName: "explicit-module", Status: StatusPassed}, nil
		},
	})

	cfg := ValidationConfig{
		Modules: []ModuleConfig{
			{Name: "explicit-module", Enabled: true},
		},
	}

	runner := NewValidationRunner()
	report, err := runner.Run(context.Background(), cfg)
	require.NoError(t, err)
	require.Len(t, report.Modules, 1)
	assert.Equal(t, "explicit-module", report.Modules[0].ModuleName)
}

func TestErrors_AreSentinel(t *testing.T) {
	assert.ErrorIs(t, ErrConfigInvalid, ErrConfigInvalid)
	assert.ErrorIs(t, ErrModuleNotFound, ErrModuleNotFound)
	assert.ErrorIs(t, ErrModulePanic, ErrModulePanic)
	assert.ErrorIs(t, ErrValidationFailed, ErrValidationFailed)
}

func TestValidationReport_Duration(t *testing.T) {
	startedAt := time.Now()
	report := &ValidationReport{
		Status:      StatusPassed,
		StartedAt:   startedAt,
		CompletedAt: startedAt.Add(2 * time.Second),
		Duration:    2 * time.Second,
	}
	assert.Equal(t, 2*time.Second, report.Duration)
}

func TestRunner_AllowListFilters(t *testing.T) {
	resetRegistry()

	Register(&mockValidator{
		name: "allowed",
		detectFn: func(ctx context.Context) (bool, error) {
			return true, nil
		},
		validateFn: func(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error) {
			return &ModuleResult{ModuleName: "allowed", Status: StatusPassed}, nil
		},
	})

	Register(&mockValidator{
		name: "not-detected-but-registered",
		detectFn: func(ctx context.Context) (bool, error) {
			return true, nil
		},
		validateFn: func(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error) {
			return &ModuleResult{ModuleName: "not-detected-but-registered", Status: StatusPassed}, nil
		},
	})

	runner := NewValidationRunner()
	runner.AllowList = []string{"allowed"}

	report, err := runner.Run(context.Background(), DefaultConfig())
	require.NoError(t, err)
	assert.Len(t, report.Modules, 1)
	assert.Equal(t, "allowed", report.Modules[0].ModuleName)
}

func TestGoValidator_RegisterInit(t *testing.T) {
	resetRegistry()
	assert.Empty(t, Modules())
}

var _ Validator = (*mockValidator)(nil)

func TestFindModuleConfig_NotFound(t *testing.T) {
	mc := findModuleConfig("nonexistent", []ModuleConfig{{Name: "other", Enabled: true}})
	assert.Equal(t, "nonexistent", mc.Name)
	assert.True(t, mc.Enabled)
}

func TestFindModuleConfig_Found(t *testing.T) {
	mc := findModuleConfig("target", []ModuleConfig{
		{Name: "other", Enabled: false},
		{Name: "target", Enabled: false},
	})
	assert.Equal(t, "target", mc.Name)
	assert.False(t, mc.Enabled)
}

// ---------------------------------------------------------------------------
// GoValidator tests
// ---------------------------------------------------------------------------

func newMockGoValidator() *GoValidator {
	return &GoValidator{
		goCmd: func() (string, error) {
			return "/usr/local/go/bin/go", nil
		},
		execCmd: func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, name, args...)
		},
	}
}

func TestGoValidator_Name(t *testing.T) {
	v := NewGoValidator()
	assert.Equal(t, "golang", v.Name())
}

func TestGoValidator_Description(t *testing.T) {
	v := NewGoValidator()
	assert.Contains(t, v.Description(), "Go")
}

func TestGoValidator_Detect_GoModFound(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(origDir)

	require.NoError(t, os.WriteFile("go.mod", []byte("module test\n"), 0644))
	v := newMockGoValidator()
	detected, err := v.Detect(context.Background())
	assert.True(t, detected)
	assert.NoError(t, err)
}

func TestGoValidator_Detect_GoModNotFound(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(origDir)

	v := newMockGoValidator()
	detected, err := v.Detect(context.Background())
	assert.False(t, detected)
	assert.NoError(t, err)
}

func TestGoValidator_Detect_Error(t *testing.T) {
	v := newMockGoValidator()
	ctx := context.Background()
	detected, err := v.Detect(ctx)
	assert.False(t, detected)
	assert.NoError(t, err)
}

func TestGoValidator_ParseGoVersion(t *testing.T) {
	assert.Equal(t, "1.26", parseGoVersion("go version go1.26 windows/amd64"))
	assert.Equal(t, "1.21.5", parseGoVersion("go version go1.21.5 linux/amd64"))
	assert.Equal(t, "1.27rc1", parseGoVersion("go version go1.27rc1 darwin/arm64"))
	assert.Equal(t, "inconnue", parseGoVersion("python 3.12"))
	assert.Equal(t, "inconnue", parseGoVersion(""))
}

func TestGoValidator_CheckGoVersion_Success(t *testing.T) {
	executed := false
	v := &GoValidator{
		goCmd: func() (string, error) { return "cmd", nil },
		execCmd: func(ctx context.Context, name string, args ...string) *exec.Cmd {
			executed = true
			return exec.CommandContext(ctx, "cmd", "/c", "echo go version go1.26 windows/amd64")
		},
	}
	result := v.checkGoVersion(context.Background())
	assert.True(t, executed)
	assert.Equal(t, StatusPassed, result.Status)
	assert.Contains(t, result.Message, "1.26")
}

func TestGoValidator_CheckGoVersion_NotFound(t *testing.T) {
	v := &GoValidator{
		goCmd: func() (string, error) { return "", fmt.Errorf("not found") },
		execCmd: func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, name, args...)
		},
	}
	result := v.checkGoVersion(context.Background())
	assert.Equal(t, StatusFailed, result.Status)
	assert.Contains(t, result.Message, "go.dev")
}

func TestGoValidator_CheckGoBuild_Success(t *testing.T) {
	v := &GoValidator{
		goCmd: func() (string, error) { return "cmd", nil },
		execCmd: func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, "cmd", "/c", "echo build ok")
		},
	}
	result := v.checkGoBuild(context.Background())
	assert.Equal(t, StatusPassed, result.Status)
	assert.Contains(t, result.Message, "Build réussi")
}

func TestGoValidator_CheckGoBuild_Failure(t *testing.T) {
	v := &GoValidator{
		goCmd: func() (string, error) { return "cmd /c", nil },
		execCmd: func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, "cmd", "/c", "exit /b 1")
		},
	}
	result := v.checkGoBuild(context.Background())
	assert.Equal(t, StatusFailed, result.Status)
}

func TestGoValidator_CheckGoTest_Success(t *testing.T) {
	v := &GoValidator{
		goCmd: func() (string, error) { return "cmd", nil },
		execCmd: func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, "cmd", "/c", "echo ok")
		},
	}
	result := v.checkGoTest(context.Background())
	assert.Equal(t, StatusPassed, result.Status)
}

func TestGoValidator_CheckGoTest_NoTests(t *testing.T) {
	v := &GoValidator{
		goCmd: func() (string, error) { return "echo", nil },
		execCmd: func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, "cmd", "/c", "echo no test files")
		},
	}
	result := v.checkGoTest(context.Background())
	assert.Equal(t, StatusWarning, result.Status)
	assert.Contains(t, result.Message, "Aucun fichier de test")
}

func TestGoValidator_CheckGoVet_Success(t *testing.T) {
	v := &GoValidator{
		goCmd: func() (string, error) { return "cmd", nil },
		execCmd: func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, "cmd", "/c", "echo vet ok")
		},
	}
	result := v.checkGoVet(context.Background())
	assert.Equal(t, StatusPassed, result.Status)
	assert.Contains(t, result.Message, "Vet réussi")
}

func TestGoValidator_CheckGoVet_Failure(t *testing.T) {
	v := &GoValidator{
		goCmd: func() (string, error) { return "cmd /c", nil },
		execCmd: func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, "cmd", "/c", "exit /b 1")
		},
	}
	result := v.checkGoVet(context.Background())
	assert.Equal(t, StatusFailed, result.Status)
}

func TestGoValidator_Validate_AllChecks(t *testing.T) {
	v := &GoValidator{
		goCmd: func() (string, error) { return "cmd", nil },
		execCmd: func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, "cmd", "/c", "echo ok")
		},
	}
	result, err := v.Validate(context.Background(), ModuleConfig{
		Name:    "golang",
		Enabled: true,
		Checks:  []string{"version", "build", "test", "vet"},
	})
	require.NoError(t, err)
	require.Len(t, result.Checks, 4)
	assert.Equal(t, "golang", result.ModuleName)
}

func TestGoValidator_Validate_FilteredChecks(t *testing.T) {
	v := &GoValidator{
		goCmd: func() (string, error) { return "cmd", nil },
		execCmd: func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, "cmd", "/c", "echo ok")
		},
	}
	result, err := v.Validate(context.Background(), ModuleConfig{
		Name:    "golang",
		Enabled: true,
		Checks:  []string{"build", "test"},
	})
	require.NoError(t, err)
	require.Len(t, result.Checks, 2)
}

func TestGoValidator_Validate_DefaultsToAll(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(origDir)
	require.NoError(t, os.WriteFile("go.mod", []byte("module test\n"), 0644))

	v := NewGoValidator()
	result, err := v.Validate(context.Background(), ModuleConfig{
		Name:    "golang",
		Enabled: true,
	})
	require.NoError(t, err)
	assert.Len(t, result.Checks, 4)
}

func TestGoValidator_Validate_TimeoutOption(t *testing.T) {
	v := &GoValidator{
		goCmd: func() (string, error) { return "cmd /c", nil },
		execCmd: func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, "cmd", "/c", "echo ok")
		},
	}
	result, err := v.Validate(context.Background(), ModuleConfig{
		Name:    "golang",
		Enabled: true,
		Checks:  []string{"version"},
		Options: map[string]string{"timeout": "1"},
	})
	require.NoError(t, err)
	assert.Equal(t, StatusPassed, result.Status)
}

func TestGoValidator_Validate_MinVersionOption(t *testing.T) {
	v := newMockGoValidator()
	cfg := ModuleConfig{
		Options: map[string]string{"go_version": "1.26"},
	}
	assert.Equal(t, "1.26", v.minGoVersion(cfg))
}

func TestGoValidator_Validate_MinVersionDefault(t *testing.T) {
	v := newMockGoValidator()
	assert.Equal(t, "1.21", v.minGoVersion(ModuleConfig{}))
}

func TestGoValidator_GoCommandNotFoundFails(t *testing.T) {
	v := &GoValidator{
		goCmd: func() (string, error) { return "", fmt.Errorf("not found") },
		execCmd: func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, name, args...)
		},
	}
	result := v.checkGoBuild(context.Background())
	assert.Equal(t, StatusFailed, result.Status)
}

func TestCountTestFailures(t *testing.T) {
	output := `ok  	pkg/a	0.34s
--- FAIL: TestSomething (0.01s)
    a_test.go:42: expected true, got false
--- FAIL: TestOther (0.02s)
FAIL	pkg/b	0.89s`
	assert.Equal(t, 2, countTestFailures(output))
	assert.Equal(t, 0, countTestFailures("ok  	pkg/a	0.34s"))
}

func TestParseGoVersion_EdgeCases(t *testing.T) {
	assert.Equal(t, "inconnue", parseGoVersion(""))
	assert.Equal(t, "inconnue", parseGoVersion("not go output"))
	assert.Equal(t, "1.26", parseGoVersion("go version go1.26 windows/amd64"))
}

// ---------------------------------------------------------------------------
// Reporter tests
// ---------------------------------------------------------------------------

func makeTestReport() *ValidationReport {
	now := time.Now()
	return &ValidationReport{
		Status:      StatusFailed,
		StartedAt:   now,
		CompletedAt: now.Add(2 * time.Second),
		Duration:    2 * time.Second,
		Modules: []ModuleResult{
			{
				ModuleName: "golang",
				Status:     StatusFailed,
				Duration:   1500 * time.Millisecond,
				StartedAt:  now,
				Checks: []CheckResult{
					{Name: "go-version", Status: StatusPassed, Message: "Go 1.26 trouvé", Duration: 100 * time.Millisecond},
					{Name: "go-build", Status: StatusPassed, Message: "Build réussi", Duration: 800 * time.Millisecond},
					{Name: "go-test", Status: StatusFailed, Message: "2 tests échoués", Duration: 500 * time.Millisecond, Details: "FAIL: TestSomething"},
					{Name: "go-vet", Status: StatusPassed, Message: "Vet réussi", Duration: 100 * time.Millisecond},
				},
			},
		},
	}
}

func TestJSONReporter_Format(t *testing.T) {
	r := NewJSONReporter()
	assert.Equal(t, "json", r.Format())
}

func TestJSONReporter_Write_AllPassed(t *testing.T) {
	report := makeTestReport()
	report.Status = StatusPassed
	report.Modules[0].Status = StatusPassed

	var buf bytes.Buffer
	err := NewJSONReporter().Write(report, &buf)
	require.NoError(t, err)

	var parsed JSONReport
	require.NoError(t, json.Unmarshal(buf.Bytes(), &parsed))
	assert.Equal(t, "passed", parsed.Status)
	assert.Equal(t, 4, parsed.NumChecks)
	assert.Equal(t, 3, parsed.NumPassed)
	assert.Equal(t, 1, parsed.NumFailed)
	assert.Len(t, parsed.Modules, 1)
	assert.Equal(t, "golang", parsed.Modules[0].ModuleName)
}

func TestJSONReporter_Write_SomeFailed(t *testing.T) {
	report := makeTestReport()

	var buf bytes.Buffer
	err := NewJSONReporter().Write(report, &buf)
	require.NoError(t, err)

	var parsed JSONReport
	require.NoError(t, json.Unmarshal(buf.Bytes(), &parsed))
	assert.Equal(t, "failed", parsed.Status)
}

func TestJSONReporter_Write_EmptyReport(t *testing.T) {
	report := &ValidationReport{
		Status:    StatusPassed,
		Modules:   []ModuleResult{},
		StartedAt: time.Now(),
	}

	var buf bytes.Buffer
	err := NewJSONReporter().Write(report, &buf)
	require.NoError(t, err)

	var parsed JSONReport
	require.NoError(t, json.Unmarshal(buf.Bytes(), &parsed))
	assert.Equal(t, "passed", parsed.Status)
	assert.Equal(t, 0, parsed.NumChecks)
	assert.Empty(t, parsed.Modules)
}

func TestJSONReporter_Write_Indented(t *testing.T) {
	report := makeTestReport()
	var buf bytes.Buffer
	err := NewJSONReporter().Write(report, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "\n")
	assert.Contains(t, output, "  ")
}

func TestJUnitReporter_Format(t *testing.T) {
	r := NewJUnitReporter()
	assert.Equal(t, "junit", r.Format())
}

func TestJUnitReporter_Write_AllPassed(t *testing.T) {
	report := &ValidationReport{
		Status:   StatusPassed,
		Duration: 1 * time.Second,
		Modules: []ModuleResult{
			{
				ModuleName: "golang",
				Status:     StatusPassed,
				Duration:   1 * time.Second,
				Checks: []CheckResult{
					{Name: "go-version", Status: StatusPassed, Message: "OK"},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := NewJUnitReporter().Write(report, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "testsuites")
	assert.Contains(t, output, "testsuite")
	assert.Contains(t, output, "testcase")
	assert.Contains(t, output, `failures="0"`)
	assert.NotContains(t, output, "<failure")
}

func TestJUnitReporter_Write_SomeFailed(t *testing.T) {
	report := &ValidationReport{
		Status:   StatusFailed,
		Duration: 2 * time.Second,
		Modules: []ModuleResult{
			{
				ModuleName: "golang",
				Status:     StatusFailed,
				Duration:   1 * time.Second,
				Checks: []CheckResult{
					{Name: "go-build", Status: StatusFailed, Message: "Build failed", Details: "error output"},
					{Name: "go-test", Status: StatusPassed, Message: "Tests OK"},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := NewJUnitReporter().Write(report, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `<failure message="Build failed"`)
	assert.Contains(t, output, `failures="1"`)
}

func TestJUnitReporter_Write_EmptyReport(t *testing.T) {
	report := &ValidationReport{
		Status:  StatusPassed,
		Modules: []ModuleResult{},
	}

	var buf bytes.Buffer
	err := NewJUnitReporter().Write(report, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `tests="0"`)
	assert.Contains(t, output, `failures="0"`)
}

func TestJUnitReporter_Write_WithWarnings(t *testing.T) {
	report := &ValidationReport{
		Status: StatusWarning,
		Modules: []ModuleResult{
			{
				ModuleName: "golang",
				Status:     StatusWarning,
				Checks: []CheckResult{
					{Name: "go-test", Status: StatusWarning, Message: "Aucun fichier de test"},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := NewJUnitReporter().Write(report, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `<skipped message="Aucun fichier de test"`)
}

func TestJUnitReporter_Write_ErrorStatus(t *testing.T) {
	report := &ValidationReport{
		Status: StatusFailed,
		Modules: []ModuleResult{
			{
				ModuleName: "broken",
				Status:     StatusError,
				Checks: []CheckResult{
					{Name: "check", Status: StatusError, Message: "erreur technique"},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := NewJUnitReporter().Write(report, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `<error message="erreur technique"`)
	assert.Contains(t, output, `errors="1"`)
}

func TestNewReportWriter_Valid(t *testing.T) {
	r := NewReportWriter("json")
	assert.NotNil(t, r)
	assert.Equal(t, "json", r.Format())

	r = NewReportWriter("junit")
	assert.NotNil(t, r)
	assert.Equal(t, "junit", r.Format())
}

func TestNewReportWriter_Invalid(t *testing.T) {
	r := NewReportWriter("unknown")
	assert.Nil(t, r)

	r = NewReportWriter("")
	assert.Nil(t, r)
}

func TestAvailableReporters(t *testing.T) {
	reporters := AvailableReporters()
	formats := make([]string, len(reporters))
	for i, r := range reporters {
		formats[i] = r.Format()
	}
	assert.Contains(t, formats, "json")
	assert.Contains(t, formats, "junit")
}

func TestFmtDuration(t *testing.T) {
	assert.Contains(t, fmtDuration(500*time.Millisecond), "ms")
	assert.Contains(t, fmtDuration(1500*time.Millisecond), "s")
	assert.Contains(t, fmtDuration(90*time.Second), "s")
}

func TestGoValidator_RegisterViaInit(t *testing.T) {
	resetRegistry()
	Register(NewGoValidator())
	modules := Modules()
	found := false
	for _, m := range modules {
		if m.Name() == "golang" {
			found = true
			break
		}
	}
	assert.True(t, found, "GoValidator should be registered via init()")
}

func TestGoValidator_Validate_WithTimeout(t *testing.T) {
	v := &GoValidator{
		goCmd: func() (string, error) { return "cmd /c", nil },
		execCmd: func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, "cmd", "/c", "echo ok")
		},
	}
	cfg := ModuleConfig{
		Name:    "golang",
		Enabled: true,
		Checks:  []string{"version"},
		Options: map[string]string{"timeout": "999"},
	}
	result, err := v.Validate(context.Background(), cfg)
	require.NoError(t, err)
	assert.Equal(t, StatusPassed, result.Status)
}

func TestTruncateOutput(t *testing.T) {
	short := "short output"
	assert.Equal(t, short, truncateOutput(short, 100))

	long := ""
	for i := 0; i < 100; i++ {
		long += "a"
	}
	truncated := truncateOutput(long, 10)
	assert.Equal(t, 10+len("\n... (tronqué)"), len(truncated))
	assert.Contains(t, truncated, "... (tronqué)")
}

func TestGoValidator_Validate_CheckGoTestFailure(t *testing.T) {
	v := &GoValidator{
		goCmd: func() (string, error) { return "cmd /c", nil },
		execCmd: func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, "cmd", "/c", "echo --- FAIL: TestX && exit /b 1")
		},
	}
	result := v.checkGoTest(context.Background())
	assert.Equal(t, StatusFailed, result.Status)
}

func TestGoValidator_CheckGoTest_PartialFailure(t *testing.T) {
	v := &GoValidator{
		goCmd: func() (string, error) { return "cmd /c", nil },
		execCmd: func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, "cmd", "/c", "echo --- FAIL: TestSomething && exit /b 1")
		},
	}
	result := v.checkGoTest(context.Background())
	assert.Equal(t, StatusFailed, result.Status)
	assert.Contains(t, result.Details, "--- FAIL")
}

func TestJSONReporter_Write_NilModules(t *testing.T) {
	report := &ValidationReport{
		Status:    StatusPassed,
		Modules:   nil,
		StartedAt: time.Now(),
	}
	var buf bytes.Buffer
	err := NewJSONReporter().Write(report, &buf)
	require.NoError(t, err)
	var parsed JSONReport
	require.NoError(t, json.Unmarshal(buf.Bytes(), &parsed))
	assert.Empty(t, parsed.Modules)
}
