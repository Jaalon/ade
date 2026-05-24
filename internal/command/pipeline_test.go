package command

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"automated_dev_environment/internal/ci"
)

type mockPipelineExecutor struct {
	executeFunc func(ctx context.Context, step ci.StepConfig) (*ci.StepResult, error)
}

func (m *mockPipelineExecutor) Execute(ctx context.Context, step ci.StepConfig) (*ci.StepResult, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, step)
	}
	return &ci.StepResult{Name: step.Name, Status: ci.StatusSucceeded, Duration: 0}, nil
}

func setDefaultGlobals() {
	plLoadConfigFn = ci.LoadConfig
	plValidateConfigFn = ci.ValidateConfig
	plNewDryRunFn = func() ci.Executor { return ci.NewDryRunExecutor() }
	plNewLocalFn = func() (ci.Executor, error) { return ci.NewLocalExecutor(), nil }
	plNewPipelineFn = func(executor ci.Executor) ci.Pipeline {
		return ci.NewPipelineRunner(executor)
	}
	plDetectOrchFn = func(_ context.Context) error { return nil }
	plRunConfigPath = "ade-pipeline.yaml"
	plRunVerbose = false
	plRunLocal = false
	plRunDryRun = false
	plInitOutput = "."
	plInitForce = false
	plInitTemplate = "generic"
}

type pipelineGlobals struct {
	LoadConfigFn     func(string) (ci.PipelineConfig, error)
	ValidateConfigFn func(ci.PipelineConfig) error
	NewDryRunFn      func() ci.Executor
	NewLocalFn       func() (ci.Executor, error)
	NewPipelineFn    func(ci.Executor) ci.Pipeline
	DetectOrchFn     func(context.Context) error
	RunConfigPath    string
	RunVerbose       bool
	RunLocal         bool
	RunDryRun        bool
	InitOutput       string
	InitForce        bool
	InitTemplate     string
}

func saveGlobals() pipelineGlobals {
	return pipelineGlobals{
		LoadConfigFn:     plLoadConfigFn,
		ValidateConfigFn: plValidateConfigFn,
		NewDryRunFn:      plNewDryRunFn,
		NewLocalFn:       plNewLocalFn,
		NewPipelineFn:    plNewPipelineFn,
		DetectOrchFn:     plDetectOrchFn,
		RunConfigPath:    plRunConfigPath,
		RunVerbose:       plRunVerbose,
		RunLocal:         plRunLocal,
		RunDryRun:        plRunDryRun,
		InitOutput:       plInitOutput,
		InitForce:        plInitForce,
		InitTemplate:     plInitTemplate,
	}
}

func restoreGlobals(g pipelineGlobals) {
	plLoadConfigFn = g.LoadConfigFn
	plValidateConfigFn = g.ValidateConfigFn
	plNewDryRunFn = g.NewDryRunFn
	plNewLocalFn = g.NewLocalFn
	plNewPipelineFn = g.NewPipelineFn
	plDetectOrchFn = g.DetectOrchFn
	plRunConfigPath = g.RunConfigPath
	plRunVerbose = g.RunVerbose
	plRunLocal = g.RunLocal
	plRunDryRun = g.RunDryRun
	plInitOutput = g.InitOutput
	plInitForce = g.InitForce
	plInitTemplate = g.InitTemplate
}

// testLock serializes tests that modify global vars.
var testLock sync.Mutex

func withFreshGlobals(body func()) {
	testLock.Lock()
	defer testLock.Unlock()
	saved := saveGlobals()
	defer restoreGlobals(saved)
	setDefaultGlobals()
	body()
}

func buildPipelineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "G\u00e8re le pipeline d'int\u00e9gration continue",
	}

	runCmd := &cobra.Command{
		Use:  "run",
		RunE: runPipeline,
	}
	runCmd.Flags().StringVar(&plRunConfigPath, "config", "ade-pipeline.yaml", "")
	runCmd.Flags().BoolVarP(&plRunVerbose, "verbose", "v", false, "")
	runCmd.Flags().BoolVar(&plRunLocal, "local", false, "")
	runCmd.Flags().BoolVar(&plRunDryRun, "dry-run", false, "")

	initCmd := &cobra.Command{
		Use:  "init",
		RunE: runPipelineInit,
	}
	initCmd.Flags().StringVarP(&plInitOutput, "output", "o", ".", "")
	initCmd.Flags().BoolVarP(&plInitForce, "force", "f", false, "")
	initCmd.Flags().StringVarP(&plInitTemplate, "template", "t", "generic", "")

	cmd.AddCommand(runCmd)
	cmd.AddCommand(initCmd)

	rootCmd := &cobra.Command{Use: "ade", SilenceUsage: true, SilenceErrors: true}
	rootCmd.AddCommand(cmd)
	return rootCmd
}

func execPipeline(args ...string) (string, error) {
	root := buildPipelineCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func TestPipelineCmd_Registered(t *testing.T) {
	withFreshGlobals(func() {
		output, err := execPipeline("pipeline", "--help")
		assert.NoError(t, err)
		assert.Contains(t, output, "pipeline")
		assert.Contains(t, output, "run")
		assert.Contains(t, output, "init")
	})
}

func TestPipelineRunCmd_DryRunDefault(t *testing.T) {
	withFreshGlobals(func() {
		usedDryRun := false
		plNewDryRunFn = func() ci.Executor {
			usedDryRun = true
			return ci.NewDryRunExecutor()
		}
		plLoadConfigFn = func(path string) (ci.PipelineConfig, error) {
			return ci.DefaultConfig(), nil
		}

		_, err := execPipeline("pipeline", "run")
		assert.NoError(t, err)
		assert.True(t, usedDryRun, "expected DryRunExecutor to be used by default")
	})
}

func TestPipelineRunCmd_WithConfigFlag(t *testing.T) {
	withFreshGlobals(func() {
		var loadedPath string
		plLoadConfigFn = func(path string) (ci.PipelineConfig, error) {
			loadedPath = path
			return ci.DefaultConfig(), nil
		}
		plNewDryRunFn = func() ci.Executor {
			e := ci.NewDryRunExecutor()
			e.Delay = 0
			return e
		}

		_, err := execPipeline("pipeline", "run", "--config", "./mon-pipeline.yaml")
		assert.NoError(t, err)
		assert.Equal(t, "./mon-pipeline.yaml", loadedPath)
	})
}

func TestPipelineRunCmd_VerboseOutput(t *testing.T) {
	withFreshGlobals(func() {
		plLoadConfigFn = func(path string) (ci.PipelineConfig, error) {
			return ci.DefaultConfig(), nil
		}
		plNewDryRunFn = func() ci.Executor {
			e := ci.NewDryRunExecutor()
			e.Delay = 0
			return e
		}

		output, err := execPipeline("pipeline", "run", "--verbose")
		assert.NoError(t, err)
		assert.Contains(t, output, "Configuration du pipeline")
		assert.Contains(t, output, "build")
	})
}

func TestPipelineRunCmd_LocalFlag(t *testing.T) {
	withFreshGlobals(func() {
		usedLocal := false
		plNewLocalFn = func() (ci.Executor, error) {
			usedLocal = true
			return ci.NewLocalExecutor(), nil
		}
		plLoadConfigFn = func(path string) (ci.PipelineConfig, error) {
			return ci.DefaultConfig(), nil
		}

		_, err := execPipeline("pipeline", "run", "--local")
		assert.NoError(t, err)
		assert.True(t, usedLocal, "expected LocalExecutor to be used with --local flag")
	})
}

func TestPipelineRunCmd_DisplayResults(t *testing.T) {
	withFreshGlobals(func() {
		plLoadConfigFn = func(path string) (ci.PipelineConfig, error) {
			return ci.DefaultConfig(), nil
		}
		plNewDryRunFn = func() ci.Executor {
			e := ci.NewDryRunExecutor()
			e.Delay = 0
			return e
		}

		output, err := execPipeline("pipeline", "run")
		assert.NoError(t, err)
		assert.Contains(t, output, "Pipeline CI")
		assert.Contains(t, output, "succ\u00e8s")
		assert.Contains(t, output, "Construction du projet")
		assert.Contains(t, output, "Tests unitaires")
	})
}

func TestPipelineRunCmd_DisplayFailure(t *testing.T) {
	withFreshGlobals(func() {
		plLoadConfigFn = func(path string) (ci.PipelineConfig, error) {
			return ci.DefaultConfig(), nil
		}

		mockExec := &mockPipelineExecutor{}
		callCount := 0
		mockExec.executeFunc = func(ctx context.Context, step ci.StepConfig) (*ci.StepResult, error) {
			callCount++
			if callCount == 2 {
				return &ci.StepResult{Name: step.Name, Status: ci.StatusFailed, Duration: 0, Err: fmt.Errorf("failure")}, nil
			}
			return &ci.StepResult{Name: step.Name, Status: ci.StatusSucceeded, Duration: 0}, nil
		}
		plNewPipelineFn = func(executor ci.Executor) ci.Pipeline {
			return ci.NewPipelineRunner(mockExec)
		}

		output, err := execPipeline("pipeline", "run")
		assert.NoError(t, err)
		assert.Contains(t, output, "\u00e9chec")
	})
}

func TestPipelineRunCmd_PriorityDryRun(t *testing.T) {
	withFreshGlobals(func() {
		usedDryRun := false
		usedLocal := false

		plNewDryRunFn = func() ci.Executor {
			usedDryRun = true
			return ci.NewDryRunExecutor()
		}
		plNewLocalFn = func() (ci.Executor, error) {
			usedLocal = true
			return ci.NewLocalExecutor(), nil
		}
		plLoadConfigFn = func(path string) (ci.PipelineConfig, error) {
			return ci.DefaultConfig(), nil
		}

		_, err := execPipeline("pipeline", "run", "--local", "--dry-run")
		assert.NoError(t, err)
		assert.True(t, usedDryRun, "expected dry-run to take priority over local")
		assert.False(t, usedLocal, "expected local not to be used when dry-run is set")
	})
}

func TestPipelineInitCmd_CreateGeneric(t *testing.T) {
	withFreshGlobals(func() {
		dir := t.TempDir()
		output, err := execPipeline("pipeline", "init", "--output", dir)
		assert.NoError(t, err)

		targetPath := filepath.Join(dir, "ade-pipeline.yaml")
		assert.FileExists(t, targetPath)

		data, _ := os.ReadFile(targetPath)
		content := string(data)
		assert.Contains(t, content, "Compilation du projet")
		assert.Contains(t, content, "Tests unitaires")
		assert.Contains(t, content, "Tests E2E")
		assert.Contains(t, output, "Fichier de configuration")
	})
}

func TestPipelineInitCmd_TemplateGo(t *testing.T) {
	withFreshGlobals(func() {
		dir := t.TempDir()
		output, err := execPipeline("pipeline", "init", "--output", dir, "--template", "go")
		assert.NoError(t, err)

		targetPath := filepath.Join(dir, "ade-pipeline.yaml")
		assert.FileExists(t, targetPath)

		data, _ := os.ReadFile(targetPath)
		content := string(data)
		assert.Contains(t, content, `["go", "build", "./..."]`)
		assert.Contains(t, content, `["go", "test", "./..."]`)
		assert.Contains(t, output, "Template : go")
	})
}

func TestPipelineInitCmd_TemplateJava(t *testing.T) {
	withFreshGlobals(func() {
		dir := t.TempDir()
		output, err := execPipeline("pipeline", "init", "--output", dir, "--template", "java")
		assert.NoError(t, err)

		targetPath := filepath.Join(dir, "ade-pipeline.yaml")
		assert.FileExists(t, targetPath)

		data, _ := os.ReadFile(targetPath)
		content := string(data)
		assert.Contains(t, content, `["mvn", "clean", "compile"]`)
		assert.Contains(t, content, `["mvn", "test"]`)
		assert.Contains(t, content, `["mvn", "verify"]`)
		assert.Contains(t, output, "Template : java")
	})
}

func TestPipelineInitCmd_ForceOverwrite(t *testing.T) {
	withFreshGlobals(func() {
		dir := t.TempDir()
		targetPath := filepath.Join(dir, "ade-pipeline.yaml")
		require.NoError(t, os.WriteFile(targetPath, []byte("original"), 0644))

		_, err := execPipeline("pipeline", "init", "--output", dir, "--force")
		assert.NoError(t, err)

		data, _ := os.ReadFile(targetPath)
		assert.NotContains(t, string(data), "original")
	})
}

func TestPipelineInitCmd_NoOverwriteWithoutForce(t *testing.T) {
	withFreshGlobals(func() {
		dir := t.TempDir()
		targetPath := filepath.Join(dir, "ade-pipeline.yaml")
		require.NoError(t, os.WriteFile(targetPath, []byte("original"), 0644))

		_, err := execPipeline("pipeline", "init", "--output", dir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "existe d\u00e9j\u00e0")

		data, _ := os.ReadFile(targetPath)
		assert.Equal(t, "original", string(data))
	})
}

func TestPipelineInitCmd_InvalidTemplate(t *testing.T) {
	withFreshGlobals(func() {
		dir := t.TempDir()
		_, err := execPipeline("pipeline", "init", "--output", dir, "--template", "unknown")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template inconnu")
		assert.Contains(t, err.Error(), "generic")
		assert.Contains(t, err.Error(), "go")
		assert.Contains(t, err.Error(), "java")
	})
}

func TestPipelineInitCmd_CustomOutput(t *testing.T) {
	withFreshGlobals(func() {
		dir := filepath.Join(t.TempDir(), "subdir")
		_, err := execPipeline("pipeline", "init", "--output", dir)
		assert.NoError(t, err)

		targetPath := filepath.Join(dir, "ade-pipeline.yaml")
		assert.FileExists(t, targetPath)
	})
}

func TestPipelineRunCmd_LoadConfigFromFile(t *testing.T) {
	withFreshGlobals(func() {
		dir := t.TempDir()
		configPath := filepath.Join(dir, "test-pipeline.yaml")
		configContent := `
stages:
  - type: build
    name: "Custom Build"
    enabled: true
    steps:
      - name: "compile"
        command: ["echo", "building"]
  - type: unit-test
    name: "Custom Tests"
    enabled: true
    steps:
      - name: "test"
        command: ["echo", "testing"]
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		plNewDryRunFn = func() ci.Executor {
			e := ci.NewDryRunExecutor()
			e.Delay = 0
			return e
		}

		output, err := execPipeline("pipeline", "run", "--config", configPath)
		assert.NoError(t, err)
		assert.Contains(t, output, "Custom Build")
		assert.Contains(t, output, "Custom Tests")
	})
}

func execPipelineWithContext(ctx context.Context, args ...string) (string, error) {
	root := buildPipelineCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.ExecuteContext(ctx)
	return buf.String(), err
}

func TestPipelineRunCmd_CancelledStatus(t *testing.T) {
	withFreshGlobals(func() {
		plLoadConfigFn = func(path string) (ci.PipelineConfig, error) {
			return ci.DefaultConfig(), nil
		}
		plNewDryRunFn = func() ci.Executor {
			e := ci.NewDryRunExecutor()
			e.Delay = 0
			return e
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		output, err := execPipelineWithContext(ctx, "pipeline", "run")
		assert.NoError(t, err)
		assert.Contains(t, output, "annul\u00e9")
	})
}
