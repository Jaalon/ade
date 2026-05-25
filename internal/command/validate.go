package command

import "github.com/spf13/cobra"

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Valide l'environnement de d\u00e9veloppement",
	Long: `Valide l'environnement de d\u00e9veloppement avec des modules de validation sp\u00e9cifiques.

Sous-commandes :
  run    Ex\u00e9cute les modules de validation
  init   G\u00e9n\u00e8re la configuration par d\u00e9faut de validation`,
}

func init() {
	validateCmd.AddCommand(validateRunCmd)
	validateCmd.AddCommand(validateInitCmd)
	rootCmd.AddCommand(validateCmd)
}
