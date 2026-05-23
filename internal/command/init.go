package command

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"automated_dev_environment/internal/agentic"
)

var (
	initForce       bool
	initOutput      string
	initConfig      string
	initSkipTools   bool
	initSkipSkills  bool
	initSkipMCP     bool
	initHaltOnError bool
)

// mockable function vars for testing
var (
	detectToolsFn  = agentic.DetectTools
	ensureSkillsFn = agentic.EnsureSkills
	configureMCPFn = agentic.ConfigureMCPServers
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise les composants du projet",
	Long:  `Initialise les composants du projet de développement agentic.`,
	RunE:  runAgenticSetup,
}

func runAgenticSetup(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	ctx := cmd.Context()
	outputDir := initOutput
	if outputDir == "" {
		outputDir = "."
	}

	hasFatal := false
	anySuccess := false

	fmt.Fprintln(out, "✓ Configuration agentic terminée")

	if !initSkipTools {
		if err := detectToolsAndReport(ctx, out, initConfig); err != nil {
			hasFatal = true
			if initHaltOnError {
				return fmt.Errorf("détection des outils: %w", err)
			}
		} else {
			anySuccess = true
		}
	}

	if !initSkipSkills {
		if err := installSkillsAndReport(ctx, out, outputDir, initForce); err != nil {
			hasFatal = true
			if initHaltOnError {
				return fmt.Errorf("installation des skills: %w", err)
			}
		} else {
			anySuccess = true
		}
	}

	if !initSkipMCP {
		if err := configureMCPAndReport(ctx, out, agentic.MCPOptions{
			OutputDir:  outputDir,
			ConfigPath: initConfig,
		}); err != nil {
			hasFatal = true
			if initHaltOnError {
				return fmt.Errorf("configuration MCP: %w", err)
			}
		} else {
			anySuccess = true
		}
	}

	if !anySuccess && hasFatal {
		return fmt.Errorf("toutes les étapes ont échoué")
	}
	return nil
}

func detectToolsAndReport(ctx context.Context, out io.Writer, configPath string) error {
	result, err := detectToolsFn(ctx, configPath)
	if err != nil {
		fmt.Fprintf(out, "\nOutils détectés :\n  ✗ Erreur : %v\n", err)
		return err
	}

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Outils détectés :")
	for _, t := range result.Tools {
		if t.Found {
			fmt.Fprintf(out, "  ✓ %s trouvé : %s\n", t.Name, t.Path)
		} else {
			inst := agentic.InstallInstructions(t.Type)
			fmt.Fprintf(out, "  ✗ %s non trouvé → %s\n", t.Name, inst.URL)
			fmt.Fprintf(out, "    %s\n", inst.Message)
		}
	}
	return nil
}

func installSkillsAndReport(ctx context.Context, out io.Writer, outputDir string, force bool) error {
	report, err := ensureSkillsFn(ctx, outputDir, force)
	if err != nil {
		fmt.Fprintf(out, "\nSkills :\n  ✗ Erreur : %v\n", err)
		return err
	}

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Skills :")
	if len(report.Installed) > 0 {
		fmt.Fprintf(out, "  ✓ %d skills installés\n", len(report.Installed))
	}
	if len(report.AlreadyExist) > 0 {
		fmt.Fprintf(out, "  ∼ %d skills déjà présents\n", len(report.AlreadyExist))
	}
	if len(report.Errors) > 0 {
		fmt.Fprintf(out, "  ✗ %d erreurs\n", len(report.Errors))
		for _, e := range report.Errors {
			fmt.Fprintf(out, "    - %s : %v\n", e.Name, e.Err)
		}
		return fmt.Errorf("%d erreur(s) lors de l'installation des skills", len(report.Errors))
	}
	if len(report.Installed) == 0 && len(report.AlreadyExist) == 0 {
		fmt.Fprintln(out, "  Aucun skill à installer")
	}
	return nil
}

func configureMCPAndReport(ctx context.Context, out io.Writer, opts agentic.MCPOptions) error {
	report, err := configureMCPFn(ctx, opts)
	if err != nil {
		fmt.Fprintf(out, "\nServeurs MCP :\n  ✗ Erreur : %v\n", err)
		return err
	}

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Serveurs MCP :")
	if report.Added > 0 || report.Updated > 0 {
		total := report.Added + report.Updated
		fmt.Fprintf(out, "  ✓ %d serveurs configurés dans .opencode/config.json\n", total)
	}
	if report.Unchanged > 0 {
		fmt.Fprintf(out, "  ∼ %d serveurs déjà existants\n", report.Unchanged)
	}
	if report.Errors > 0 {
		fmt.Fprintf(out, "  ✗ %d erreurs\n", report.Errors)
		return fmt.Errorf("%d erreur(s) lors de la configuration MCP", report.Errors)
	}
	if report.Added == 0 && report.Updated == 0 && report.Unchanged == 0 {
		fmt.Fprintln(out, "  Aucun serveur MCP configuré")
	}
	return nil
}

func init() {
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Écraser les fichiers existants sans confirmation")
	initCmd.Flags().StringVarP(&initOutput, "output", "o", ".", "Répertoire du projet")
	initCmd.Flags().StringVar(&initConfig, "config", "", "Chemin du fichier de configuration YAML (par défaut : auto-détection)")
	initCmd.Flags().BoolVar(&initSkipTools, "skip-tools", false, "Ignorer la détection des outils")
	initCmd.Flags().BoolVar(&initSkipSkills, "skip-skills", false, "Ignorer l'installation des skills")
	initCmd.Flags().BoolVar(&initSkipMCP, "skip-mcp", false, "Ignorer la configuration MCP")
	initCmd.Flags().BoolVar(&initHaltOnError, "halt-on-error", false, "Arrêter le setup à la première erreur")

	initCmd.AddCommand(initSpecsCmd)
	initCmd.AddCommand(initCiCmd)
	rootCmd.AddCommand(initCmd)
}
