package command

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"automated_dev_environment/internal/docker"
	"automated_dev_environment/internal/generator"
	"automated_dev_environment/internal/templates"
)

var (
	dockerCheckFn       = docker.Check
	dockerNewClientFn   = docker.NewClient
	execLookPathFn      = exec.LookPath
	execCommandFn       = exec.Command
	executeCmdFn        = func(cmd *exec.Cmd) error { return cmd.Run() }
	checkDockerPluginFn = func() bool {
		cmd := execCommandFn("docker", "compose", "version")
		return executeCmdFn(cmd) == nil
	}
)

var (
	ciOutput  string
	ciForce   bool
	ciName    string
	ciPort    string
	ciNetwork string
)

var initCiCmd = &cobra.Command{
	Use:   "ci",
	Short: "D\u00e9ploie l'environnement de pr\u00e9production",
	Long:  `D\u00e9tecte Docker/Podman, g\u00e9n\u00e8re docker-compose.yml et .env, d\u00e9ploie les conteneurs.`,
	RunE:  runInitCi,
}

type ciOptions struct {
	OutputDir      string
	Force          bool
	ProjectName    string
	ConfigPort     string
	ComposeNetwork string
}

func (o ciOptions) toTemplateData() templates.TemplateData {
	return templates.TemplateData{
		ProjectName: o.ProjectName,
		Compose: templates.ComposeConfig{
			ConfigPort: o.ConfigPort,
			Network:    o.ComposeNetwork,
		},
	}
}

func runInitCi(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	ctx := cmd.Context()

	outputDir := ciOutput
	if outputDir == "" {
		outputDir = "."
	}

	binary, err := detectAndPing(ctx, out)
	if err != nil {
		return err
	}

	projectName := ciName
	if projectName == "" {
		absDir, absErr := filepath.Abs(outputDir)
		if absErr == nil {
			projectName = filepath.Base(absDir)
		}
	}
	if projectName == "" || projectName == "." {
		projectName = "preprod"
	}

	opts := ciOptions{
		OutputDir:      outputDir,
		Force:          ciForce,
		ProjectName:    projectName,
		ConfigPort:     ciPort,
		ComposeNetwork: ciNetwork,
	}

	fmt.Fprintln(out, "")
	fmt.Fprintf(out, "  G\u00e9n\u00e9ration des fichiers de d\u00e9ploiement...\n")
	report, genErr := generateComposeFiles(ctx, out, opts)
	if genErr != nil {
		return fmt.Errorf("g\u00e9n\u00e9ration des fichiers: %w", genErr)
	}
	for _, f := range report.Files {
		status := "?"
		switch f.Status {
		case generator.StatusCreated:
			status = "cr\u00e9\u00e9"
		case generator.StatusOverwritten:
			status = "mis \u00e0 jour"
		case generator.StatusSkipped:
			status = "ignor\u00e9"
		}
		fmt.Fprintf(out, "    %s (%s)\n", f.TargetPath, status)
	}

	fmt.Fprintln(out, "")
	fmt.Fprintf(out, "  D\u00e9ploiement avec %s...\n", binary)
	deployCompose(ctx, out, binary, outputDir)

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "  Statut des conteneurs:")
	showComposeStatus(ctx, out, binary, outputDir)

	return nil
}

func detectAndPing(ctx context.Context, out io.Writer) (string, error) {
	binary, err := dockerCheckFn()
	if err != nil {
		return "", fmt.Errorf("Docker ou Podman requis. Veuillez installer Docker Desktop (https://www.docker.com/products/docker-desktop/) ou Podman (https://podman.io/).")
	}
	fmt.Fprintf(out, "  \u2713 %s trouv\u00e9\n", binary)

	client, cliErr := dockerNewClientFn()
	if cliErr != nil {
		return binary, fmt.Errorf("connexion \u00e0 %s impossible: %v. V\u00e9rifiez que le service est install\u00e9.", binary, cliErr)
	}
	defer client.Close()

	if pingErr := client.Ping(ctx); pingErr != nil {
		return binary, fmt.Errorf("le d\u00e9mon %s n'est pas accessible: %v. D\u00e9marrez Docker Desktop ou v\u00e9rifiez le service Podman.", binary, pingErr)
	}
	fmt.Fprintf(out, "  \u2713 D\u00e9mon %s actif\n", binary)

	return binary, nil
}

func generateComposeFiles(_ context.Context, _ io.Writer, opts ciOptions) (*generator.Report, error) {
	absDir, err := filepath.Abs(opts.OutputDir)
	if err != nil {
		return nil, err
	}

	tmplData := opts.toTemplateData()

	genOpts := generator.Options{
		OutputDir:       absDir,
		Force:           opts.Force,
		TemplateData:    tmplData,
		TemplatesFilter: []string{"docker-compose", "env"},
		Prompter:        &generator.StdPrompter{},
	}

	return generator.Generate(context.Background(), genOpts)
}

func deployCompose(_ context.Context, out io.Writer, binary string, workDir string) {
	composeCmd, composeArgs := getComposeCommand(binary)
	if composeCmd == "" {
		fmt.Fprintf(out, "  \u26a0 Commande compose non trouv\u00e9e pour %s. Les fichiers sont g\u00e9n\u00e9r\u00e9s. D\u00e9ployez manuellement avec 'docker compose up -d'.\n", binary)
		return
	}

	args := append(append([]string(nil), composeArgs...), "up", "-d")
	cmd := execCommandFn(composeCmd, args...)
	cmd.Dir = workDir
	cmd.Stdout = out
	cmd.Stderr = out
	if err := executeCmdFn(cmd); err != nil {
		fmt.Fprintf(out, "  \u26a0 D\u00e9ploiement impossible. Les fichiers sont g\u00e9n\u00e9r\u00e9s. D\u00e9ployez manuellement avec '%s up -d' une fois l'image configur\u00e9e.\n", composeCmd)
	}
}

func showComposeStatus(_ context.Context, out io.Writer, binary string, workDir string) {
	composeCmd, composeArgs := getComposeCommand(binary)
	if composeCmd == "" {
		return
	}

	args := append(append([]string(nil), composeArgs...), "ps")
	cmd := execCommandFn(composeCmd, args...)
	cmd.Dir = workDir
	cmd.Stdout = out
	cmd.Stderr = out
	_ = executeCmdFn(cmd)
}

func getComposeCommand(binary string) (string, []string) {
	switch binary {
	case "docker":
		if _, err := execLookPathFn("docker"); err == nil {
			if checkDockerPluginFn() {
				return "docker", []string{"compose"}
			}
		}
		if _, err := execLookPathFn("docker-compose"); err == nil {
			return "docker-compose", nil
		}
	case "podman":
		if _, err := execLookPathFn("podman-compose"); err == nil {
			return "podman-compose", nil
		}
	}
	return "", nil
}

func init() {
	initCiCmd.Flags().StringVarP(&ciOutput, "output", "o", ".", "R\u00e9pertoire de sortie pour les fichiers g\u00e9n\u00e9r\u00e9s")
	initCiCmd.Flags().BoolVarP(&ciForce, "force", "f", false, "\u00c9craser les fichiers existants sans confirmation")
	initCiCmd.Flags().StringVar(&ciName, "name", "", "Nom du projet (d\u00e9faut: nom du r\u00e9pertoire)")
	initCiCmd.Flags().StringVar(&ciPort, "port", "8080", "Port du conteneur de configuration (web UI)")
	initCiCmd.Flags().StringVar(&ciNetwork, "network", "ade-network", "Nom du r\u00e9seau Docker")
}
