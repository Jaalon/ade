package command

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"automated_dev_environment/internal/ci"
	"automated_dev_environment/internal/docker"
)

var (
	plLoadConfigFn     = ci.LoadConfig
	plValidateConfigFn = ci.ValidateConfig
	plNewDryRunFn      = func() ci.Executor { return ci.NewDryRunExecutor() }
	plNewLocalFn       = func() (ci.Executor, error) { return ci.NewLocalExecutor(), nil }
	plNewPipelineFn    = func(executor ci.Executor) ci.Pipeline {
		return ci.NewPipelineRunner(executor)
	}
	plDetectOrchFn = docker.EnsureConfigContainer
)

var (
	plRunConfigPath string
	plRunVerbose    bool
	plRunLocal      bool
	plRunDryRun     bool
)

var pipelineRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Ex\u00e9cute le pipeline d'int\u00e9gration continue",
	Long: `Ex\u00e9cute le pipeline d'int\u00e9gration continue locale.

Le pipeline se compose de 7 \u00e9tages ex\u00e9cut\u00e9s s\u00e9quentiellement :
build \u2192 validation \u2192 tests unitaires \u2192 tests d'int\u00e9gration \u2192 d\u00e9ploiement test \u2192 E2E \u2192 pr\u00e9production.

Par d\u00e9faut, le pipeline s'ex\u00e9cute en mode dry-run (simulation).
Utilisez --local pour ex\u00e9cuter les commandes directement sur la machine.`,
	RunE: runPipeline,
}

type pipelineRunOptions struct {
	ConfigPath string
	Verbose    bool
	Local      bool
	DryRun     bool
}

func runPipeline(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	ctx := cmd.Context()

	opts := pipelineRunOptions{
		ConfigPath: plRunConfigPath,
		Verbose:    plRunVerbose,
		Local:      plRunLocal,
		DryRun:     plRunDryRun,
	}

	cfg, err := loadPipelineConfig(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("chargement de la configuration: %w", err)
	}

	detectOrchestrator(ctx, out, opts.Verbose)

	executor, err := selectPipelineExecutor(ctx, opts)
	if err != nil {
		return fmt.Errorf("s\u00e9lection de l'ex\u00e9cuteur: %w", err)
	}

	pipeline := plNewPipelineFn(executor)

	if opts.Verbose {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Configuration du pipeline:")
		for _, s := range cfg.Stages {
			status := "activ\u00e9"
			if !s.Enabled {
				status = "d\u00e9sactiv\u00e9"
			}
			fmt.Fprintf(out, "  %s (%s)\n", s.Type, status)
		}
	}

	result, err := pipeline.Run(ctx, cfg)
	if err != nil {
		return fmt.Errorf("ex\u00e9cution du pipeline: %w", err)
	}

	fmt.Fprintln(out, "")
	displayPipelineResult(out, result, opts.Verbose)

	if result.Status == ci.StatusCancelled {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "\u26a1 Pipeline annul\u00e9 par l'utilisateur")
	}

	return nil
}

func detectOrchestrator(ctx context.Context, out io.Writer, verbose bool) {
	if !verbose {
		return
	}
	err := plDetectOrchFn(ctx)
	if err != nil {
		fmt.Fprintf(out, "  Orchestrateur: \u26a0 %v\n", err)
	} else {
		fmt.Fprintln(out, "  Orchestrateur: disponible")
	}
}

func selectPipelineExecutor(ctx context.Context, opts pipelineRunOptions) (ci.Executor, error) {
	if opts.DryRun {
		return plNewDryRunFn(), nil
	}

	if opts.Local {
		return plNewLocalFn()
	}

	return plNewDryRunFn(), nil
}

func loadPipelineConfig(path string) (ci.PipelineConfig, error) {
	return plLoadConfigFn(path)
}

func displayPipelineResult(out io.Writer, result *ci.PipelineResult, verbose bool) {
	statusLabel := statusText(result.Status)
	fmt.Fprintf(out, "Pipeline CI \u2014 %s (%s)\n", statusLabel, formatDuration(result.Duration))

	fmt.Fprintln(out, "\u250c\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2510")
	for _, stage := range result.Stages {
		s := stageStatusSymbol(stage.Status)
		name := stageName(stage)
		dur := ""
		if stage.Status != ci.StatusSkipped {
			dur = formatDuration(stage.Duration)
		}
		fmt.Fprintf(out, "\u2502 %-20s \u2502 %-9s \u2502 %8s \u2502\n", name, s, dur)
	}
	fmt.Fprintln(out, "\u2514\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2518")

	if verbose {
		fmt.Fprintln(out, "")
		for _, stage := range result.Stages {
			if stage.Status == ci.StatusSkipped {
				continue
			}
			fmt.Fprintf(out, "=== %s: %s ===\n", stage.Type, stage.Name)
			for _, step := range stage.Steps {
				if step.Output != "" {
					fmt.Fprintln(out, step.Output)
				}
				fmt.Fprintf(out, "%s (%s)\n", stageStatusSymbol(step.Status), formatDuration(step.Duration))
			}
			fmt.Fprintln(out, "")
		}
	}
}

func stageName(s ci.StageResult) string {
	if s.Name != "" {
		return s.Name
	}
	return string(s.Type)
}

func stageStatusSymbol(status ci.StageStatus) string {
	switch status {
	case ci.StatusSucceeded:
		return "\u2713 succ\u00e8s"
	case ci.StatusFailed:
		return "\u2717 \u00e9chec"
	case ci.StatusSkipped:
		return "\u2014 ignor\u00e9"
	case ci.StatusCancelled:
		return "\u26a1 annul\u00e9"
	case ci.StatusRunning:
		return "\u25b6 en cours"
	default:
		return "?"
	}
}

func statusText(s ci.StageStatus) string {
	switch s {
	case ci.StatusSucceeded:
		return "termin\u00e9"
	case ci.StatusFailed:
		return "termin\u00e9 avec \u00e9checs"
	case ci.StatusCancelled:
		return "annul\u00e9"
	default:
		return string(s)
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	sec := d.Seconds()
	if sec < 60 {
		return fmt.Sprintf("%.2fs", sec)
	}
	return fmt.Sprintf("%.1fmin", sec/60)
}

func init() {
	pipelineRunCmd.Flags().StringVar(&plRunConfigPath, "config", "ade-pipeline.yaml", "Chemin du fichier de configuration du pipeline")
	pipelineRunCmd.Flags().BoolVarP(&plRunVerbose, "verbose", "v", false, "Afficher les logs d\u00e9taill\u00e9s")
	pipelineRunCmd.Flags().BoolVar(&plRunLocal, "local", false, "Ex\u00e9cuter les commandes localement")
	pipelineRunCmd.Flags().BoolVar(&plRunDryRun, "dry-run", false, "Forcer le mode dry-run (simulation)")
}
