package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initSpecsCmd = &cobra.Command{
	Use:   "specs",
	Short: "Initialise les fichiers de spécification",
	Long:  `Génère les fichiers de configuration du projet (.gitignore, IDE, skills, etc.).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), "Initialisation des spécifications... (à implémenter)")
		return nil
	},
}
