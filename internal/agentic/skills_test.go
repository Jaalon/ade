package agentic

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListAvailableSkills_ReturnsAtLeast6(t *testing.T) {
	skills, err := ListAvailableSkills()
	assert.NoError(t, err)
	assert.NotEmpty(t, skills)
	assert.GreaterOrEqual(t, len(skills), 6)
	for _, s := range skills {
		assert.Contains(t, s.Name, "skill-")
	}
}

func TestListProjectSkills_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	skills, err := ListProjectSkills(dir)
	assert.NoError(t, err)
	assert.Empty(t, skills)
}

func TestMissingSkills_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	missing, err := MissingSkills(dir)
	assert.NoError(t, err)
	assert.NotEmpty(t, missing)

	all, _ := ListAvailableSkills()
	assert.Equal(t, len(all), len(missing))
}

func TestMissingSkills_AllInstalled(t *testing.T) {
	dir := t.TempDir()

	all, _ := ListAvailableSkills()
	for _, s := range all {
		fullPath := filepath.Join(dir, s.TargetPath)
		osMkdirAll(filepath.Dir(fullPath), 0755)
		os.WriteFile(fullPath, []byte("test"), 0644)
	}

	missing, err := MissingSkills(dir)
	assert.NoError(t, err)
	assert.Empty(t, missing)
}

func TestMissingSkills_PartiallyInstalled(t *testing.T) {
	dir := t.TempDir()

	all, _ := ListAvailableSkills()
	if len(all) == 0 {
		t.Skip("no skills available")
	}

	first := all[0]
	fullPath := filepath.Join(dir, first.TargetPath)
	osMkdirAll(filepath.Dir(fullPath), 0755)
	os.WriteFile(fullPath, []byte("test"), 0644)

	missing, err := MissingSkills(dir)
	assert.NoError(t, err)
	assert.Equal(t, len(all)-1, len(missing))
}

func TestEnsureSkills_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	report, err := EnsureSkills(context.Background(), dir, false)
	assert.NoError(t, err)
	assert.Equal(t, report.Total, len(report.Installed))
	assert.Empty(t, report.AlreadyExist)
	assert.Empty(t, report.Errors)

	for _, s := range report.Installed {
		fullPath := filepath.Join(dir, skillTargetPath(s))
		_, err := os.Stat(fullPath)
		assert.NoError(t, err, "skill file should exist: %s", fullPath)
	}
}

func TestEnsureSkills_AlreadyExistNoForce(t *testing.T) {
	dir := t.TempDir()

	all, _ := ListAvailableSkills()
	if len(all) == 0 {
		t.Skip("no skills available")
	}

	first := all[0]
	fullPath := filepath.Join(dir, first.TargetPath)
	osMkdirAll(filepath.Dir(fullPath), 0755)
	os.WriteFile(fullPath, []byte("existing"), 0644)

	report, err := EnsureSkills(context.Background(), dir, false)
	assert.NoError(t, err)
	assert.Contains(t, report.AlreadyExist, first.Name)
}

func TestEnsureSkills_AlreadyExistWithForce(t *testing.T) {
	dir := t.TempDir()

	all, _ := ListAvailableSkills()
	if len(all) == 0 {
		t.Skip("no skills available")
	}

	first := all[0]
	fullPath := filepath.Join(dir, first.TargetPath)
	osMkdirAll(filepath.Dir(fullPath), 0755)
	os.WriteFile(fullPath, []byte("existing"), 0644)

	report, err := EnsureSkills(context.Background(), dir, true)
	assert.NoError(t, err)
	assert.Contains(t, report.Installed, first.Name)
}

func TestEnsureSkills_ContextCancel(t *testing.T) {
	dir := t.TempDir()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	report, err := EnsureSkills(ctx, dir, false)
	assert.Error(t, err)
	assert.NotNil(t, report)
}

func TestEnsureSkills_WriteError(t *testing.T) {
	savedWrite := osWriteFile
	defer func() { osWriteFile = savedWrite }()
	osWriteFile = func(name string, data []byte, perm os.FileMode) error {
		return fmt.Errorf("simulated write error")
	}

	dir := t.TempDir()
	report, err := EnsureSkills(context.Background(), dir, false)
	assert.NoError(t, err)
	assert.NotEmpty(t, report.Errors)
	assert.Empty(t, report.Installed)
}

func skillTargetPath(name string) string {
	all, _ := ListAvailableSkills()
	for _, s := range all {
		if s.Name == name {
			return s.TargetPath
		}
	}
	return ""
}
