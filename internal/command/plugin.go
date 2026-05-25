package command

import (
	"github.com/spf13/cobra"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Gère les plugins Docker",
	Long: `Gère les plugins Docker de l'environnement de développement.
	
Sous-commandes :
  list      Liste les plugins enregistrés
  info      Affiche les détails d'un plugin
  install   Installe un conteneur plugin Docker
  uninstall Supprime un conteneur plugin Docker`,
}

func init() {
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginInfoCmd)
	pluginCmd.AddCommand(pluginInstallCmd)
	pluginCmd.AddCommand(pluginUninstallCmd)
	rootCmd.AddCommand(pluginCmd)
}
