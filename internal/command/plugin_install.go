package command

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"automated_dev_environment/internal/docker"
)

var installPluginFn = func(ctx context.Context, image string) error {
	dockerClient, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("connexion à Docker: %w", err)
	}
	defer dockerClient.Close()

	log.Printf("Téléchargement de l'image %s...", image)
	if err := dockerClient.PullImage(ctx, image); err != nil {
		return fmt.Errorf("téléchargement de %s: %w", image, err)
	}

	cfg := docker.ContainerExecConfig{
		Image: image,
	}

	log.Printf("Démarrage du conteneur %s...", image)
	_, err = dockerClient.RunContainer(ctx, cfg)
	if err != nil {
		return fmt.Errorf("démarrage de %s: %w", image, err)
	}

	log.Printf("Plugin %s installé avec succès.", image)
	return nil
}

var pluginInstallCmd = &cobra.Command{
	Use:   "install <image>",
	Short: "Installe un conteneur plugin Docker",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := installPluginFn(cmd.Context(), args[0]); err != nil {
			cmd.PrintErrln("⚠", err)
			return nil
		}
		return nil
	},
}
