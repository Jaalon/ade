package command

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestRootCmdExists(t *testing.T) {
	assert.NotNil(t, rootCmd)
	assert.Equal(t, "ade", rootCmd.Use)
}

func TestHelpContainsDescription(t *testing.T) {
	output, err := executeCommand("--help")
	assert.NoError(t, err)
	assert.Contains(t, output, "développement agentic")
}

func TestVersionFlag(t *testing.T) {
	savedArgs := os.Args
	defer func() { os.Args = savedArgs }()
	versionOutput := captureStdout(t, func() {
		os.Args = []string{"ade", "--version"}
		Execute()
	})
	assert.Equal(t, "0.1.0\n", versionOutput)
}

func TestVersionCommand(t *testing.T) {
	output, err := executeCommand("version")
	assert.NoError(t, err)
	assert.Contains(t, output, "0.1.0")
}

func TestUnknownCommand(t *testing.T) {
	_, err := executeCommand("inconnu")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
}

func TestNoArgsShowsHelp(t *testing.T) {
	output, err := executeCommand()
	assert.NoError(t, err)
	assert.Contains(t, output, "Usage")
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	stdout := os.Stdout
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = stdout
	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	assert.NoError(t, err)
	return buf.String()
}

func TestExecuteVersionDoesNotBlock(t *testing.T) {
	savedArgs := os.Args
	defer func() { os.Args = savedArgs }()
	os.Args = []string{"ade", "--version"}

	done := make(chan struct{})
	go func() {
		Execute()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Execute() blocked for more than 5 seconds")
	}
}
