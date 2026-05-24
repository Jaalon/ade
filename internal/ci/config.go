package ci

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

func DefaultConfig() PipelineConfig {
	stageName := func(t StageType) string {
		if n := stageDescriptions[t]; n != "" {
			return n
		}
		return string(t)
	}

	stages := make([]StageConfig, len(AllStages()))
	for i, t := range AllStages() {
		stages[i] = StageConfig{
			Type:    t,
			Name:    stageName(t),
			Enabled: true,
			Steps: []StepConfig{
				{Name: stageName(t), Command: []string{"echo", "Configure me in ade-pipeline.yaml"}},
			},
		}
	}
	return PipelineConfig{Stages: stages}
}

func LoadConfig(path string) (PipelineConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return PipelineConfig{}, fmt.Errorf("%w: lecture de %s: %v", ErrConfigInvalid, path, err)
	}

	var cfg PipelineConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return PipelineConfig{}, fmt.Errorf("%w: parsing YAML de %s: %v", ErrConfigInvalid, path, err)
	}

	if cfg.Stages == nil {
		return DefaultConfig(), nil
	}

	return cfg, nil
}

func ValidateConfig(cfg PipelineConfig) error {
	validTypes := make(map[StageType]int)
	for i, t := range AllStages() {
		validTypes[t] = i
	}

	var errs []string
	seen := make(map[StageType]bool)

	for _, s := range cfg.Stages {
		order, ok := validTypes[s.Type]
		if !ok {
			errs = append(errs, fmt.Sprintf("stage %q: %v", s.Type, ErrUnknownStage))
			continue
		}

		if seen[s.Type] {
			errs = append(errs, fmt.Sprintf("stage %q: doublon", s.Type))
			continue
		}
		seen[s.Type] = true

		if !s.Enabled {
			continue
		}

		if len(s.Steps) == 0 {
			errs = append(errs, fmt.Sprintf("stage %q: aucun step défini", s.Type))
			continue
		}

		for j, step := range s.Steps {
			if len(step.Command) == 0 {
				errs = append(errs, fmt.Sprintf("stage %q step %d: %v (command vide)", s.Type, j, ErrInvalidStep))
			}
		}

		_ = order
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: %s", ErrConfigInvalid, strings.Join(errs, "; "))
	}

	reorderStages(cfg)

	return nil
}

func reorderStages(cfg PipelineConfig) {
	order := make(map[StageType]int)
	for i, t := range AllStages() {
		order[t] = i
	}

	sort.SliceStable(cfg.Stages, func(i, j int) bool {
		return order[cfg.Stages[i].Type] < order[cfg.Stages[j].Type]
	})
}
