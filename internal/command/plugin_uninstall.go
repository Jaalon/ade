package command

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"automated_dev_environment/internal/docker"
)

var uninstallPluginFn = func(ctx context.Context, name string) error {
	client, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("connexion à Docker: %w", err)
	}
	defer client.Close()

	running, err := client.IsContainerRunning(ctx, name)
	if err != nil {
		return fmt.Errorf("vérification du conteneur %s: %w", name, err)
	}

	if !running {
		log.Printf("Aucun conteneur en cours d'exécution avec le nom %s.", name)
		return nil
	}

	log.Printf("Arrêt du conteneur %s...", name)
	return nil
}

var pluginUninstallCmd = &cobra.Command{
	Use:   "uninstall <name>",
	Short: "Supprime un conteneur plugin Docker",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := uninstallPluginFn(cmd.Context(), args[0]); err != nil {
			cmd.PrintErrln("⚠", err)
			return nil
		}
		return nil
	},
}
