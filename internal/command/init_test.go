package command

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"automated_dev_environment/internal/agentic"
	"automated_dev_environment/internal/generator"
	"automated_dev_environment/internal/templates"
)

func buildSpecsCmd() *cobra.Command {
	var force bool
	var output, name, lang, module string

	cmd := &cobra.Command{
		Use:   "specs",
		Short: "Génère les fichiers de configuration du projet",
		RunE: func(c *cobra.Command, args []string) error {
			absDir, err := filepath.Abs(output)
			if err != nil {
				return err
			}
			info, err := os.Stat(absDir)
			if err == nil && !info.IsDir() {
				return nil
			}
			if name == "" {
				name = filepath.Base(absDir)
			}
			if module == "" {
				module = name
			}
			tmplData := templates.TemplateData{
				ProjectName: name,
				GoVersion:   "1.26",
				ModulePath:  module,
				Lang:        lang,
			}
			opts := generator.Options{
				OutputDir:    absDir,
				Force:        force,
				TemplateData: tmplData,
				Prompter:     &generator.StdPrompter{},
			}
			report, genErr := generator.Generate(c.Context(), opts)
			if genErr != nil {
				return genErr
			}
			out := c.OutOrStdout()
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
			fmt.Fprintf(out, "\n  Fichiers créés : %d\n", report.Success)
			fmt.Fprintf(out, "  Fichiers ignorés : %d\n", report.Skipped)
			if report.Errors > 0 {
				fmt.Fprintf(out, "  Erreurs : %d\n", report.Errors)
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "")
	cmd.Flags().StringVarP(&output, "output", "o", ".", "")
	cmd.Flags().StringVar(&name, "name", "", "")
	cmd.Flags().StringVar(&lang, "lang", "fr", "")
	cmd.Flags().StringVar(&module, "module", "", "")
	return cmd
}

func resetInitFlags() {
	initForce = false
	initOutput = ""
	initConfig = ""
	initSkipTools = false
	initSkipSkills = false
	initSkipMCP = false
	initHaltOnError = false
}

func buildInitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:  "init",
		RunE: runAgenticSetup,
	}
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "")
	initCmd.Flags().StringVarP(&initOutput, "output", "o", ".", "")
	initCmd.Flags().StringVar(&initConfig, "config", "", "")
	initCmd.Flags().BoolVar(&initSkipTools, "skip-tools", false, "")
	initCmd.Flags().BoolVar(&initSkipSkills, "skip-skills", false, "")
	initCmd.Flags().BoolVar(&initSkipMCP, "skip-mcp", false, "")
	initCmd.Flags().BoolVar(&initHaltOnError, "halt-on-error", false, "")
	initCmd.AddCommand(buildSpecsCmd())
	ciCmd := &cobra.Command{
		Use: "ci",
		RunE: func(c *cobra.Command, args []string) error {
			fmt.Fprintln(c.OutOrStdout(), "Initialisation de l'intégration continue... (à implémenter)")
			return nil
		},
	}
	initCmd.AddCommand(ciCmd)
	return initCmd
}

func execInit(args ...string) (string, error) {
	resetInitFlags()
	rootCmd := &cobra.Command{Use: "ade", SilenceUsage: true, SilenceErrors: true}
	rootCmd.AddCommand(buildInitCmd())
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestInitHelpContainsSpecsAndCi(t *testing.T) {
	output, err := execInit("init", "--help")
	assert.NoError(t, err)
	assert.Contains(t, output, "specs")
	assert.Contains(t, output, "ci")
}

func TestInitSpecsHelpContainsFlags(t *testing.T) {
	output, err := execInit("init", "specs", "--help")
	assert.NoError(t, err)
	assert.Contains(t, output, "--force")
	assert.Contains(t, output, "--output")
}

func TestInitSpecs_GeneratesFiles(t *testing.T) {
	dir := t.TempDir()
	output, err := execInit("init", "specs", "--output", dir, "--force")
	assert.NoError(t, err)
	assert.Contains(t, output, "Fichiers créés")
	assert.FileExists(t, filepath.Join(dir, ".gitignore"))
	assert.FileExists(t, filepath.Join(dir, ".opencode", "config.json"))
	assert.FileExists(t, filepath.Join(dir, ".opencode", "workflow.yaml"))
}

func TestInitSpecs_Integration(t *testing.T) {
	dir := t.TempDir()
	_, err := execInit("init", "specs", "--output", dir, "--force")
	assert.NoError(t, err)

	gitignoreData, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	assert.NoError(t, err)
	assert.Contains(t, string(gitignoreData), "*.exe")

	configData, err := os.ReadFile(filepath.Join(dir, ".opencode", "config.json"))
	assert.NoError(t, err)
	assert.Contains(t, string(configData), "mcp_servers")

	workflowData, err := os.ReadFile(filepath.Join(dir, ".opencode", "workflow.yaml"))
	assert.NoError(t, err)
	assert.Contains(t, string(workflowData), "specification")

	skillDirs := []string{
		"specification-en", "specification-fr",
		"story-en", "story-fr",
		"tasks-fr", "tasks-en",
		"feedback-fr",
	}
	for _, s := range skillDirs {
		skillPath := filepath.Join(dir, ".opencode", "skills", s, "SKILL.md")
		assert.FileExists(t, skillPath, "missing skill: %s", s)
	}
}

func TestInitSpecs_ForceFlag(t *testing.T) {
	dir := t.TempDir()
	gitignorePath := filepath.Join(dir, ".gitignore")
	err := os.WriteFile(gitignorePath, []byte("old"), 0644)
	assert.NoError(t, err)

	output, err := execInit("init", "specs", "--output", dir, "--force")
	assert.NoError(t, err)
	assert.Contains(t, output, ".gitignore")

	data, err := os.ReadFile(gitignorePath)
	assert.NoError(t, err)
	assert.NotEqual(t, "old", string(data))
	assert.Contains(t, string(data), "*.exe")
}

func TestInitSpecs_SkillsAreCopied(t *testing.T) {
	dir := t.TempDir()
	_, err := execInit("init", "specs", "--output", dir, "--force")
	assert.NoError(t, err)

	skillPath := filepath.Join(dir, ".opencode", "skills", "specification-en", "SKILL.md")
	data, err := os.ReadFile(skillPath)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "Specification")
}

func TestInitSpecs_ShowsCreatedFiles(t *testing.T) {
	dir := t.TempDir()
	output, err := execInit("init", "specs", "--output", dir, "--force")
	assert.NoError(t, err)
	assert.Contains(t, output, ".gitignore")
	assert.Contains(t, output, ".opencode/config.json")
}

func TestInitSpecs_WithNameAndLang(t *testing.T) {
	dir := t.TempDir()
	output, err := execInit("init", "specs",
		"--output", dir,
		"--force",
		"--name", "mon-app",
		"--lang", "en",
		"--module", "github.com/user/mon-app")
	assert.NoError(t, err)
	assert.Contains(t, output, "Fichiers créés")
}

func TestInitCi(t *testing.T) {
	output, err := execInit("init", "ci")
	assert.NoError(t, err)
	assert.Contains(t, output, "intégration continue")
}

func TestInitUnknownCommand(t *testing.T) {
	savedDetect := detectToolsFn
	defer func() { detectToolsFn = savedDetect }()
	savedSkills := ensureSkillsFn
	defer func() { ensureSkillsFn = savedSkills }()
	savedMCP := configureMCPFn
	defer func() { configureMCPFn = savedMCP }()

	detectToolsFn = func(_ context.Context, _ string) (*agentic.DetectionResult, error) {
		return &agentic.DetectionResult{}, nil
	}
	ensureSkillsFn = func(_ context.Context, _ string, _ bool) (*agentic.SkillsReport, error) {
		return &agentic.SkillsReport{}, nil
	}
	configureMCPFn = func(_ context.Context, _ agentic.MCPOptions) (*agentic.MCPReport, error) {
		return &agentic.MCPReport{}, nil
	}

	output, err := execInit("init", "inconnu")
	assert.NoError(t, err)
	assert.Contains(t, output, "Configuration agentic terminée")
}

// --- Agentic setup tests ---

func TestInit_RunsAgenticSetup(t *testing.T) {
	savedDetect := detectToolsFn
	defer func() { detectToolsFn = savedDetect }()
	savedSkills := ensureSkillsFn
	defer func() { ensureSkillsFn = savedSkills }()
	savedMCP := configureMCPFn
	defer func() { configureMCPFn = savedMCP }()

	detectCalled := false
	detectToolsFn = func(_ context.Context, _ string) (*agentic.DetectionResult, error) {
		detectCalled = true
		return &agentic.DetectionResult{
			Tools: []agentic.ToolInfo{
				{Name: "OpenCode", Path: "/usr/bin/opencode", Found: true},
				{Name: "Cursor", Found: false},
			},
		}, nil
	}

	skillsCalled := false
	ensureSkillsFn = func(_ context.Context, _ string, _ bool) (*agentic.SkillsReport, error) {
		skillsCalled = true
		return &agentic.SkillsReport{
			Installed: []string{"specification-en", "specification-fr", "story-en"},
			Total:     3,
		}, nil
	}

	mcpCalled := false
	configureMCPFn = func(_ context.Context, _ agentic.MCPOptions) (*agentic.MCPReport, error) {
		mcpCalled = true
		return &agentic.MCPReport{Added: 2}, nil
	}

	output, err := execInit("init")
	assert.NoError(t, err)
	assert.True(t, detectCalled)
	assert.True(t, skillsCalled)
	assert.True(t, mcpCalled)
	assert.Contains(t, output, "Configuration agentic terminée")
	assert.Contains(t, output, "Outils détectés")
	assert.Contains(t, output, "OpenCode")
	assert.Contains(t, output, "Skills")
	assert.Contains(t, output, "3 skills installés")
	assert.Contains(t, output, "Serveurs MCP")
	assert.Contains(t, output, "2 serveurs configurés")
}

func TestInit_SkipTools(t *testing.T) {
	savedDetect := detectToolsFn
	defer func() { detectToolsFn = savedDetect }()
	savedSkills := ensureSkillsFn
	defer func() { ensureSkillsFn = savedSkills }()
	savedMCP := configureMCPFn
	defer func() { configureMCPFn = savedMCP }()

	detectCalled := false
	detectToolsFn = func(_ context.Context, _ string) (*agentic.DetectionResult, error) {
		detectCalled = true
		return &agentic.DetectionResult{}, nil
	}
	ensureSkillsFn = func(_ context.Context, _ string, _ bool) (*agentic.SkillsReport, error) {
		return &agentic.SkillsReport{}, nil
	}
	configureMCPFn = func(_ context.Context, _ agentic.MCPOptions) (*agentic.MCPReport, error) {
		return &agentic.MCPReport{}, nil
	}

	output, err := execInit("init", "--skip-tools")
	assert.NoError(t, err)
	assert.False(t, detectCalled, "detection should be skipped")
	assert.NotContains(t, output, "Outils détectés")
	assert.Contains(t, output, "Skills")
	assert.Contains(t, output, "Serveurs MCP")
}

func TestInit_SkipSkills(t *testing.T) {
	savedDetect := detectToolsFn
	defer func() { detectToolsFn = savedDetect }()
	savedSkills := ensureSkillsFn
	defer func() { ensureSkillsFn = savedSkills }()
	savedMCP := configureMCPFn
	defer func() { configureMCPFn = savedMCP }()

	detectToolsFn = func(_ context.Context, _ string) (*agentic.DetectionResult, error) {
		return &agentic.DetectionResult{}, nil
	}
	skillsCalled := false
	ensureSkillsFn = func(_ context.Context, _ string, _ bool) (*agentic.SkillsReport, error) {
		skillsCalled = true
		return &agentic.SkillsReport{}, nil
	}
	configureMCPFn = func(_ context.Context, _ agentic.MCPOptions) (*agentic.MCPReport, error) {
		return &agentic.MCPReport{}, nil
	}

	output, err := execInit("init", "--skip-skills")
	assert.NoError(t, err)
	assert.False(t, skillsCalled, "skills should be skipped")
	assert.Contains(t, output, "Outils détectés")
	assert.NotContains(t, output, "Skills")
	assert.Contains(t, output, "Serveurs MCP")
}

func TestInit_SkipMCP(t *testing.T) {
	savedDetect := detectToolsFn
	defer func() { detectToolsFn = savedDetect }()
	savedSkills := ensureSkillsFn
	defer func() { ensureSkillsFn = savedSkills }()
	savedMCP := configureMCPFn
	defer func() { configureMCPFn = savedMCP }()

	detectToolsFn = func(_ context.Context, _ string) (*agentic.DetectionResult, error) {
		return &agentic.DetectionResult{}, nil
	}
	ensureSkillsFn = func(_ context.Context, _ string, _ bool) (*agentic.SkillsReport, error) {
		return &agentic.SkillsReport{}, nil
	}
	mcpCalled := false
	configureMCPFn = func(_ context.Context, _ agentic.MCPOptions) (*agentic.MCPReport, error) {
		mcpCalled = true
		return &agentic.MCPReport{}, nil
	}

	output, err := execInit("init", "--skip-mcp")
	assert.NoError(t, err)
	assert.False(t, mcpCalled, "mcp should be skipped")
	assert.Contains(t, output, "Outils détectés")
	assert.Contains(t, output, "Skills")
	assert.NotContains(t, output, "Serveurs MCP")
}

func TestInit_SkipAll(t *testing.T) {
	savedDetect := detectToolsFn
	defer func() { detectToolsFn = savedDetect }()
	savedSkills := ensureSkillsFn
	defer func() { ensureSkillsFn = savedSkills }()
	savedMCP := configureMCPFn
	defer func() { configureMCPFn = savedMCP }()

	detectCalled := false
	detectToolsFn = func(_ context.Context, _ string) (*agentic.DetectionResult, error) {
		detectCalled = true
		return &agentic.DetectionResult{}, nil
	}
	skillsCalled := false
	ensureSkillsFn = func(_ context.Context, _ string, _ bool) (*agentic.SkillsReport, error) {
		skillsCalled = true
		return &agentic.SkillsReport{}, nil
	}
	mcpCalled := false
	configureMCPFn = func(_ context.Context, _ agentic.MCPOptions) (*agentic.MCPReport, error) {
		mcpCalled = true
		return &agentic.MCPReport{}, nil
	}

	output, err := execInit("init", "--skip-tools", "--skip-skills", "--skip-mcp")
	assert.NoError(t, err)
	assert.False(t, detectCalled)
	assert.False(t, skillsCalled)
	assert.False(t, mcpCalled)
	assert.Equal(t, "✓ Configuration agentic terminée\n", output)
}

func TestInit_HaltOnError(t *testing.T) {
	savedDetect := detectToolsFn
	defer func() { detectToolsFn = savedDetect }()
	savedSkills := ensureSkillsFn
	defer func() { ensureSkillsFn = savedSkills }()
	savedMCP := configureMCPFn
	defer func() { configureMCPFn = savedMCP }()

	detectToolsFn = func(_ context.Context, _ string) (*agentic.DetectionResult, error) {
		return nil, fmt.Errorf("outils non disponibles")
	}
	skillsCalled := false
	ensureSkillsFn = func(_ context.Context, _ string, _ bool) (*agentic.SkillsReport, error) {
		skillsCalled = true
		return &agentic.SkillsReport{}, nil
	}
	mcpCalled := false
	configureMCPFn = func(_ context.Context, _ agentic.MCPOptions) (*agentic.MCPReport, error) {
		mcpCalled = true
		return &agentic.MCPReport{}, nil
	}

	_, err := execInit("init", "--halt-on-error")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "outils non disponibles")
	assert.False(t, skillsCalled, "skills should NOT be called after error with halt-on-error")
	assert.False(t, mcpCalled, "mcp should NOT be called after error with halt-on-error")
}

func TestInit_ContinueOnError(t *testing.T) {
	savedDetect := detectToolsFn
	defer func() { detectToolsFn = savedDetect }()
	savedSkills := ensureSkillsFn
	defer func() { ensureSkillsFn = savedSkills }()
	savedMCP := configureMCPFn
	defer func() { configureMCPFn = savedMCP }()

	detectToolsFn = func(_ context.Context, _ string) (*agentic.DetectionResult, error) {
		return nil, fmt.Errorf("outils non disponibles")
	}
	skillsCalled := false
	ensureSkillsFn = func(_ context.Context, _ string, _ bool) (*agentic.SkillsReport, error) {
		skillsCalled = true
		return &agentic.SkillsReport{
			Installed: []string{"specification-en"},
			Total:     1,
		}, nil
	}
	mcpCalled := false
	configureMCPFn = func(_ context.Context, _ agentic.MCPOptions) (*agentic.MCPReport, error) {
		mcpCalled = true
		return &agentic.MCPReport{Added: 1}, nil
	}

	output, err := execInit("init")
	assert.NoError(t, err)
	assert.True(t, skillsCalled, "skills should still be called after detection error")
	assert.True(t, mcpCalled, "mcp should still be called after detection error")
	assert.Contains(t, output, "Erreur")
	assert.Contains(t, output, "1 skills installés")
	assert.Contains(t, output, "1 serveurs configurés")
}

func TestInit_AllStepsFailed(t *testing.T) {
	savedDetect := detectToolsFn
	defer func() { detectToolsFn = savedDetect }()
	savedSkills := ensureSkillsFn
	defer func() { ensureSkillsFn = savedSkills }()
	savedMCP := configureMCPFn
	defer func() { configureMCPFn = savedMCP }()

	detectToolsFn = func(_ context.Context, _ string) (*agentic.DetectionResult, error) {
		return nil, fmt.Errorf("detection error")
	}
	ensureSkillsFn = func(_ context.Context, _ string, _ bool) (*agentic.SkillsReport, error) {
		return nil, fmt.Errorf("skills error")
	}
	configureMCPFn = func(_ context.Context, _ agentic.MCPOptions) (*agentic.MCPReport, error) {
		return nil, fmt.Errorf("mcp error")
	}

	_, err := execInit("init")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "toutes les étapes ont échoué")
}

func TestInit_HelpContainsNewFlags(t *testing.T) {
	output, err := execInit("init", "--help")
	assert.NoError(t, err)
	assert.Contains(t, output, "--force")
	assert.Contains(t, output, "--output")
	assert.Contains(t, output, "--config")
	assert.Contains(t, output, "--skip-tools")
	assert.Contains(t, output, "--skip-skills")
	assert.Contains(t, output, "--skip-mcp")
	assert.Contains(t, output, "--halt-on-error")
}

func TestInit_SpecsStillWorks(t *testing.T) {
	dir := t.TempDir()
	output, err := execInit("init", "specs", "--output", dir, "--force")
	assert.NoError(t, err)
	assert.Contains(t, output, "Fichiers créés")
	assert.FileExists(t, filepath.Join(dir, ".gitignore"))
}

func TestInit_CiStillWorks(t *testing.T) {
	output, err := execInit("init", "ci")
	assert.NoError(t, err)
	assert.Contains(t, output, "intégration continue")
}
