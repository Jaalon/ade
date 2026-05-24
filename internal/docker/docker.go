package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	dockerClient "github.com/docker/docker/client"
)

var execLookPath = exec.LookPath

type ContainerExecConfig struct {
	Image   string
	Command []string
	Env     map[string]string
	WorkDir string
}

type ContainerExecResult struct {
	Output   string
	ExitCode int64
}

type Client interface {
	Ping(ctx context.Context) error
	IsContainerRunning(ctx context.Context, containerName string) (bool, error)
	Close() error

	PullImage(ctx context.Context, image string) error
	RunContainer(ctx context.Context, cfg ContainerExecConfig) (*ContainerExecResult, error)
}

type realClient struct {
	client *dockerClient.Client
}

func (c *realClient) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx)
	return err
}

func (c *realClient) IsContainerRunning(ctx context.Context, containerName string) (bool, error) {
	containers, err := c.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("name", containerName)),
	})
	if err != nil {
		return false, err
	}
	for _, ctr := range containers {
		for _, name := range ctr.Names {
			if name == "/"+containerName && ctr.State == "running" {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *realClient) Close() error {
	return c.client.Close()
}

func (c *realClient) PullImage(ctx context.Context, imageRef string) error {
	reader, err := c.client.ImagePull(ctx, imageRef, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("pull image %s: %w", imageRef, err)
	}
	defer reader.Close()
	_, _ = io.Copy(io.Discard, reader)
	return nil
}

func (c *realClient) RunContainer(ctx context.Context, cfg ContainerExecConfig) (*ContainerExecResult, error) {
	env := make([]string, 0, len(cfg.Env))
	for k, v := range cfg.Env {
		env = append(env, k+"="+v)
	}

	created, err := c.client.ContainerCreate(ctx, &container.Config{
		Image:      cfg.Image,
		Cmd:        cfg.Command,
		Env:        env,
		WorkingDir: cfg.WorkDir,
	}, nil, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("create container from %s: %w", cfg.Image, err)
	}

	if err := c.client.ContainerStart(ctx, created.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("start container %s: %w", created.ID, err)
	}

	statusCh, errCh := c.client.ContainerWait(ctx, created.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		return nil, fmt.Errorf("wait container %s: %w", created.ID, err)
	case <-statusCh:
	}

	logReader, err := c.client.ContainerLogs(ctx, created.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("logs container %s: %w", created.ID, err)
	}
	defer logReader.Close()

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, logReader)

	_ = c.client.ContainerRemove(ctx, created.ID, container.RemoveOptions{})

	info, err := c.client.ContainerInspect(ctx, created.ID)
	if err != nil {
		return &ContainerExecResult{Output: buf.String(), ExitCode: 0}, nil
	}
	exitCode := int64(0)
	if info.State != nil {
		exitCode = int64(info.State.ExitCode)
	}

	return &ContainerExecResult{Output: buf.String(), ExitCode: exitCode}, nil
}

func Check() (string, error) {
	if _, err := execLookPath("docker"); err == nil {
		return "docker", nil
	}
	if _, err := execLookPath("podman"); err == nil {
		return "podman", nil
	}
	return "", fmt.Errorf("Docker/Podman non trouvé dans le PATH")
}

func NewClient() (Client, error) {
	client, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("création du client Docker: %w", err)
	}
	return &realClient{client: client}, nil
}

func EnsureConfigContainer(ctx context.Context) error {
	binary, err := Check()
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠ %v. Les fonctionnalités conteneurisées ne sont pas disponibles.\n", err)
		return nil
	}

	client, err := NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠ Impossible de se connecter à %s: %v\n", binary, err)
		return nil
	}
	defer client.Close()

	err = client.Ping(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠ Le démon %s n'est pas en cours d'exécution.\n", binary)
		return nil
	}

	running, err := client.IsContainerRunning(ctx, "ade-config")
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠ Impossible de vérifier le conteneur de configuration.\n")
		return nil
	}

	if !running {
		fmt.Fprintf(os.Stderr, "Conteneur de configuration non trouvé. Les fonctionnalités avancées nécessitent le déploiement via 'ade init ci'.\n")
	}

	return nil
}
