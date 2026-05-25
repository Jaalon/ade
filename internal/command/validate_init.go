package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	valInitOutput string
	valInitForce  bool
)

var validateInitCmd = &cobra.Command{
	Use:   "init",
	Short: "G\u00e9n\u00e8re la configuration par d\u00e9faut de validation",
	Long:  `G\u00e9n\u00e8re un fichier ade-validate.yaml avec les modules disponibles.`,
	RunE:  runValidateInit,
}

func runValidateInit(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	outputDir := valInitOutput
	if outputDir == "" {
		outputDir = "."
	}

	targetPath := filepath.Join(outputDir, "ade-validate.yaml")

	if _, statErr := os.Stat(targetPath); statErr == nil && !valInitForce {
		return fmt.Errorf("le fichier %s existe d\u00e9j\u00e0. Utilisez --force pour \u00e9craser", targetPath)
	}

	dir := filepath.Dir(targetPath)
	if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
		return fmt.Errorf("impossible de cr\u00e9er le r\u00e9pertoire %s: %w", dir, mkErr)
	}

	if err := os.WriteFile(targetPath, []byte(validateTemplateYAML), 0644); err != nil {
		return fmt.Errorf("impossible d'\u00e9crire %s: %w", targetPath, err)
	}

	fmt.Fprintf(out, "\u2713 Fichier de configuration cr\u00e9\u00e9 : %s\n", targetPath)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "  Prochaines \u00e9tapes :")
	fmt.Fprintln(out, "  1. Personnalisez les modules dans ade-validate.yaml")
	fmt.Fprintln(out, "  2. Ex\u00e9cutez 'ade validate run' pour valider l'environnement")

	return nil
}

const validateTemplateYAML = `# Configuration de validation de l'environnement
# G\u00e9n\u00e9r\u00e9 par 'ade validate init'
output_dir: .ade/validation
formats:
  - json
  - junit

modules:
  - name: golang
    enabled: true
    checks:
      - version
      - build
      - test
      - vet
`

func init() {
	validateInitCmd.Flags().StringVarP(&valInitOutput, "output", "o", ".", "R\u00e9pertoire de sortie")
	validateInitCmd.Flags().BoolVarP(&valInitForce, "force", "f", false, "\u00c9craser sans confirmation")
}
