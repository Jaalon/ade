package ci

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockExecutor struct {
	executeFunc func(ctx context.Context, step StepConfig) (*StepResult, error)
}

func (m *mockExecutor) Execute(ctx context.Context, step StepConfig) (*StepResult, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, step)
	}
	return &StepResult{Name: step.Name, Status: StatusSucceeded, Duration: time.Millisecond}, nil
}

func TestStageTypeOrder(t *testing.T) {
	stages := AllStages()
	require.Len(t, stages, 6)
	assert.Equal(t, StageBuild, stages[0])
	assert.Equal(t, StageUnitTest, stages[1])
	assert.Equal(t, StageIntegrationTest, stages[2])
	assert.Equal(t, StageTestDeploy, stages[3])
	assert.Equal(t, StageE2E, stages[4])
	assert.Equal(t, StagePreprod, stages[5])
}

func TestStageDescription(t *testing.T) {
	assert.NotEmpty(t, StageDescription(StageBuild))
	assert.NotEmpty(t, StageDescription(StageUnitTest))
	assert.NotEmpty(t, StageDescription(StageIntegrationTest))
	assert.NotEmpty(t, StageDescription(StageTestDeploy))
	assert.NotEmpty(t, StageDescription(StageE2E))
	assert.NotEmpty(t, StageDescription(StagePreprod))
	assert.Empty(t, StageDescription("inconnu"))
	assert.Contains(t, StageDescription(StageBuild), "Construction")
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	require.Len(t, cfg.Stages, 6)
	for _, s := range cfg.Stages {
		assert.True(t, s.Enabled, "stage %s should be enabled", s.Type)
		assert.GreaterOrEqual(t, len(s.Steps), 1, "stage %s should have at least 1 step", s.Type)
	}
}

func TestValidateConfig_Valid(t *testing.T) {
	cfg := DefaultConfig()
	err := ValidateConfig(cfg)
	assert.NoError(t, err)
}

func TestValidateConfig_UnknownStage(t *testing.T) {
	cfg := PipelineConfig{
		Stages: []StageConfig{
			{Type: "unknown", Enabled: true, Steps: []StepConfig{{Name: "test", Command: []string{"echo"}}}},
		},
	}
	err := ValidateConfig(cfg)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigInvalid)
}

func TestValidateConfig_EmptyCommand(t *testing.T) {
	cfg := PipelineConfig{
		Stages: []StageConfig{
			{Type: StageBuild, Enabled: true, Steps: []StepConfig{{Name: "test", Command: nil}}},
		},
	}
	err := ValidateConfig(cfg)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigInvalid)
}

func TestValidateConfig_DisabledNoSteps(t *testing.T) {
	cfg := PipelineConfig{
		Stages: []StageConfig{
			{Type: StageBuild, Enabled: false, Steps: nil},
		},
	}
	err := ValidateConfig(cfg)
	assert.NoError(t, err)
}

func TestValidateConfig_Reordering(t *testing.T) {
	cfg := PipelineConfig{
		Stages: []StageConfig{
			{Type: StagePreprod, Enabled: true, Steps: []StepConfig{{Name: "a", Command: []string{"echo"}}}},
			{Type: StageBuild, Enabled: true, Steps: []StepConfig{{Name: "b", Command: []string{"echo"}}}},
		},
	}
	err := ValidateConfig(cfg)
	assert.NoError(t, err)
	assert.Equal(t, StageBuild, cfg.Stages[0].Type)
	assert.Equal(t, StagePreprod, cfg.Stages[1].Type)
}

func TestPipelineResult_Failed(t *testing.T) {
	r := &PipelineResult{
		Stages: []StageResult{
			{Status: StatusSucceeded},
			{Status: StatusFailed},
		},
	}
	assert.True(t, r.Failed())
	assert.False(t, r.Succeeded())
}

func TestPipelineResult_Succeeded(t *testing.T) {
	r := &PipelineResult{
		Stages: []StageResult{
			{Status: StatusSucceeded},
			{Status: StatusSucceeded},
		},
	}
	assert.False(t, r.Failed())
	assert.True(t, r.Succeeded())
}

func TestPipelineResult_SucceededWithSkipped(t *testing.T) {
	r := &PipelineResult{
		Stages: []StageResult{
			{Status: StatusSucceeded},
			{Status: StatusSkipped},
		},
	}
	assert.False(t, r.Failed())
	assert.True(t, r.Succeeded())
}

func TestPipelineRunner_AllSuccessful(t *testing.T) {
	executor := &mockExecutor{
		executeFunc: func(ctx context.Context, step StepConfig) (*StepResult, error) {
			return &StepResult{Name: step.Name, Status: StatusSucceeded, Duration: time.Millisecond}, nil
		},
	}
	runner := NewPipelineRunner(executor)
	result, err := runner.Run(context.Background(), DefaultConfig())
	require.NoError(t, err)
	require.Len(t, result.Stages, 6)
	for _, s := range result.Stages {
		assert.Equal(t, StatusSucceeded, s.Status, "stage %s should be succeeded", s.Type)
	}
	assert.True(t, result.Succeeded())
}

func TestPipelineRunner_StageFailure(t *testing.T) {
	callCount := 0
	executor := &mockExecutor{
		executeFunc: func(ctx context.Context, step StepConfig) (*StepResult, error) {
			callCount++
			if callCount == 3 {
				return &StepResult{Name: step.Name, Status: StatusFailed, Duration: time.Millisecond, Err: fmt.Errorf("step failed")}, nil
			}
			return &StepResult{Name: step.Name, Status: StatusSucceeded, Duration: time.Millisecond}, nil
		},
	}

	cfg := PipelineConfig{
		Stages: []StageConfig{
			{Type: StageBuild, Enabled: true, Steps: []StepConfig{{Name: "build", Command: []string{"go"}}}},
			{Type: StageUnitTest, Enabled: true, Steps: []StepConfig{{Name: "test", Command: []string{"go"}}}},
			{Type: StageIntegrationTest, Enabled: true, Steps: []StepConfig{{Name: "it", Command: []string{"go"}}}},
			{Type: StageTestDeploy, Enabled: true, Steps: []StepConfig{{Name: "deploy", Command: []string{"docker"}}}},
			{Type: StageE2E, Enabled: true, Steps: []StepConfig{{Name: "e2e", Command: []string{"go"}}}},
			{Type: StagePreprod, Enabled: true, Steps: []StepConfig{{Name: "preprod", Command: []string{"./deploy"}}}},
		},
	}

	runner := NewPipelineRunner(executor)
	result, err := runner.Run(context.Background(), cfg)
	require.NoError(t, err)
	require.Len(t, result.Stages, 6)

	assert.Equal(t, StatusSucceeded, result.Stages[0].Status)
	assert.Equal(t, StatusSucceeded, result.Stages[1].Status)
	assert.Equal(t, StatusFailed, result.Stages[2].Status)
	assert.Equal(t, StatusSkipped, result.Stages[3].Status)
	assert.Equal(t, StatusSkipped, result.Stages[4].Status)
	assert.Equal(t, StatusSkipped, result.Stages[5].Status)
}

func TestPipelineRunner_AllDisabled(t *testing.T) {
	executor := &mockExecutor{}
	cfg := PipelineConfig{
		Stages: []StageConfig{
			{Type: StageBuild, Enabled: false, Steps: nil},
			{Type: StageUnitTest, Enabled: false, Steps: nil},
		},
	}
	runner := NewPipelineRunner(executor)
	result, err := runner.Run(context.Background(), cfg)
	require.NoError(t, err)
	assert.True(t, result.Succeeded())
}

func TestPipelineRunner_Cancellation(t *testing.T) {
	executor := &mockExecutor{
		executeFunc: func(ctx context.Context, step StepConfig) (*StepResult, error) {
			return &StepResult{Name: step.Name, Status: StatusSucceeded, Duration: time.Millisecond}, nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := DefaultConfig()
	runner := NewPipelineRunner(executor)
	result, err := runner.Run(ctx, cfg)
	require.NoError(t, err)
	assert.Equal(t, StatusCancelled, result.Status)
}

func TestPipelineRunner_CancelDuringExecution(t *testing.T) {
	block := make(chan struct{})
	executor := &mockExecutor{
		executeFunc: func(ctx context.Context, step StepConfig) (*StepResult, error) {
			<-block
			return &StepResult{Name: step.Name, Status: StatusSucceeded, Duration: time.Millisecond}, nil
		},
	}

	cfg := PipelineConfig{
		Stages: []StageConfig{
			{Type: StageBuild, Enabled: true, Steps: []StepConfig{{Name: "build", Command: []string{"go"}}}},
			{Type: StageUnitTest, Enabled: true, Steps: []StepConfig{{Name: "test", Command: []string{"go"}}}},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	runner := NewPipelineRunner(executor)

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
		close(block)
	}()

	result, err := runner.Run(ctx, cfg)
	require.NoError(t, err)
	assert.Equal(t, StatusCancelled, result.Status)
}

func TestPipelineRunner_ExecutorError(t *testing.T) {
	executor := &mockExecutor{
		executeFunc: func(ctx context.Context, step StepConfig) (*StepResult, error) {
			return nil, fmt.Errorf("executor crashed")
		},
	}

	cfg := PipelineConfig{
		Stages: []StageConfig{
			{Type: StageBuild, Enabled: true, Steps: []StepConfig{{Name: "build", Command: []string{"go"}}}},
			{Type: StageUnitTest, Enabled: true, Steps: []StepConfig{{Name: "test", Command: []string{"go"}}}},
			{Type: StageIntegrationTest, Enabled: true, Steps: []StepConfig{{Name: "it", Command: []string{"go"}}}},
			{Type: StageTestDeploy, Enabled: true, Steps: []StepConfig{{Name: "deploy", Command: []string{"docker"}}}},
			{Type: StageE2E, Enabled: true, Steps: []StepConfig{{Name: "e2e", Command: []string{"go"}}}},
			{Type: StagePreprod, Enabled: true, Steps: []StepConfig{{Name: "preprod", Command: []string{"./deploy"}}}},
		},
	}

	runner := NewPipelineRunner(executor)
	result, err := runner.Run(context.Background(), cfg)
	require.NoError(t, err)
	require.Len(t, result.Stages, 6)
	assert.Equal(t, StatusFailed, result.Stages[0].Status)
	for i := 1; i < 6; i++ {
		assert.Equal(t, StatusSkipped, result.Stages[i].Status, "stage %d should be skipped", i)
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	cfg, err := LoadConfig("./nonexistent-file-12345.yaml")
	assert.NoError(t, err)
	assert.Len(t, cfg.Stages, 6)
}

func TestLoadConfig_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pipeline.yaml")
	yamlContent := `
stages:
  - type: build
    name: "Build"
    enabled: true
    steps:
      - name: "Compile"
        command: ["go", "build"]
  - type: unit-test
    name: "Tests"
    enabled: true
    steps:
      - name: "Test"
        command: ["go", "test"]
`
	require.NoError(t, os.WriteFile(path, []byte(yamlContent), 0644))

	cfg, err := LoadConfig(path)
	assert.NoError(t, err)
	require.Len(t, cfg.Stages, 2)
	assert.Equal(t, StageBuild, cfg.Stages[0].Type)
	assert.Equal(t, "Compile", cfg.Stages[0].Steps[0].Name)
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	require.NoError(t, os.WriteFile(path, []byte("{{invalid"), 0644))

	_, err := LoadConfig(path)
	assert.Error(t, err)
}

func TestDockerStepExecutor_NotImplemented(t *testing.T) {
	exec := NewDockerStepExecutor()
	result, err := exec.Execute(context.Background(), StepConfig{Name: "test"})
	assert.ErrorIs(t, err, ErrNotImplemented)
	assert.Nil(t, result)
}

func TestPipelineRunner_ExecutorIntegration(t *testing.T) {
	var executedSteps []string
	executor := &mockExecutor{
		executeFunc: func(ctx context.Context, step StepConfig) (*StepResult, error) {
			executedSteps = append(executedSteps, step.Name)
			return &StepResult{Name: step.Name, Status: StatusSucceeded, Duration: time.Millisecond}, nil
		},
	}

	cfg := PipelineConfig{
		Stages: []StageConfig{
			{Type: StageBuild, Enabled: true, Steps: []StepConfig{{Name: "compile", Command: []string{"go"}}}},
			{Type: StageUnitTest, Enabled: true, Steps: []StepConfig{{Name: "test", Command: []string{"go"}}}},
		},
	}

	runner := NewPipelineRunner(executor)
	_, err := runner.Run(context.Background(), cfg)
	require.NoError(t, err)
	assert.Equal(t, []string{"compile", "test"}, executedSteps)
}

func TestPipelineRunner_InvalidConfig(t *testing.T) {
	executor := &mockExecutor{}
	cfg := PipelineConfig{
		Stages: []StageConfig{
			{Type: "invalid", Enabled: true, Steps: []StepConfig{{Name: "x", Command: []string{"echo"}}}},
		},
	}
	runner := NewPipelineRunner(executor)
	_, err := runner.Run(context.Background(), cfg)
	assert.Error(t, err)
}

func TestValidateConfig_DuplicateStage(t *testing.T) {
	cfg := PipelineConfig{
		Stages: []StageConfig{
			{Type: StageBuild, Enabled: true, Steps: []StepConfig{{Name: "a", Command: []string{"echo"}}}},
			{Type: StageBuild, Enabled: true, Steps: []StepConfig{{Name: "b", Command: []string{"echo"}}}},
		},
	}
	err := ValidateConfig(cfg)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigInvalid)
}

func TestPipelineRunner_ValidatePropagation(t *testing.T) {
	executor := &mockExecutor{}
	runner := NewPipelineRunner(executor)

	cfg := PipelineConfig{
		Stages: []StageConfig{
			{Type: StageBuild, Enabled: true, Steps: []StepConfig{{Name: "ok", Command: []string{"echo"}}}},
		},
	}
	assert.NoError(t, runner.Validate(cfg))

	badCfg := PipelineConfig{
		Stages: []StageConfig{
			{Type: "bad", Enabled: true, Steps: []StepConfig{{Name: "x", Command: []string{"echo"}}}},
		},
	}
	err := runner.Validate(badCfg)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrConfigInvalid) || errors.Is(err, ErrUnknownStage), "error should wrap ErrConfigInvalid or ErrUnknownStage")
}
