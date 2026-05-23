package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerClient "github.com/docker/docker/client"
)

var execLookPath = exec.LookPath

type Client interface {
	Ping(ctx context.Context) error
	IsContainerRunning(ctx context.Context, containerName string) (bool, error)
	Close() error
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
