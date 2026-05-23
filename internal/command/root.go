package command

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"automated_dev_environment/internal/docker"
)

var rootCmd = &cobra.Command{
	Use:   "ade",
	Short: "Automated Dev Environment - CLI tool",
	Long: `Automated Dev Environment (ade) est un outil en ligne de commande pour
initialiser un environnement de développement agentic robuste sur Windows.

Il configure l'outillage agentic (OpenCode, Cursor), les IDE JetBrains,
l'intégration continue locale, le déploiement de préproduction via Docker,
et un système de validation modulaire de l'environnement.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func Execute() {
	args := os.Args[1:]
	if len(args) >= 1 && args[0] == "--version" {
		fmt.Println(Version)
		return
	}
	startDockerCheck()
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func startDockerCheck() {
	ctx := context.Background()
	err := docker.EnsureConfigContainer(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠ avertissement: %v\n", err)
	}
}
