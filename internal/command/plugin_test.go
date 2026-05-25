package command

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPluginList_NoPlugins(t *testing.T) {
	saved := listPluginsFn
	defer func() { listPluginsFn = saved }()

	listPluginsFn = func(ctx context.Context) (string, error) {
		return "Aucun plugin enregistré.", nil
	}

	output, err := executeCommand("plugin", "list")
	assert.NoError(t, err)
	assert.Contains(t, output, "Aucun plugin enregistré.")
}

func TestPluginList_WithPlugins(t *testing.T) {
	saved := listPluginsFn
	defer func() { listPluginsFn = saved }()

	listPluginsFn = func(ctx context.Context) (string, error) {
		return "NOM\tVERSION\tSTATUT\np1\t1.0.0\tHEALTHY\n", nil
	}

	output, err := executeCommand("plugin", "list")
	assert.NoError(t, err)
	assert.Contains(t, output, "p1")
}

func TestPluginList_OrchestratorUnreachable(t *testing.T) {
	saved := listPluginsFn
	defer func() { listPluginsFn = saved }()

	listPluginsFn = func(ctx context.Context) (string, error) {
		return "", errors.New("orchestrateur non disponible")
	}

	_, err := executeCommand("plugin", "list")
	assert.NoError(t, err)
}

func TestPluginInfo_Existing(t *testing.T) {
	saved := getPluginInfoFn
	defer func() { getPluginInfoFn = saved }()

	getPluginInfoFn = func(ctx context.Context, name string) (string, error) {
		return "Nom         : test-plugin\nVersion     : 1.0.0\n", nil
	}

	output, err := executeCommand("plugin", "info", "test-plugin")
	assert.NoError(t, err)
	assert.Contains(t, output, "test-plugin")
}

func TestPluginInfo_NotFound(t *testing.T) {
	saved := getPluginInfoFn
	defer func() { getPluginInfoFn = saved }()

	getPluginInfoFn = func(ctx context.Context, name string) (string, error) {
		return "", errors.New("plugin \"unknown\" introuvable")
	}

	_, err := executeCommand("plugin", "info", "unknown")
	assert.NoError(t, err)
}

func TestPluginInfo_MissingArg(t *testing.T) {
	_, err := executeCommand("plugin", "info")
	assert.Error(t, err)
}

func TestPluginCmd_NoSubcommand(t *testing.T) {
	output, err := executeCommand("plugin")
	assert.NoError(t, err)
	assert.Contains(t, output, "Usage")
}

func TestPluginInstall(t *testing.T) {
	saved := installPluginFn
	defer func() { installPluginFn = saved }()

	installed := ""
	installPluginFn = func(ctx context.Context, image string) error {
		installed = image
		return nil
	}

	_, err := executeCommand("plugin", "install", "my-plugin:latest")
	assert.NoError(t, err)
	assert.Equal(t, "my-plugin:latest", installed)
}

func TestPluginInstall_MissingArg(t *testing.T) {
	_, err := executeCommand("plugin", "install")
	assert.Error(t, err)
}

func TestPluginUninstall(t *testing.T) {
	saved := uninstallPluginFn
	defer func() { uninstallPluginFn = saved }()

	uninstalled := ""
	uninstallPluginFn = func(ctx context.Context, name string) error {
		uninstalled = name
		return nil
	}

	_, err := executeCommand("plugin", "uninstall", "my-plugin")
	assert.NoError(t, err)
	assert.Equal(t, "my-plugin", uninstalled)
}

func TestPluginUninstall_MissingArg(t *testing.T) {
	_, err := executeCommand("plugin", "uninstall")
	assert.Error(t, err)
}

func TestPluginHelp(t *testing.T) {
	output, err := executeCommand("plugin", "--help")
	assert.NoError(t, err)
	assert.True(t, strings.Contains(output, "list") || strings.Contains(output, "Liste"))
}
