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

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
