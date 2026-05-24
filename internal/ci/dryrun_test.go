package ci

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDryRunExecutor_DefaultConfig(t *testing.T) {
	e := NewDryRunExecutor()
	assert.Equal(t, 500*time.Millisecond, e.Delay)
	assert.Equal(t, 1.0, e.SuccessRate)

	result, err := e.Execute(context.Background(), StepConfig{Name: "test", Command: []string{"echo"}})
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, StatusSucceeded, result.Status)
}

func TestDryRunExecutor_AlwaysFails(t *testing.T) {
	e := NewDryRunExecutor()
	e.SuccessRate = 0.0
	e.Delay = time.Millisecond

	result, err := e.Execute(context.Background(), StepConfig{Name: "test", Command: []string{"echo"}})
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, StatusFailed, result.Status)
	assert.NotNil(t, result.Err)
	assert.Contains(t, result.Err.Error(), "dry-run")
}

func TestDryRunExecutor_DeterministicFailure(t *testing.T) {
	e := &DryRunExecutor{
		Delay:       time.Millisecond,
		SuccessRate: 0.5,
		rng:         rand.New(rand.NewSource(42)),
	}

	result, err := e.Execute(context.Background(), StepConfig{Name: "test", Command: []string{"echo"}})
	assert.NoError(t, err)
	require.NotNil(t, result)

	e2 := &DryRunExecutor{
		Delay:       time.Millisecond,
		SuccessRate: 0.5,
		rng:         rand.New(rand.NewSource(42)),
	}

	result2, err2 := e2.Execute(context.Background(), StepConfig{Name: "test", Command: []string{"echo"}})
	assert.NoError(t, err2)
	require.NotNil(t, result2)

	assert.Equal(t, result.Status, result2.Status)
}

func TestDryRunExecutor_CancelledContext(t *testing.T) {
	e := NewDryRunExecutor()
	e.Delay = 100 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := e.Execute(ctx, StepConfig{Name: "test", Command: []string{"echo"}})
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, StatusCancelled, result.Status)
}

func TestDryRunExecutor_CancelDuringDelay(t *testing.T) {
	e := NewDryRunExecutor()
	e.Delay = 1 * time.Second

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	result, err := e.Execute(ctx, StepConfig{Name: "test", Command: []string{"echo"}})
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, StatusCancelled, result.Status)
}

func TestDryRunExecutor_ZeroDelay(t *testing.T) {
	e := NewDryRunExecutor()
	e.Delay = 0

	start := time.Now()
	result, err := e.Execute(context.Background(), StepConfig{Name: "test", Command: []string{"echo"}})
	duration := time.Since(start)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, StatusSucceeded, result.Status)
	assert.Less(t, duration, 50*time.Millisecond)
}

func TestDryRunExecutor_Delay(t *testing.T) {
	e := NewDryRunExecutor()
	e.Delay = 50 * time.Millisecond

	start := time.Now()
	result, err := e.Execute(context.Background(), StepConfig{Name: "test", Command: []string{"echo"}})
	duration := time.Since(start)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.GreaterOrEqual(t, duration, 40*time.Millisecond)
	assert.GreaterOrEqual(t, result.Duration, 40*time.Millisecond)
}

func TestDryRunExecutor_InvalidStep(t *testing.T) {
	e := NewDryRunExecutor()
	e.Delay = time.Millisecond

	result, err := e.Execute(context.Background(), StepConfig{Name: "empty"})
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, StatusFailed, result.Status)
	assert.ErrorIs(t, result.Err, ErrInvalidStep)
}

func TestSimulatedOutput_Command(t *testing.T) {
	out := SimulatedOutput(StepConfig{Command: []string{"go", "build", "./..."}})
	assert.Contains(t, out, "go build")
	assert.Contains(t, out, "downloading")

	out2 := SimulatedOutput(StepConfig{Command: []string{"go", "test", "./..."}})
	assert.Contains(t, out2, "go test")
	assert.Contains(t, out2, "passed")
}

func TestSimulatedOutput_Image(t *testing.T) {
	out := SimulatedOutput(StepConfig{
		Image:   "golang:1.26-alpine",
		Command: []string{"go", "test", "./..."},
	})
	assert.Contains(t, out, "docker run")
	assert.Contains(t, out, "golang:1.26-alpine")
	assert.Contains(t, out, "Pulling from")
}

func TestSimulatedOutput_DeployStep(t *testing.T) {
	out := SimulatedOutput(StepConfig{Command: []string{"docker", "compose", "up", "-d"}})
	assert.Contains(t, out, "docker compose up")
	assert.Contains(t, out, "Container")
	assert.Contains(t, out, "Started")

	out2 := SimulatedOutput(StepConfig{Command: []string{"docker-compose", "up", "-d"}})
	assert.Contains(t, out2, "Container")
}

func TestDryRunExecutor_WithPipelineRunner(t *testing.T) {
	e := NewDryRunExecutor()
	e.Delay = 0
	e.SuccessRate = 1.0

	runner := NewPipelineRunner(e)
	result, err := runner.Run(context.Background(), DefaultConfig())
	require.NoError(t, err)
	require.Len(t, result.Stages, 6)
	for _, s := range result.Stages {
		assert.Equal(t, StatusSucceeded, s.Status)
		for _, step := range s.Steps {
			assert.Equal(t, StatusSucceeded, step.Status)
			assert.NotEmpty(t, step.Output)
		}
	}
	assert.True(t, result.Succeeded())
	assert.Less(t, result.Duration, 1*time.Second)
}

func findPartialFailureSeed() int64 {
	for seed := int64(0); seed < 10000; seed++ {
		rng := rand.New(rand.NewSource(seed))
		firstOk := rng.Float64() < 0.5
		hasFail := false
		for i := 0; i < 10; i++ {
			if rng.Float64() >= 0.5 {
				hasFail = true
				break
			}
		}
		if firstOk && hasFail {
			return seed
		}
	}
	return -1
}

func TestDryRunExecutor_WithPipelineRunnerPartialFailure(t *testing.T) {
	seed := findPartialFailureSeed()
	require.NotEqual(t, -1, seed, "no suitable seed found for partial failure test")

	e := &DryRunExecutor{
		Delay:       0,
		SuccessRate: 0.5,
		rng:         rand.New(rand.NewSource(seed)),
	}

	runner := NewPipelineRunner(e)
	result, err := runner.Run(context.Background(), DefaultConfig())
	require.NoError(t, err)
	require.Len(t, result.Stages, 6)

	hasSuccess := false
	hasFailure := false
	for _, s := range result.Stages {
		if s.Status == StatusSucceeded {
			hasSuccess = true
		}
		if s.Status == StatusFailed {
			hasFailure = true
		}
	}
	assert.True(t, hasSuccess, "expected at least one successful stage")
	assert.True(t, hasFailure, "expected at least one failed stage")
}

func TestSimulatedOutput_Maven(t *testing.T) {
	out := SimulatedOutput(StepConfig{Command: []string{"mvn", "clean", "compile"}})
	assert.Contains(t, out, "BUILD SUCCESS")
	assert.Contains(t, out, "maven-compiler-plugin")
}

func TestSimulatedOutput_Npm(t *testing.T) {
	out := SimulatedOutput(StepConfig{Command: []string{"npm", "run", "build"}})
	assert.Contains(t, out, "npm run build")
	assert.Contains(t, out, "webpack")
}

func TestSimulatedOutput_Generic(t *testing.T) {
	out := SimulatedOutput(StepConfig{Command: []string{"./deploy.sh"}})
	assert.Contains(t, out, "./deploy.sh")
	assert.Contains(t, out, "Command completed")

	out2 := SimulatedOutput(StepConfig{Command: []string{"echo", "hello"}})
	assert.Contains(t, out2, "echo hello")
}

func TestSimulatedOutput_ImageOnly(t *testing.T) {
	out := SimulatedOutput(StepConfig{Image: "nginx:alpine"})
	assert.Contains(t, out, "docker run")
	assert.Contains(t, out, "nginx:alpine")
	assert.Contains(t, out, "Container execution completed")
}

func TestDryRunExecutor_OutputNotEmptyOnSuccess(t *testing.T) {
	e := NewDryRunExecutor()
	e.Delay = 0

	result, err := e.Execute(context.Background(), StepConfig{Name: "build", Command: []string{"go", "build", "./..."}})
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.Output)
	assert.Contains(t, result.Output, "go build")
}

func TestDryRunExecutor_OutputNotEmptyOnFailure(t *testing.T) {
	e := NewDryRunExecutor()
	e.Delay = 0
	e.SuccessRate = 0.0

	result, err := e.Execute(context.Background(), StepConfig{Name: "build", Command: []string{"go", "build", "./..."}})
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, StatusFailed, result.Status)
	assert.NotEmpty(t, result.Output)
}

func TestSimulatedOutput_LongCommand(t *testing.T) {
	cmd := []string{"sh", "-c", "echo hello && sleep 1 && echo world"}
	out := SimulatedOutput(StepConfig{Command: cmd})
	assert.Contains(t, out, strings.Join(cmd, " "))
	assert.Contains(t, out, "Command completed")
}

func TestDryRunExecutor_MultipleExecutions(t *testing.T) {
	e := NewDryRunExecutor()
	e.Delay = 0
	e.SuccessRate = 1.0

	for i := 0; i < 10; i++ {
		result, err := e.Execute(context.Background(), StepConfig{
			Name:    fmt.Sprintf("step-%d", i),
			Command: []string{"echo", "test"},
		})
		assert.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, StatusSucceeded, result.Status)
		assert.NotEmpty(t, result.Output)
	}
}
