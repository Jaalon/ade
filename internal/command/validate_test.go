package command

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"automated_dev_environment/internal/validation"
)

var valTestLock sync.Mutex

type mockValidateModule struct{}

func (m *mockValidateModule) Name() string                             { return "mock-test" }
func (m *mockValidateModule) Description() string                      { return "Module de test mock" }
func (m *mockValidateModule) Detect(ctx context.Context) (bool, error) { return true, nil }
func (m *mockValidateModule) Validate(ctx context.Context, cfg validation.ModuleConfig) (*validation.ModuleResult, error) {
	return &validation.ModuleResult{
		ModuleName: "mock-test",
		Status:     validation.StatusPassed,
		Checks: []validation.CheckResult{
			{Name: "mock-check", Status: validation.StatusPassed, Message: "OK"},
		},
	}, nil
}

type noopReportWriter struct{}

func (n *noopReportWriter) Format() string { return "_noop" }
func (n *noopReportWriter) Write(report *validation.ValidationReport, w io.Writer) error {
	return nil
}

var noopWriter = &noopReportWriter{}

func setDefaultValidateGlobals() {
	valLoadConfigFn = validation.LoadConfig
	valValidateConfigFn = validation.ValidateConfig
	valNewRunnerFn = func() *validation.ValidationRunner { return validation.NewValidationRunner() }
	valDetectModulesFn = validation.DetectModules
	valNewReportWriterFn = func(format string) validation.ReportWriter {
		return noopWriter
	}
	valStoreReportFn = func(_ context.Context, _ *validation.ValidationReport, _ string) error {
		return nil
	}
	valRunConfig = "ade-validate.yaml"
	valRunOutput = ""
	valRunFormat = "json"
	valRunVerbose = false
	valInitOutput = "."
	valInitForce = false
}

type validateGlobals struct {
	LoadConfigFn      func(string) (validation.ValidationConfig, error)
	ValidateConfigFn  func(validation.ValidationConfig) error
	NewRunnerFn       func() *validation.ValidationRunner
	DetectModulesFn   func(context.Context) ([]validation.Validator, error)
	NewReportWriterFn func(string) validation.ReportWriter
	StoreReportFn     func(context.Context, *validation.ValidationReport, string) error
	RunConfig         string
	RunOutput         string
	RunFormat         string
	RunVerbose        bool
	InitOutput        string
	InitForce         bool
}

func saveValidateGlobals() validateGlobals {
	return validateGlobals{
		LoadConfigFn:      valLoadConfigFn,
		ValidateConfigFn:  valValidateConfigFn,
		NewRunnerFn:       valNewRunnerFn,
		DetectModulesFn:   valDetectModulesFn,
		NewReportWriterFn: valNewReportWriterFn,
		StoreReportFn:     valStoreReportFn,
		RunConfig:         valRunConfig,
		RunOutput:         valRunOutput,
		RunFormat:         valRunFormat,
		RunVerbose:        valRunVerbose,
		InitOutput:        valInitOutput,
		InitForce:         valInitForce,
	}
}

func restoreValidateGlobals(g validateGlobals) {
	valLoadConfigFn = g.LoadConfigFn
	valValidateConfigFn = g.ValidateConfigFn
	valNewRunnerFn = g.NewRunnerFn
	valDetectModulesFn = g.DetectModulesFn
	valNewReportWriterFn = g.NewReportWriterFn
	valStoreReportFn = g.StoreReportFn
	valRunConfig = g.RunConfig
	valRunOutput = g.RunOutput
	valRunFormat = g.RunFormat
	valRunVerbose = g.RunVerbose
	valInitOutput = g.InitOutput
	valInitForce = g.InitForce
}

func registerTestModule() {
	for _, m := range validation.Modules() {
		if m.Name() == "mock-test" {
			return
		}
	}
	validation.Register(&mockValidateModule{})
}

func withFreshValidateGlobals(body func()) {
	valTestLock.Lock()
	defer valTestLock.Unlock()
	saved := saveValidateGlobals()
	defer restoreValidateGlobals(saved)
	setDefaultValidateGlobals()
	registerTestModule()
	body()
}

func buildValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Valide l'environnement",
	}

	runCmd := &cobra.Command{
		Use:  "run",
		RunE: runValidate,
	}
	runCmd.Flags().StringVarP(&valRunConfig, "config", "c", "ade-validate.yaml", "")
	runCmd.Flags().StringVarP(&valRunOutput, "output", "o", ".ade/validation", "")
	runCmd.Flags().StringVarP(&valRunFormat, "format", "f", "json", "")
	runCmd.Flags().BoolVarP(&valRunVerbose, "verbose", "v", false, "")

	initCmd := &cobra.Command{
		Use:  "init",
		RunE: runValidateInit,
	}
	initCmd.Flags().StringVarP(&valInitOutput, "output", "o", ".", "")
	initCmd.Flags().BoolVarP(&valInitForce, "force", "f", false, "")

	cmd.AddCommand(runCmd)
	cmd.AddCommand(initCmd)

	rootCmd := &cobra.Command{Use: "ade", SilenceUsage: true, SilenceErrors: true}
	rootCmd.AddCommand(cmd)
	return rootCmd
}

func execValidate(args ...string) (string, error) {
	root := buildValidateCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func TestValidateCmd_Registered(t *testing.T) {
	withFreshValidateGlobals(func() {
		output, err := execValidate("validate", "--help")
		assert.NoError(t, err)
		assert.Contains(t, output, "validate")
		assert.Contains(t, output, "run")
		assert.Contains(t, output, "init")
	})
}

func TestValidateRunCmd_Default(t *testing.T) {
	withFreshValidateGlobals(func() {
		valLoadConfigFn = func(path string) (validation.ValidationConfig, error) {
			return validation.DefaultConfig(), nil
		}

		output, err := execValidate("validate", "run")
		assert.NoError(t, err)
		assert.Contains(t, output, "mock-test")
	})
}

func TestValidateRunCmd_WithConfigFlag(t *testing.T) {
	withFreshValidateGlobals(func() {
		var loadedPath string
		valLoadConfigFn = func(path string) (validation.ValidationConfig, error) {
			loadedPath = path
			return validation.DefaultConfig(), nil
		}

		_, err := execValidate("validate", "run", "--config", "./custom.yaml")
		assert.NoError(t, err)
		assert.Equal(t, "./custom.yaml", loadedPath)
	})
}

func TestValidateRunCmd_VerboseOutput(t *testing.T) {
	withFreshValidateGlobals(func() {
		valLoadConfigFn = func(path string) (validation.ValidationConfig, error) {
			return validation.DefaultConfig(), nil
		}

		output, err := execValidate("validate", "run", "--verbose")
		assert.NoError(t, err)
		assert.Contains(t, output, "OK")
	})
}

func TestValidateRunCmd_FormatJSON(t *testing.T) {
	withFreshValidateGlobals(func() {
		valLoadConfigFn = func(path string) (validation.ValidationConfig, error) {
			return validation.DefaultConfig(), nil
		}

		output, err := execValidate("validate", "run", "--format", "json")
		assert.NoError(t, err)
		assert.Contains(t, output, "report.json")
	})
}

func TestValidateRunCmd_FormatJUnit(t *testing.T) {
	withFreshValidateGlobals(func() {
		valLoadConfigFn = func(path string) (validation.ValidationConfig, error) {
			return validation.DefaultConfig(), nil
		}

		output, err := execValidate("validate", "run", "--format", "junit")
		assert.NoError(t, err)
		assert.Contains(t, output, "report.xml")
	})
}

func TestValidateRunCmd_FormatBoth(t *testing.T) {
	withFreshValidateGlobals(func() {
		valLoadConfigFn = func(path string) (validation.ValidationConfig, error) {
			return validation.DefaultConfig(), nil
		}

		output, err := execValidate("validate", "run", "--format", "json,junit")
		assert.NoError(t, err)
		assert.Contains(t, output, "report.json")
		assert.Contains(t, output, "report.xml")
	})
}

func TestValidateRunCmd_CustomOutput(t *testing.T) {
	withFreshValidateGlobals(func() {
		tmpDir := t.TempDir()
		valLoadConfigFn = func(path string) (validation.ValidationConfig, error) {
			return validation.DefaultConfig(), nil
		}

		output, err := execValidate("validate", "run", "--output", tmpDir)
		assert.NoError(t, err)
		assert.Contains(t, output, "Int\u00e9gration orchestrateur")
	})
}

func TestValidateRunCmd_OutputDirCreated(t *testing.T) {
	withFreshValidateGlobals(func() {
		tmpDir := filepath.Join(t.TempDir(), "nested", "reports")
		valLoadConfigFn = func(path string) (validation.ValidationConfig, error) {
			return validation.DefaultConfig(), nil
		}

		_, err := execValidate("validate", "run", "--output", tmpDir)
		assert.NoError(t, err)
		_, statErr := os.Stat(tmpDir)
		assert.NoError(t, statErr, "le r\u00e9pertoire doit \u00eatre cr\u00e9\u00e9")
	})
}

func TestValidateInitCmd_CreateFile(t *testing.T) {
	withFreshValidateGlobals(func() {
		tmpDir := t.TempDir()
		valInitOutput = tmpDir

		_, err := execValidate("validate", "init", "--output", tmpDir)
		assert.NoError(t, err)

		expectedPath := filepath.Join(tmpDir, "ade-validate.yaml")
		_, statErr := os.Stat(expectedPath)
		assert.NoError(t, statErr, "le fichier doit exister")
	})
}

func TestValidateInitCmd_ForceOverwrite(t *testing.T) {
	withFreshValidateGlobals(func() {
		tmpDir := t.TempDir()
		targetPath := filepath.Join(tmpDir, "ade-validate.yaml")
		require.NoError(t, os.WriteFile(targetPath, []byte("ancien"), 0644))

		_, err := execValidate("validate", "init", "--output", tmpDir, "--force")
		assert.NoError(t, err)

		data, _ := os.ReadFile(targetPath)
		assert.Contains(t, string(data), "Configuration de validation")
	})
}

func TestValidateInitCmd_NoOverwriteWithoutForce(t *testing.T) {
	withFreshValidateGlobals(func() {
		tmpDir := t.TempDir()
		targetPath := filepath.Join(tmpDir, "ade-validate.yaml")
		require.NoError(t, os.WriteFile(targetPath, []byte("ancien"), 0644))

		_, err := execValidate("validate", "init", "--output", tmpDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "existe d\u00e9j\u00e0")
	})
}

func TestValidateInitCmd_CustomOutput(t *testing.T) {
	withFreshValidateGlobals(func() {
		tmpDir := filepath.Join(t.TempDir(), "custom", "dir")

		_, err := execValidate("validate", "init", "--output", tmpDir)
		assert.NoError(t, err)

		expectedPath := filepath.Join(tmpDir, "ade-validate.yaml")
		_, statErr := os.Stat(expectedPath)
		assert.NoError(t, statErr)
	})
}

func TestParseFormats_Single(t *testing.T) {
	formats := parseFormats("json")
	assert.Equal(t, []string{"json"}, formats)
}

func TestParseFormats_Multiple(t *testing.T) {
	formats := parseFormats("json,junit")
	assert.Equal(t, []string{"json", "junit"}, formats)
}

func TestParseFormats_Empty(t *testing.T) {
	formats := parseFormats("")
	assert.Equal(t, []string{"json"}, formats)
}
