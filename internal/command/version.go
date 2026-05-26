package command

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"automated_dev_environment/internal/orchestrator"
)

const Version = "0.1.0"

var showOrchestratorStatus bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Affiche la version de l'outil",
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "ade version %s\n", Version)

		if showOrchestratorStatus {
			fmt.Fprintln(out, "")
			printOrchestratorStatus(out)
		}

		return nil
	},
}

func printOrchestratorStatus(out io.Writer) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	client := orchestrator.NewClient()
	if err := client.Health(ctx); err != nil {
		fmt.Fprintf(out, "Orchestrateur: indisponible (%v)\n", err)
		return
	}

	cfg, err := client.GetConfig(ctx)
	if err != nil {
		fmt.Fprintf(out, "Orchestrateur: connect\u00e9 (version inconnue)\n")
		return
	}
	fmt.Fprintf(out, "Orchestrateur: connect\u00e9 (version %s, projets: %s)\n",
		cfg.OrchestratorVersion, cfg.ProjectName)
}

func init() {
	versionCmd.Flags().BoolVarP(&showOrchestratorStatus, "orchestrator", "o", false, "Afficher le statut de l'orchestrateur")
}
