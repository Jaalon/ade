package command

import (
	"context"

	"github.com/spf13/cobra"

	"automated_dev_environment/internal/orchestrator"
)

var getPluginInfoFn = func(ctx context.Context, name string) (string, error) {
	client := orchestrator.NewClient()
	plugin, err := client.GetPlugin(ctx, name)
	if err != nil {
		return "", err
	}
	return orchestrator.FormatPluginInfo(plugin), nil
}

var pluginInfoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Affiche les détails d'un plugin",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		output, err := getPluginInfoFn(cmd.Context(), args[0])
		if err != nil {
			cmd.PrintErrln("⚠", err)
			return nil
		}
		cmd.Print(output)
		return nil
	},
}
