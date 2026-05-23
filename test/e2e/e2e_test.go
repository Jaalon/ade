//go:build e2e

package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func moduleRoot() string {
	out, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(strings.TrimSpace(string(out)))
}

func buildAde(t *testing.T, modRoot, binPath string) {
	t.Helper()
	cmd := exec.Command("go", "build", "-ldflags=-s -w", "-o", binPath, "./cmd/ade")
	cmd.Dir = modRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
}

func TestAdeHelpContainsDescription(t *testing.T) {
	modRoot := moduleRoot()
	binPath := filepath.Join(modRoot, "ade_test.exe")
	buildAde(t, modRoot, binPath)
	defer os.Remove(binPath)

	cmd := exec.Command(binPath, "--help")
	cmd.Dir = modRoot
	helpOut, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ade --help failed: %v\n%s", err, helpOut)
	}
	output := string(helpOut)

	if !strings.Contains(output, "Automated Dev Environment") {
		t.Fatalf("help output does not contain 'Automated Dev Environment':\n%s", output)
	}
}

func TestE2E_InitSpecs(t *testing.T) {
	modRoot := moduleRoot()
	binPath := filepath.Join(modRoot, "ade_test.exe")
	buildAde(t, modRoot, binPath)
	defer os.Remove(binPath)

	tmpDir := t.TempDir()

	cmd := exec.Command(binPath, "init", "specs", "--force")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ade init specs --force failed: %v\n%s", err, output)
	}

	outStr := string(output)
	if !strings.Contains(outStr, "Fichiers créés") {
		t.Fatalf("output does not contain 'Fichiers créés':\n%s", outStr)
	}

	expectedFiles := []string{
		".gitignore",
		".opencode/config.json",
		".opencode/workflow.yaml",
		".opencode/skills/specification-en/SKILL.md",
		".opencode/skills/specification-fr/SKILL.md",
		".opencode/skills/story-en/SKILL.md",
		".opencode/skills/story-fr/SKILL.md",
		".opencode/skills/tasks-fr/SKILL.md",
		".opencode/skills/tasks-en/SKILL.md",
		".opencode/skills/feedback-fr/SKILL.md",
	}
	for _, f := range expectedFiles {
		fullPath := filepath.Join(tmpDir, f)
		if _, statErr := os.Stat(fullPath); statErr != nil {
			t.Errorf("expected file %s does not exist: %v", f, statErr)
		}
	}
}
