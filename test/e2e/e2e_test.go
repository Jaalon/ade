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

func TestAdeHelpContainsDescription(t *testing.T) {
	modRoot := moduleRoot()

	cmd := exec.Command("go", "build", "-ldflags=-s -w", "-o", filepath.Join(modRoot, "ade_test.exe"), "./cmd/ade")
	cmd.Dir = modRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	defer os.Remove(filepath.Join(modRoot, "ade_test.exe"))

	cmd = exec.Command(filepath.Join(modRoot, "ade_test.exe"), "--help")
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
