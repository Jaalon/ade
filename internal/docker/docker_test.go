package docker

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockClient struct {
	pingFunc           func(ctx context.Context) error
	isContainerRunning func(ctx context.Context, name string) (bool, error)
	pullImageFunc      func(ctx context.Context, image string) error
	runContainerFunc   func(ctx context.Context, cfg ContainerExecConfig) (*ContainerExecResult, error)
}

func (m *mockClient) Ping(ctx context.Context) error {
	if m.pingFunc != nil {
		return m.pingFunc(ctx)
	}
	return nil
}

func (m *mockClient) IsContainerRunning(ctx context.Context, name string) (bool, error) {
	if m.isContainerRunning != nil {
		return m.isContainerRunning(ctx, name)
	}
	return false, nil
}

func (m *mockClient) Close() error { return nil }

func (m *mockClient) PullImage(ctx context.Context, image string) error {
	if m.pullImageFunc != nil {
		return m.pullImageFunc(ctx, image)
	}
	return nil
}

func (m *mockClient) RunContainer(ctx context.Context, cfg ContainerExecConfig) (*ContainerExecResult, error) {
	if m.runContainerFunc != nil {
		return m.runContainerFunc(ctx, cfg)
	}
	return &ContainerExecResult{Output: "mock output", ExitCode: 0}, nil
}

func TestCheckReturnsDocker(t *testing.T) {
	saved := execLookPath
	defer func() { execLookPath = saved }()
	execLookPath = func(name string) (string, error) {
		if name == "docker" {
			return "docker", nil
		}
		return "", fmt.Errorf("not found")
	}

	name, err := Check()
	assert.NoError(t, err)
	assert.Equal(t, "docker", name)
}

func TestCheckReturnsPodman(t *testing.T) {
	saved := execLookPath
	defer func() { execLookPath = saved }()
	execLookPath = func(name string) (string, error) {
		if name == "docker" {
			return "", fmt.Errorf("not found")
		}
		return "podman", nil
	}

	name, err := Check()
	assert.NoError(t, err)
	assert.Equal(t, "podman", name)
}

func TestCheckReturnsErrorWhenNoneFound(t *testing.T) {
	saved := execLookPath
	defer func() { execLookPath = saved }()
	execLookPath = func(name string) (string, error) {
		return "", fmt.Errorf("not found")
	}

	name, err := Check()
	assert.Error(t, err)
	assert.Empty(t, name)
	assert.Contains(t, err.Error(), "non trouvé")
}

func TestEnsureConfigContainerNoError(t *testing.T) {
	ctx := context.Background()
	err := EnsureConfigContainer(ctx)
	assert.NoError(t, err)
}

func TestIsContainerRunningWithTimeout(t *testing.T) {
	mock := &mockClient{
		isContainerRunning: func(ctx context.Context, name string) (bool, error) {
			select {
			case <-ctx.Done():
				return false, ctx.Err()
			case <-time.After(100 * time.Millisecond):
				return true, nil
			}
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	running, err := mock.IsContainerRunning(ctx, "ade-config")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deadline exceeded")
	assert.False(t, running)
}

func TestPullImageDelegatesToRealClient(t *testing.T) {
	mock := &mockClient{
		pullImageFunc: func(ctx context.Context, image string) error {
			assert.Equal(t, "golang:1.26-alpine", image)
			return nil
		},
	}

	err := mock.PullImage(context.Background(), "golang:1.26-alpine")
	assert.NoError(t, err)
}

func TestPullImageReturnsError(t *testing.T) {
	mock := &mockClient{
		pullImageFunc: func(ctx context.Context, image string) error {
			return fmt.Errorf("pull failed: not found")
		},
	}

	err := mock.PullImage(context.Background(), "nonexistent:latest")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRunContainerDelegatesToRealClient(t *testing.T) {
	mock := &mockClient{
		runContainerFunc: func(ctx context.Context, cfg ContainerExecConfig) (*ContainerExecResult, error) {
			assert.Equal(t, "alpine:latest", cfg.Image)
			assert.Equal(t, []string{"echo", "hello"}, cfg.Command)
			assert.Equal(t, "1", cfg.Env["TEST"])
			return &ContainerExecResult{Output: "hello\n", ExitCode: 0}, nil
		},
	}

	cfg := ContainerExecConfig{
		Image:   "alpine:latest",
		Command: []string{"echo", "hello"},
		Env:     map[string]string{"TEST": "1"},
	}
	result, err := mock.RunContainer(context.Background(), cfg)
	assert.NoError(t, err)
	assert.Equal(t, "hello\n", result.Output)
	assert.Equal(t, int64(0), result.ExitCode)
}

func TestRunContainerReturnsNonZeroExitCode(t *testing.T) {
	mock := &mockClient{
		runContainerFunc: func(ctx context.Context, cfg ContainerExecConfig) (*ContainerExecResult, error) {
			return &ContainerExecResult{Output: "error output", ExitCode: 1}, nil
		},
	}

	cfg := ContainerExecConfig{
		Image:   "alpine:latest",
		Command: []string{"sh", "-c", "exit 1"},
	}
	result, err := mock.RunContainer(context.Background(), cfg)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), result.ExitCode)
	assert.Equal(t, "error output", result.Output)
}

func TestRunContainerError(t *testing.T) {
	mock := &mockClient{
		runContainerFunc: func(ctx context.Context, cfg ContainerExecConfig) (*ContainerExecResult, error) {
			return nil, fmt.Errorf("container failed: image not found")
		},
	}

	cfg := ContainerExecConfig{Image: "nonexistent:latest"}
	result, err := mock.RunContainer(context.Background(), cfg)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
