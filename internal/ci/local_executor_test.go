package ci

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalExecutor_Success(t *testing.T) {
	e := NewLocalExecutor()

	var cmd []string
	if runtime.GOOS == "windows" {
		cmd = []string{"cmd", "/c", "echo hello"}
	} else {
		cmd = []string{"sh", "-c", "echo hello"}
	}

	result, err := e.Execute(context.Background(), StepConfig{Name: "echo", Command: cmd})
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, StatusSucceeded, result.Status)
	assert.Contains(t, result.Output, "hello")
}

func TestLocalExecutor_Failure(t *testing.T) {
	e := NewLocalExecutor()

	var cmd []string
	if runtime.GOOS == "windows" {
		cmd = []string{"cmd", "/c", "exit 1"}
	} else {
		cmd = []string{"sh", "-c", "exit 1"}
	}

	result, err := e.Execute(context.Background(), StepConfig{Name: "fail", Command: cmd})
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, StatusFailed, result.Status)
	assert.NotNil(t, result.Err)
}

func TestLocalExecutor_Cancellation(t *testing.T) {
	e := NewLocalExecutor()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var cmd []string
	if runtime.GOOS == "windows" {
		cmd = []string{"cmd", "/c", "echo never"}
	} else {
		cmd = []string{"sh", "-c", "echo never"}
	}

	result, err := e.Execute(ctx, StepConfig{Name: "cancelled", Command: cmd})
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, StatusCancelled, result.Status)
}

func TestLocalExecutor_WorkDir(t *testing.T) {
	e := NewLocalExecutor()

	dir := t.TempDir()

	var cmd []string
	if runtime.GOOS == "windows" {
		cmd = []string{"cmd", "/c", "echo %CD%"}
	} else {
		cmd = []string{"sh", "-c", "pwd"}
	}

	result, err := e.Execute(context.Background(), StepConfig{
		Name:    "pwd",
		Command: cmd,
		WorkDir: dir,
	})
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, StatusSucceeded, result.Status)

	absDir, _ := filepath.Abs(dir)
	assert.Contains(t, result.Output, filepath.Base(absDir))
}

func TestLocalExecutor_Env(t *testing.T) {
	e := NewLocalExecutor()

	var cmd []string
	if runtime.GOOS == "windows" {
		cmd = []string{"cmd", "/c", "echo %ADE_TEST_VAR%"}
	} else {
		cmd = []string{"sh", "-c", "echo $ADE_TEST_VAR"}
	}

	result, err := e.Execute(context.Background(), StepConfig{
		Name:    "env",
		Command: cmd,
		Env:     map[string]string{"ADE_TEST_VAR": "custom_value"},
	})
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, StatusSucceeded, result.Status)
	assert.Contains(t, result.Output, "custom_value")
}

func TestLocalExecutor_EmptyCommand(t *testing.T) {
	e := NewLocalExecutor()

	result, err := e.Execute(context.Background(), StepConfig{Name: "empty"})
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, StatusFailed, result.Status)
	assert.ErrorIs(t, result.Err, ErrInvalidStep)
}

func TestLocalExecutor_CommandNotFound(t *testing.T) {
	e := NewLocalExecutor()

	result, err := e.Execute(context.Background(), StepConfig{
		Name:    "nonexistent",
		Command: []string{"nonexistent-command-12345"},
	})
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, StatusFailed, result.Status)
	assert.NotNil(t, result.Err)
}

func TestLocalExecutor_CancelDuringLongCommand(t *testing.T) {
	e := NewLocalExecutor()

	ctx, cancel := context.WithCancel(context.Background())

	var cmd []string
	if runtime.GOOS == "windows" {
		cmd = []string{"cmd", "/c", "ping -n 10 127.0.0.1 > nul"}
	} else {
		cmd = []string{"sh", "-c", "sleep 30"}
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	result, err := e.Execute(ctx, StepConfig{Name: "long", Command: cmd})
	assert.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, StatusCancelled, result.Status,
		"expected cancelled status for long command cancelled mid-execution")
}

func TestLocalExecutor_Duration(t *testing.T) {
	e := NewLocalExecutor()

	var cmd []string
	if runtime.GOOS == "windows" {
		cmd = []string{"cmd", "/c", "echo test"}
	} else {
		cmd = []string{"sh", "-c", "echo test"}
	}

	start := time.Now()
	result, err := e.Execute(context.Background(), StepConfig{Name: "echo", Command: cmd})
	elapsed := time.Since(start)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, StatusSucceeded, result.Status)
	assert.Greater(t, elapsed, time.Duration(0))
	assert.Greater(t, result.Duration, time.Duration(0))
	assert.LessOrEqual(t, result.Duration, elapsed+100*time.Millisecond)
	assert.GreaterOrEqual(t, result.Duration, time.Duration(0))
}

func TestLocalExecutor_WithPipelineRunner(t *testing.T) {
	e := NewLocalExecutor()

	var cmd []string
	if runtime.GOOS == "windows" {
		cmd = []string{"cmd", "/c", "echo ok"}
	} else {
		cmd = []string{"sh", "-c", "echo ok"}
	}

	cfg := PipelineConfig{
		Stages: []StageConfig{
			{Type: StageBuild, Enabled: true, Steps: []StepConfig{{Name: "test", Command: cmd}}},
			{Type: StageUnitTest, Enabled: true, Steps: []StepConfig{{Name: "test2", Command: cmd}}},
		},
	}

	runner := NewPipelineRunner(e)
	result, err := runner.Run(context.Background(), cfg)
	require.NoError(t, err)
	require.Len(t, result.Stages, 7)
	assert.Equal(t, StatusSucceeded, result.Stages[0].Status)
	assert.Equal(t, StatusSkipped, result.Stages[1].Status)
	assert.Equal(t, StatusSucceeded, result.Stages[2].Status)
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func init() {
	fmt.Fprint(os.Stderr, "")
}
