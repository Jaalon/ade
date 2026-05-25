package command

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"automated_dev_environment/internal/validation"
)

var (
	valLoadConfigFn      = validation.LoadConfig
	valValidateConfigFn  = validation.ValidateConfig
	valNewRunnerFn       = func() *validation.ValidationRunner { return validation.NewValidationRunner() }
	valDetectModulesFn   = validation.DetectModules
	valNewReportWriterFn = validation.NewReportWriter
	valStoreReportFn     = validation.StoreValidationReport
)

var (
	valRunConfig  string
	valRunOutput  string
	valRunFormat  string
	valRunVerbose bool
)

var validateRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Ex\u00e9cute les modules de validation",
	Long:  `Ex\u00e9cute les modules de validation de l'environnement et g\u00e9n\u00e8re les rapports.`,
	RunE:  runValidate,
}

type validateRunOptions struct {
	ConfigPath string
	OutputDir  string
	Formats    []string
	Verbose    bool
}

func runValidate(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	opts := validateRunOptions{
		ConfigPath: valRunConfig,
		OutputDir:  valRunOutput,
		Formats:    parseFormats(valRunFormat),
		Verbose:    valRunVerbose,
	}

	cfg, err := loadValidateConfig(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("chargement de la configuration: %w", err)
	}

	if opts.OutputDir != "" {
		cfg.OutputDir = opts.OutputDir
	}
	if len(opts.Formats) > 0 {
		cfg.Formats = opts.Formats
	}

	if err := ensureOutputDir(cfg.OutputDir); err != nil {
		return fmt.Errorf("pr\u00e9paration du r\u00e9pertoire de sortie: %w", err)
	}

	runner := valNewRunnerFn()
	report, err := runner.Run(cmd.Context(), cfg)
	if err != nil {
		return fmt.Errorf("ex\u00e9cution de la validation: %w", err)
	}

	fmt.Fprintln(out, "")
	displayValidationResult(out, report, opts.Verbose)

	if errs := writeValidationReports(report, opts); len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(out, "  \u26a0 Erreur d'\u00e9criture du rapport: %v\n", e)
		}
	}

	if err := storeOrchestratorReport(cmd.Context(), out, report, cfg.OutputDir); err != nil {
		fmt.Fprintf(out, "  \u26a0 Rapport orchestrateur: %v\n", err)
	}

	return nil
}

func loadValidateConfig(path string) (validation.ValidationConfig, error) {
	return valLoadConfigFn(path)
}

func displayValidationResult(out io.Writer, report *validation.ValidationReport, verbose bool) {
	statusLabel := "termin\u00e9"
	if report.Failed() {
		statusLabel = "termin\u00e9 avec \u00e9checs"
	}
	fmt.Fprintf(out, "\u2713 Validation %s (%d modules, %d checks, %d passed, %d failed)\n\n",
		statusLabel, len(report.Modules), report.NumChecks(), report.NumPassed(), report.NumFailed())

	fmt.Fprintf(out, "  %-16s %-10s %-10s %s\n", "Module", "Statut", "Checks", "Dur\u00e9e")
	for _, m := range report.Modules {
		symbol := statusSymbol(m.Status)
		dur := formatDuration(m.Duration)
		checks := fmt.Sprintf("%d/%d", passedCount(m), len(m.Checks))
		fmt.Fprintf(out, "  %-16s %-10s %-10s %s\n", m.ModuleName, symbol, checks, dur)

		if verbose {
			for _, c := range m.Checks {
				cs := statusSymbol(c.Status)
				cd := formatDuration(c.Duration)
				fmt.Fprintf(out, "    %-20s %-10s %-10s %s\n", c.Name, cs, cd, c.Message)
				if c.Details != "" {
					fmt.Fprintf(out, "      %s\n", c.Details)
				}
			}
		}
	}

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "  Rapports g\u00e9n\u00e9r\u00e9s :")
	for _, f := range reportFormats(report) {
		fmt.Fprintf(out, "    \u2713 %s\n", f)
	}
}

func passedCount(m validation.ModuleResult) int {
	count := 0
	for _, c := range m.Checks {
		if c.Status == validation.StatusPassed {
			count++
		}
	}
	return count
}

func statusSymbol(s validation.CheckStatus) string {
	switch s {
	case validation.StatusPassed:
		return "\u2713 pass\u00e9"
	case validation.StatusFailed:
		return "\u2717 \u00e9chec"
	case validation.StatusWarning:
		return "\u26a0 avertissement"
	case validation.StatusSkipped:
		return "\u2014 ignor\u00e9"
	case validation.StatusError:
		return "\u2717 erreur"
	default:
		return "?"
	}
}

func reportFormats(report *validation.ValidationReport) []string {
	if report == nil {
		return nil
	}
	dir := report.Config.OutputDir
	if dir == "" {
		dir = ".ade/validation"
	}
	var formats []string
	for _, f := range report.Config.Formats {
		switch f {
		case "json":
			formats = append(formats, filepath.Join(dir, "report.json"))
		case "junit":
			formats = append(formats, filepath.Join(dir, "report.xml"))
		}
	}
	return formats
}

func writeValidationReports(report *validation.ValidationReport, opts validateRunOptions) []error {
	var errs []error
	for _, format := range opts.Formats {
		w := valNewReportWriterFn(format)
		if w == nil {
			errs = append(errs, fmt.Errorf("format inconnu: %s", format))
			continue
		}

		var filename string
		switch format {
		case "json":
			filename = "report.json"
		case "junit":
			filename = "report.xml"
		default:
			filename = "report." + format
		}

		path := filepath.Join(opts.OutputDir, filename)
		f, err := os.Create(path)
		if err != nil {
			errs = append(errs, fmt.Errorf("cr\u00e9ation de %s: %w", path, err))
			continue
		}

		if err := w.Write(report, f); err != nil {
			f.Close()
			errs = append(errs, fmt.Errorf("\u00e9criture de %s: %w", path, err))
			continue
		}
		f.Close()
	}
	return errs
}

func ensureOutputDir(path string) error {
	if path == "" {
		return nil
	}
	return os.MkdirAll(path, 0755)
}

func parseFormats(formatFlag string) []string {
	if formatFlag == "" {
		return []string{"json"}
	}
	parts := strings.Split(formatFlag, ",")
	var formats []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			formats = append(formats, p)
		}
	}
	if len(formats) == 0 {
		return []string{"json"}
	}
	return formats
}

func storeOrchestratorReport(ctx context.Context, out io.Writer, report *validation.ValidationReport, outputDir string) error {
	if err := valStoreReportFn(ctx, report, outputDir); err != nil {
		return err
	}
	fmt.Fprintf(out, "  \u2248 Int\u00e9gration orchestrateur : disponible dans Story #008\n")
	return nil
}

func init() {
	validateRunCmd.Flags().StringVarP(&valRunConfig, "config", "c", "ade-validate.yaml", "Chemin du fichier de configuration de validation")
	validateRunCmd.Flags().StringVarP(&valRunOutput, "output", "o", ".ade/validation", "R\u00e9pertoire de sortie des rapports")
	validateRunCmd.Flags().StringVarP(&valRunFormat, "format", "f", "json", "Format de rapport (json, junit, ou json,junit)")
	validateRunCmd.Flags().BoolVarP(&valRunVerbose, "verbose", "v", false, "Afficher les d\u00e9tails de chaque check")
}
