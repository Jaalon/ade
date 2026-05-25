package validation

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type ValidationConfig struct {
	Modules   []ModuleConfig `yaml:"modules" json:"modules"`
	OutputDir string         `yaml:"output_dir,omitempty" json:"output_dir,omitempty"`
	Formats   []string       `yaml:"formats,omitempty" json:"formats,omitempty"`
}

type ModuleConfig struct {
	Name    string            `yaml:"name" json:"name"`
	Enabled bool              `yaml:"enabled" json:"enabled"`
	Options map[string]string `yaml:"options,omitempty" json:"options,omitempty"`
	Checks  []string          `yaml:"checks,omitempty" json:"checks,omitempty"`
}

func DefaultConfig() ValidationConfig {
	return ValidationConfig{
		Modules:   nil,
		OutputDir: ".ade/validation",
		Formats:   []string{"json"},
	}
}

func LoadConfig(path string) (ValidationConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return ValidationConfig{}, fmt.Errorf("%w: lecture de %s: %v", ErrConfigInvalid, path, err)
	}

	var cfg ValidationConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return ValidationConfig{}, fmt.Errorf("%w: parsing YAML de %s: %v", ErrConfigInvalid, path, err)
	}

	if cfg.Formats == nil {
		cfg.Formats = []string{"json"}
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = ".ade/validation"
	}
	if cfg.Modules == nil {
		cfg.Modules = []ModuleConfig{}
	}

	return cfg, nil
}

func ValidateConfig(cfg ValidationConfig) error {
	validFormats := map[string]bool{"json": true, "junit": true}
	var errs []string

	for _, f := range cfg.Formats {
		if !validFormats[f] {
			errs = append(errs, fmt.Sprintf("format inconnu: %q (formats supportés: json, junit)", f))
		}
	}

	seen := make(map[string]bool)
	for _, m := range cfg.Modules {
		if m.Name != "" {
			if seen[m.Name] {
				errs = append(errs, fmt.Sprintf("module %q: doublon dans la configuration", m.Name))
			}
			seen[m.Name] = true
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: %s", ErrConfigInvalid, strings.Join(errs, "; "))
	}

	return nil
}
