package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func executeInit(args ...string) (string, error) {
	fullArgs := append([]string{"init"}, args...)
	return executeCommand(fullArgs...)
}

func TestInitHelpContainsSpecsAndCi(t *testing.T) {
	output, err := executeInit("--help")
	assert.NoError(t, err)
	assert.Contains(t, output, "specs")
	assert.Contains(t, output, "ci")
}

func TestInitSpecs(t *testing.T) {
	output, err := executeInit("specs")
	assert.NoError(t, err)
	assert.Contains(t, output, "spécifications")
}

func TestInitCi(t *testing.T) {
	output, err := executeInit("ci")
	assert.NoError(t, err)
	assert.Contains(t, output, "intégration continue")
}

func TestInitUnknownCommand(t *testing.T) {
	output, err := executeInit("inconnu")
	assert.NoError(t, err)
	assert.Contains(t, output, "specs")
	assert.Contains(t, output, "ci")
}
