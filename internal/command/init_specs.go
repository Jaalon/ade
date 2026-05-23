package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"automated_dev_environment/internal/generator"
	"automated_dev_environment/internal/templates"
)

var (
	initSpecsForce  bool
	initSpecsOutput string
	initSpecsName   string
	initSpecsLang   string
	initSpecsModule string
)

var initSpecsCmd = &cobra.Command{
	Use:   "specs",
	Short: "Génère les fichiers de configuration du projet",
	Long: `Génère les fichiers de configuration locaux (.gitignore, skills,
serveurs MCP, workflow de développement) depuis des templates intégrés.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		outputDir := initSpecsOutput
		absDir, err := filepath.Abs(outputDir)
		if err != nil {
			return fmt.Errorf("chemin de sortie invalide: %w", err)
		}

		info, err := os.Stat(absDir)
		if err == nil && !info.IsDir() {
			return fmt.Errorf("le chemin %s existe et n'est pas un répertoire", absDir)
		}

		projectName := initSpecsName
		if projectName == "" {
			projectName = filepath.Base(absDir)
		}

		modulePath := initSpecsModule
		if modulePath == "" {
			modulePath = projectName
		}

		tmplData := templates.TemplateData{
			ProjectName: projectName,
			GoVersion:   "1.26",
			ModulePath:  modulePath,
			Lang:        initSpecsLang,
		}

		opts := generator.Options{
			OutputDir:    absDir,
			Force:        initSpecsForce,
			TemplateData: tmplData,
			Prompter:     &generator.StdPrompter{},
		}

		report, err := generator.Generate(cmd.Context(), opts)
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		fmt.Fprintln(out)

		if len(report.Files) == 0 {
			fmt.Fprintln(out, "Aucun template à générer.")
			return nil
		}

		for _, f := range report.Files {
			switch f.Status {
			case generator.StatusCreated:
				fmt.Fprintf(out, "  ✓ %s\n", f.TargetPath)
			case generator.StatusOverwritten:
				fmt.Fprintf(out, "  ✓ %s (écrasé)\n", f.TargetPath)
			case generator.StatusSkipped:
				fmt.Fprintf(out, "  ∼ %s (ignoré)\n", f.TargetPath)
			case generator.StatusError:
				fmt.Fprintf(out, "  ✗ %s: %v\n", f.TargetPath, f.Err)
			}
		}

		fmt.Fprintln(out)
		fmt.Fprintf(out, "  Fichiers créés : %d\n", report.Success)
		fmt.Fprintf(out, "  Fichiers ignorés : %d\n", report.Skipped)
		if report.Errors > 0 {
			fmt.Fprintf(out, "  Erreurs : %d\n", report.Errors)
		}

		if report.Errors > 0 {
			return fmt.Errorf("%d erreur(s) lors de la génération", report.Errors)
		}
		return nil
	},
}

func init() {
	initSpecsCmd.Flags().BoolVarP(&initSpecsForce, "force", "f", false,
		"Écraser les fichiers existants sans confirmation")
	initSpecsCmd.Flags().StringVarP(&initSpecsOutput, "output", "o", ".",
		"Répertoire de destination")
	initSpecsCmd.Flags().StringVar(&initSpecsName, "name", "",
		"Nom du projet (défaut: nom du répertoire)")
	initSpecsCmd.Flags().StringVar(&initSpecsLang, "lang", "fr",
		"Langue pour les skills (fr ou en)")
	initSpecsCmd.Flags().StringVar(&initSpecsModule, "module", "",
		"Module Go (défaut: nom du projet)")
}
