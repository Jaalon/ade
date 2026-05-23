package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCiCmd = &cobra.Command{
	Use:   "ci",
	Short: "Initialise l'intégration continue",
	Long:  `Déploie l'environnement de préproduction locale via Docker.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), "Initialisation de l'intégration continue... (à implémenter)")
		return nil
	},
}
