package command

import (
	"github.com/spf13/cobra"
)

var pipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "G\u00e8re le pipeline d'int\u00e9gration continue",
	Long: `G\u00e8re le pipeline d'int\u00e9gration continue locale.

Sous-commandes :
  run    Ex\u00e9cute le pipeline CI
  init   G\u00e9n\u00e8re la configuration par d\u00e9faut du pipeline`,
}

func init() {
	pipelineCmd.AddCommand(pipelineRunCmd)
	pipelineCmd.AddCommand(pipelineInitCmd)
	rootCmd.AddCommand(pipelineCmd)
}
