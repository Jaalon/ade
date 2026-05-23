package command

import "github.com/spf13/cobra"

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise les composants du projet",
	Long:  `Initialise les composants du projet de développement agentic.`,
}

func init() {
	initCmd.AddCommand(initSpecsCmd)
	initCmd.AddCommand(initCiCmd)
	rootCmd.AddCommand(initCmd)
}
