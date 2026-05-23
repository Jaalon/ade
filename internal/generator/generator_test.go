package generator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"automated_dev_environment/internal/templates"
)

type mockPrompter struct {
	response bool
}

func (m *mockPrompter) Confirm(_ string) (bool, error) {
	return m.response, nil
}

func TestGenerate_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	opts := Options{
		OutputDir: dir,
		Force:     true,
	}
	report, err := Generate(context.Background(), opts)
	if err != nil {
		t.Fatalf("Generate() returned unexpected error: %v", err)
	}
	if report.Success == 0 {
		t.Fatal("expected at least 1 created file, got 0")
	}
	for _, f := range report.Files {
		if f.Status == StatusError {
			t.Errorf("file %s has error status: %v", f.TargetPath, f.Err)
		}
		fullPath := filepath.Join(dir, f.TargetPath)
		if _, statErr := os.Stat(fullPath); statErr != nil {
			t.Errorf("expected file %s to exist: %v", f.TargetPath, statErr)
		}
	}
}

func TestGenerate_SkipExistingWhenNotForced(t *testing.T) {
	dir := t.TempDir()

	gitignorePath := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}

	opts := Options{
		OutputDir: dir,
		Force:     false,
		Prompter:  &mockPrompter{response: false},
	}
	report, err := Generate(context.Background(), opts)
	if err != nil {
		t.Fatalf("Generate() returned unexpected error: %v", err)
	}

	found := false
	for _, f := range report.Files {
		if f.TargetPath == ".gitignore" {
			found = true
			if f.Status != StatusSkipped {
				t.Errorf("expected .gitignore to be skipped, got %s", f.Status)
			}
			break
		}
	}
	if !found {
		t.Error(".gitignore not found in report")
	}
	if report.Skipped == 0 {
		t.Error("expected at least 1 skipped file")
	}
}

func TestGenerate_ForceOverwrites(t *testing.T) {
	dir := t.TempDir()

	gitignorePath := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	opts := Options{
		OutputDir: dir,
		Force:     true,
	}
	report, err := Generate(context.Background(), opts)
	if err != nil {
		t.Fatalf("Generate() returned unexpected error: %v", err)
	}

	found := false
	for _, f := range report.Files {
		if f.TargetPath == ".gitignore" {
			found = true
			if f.Status != StatusCreated && f.Status != StatusOverwritten {
				t.Errorf("expected .gitignore to be created/overwritten, got %s", f.Status)
			}
			break
		}
	}
	if !found {
		t.Error(".gitignore not found in report")
	}

	data, _ := os.ReadFile(gitignorePath)
	if string(data) == "old" {
		t.Error(".gitignore was not overwritten")
	}
}

func TestGenerate_TemplatesFilter(t *testing.T) {
	dir := t.TempDir()
	opts := Options{
		OutputDir:       dir,
		Force:           true,
		TemplatesFilter: []string{"gitignore"},
	}
	report, err := Generate(context.Background(), opts)
	if err != nil {
		t.Fatalf("Generate() returned unexpected error: %v", err)
	}
	if len(report.Files) != 1 {
		t.Fatalf("expected exactly 1 file, got %d", len(report.Files))
	}
	if report.Files[0].TemplateName != "gitignore" {
		t.Errorf("expected file 'gitignore', got %s", report.Files[0].TemplateName)
	}
}

func TestGenerate_OutputDirIsFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "afile")
	if err := os.WriteFile(filePath, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	opts := Options{
		OutputDir: filePath,
		Force:     true,
	}
	_, err := Generate(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error when output is a file, got nil")
	}
}

func defaultTemplateData() templates.TemplateData {
	return templates.TemplateData{
		ProjectName: "test-project",
		GoVersion:   "1.26",
		ModulePath:  "github.com/user/test",
		Lang:        "fr",
	}
}

func TestGenerate_AllTemplatesCreateFiles(t *testing.T) {
	allTemplates := templates.ListTemplates()

	for _, tmpl := range allTemplates {
		t.Run(tmpl.Name, func(t *testing.T) {
			sub := t.TempDir()
			opts := Options{
				OutputDir:       sub,
				Force:           true,
				TemplateData:    defaultTemplateData(),
				TemplatesFilter: []string{tmpl.Name},
			}
			report, err := Generate(context.Background(), opts)
			if err != nil {
				t.Fatalf("Generate(%q) failed: %v", tmpl.Name, err)
			}
			if len(report.Files) != 1 {
				t.Fatalf("expected 1 file, got %d", len(report.Files))
			}
			if report.Files[0].Status == StatusError {
				t.Fatalf("file %s has error: %v", tmpl.TargetPath, report.Files[0].Err)
			}
			fullPath := filepath.Join(sub, tmpl.TargetPath)
			if _, statErr := os.Stat(fullPath); statErr != nil {
				t.Errorf("file %s does not exist: %v", fullPath, statErr)
			}
		})
	}
}

func TestGenerate_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	dir := t.TempDir()
	opts := Options{
		OutputDir: dir,
		Force:     true,
	}
	_, err := Generate(ctx, opts)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}
