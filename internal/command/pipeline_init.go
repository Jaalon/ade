package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	plInitOutput   string
	plInitForce    bool
	plInitTemplate string
)

var pipelineInitCmd = &cobra.Command{
	Use:   "init",
	Short: "G\u00e9n\u00e8re la configuration par d\u00e9faut du pipeline",
	Long: `G\u00e9n\u00e8re un fichier ade-pipeline.yaml avec la configuration par d\u00e9faut.

Templates disponibles :
  generic   Configuration g\u00e9n\u00e9rique avec commentaires (d\u00e9faut)
  go        Configuration pr\u00eate pour un projet Go
  java      Configuration pr\u00eate pour un projet Java/Maven`,
	RunE: runPipelineInit,
}

func runPipelineInit(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	content, err := pipelineTemplateContent(plInitTemplate)
	if err != nil {
		return err
	}

	outputDir := plInitOutput
	if outputDir == "" {
		outputDir = "."
	}

	targetPath := filepath.Join(outputDir, "ade-pipeline.yaml")

	if _, statErr := os.Stat(targetPath); statErr == nil && !plInitForce {
		return fmt.Errorf("le fichier %s existe d\u00e9j\u00e0. Utilisez --force pour \u00e9craser", targetPath)
	}

	dir := filepath.Dir(targetPath)
	if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
		return fmt.Errorf("impossible de cr\u00e9er le r\u00e9pertoire %s: %w", dir, mkErr)
	}

	if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("impossible d'\u00e9crire %s: %w", targetPath, err)
	}

	fmt.Fprintf(out, "\u2713 Fichier de configuration cr\u00e9\u00e9 : %s\n", targetPath)
	if plInitTemplate != "generic" {
		fmt.Fprintf(out, "  Template : %s\n", plInitTemplate)
	}
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "  Prochaines \u00e9tapes :")
	fmt.Fprintln(out, "  1. Personnalisez les commandes dans ade-pipeline.yaml")
	fmt.Fprintln(out, "  2. Ex\u00e9cutez 'ade pipeline run' pour tester le pipeline")

	return nil
}

func pipelineTemplateContent(template string) (string, error) {
	switch strings.ToLower(template) {
	case "generic":
		return pipelineTemplateGeneric, nil
	case "go":
		return pipelineTemplateGo, nil
	case "java":
		return pipelineTemplateJava, nil
	default:
		avail := strings.Join(availableTemplates(), ", ")
		return "", fmt.Errorf("template inconnu : %q. Templates disponibles : %s", template, avail)
	}
}

func availableTemplates() []string {
	return []string{"generic", "go", "java"}
}

const pipelineTemplateGeneric = `# Configuration du pipeline CI
# G\u00e9n\u00e9r\u00e9 par 'ade pipeline init'
# Personnalisez les commandes selon votre langage et vos besoins.

stages:
  - type: build
    name: "Compilation du projet"
    enabled: true
    steps:
      - name: "Compiler"
        # Exemples : ["go", "build", "./..."]
        #           ["mvn", "clean", "compile"]
        #           ["npm", "run", "build"]
        command: []

  - type: unit-test
    name: "Tests unitaires"
    enabled: true
    steps:
      - name: "Tests unitaires"
        command: []

  - type: integration-test
    name: "Tests d'int\u00e9gration"
    enabled: true
    steps:
      - name: "Tests d'int\u00e9gration"
        command: []

  - type: test-deploy
    name: "D\u00e9ploiement test"
    enabled: false
    steps:
      - name: "D\u00e9ployer"
        command: ["docker", "compose", "-f", "docker-compose.test.yml", "up", "-d"]

  - type: e2e
    name: "Tests E2E"
    enabled: true
    steps:
      - name: "Tests E2E"
        command: []

  - type: preprod
    name: "Pr\u00e9production"
    enabled: false
    steps:
      - name: "D\u00e9ployer"
        command: ["./deploy-preprod.sh"]
`

const pipelineTemplateGo = `# Configuration du pipeline CI pour un projet Go

stages:
  - type: build
    name: "Compilation du projet"
    enabled: true
    steps:
      - name: "Build Go"
        command: ["go", "build", "./..."]

  - type: unit-test
    name: "Tests unitaires"
    enabled: true
    steps:
      - name: "Tests unitaires Go"
        command: ["go", "test", "./..."]

  - type: integration-test
    name: "Tests d'int\u00e9gration"
    enabled: true
    steps:
      - name: "Tests d'int\u00e9gration Go"
        command: ["go", "test", "-tags=integration", "./..."]

  - type: test-deploy
    name: "D\u00e9ploiement test"
    enabled: false
    steps:
      - name: "D\u00e9ployer"
        command: ["docker", "compose", "-f", "docker-compose.test.yml", "up", "-d"]

  - type: e2e
    name: "Tests E2E"
    enabled: true
    steps:
      - name: "Tests E2E Go"
        command: ["go", "test", "-tags=e2e", "./..."]

  - type: preprod
    name: "Pr\u00e9production"
    enabled: false
    steps:
      - name: "D\u00e9ployer"
        command: ["./deploy-preprod.sh"]
`

const pipelineTemplateJava = `# Configuration du pipeline CI pour un projet Java/Maven

stages:
  - type: build
    name: "Compilation du projet"
    enabled: true
    steps:
      - name: "Build Maven"
        command: ["mvn", "clean", "compile"]

  - type: unit-test
    name: "Tests unitaires"
    enabled: true
    steps:
      - name: "Tests unitaires Maven"
        command: ["mvn", "test"]

  - type: integration-test
    name: "Tests d'int\u00e9gration"
    enabled: true
    steps:
      - name: "Tests d'int\u00e9gration Maven"
        command: ["mvn", "verify"]

  - type: test-deploy
    name: "D\u00e9ploiement test"
    enabled: false
    steps:
      - name: "D\u00e9ployer"
        command: ["docker", "compose", "-f", "docker-compose.test.yml", "up", "-d"]

  - type: e2e
    name: "Tests E2E"
    enabled: false
    steps:
      - name: "Tests E2E"
        command: []

  - type: preprod
    name: "Pr\u00e9production"
    enabled: false
    steps:
      - name: "D\u00e9ployer"
        command: ["./deploy-preprod.sh"]
`

func init() {
	pipelineInitCmd.Flags().StringVarP(&plInitOutput, "output", "o", ".", "R\u00e9pertoire de sortie")
	pipelineInitCmd.Flags().BoolVarP(&plInitForce, "force", "f", false, "\u00c9craser le fichier existant")
	pipelineInitCmd.Flags().StringVarP(&plInitTemplate, "template", "t", "generic", "Template \u00e0 utiliser (generic, go, java)")
}
