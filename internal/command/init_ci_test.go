package command

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"automated_dev_environment/internal/docker"
	"automated_dev_environment/internal/templates"
)

type mockDockerClient struct {
	pingFunc func(ctx context.Context) error
}

func (m *mockDockerClient) Ping(ctx context.Context) error {
	if m.pingFunc != nil {
		return m.pingFunc(ctx)
	}
	return nil
}

func (m *mockDockerClient) IsContainerRunning(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (m *mockDockerClient) Close() error {
	return nil
}

func resetCiFlags() {
	ciOutput = ""
	ciForce = false
	ciName = ""
	ciPort = "8080"
	ciNetwork = "ade-network"
}

func buildCiInitCmd() *cobra.Command {
	resetCiFlags()
	ciCmd := &cobra.Command{
		Use:   "ci",
		Short: "D\u00e9ploie l'environnement de pr\u00e9production",
		Long:  `D\u00e9tecte Docker/Podman, g\u00e9n\u00e8re docker-compose.yml et .env, d\u00e9ploie les conteneurs.`,
		RunE:  runInitCi,
	}
	ciCmd.Flags().StringVarP(&ciOutput, "output", "o", ".", "R\u00e9pertoire de sortie pour les fichiers g\u00e9n\u00e9r\u00e9s")
	ciCmd.Flags().BoolVarP(&ciForce, "force", "f", false, "\u00c9craser les fichiers existants sans confirmation")
	ciCmd.Flags().StringVar(&ciName, "name", "", "Nom du projet (d\u00e9faut: nom du r\u00e9pertoire)")
	ciCmd.Flags().StringVar(&ciPort, "port", "8080", "Port du conteneur de configuration (web UI)")
	ciCmd.Flags().StringVar(&ciNetwork, "network", "ade-network", "Nom du r\u00e9seau Docker")

	initCmd := &cobra.Command{Use: "init", SilenceUsage: true, SilenceErrors: true}
	initCmd.AddCommand(ciCmd)
	rootCmd := &cobra.Command{Use: "ade", SilenceUsage: true, SilenceErrors: true}
	rootCmd.AddCommand(initCmd)
	return rootCmd
}

func execCi(args ...string) (string, error) {
	root := buildCiInitCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func TestInitCi_DetectDocker(t *testing.T) {
	savedCheck := dockerCheckFn
	defer func() { dockerCheckFn = savedCheck }()
	savedClient := dockerNewClientFn
	defer func() { dockerNewClientFn = savedClient }()
	savedLookPath := execLookPathFn
	defer func() { execLookPathFn = savedLookPath }()
	savedPlugin := checkDockerPluginFn
	defer func() { checkDockerPluginFn = savedPlugin }()

	dockerCheckFn = func() (string, error) { return "docker", nil }
	dockerNewClientFn = func() (docker.Client, error) {
		return &mockDockerClient{
			pingFunc: func(_ context.Context) error { return nil },
		}, nil
	}
	execLookPathFn = func(_ string) (string, error) { return "", fmt.Errorf("not found") }
	checkDockerPluginFn = func() bool { return false }

	output, err := execCi("init", "ci")
	assert.NoError(t, err)
	assert.Contains(t, output, "docker trouv\u00e9")
}

func TestInitCi_DetectPodman(t *testing.T) {
	savedCheck := dockerCheckFn
	defer func() { dockerCheckFn = savedCheck }()
	savedClient := dockerNewClientFn
	defer func() { dockerNewClientFn = savedClient }()
	savedLookPath := execLookPathFn
	defer func() { execLookPathFn = savedLookPath }()
	savedPlugin := checkDockerPluginFn
	defer func() { checkDockerPluginFn = savedPlugin }()

	dockerCheckFn = func() (string, error) { return "podman", nil }
	dockerNewClientFn = func() (docker.Client, error) {
		return &mockDockerClient{
			pingFunc: func(_ context.Context) error { return nil },
		}, nil
	}
	execLookPathFn = func(_ string) (string, error) { return "", fmt.Errorf("not found") }
	checkDockerPluginFn = func() bool { return false }

	output, err := execCi("init", "ci")
	assert.NoError(t, err)
	assert.Contains(t, output, "podman trouv\u00e9")
}

func TestInitCi_NoDockerFound(t *testing.T) {
	savedCheck := dockerCheckFn
	defer func() { dockerCheckFn = savedCheck }()

	dockerCheckFn = func() (string, error) { return "", fmt.Errorf("not found") }

	_, err := execCi("init", "ci")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Docker ou Podman requis")
}

func TestInitCi_DaemonUnreachable(t *testing.T) {
	savedCheck := dockerCheckFn
	defer func() { dockerCheckFn = savedCheck }()
	savedClient := dockerNewClientFn
	defer func() { dockerNewClientFn = savedClient }()

	dockerCheckFn = func() (string, error) { return "docker", nil }
	dockerNewClientFn = func() (docker.Client, error) {
		return &mockDockerClient{
			pingFunc: func(_ context.Context) error { return fmt.Errorf("connection refused") },
		}, nil
	}

	_, err := execCi("init", "ci")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "d\u00e9mon")
}

func TestInitCi_GeneratesFiles(t *testing.T) {
	savedCheck := dockerCheckFn
	defer func() { dockerCheckFn = savedCheck }()
	savedClient := dockerNewClientFn
	defer func() { dockerNewClientFn = savedClient }()
	savedLookPath := execLookPathFn
	defer func() { execLookPathFn = savedLookPath }()
	savedPlugin := checkDockerPluginFn
	defer func() { checkDockerPluginFn = savedPlugin }()

	dockerCheckFn = func() (string, error) { return "docker", nil }
	dockerNewClientFn = func() (docker.Client, error) {
		return &mockDockerClient{
			pingFunc: func(_ context.Context) error { return nil },
		}, nil
	}
	execLookPathFn = func(_ string) (string, error) { return "", fmt.Errorf("not found") }
	checkDockerPluginFn = func() bool { return false }

	dir := t.TempDir()
	output, err := execCi("init", "ci", "--output", dir, "--force", "--name", "test-proj")
	assert.NoError(t, err)

	composePath := filepath.Join(dir, "docker-compose.yml")
	envPath := filepath.Join(dir, ".env")

	assert.FileExists(t, composePath, "docker-compose.yml should exist")
	assert.FileExists(t, envPath, ".env should exist")

	data, _ := os.ReadFile(envPath)
	assert.Contains(t, string(data), "ADE_PROJECT_NAME=test-proj")

	assert.Contains(t, output, "docker-compose.yml")
	assert.Contains(t, output, ".env")
}

func TestInitCi_ProjectNameDefaults(t *testing.T) {
	savedCheck := dockerCheckFn
	defer func() { dockerCheckFn = savedCheck }()
	savedClient := dockerNewClientFn
	defer func() { dockerNewClientFn = savedClient }()
	savedLookPath := execLookPathFn
	defer func() { execLookPathFn = savedLookPath }()
	savedPlugin := checkDockerPluginFn
	defer func() { checkDockerPluginFn = savedPlugin }()

	dockerCheckFn = func() (string, error) { return "docker", nil }
	dockerNewClientFn = func() (docker.Client, error) {
		return &mockDockerClient{
			pingFunc: func(_ context.Context) error { return nil },
		}, nil
	}
	execLookPathFn = func(_ string) (string, error) { return "", fmt.Errorf("not found") }
	checkDockerPluginFn = func() bool { return false }

	dir := t.TempDir()
	_, err := execCi("init", "ci", "--output", dir, "--force", "--name", "mon-projet")
	assert.NoError(t, err)

	envPath := filepath.Join(dir, ".env")
	data, _ := os.ReadFile(envPath)
	assert.Contains(t, string(data), "ADE_PROJECT_NAME=mon-projet")
}

func TestInitCi_CustomPort(t *testing.T) {
	savedCheck := dockerCheckFn
	defer func() { dockerCheckFn = savedCheck }()
	savedClient := dockerNewClientFn
	defer func() { dockerNewClientFn = savedClient }()
	savedLookPath := execLookPathFn
	defer func() { execLookPathFn = savedLookPath }()
	savedPlugin := checkDockerPluginFn
	defer func() { checkDockerPluginFn = savedPlugin }()

	dockerCheckFn = func() (string, error) { return "docker", nil }
	dockerNewClientFn = func() (docker.Client, error) {
		return &mockDockerClient{
			pingFunc: func(_ context.Context) error { return nil },
		}, nil
	}
	execLookPathFn = func(_ string) (string, error) { return "", fmt.Errorf("not found") }
	checkDockerPluginFn = func() bool { return false }

	dir := t.TempDir()
	_, err := execCi("init", "ci", "--output", dir, "--force", "--name", "p", "--port", "9090")
	assert.NoError(t, err)

	composePath := filepath.Join(dir, "docker-compose.yml")
	data, _ := os.ReadFile(composePath)
	assert.Contains(t, string(data), "9090:80")
}

func TestInitCi_ComposeNotFoundContinues(t *testing.T) {
	savedCheck := dockerCheckFn
	defer func() { dockerCheckFn = savedCheck }()
	savedClient := dockerNewClientFn
	defer func() { dockerNewClientFn = savedClient }()
	savedLookPath := execLookPathFn
	defer func() { execLookPathFn = savedLookPath }()
	savedPlugin := checkDockerPluginFn
	defer func() { checkDockerPluginFn = savedPlugin }()

	dockerCheckFn = func() (string, error) { return "docker", nil }
	dockerNewClientFn = func() (docker.Client, error) {
		return &mockDockerClient{
			pingFunc: func(_ context.Context) error { return nil },
		}, nil
	}
	execLookPathFn = func(_ string) (string, error) { return "", fmt.Errorf("not found") }
	checkDockerPluginFn = func() bool { return false }

	dir := t.TempDir()
	output, err := execCi("init", "ci", "--output", dir, "--force", "--name", "test")
	assert.NoError(t, err)
	assert.Contains(t, output, "Commande compose non trouv\u00e9e")
}

func TestInitCi_HelpContainsNewFlags(t *testing.T) {
	output, err := execCi("init", "ci", "--help")
	assert.NoError(t, err)
	assert.Contains(t, output, "--output")
	assert.Contains(t, output, "--force")
	assert.Contains(t, output, "--name")
	assert.Contains(t, output, "--port")
	assert.Contains(t, output, "--network")
}

func TestInitCi_DeployAndShowStatus(t *testing.T) {
	savedCheck := dockerCheckFn
	defer func() { dockerCheckFn = savedCheck }()
	savedClient := dockerNewClientFn
	defer func() { dockerNewClientFn = savedClient }()
	savedLookPath := execLookPathFn
	defer func() { execLookPathFn = savedLookPath }()
	savedPlugin := checkDockerPluginFn
	defer func() { checkDockerPluginFn = savedPlugin }()
	savedExecCmd := executeCmdFn
	defer func() { executeCmdFn = savedExecCmd }()

	dockerCheckFn = func() (string, error) { return "docker", nil }
	dockerNewClientFn = func() (docker.Client, error) {
		return &mockDockerClient{
			pingFunc: func(_ context.Context) error { return nil },
		}, nil
	}
	execLookPathFn = func(name string) (string, error) {
		if name == "docker" {
			return "docker", nil
		}
		return "", fmt.Errorf("not found")
	}
	checkDockerPluginFn = func() bool { return true }
	executeCmdFn = func(_ *exec.Cmd) error { return nil }

	dir := t.TempDir()
	output, err := execCi("init", "ci", "--output", dir, "--force", "--name", "test")
	assert.NoError(t, err)
	assert.Contains(t, output, "D\u00e9ploiement avec docker")
	assert.Contains(t, output, "Statut des conteneurs")
}

func TestInitCi_FallbackToDockerCompose(t *testing.T) {
	savedCheck := dockerCheckFn
	defer func() { dockerCheckFn = savedCheck }()
	savedClient := dockerNewClientFn
	defer func() { dockerNewClientFn = savedClient }()
	savedLookPath := execLookPathFn
	defer func() { execLookPathFn = savedLookPath }()
	savedPlugin := checkDockerPluginFn
	defer func() { checkDockerPluginFn = savedPlugin }()
	savedExecCmd := executeCmdFn
	defer func() { executeCmdFn = savedExecCmd }()

	dockerCheckFn = func() (string, error) { return "docker", nil }
	dockerNewClientFn = func() (docker.Client, error) {
		return &mockDockerClient{
			pingFunc: func(_ context.Context) error { return nil },
		}, nil
	}
	checkDockerPluginFn = func() bool { return false }
	execLookPathFn = func(name string) (string, error) {
		if name == "docker" {
			return "docker", nil
		}
		if name == "docker-compose" {
			return "docker-compose", nil
		}
		return "", fmt.Errorf("not found")
	}
	executeCmdFn = func(_ *exec.Cmd) error { return nil }

	dir := t.TempDir()
	output, err := execCi("init", "ci", "--output", dir, "--force", "--name", "test")
	assert.NoError(t, err)
	assert.Contains(t, output, "D\u00e9ploiement avec docker")
}

func TestInitCi_DeployFailureNonFatal(t *testing.T) {
	savedCheck := dockerCheckFn
	defer func() { dockerCheckFn = savedCheck }()
	savedClient := dockerNewClientFn
	defer func() { dockerNewClientFn = savedClient }()
	savedLookPath := execLookPathFn
	defer func() { execLookPathFn = savedLookPath }()
	savedPlugin := checkDockerPluginFn
	defer func() { checkDockerPluginFn = savedPlugin }()
	savedExecCmd := executeCmdFn
	defer func() { executeCmdFn = savedExecCmd }()

	dockerCheckFn = func() (string, error) { return "docker", nil }
	dockerNewClientFn = func() (docker.Client, error) {
		return &mockDockerClient{
			pingFunc: func(_ context.Context) error { return nil },
		}, nil
	}
	execLookPathFn = func(name string) (string, error) {
		if name == "docker" {
			return "docker", nil
		}
		return "", fmt.Errorf("not found")
	}
	checkDockerPluginFn = func() bool { return true }
	executeCmdFn = func(_ *exec.Cmd) error { return fmt.Errorf("compose unavailable") }

	dir := t.TempDir()
	output, err := execCi("init", "ci", "--output", dir, "--force", "--name", "test")
	assert.NoError(t, err)
	assert.Contains(t, output, "Les fichiers sont g\u00e9n\u00e9r\u00e9s")
}

func TestInitCi_DefaultPortAndNetwork(t *testing.T) {
	savedCheck := dockerCheckFn
	defer func() { dockerCheckFn = savedCheck }()
	savedClient := dockerNewClientFn
	defer func() { dockerNewClientFn = savedClient }()
	savedLookPath := execLookPathFn
	defer func() { execLookPathFn = savedLookPath }()
	savedPlugin := checkDockerPluginFn
	defer func() { checkDockerPluginFn = savedPlugin }()

	dockerCheckFn = func() (string, error) { return "docker", nil }
	dockerNewClientFn = func() (docker.Client, error) {
		return &mockDockerClient{
			pingFunc: func(_ context.Context) error { return nil },
		}, nil
	}
	execLookPathFn = func(_ string) (string, error) { return "", fmt.Errorf("not found") }
	checkDockerPluginFn = func() bool { return false }

	dir := t.TempDir()
	_, err := execCi("init", "ci", "--output", dir, "--force", "--name", "test")
	assert.NoError(t, err)

	composePath := filepath.Join(dir, "docker-compose.yml")
	data, _ := os.ReadFile(composePath)
	assert.Contains(t, string(data), "8080:80")
	assert.Contains(t, string(data), "ade-network")
}

func TestInitCi_GetComposeCommand(t *testing.T) {
	savedLookPath := execLookPathFn
	defer func() { execLookPathFn = savedLookPath }()
	savedPlugin := checkDockerPluginFn
	defer func() { checkDockerPluginFn = savedPlugin }()

	execLookPathFn = func(name string) (string, error) { return "", fmt.Errorf("not found") }
	checkDockerPluginFn = func() bool { return false }

	cmd, args := getComposeCommand("unknown")
	assert.Empty(t, cmd)
	assert.Nil(t, args)
}

func TestInitCi_GetComposeCommandDockerFallback(t *testing.T) {
	savedLookPath := execLookPathFn
	defer func() { execLookPathFn = savedLookPath }()
	savedPlugin := checkDockerPluginFn
	defer func() { checkDockerPluginFn = savedPlugin }()

	checkDockerPluginFn = func() bool { return false }
	calls := 0
	execLookPathFn = func(name string) (string, error) {
		calls++
		if name == "docker" {
			return "docker", nil
		}
		if name == "docker-compose" {
			return "docker-compose", nil
		}
		return "", fmt.Errorf("not found")
	}

	cmd, args := getComposeCommand("docker")
	assert.Equal(t, "docker-compose", cmd)
	assert.Nil(t, args)
	assert.Equal(t, 2, calls)
}

func TestInitCi_GetComposeCommandPodman(t *testing.T) {
	savedLookPath := execLookPathFn
	defer func() { execLookPathFn = savedLookPath }()

	execLookPathFn = func(name string) (string, error) {
		if name == "podman-compose" {
			return "podman-compose", nil
		}
		return "", fmt.Errorf("not found")
	}

	cmd, args := getComposeCommand("podman")
	assert.Equal(t, "podman-compose", cmd)
	assert.Nil(t, args)
}

func TestInitCi_TemplateDataMapping(t *testing.T) {
	opts := ciOptions{
		OutputDir:      "/tmp",
		Force:          true,
		ProjectName:    "mon-app",
		ConfigPort:     "9090",
		ComposeNetwork: "my-net",
	}

	data := opts.toTemplateData()
	assert.Equal(t, "mon-app", data.ProjectName)
	assert.Equal(t, "9090", data.Compose.ConfigPort)
	assert.Equal(t, "my-net", data.Compose.Network)
}

func TestInitCi_RenderComposeFromOptions(t *testing.T) {
	opts := ciOptions{
		ProjectName:    "test-app",
		ConfigPort:     "7070",
		ComposeNetwork: "test-net",
	}

	data := opts.toTemplateData()
	out, err := templates.Render("docker-compose", data)
	assert.NoError(t, err)
	assert.Contains(t, out, "nginx:alpine")
	assert.Contains(t, out, "7070:80")
	assert.Contains(t, out, "test-net")
}
