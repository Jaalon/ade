package command

import (
	"bytes"
	"context"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"automated_dev_environment/internal/orchestrator"
)

var listPluginsFn = func(ctx context.Context) (string, error) {
	client := orchestrator.NewClient()
	plugins, err := client.ListPlugins(ctx)
	if err != nil {
		return "", err
	}
	raw := orchestrator.FormatPluginList(plugins)

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', 0)
	w.Write([]byte(raw))
	w.Flush()
	return buf.String(), nil
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "Liste les plugins enregistrés",
	RunE: func(cmd *cobra.Command, args []string) error {
		output, err := listPluginsFn(cmd.Context())
		if err != nil {
			cmd.PrintErrln("⚠", err)
			return nil
		}
		cmd.Print(output)
		return nil
	},
}
