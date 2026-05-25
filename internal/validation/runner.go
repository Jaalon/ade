package validation

import (
	"context"
	"fmt"
	"time"
)

type ValidationRunner struct {
	AllowList []string
}

func NewValidationRunner() *ValidationRunner {
	return &ValidationRunner{}
}

func (r *ValidationRunner) Run(ctx context.Context, cfg ValidationConfig) (*ValidationReport, error) {
	startedAt := time.Now()

	report := &ValidationReport{
		Status:    StatusPassed,
		StartedAt: startedAt,
		Config:    cfg,
	}

	modules, err := r.selectModules(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if len(modules) == 0 {
		report.CompletedAt = time.Now()
		report.Duration = report.CompletedAt.Sub(report.StartedAt)
		return report, nil
	}

	anyFailed := false

	for _, v := range modules {
		select {
		case <-ctx.Done():
			report.Modules = append(report.Modules, ModuleResult{
				ModuleName: v.Name(),
				Status:     StatusSkipped,
			})
			continue
		default:
		}

		moduleResult := r.runModule(ctx, v, cfg)
		report.Modules = append(report.Modules, *moduleResult)

		if moduleResult.Status == StatusFailed || moduleResult.Status == StatusError {
			anyFailed = true
		}
	}

	completedAt := time.Now()
	report.CompletedAt = completedAt
	report.Duration = completedAt.Sub(report.StartedAt)

	if anyFailed {
		report.Status = StatusFailed
	} else {
		select {
		case <-ctx.Done():
			report.Status = StatusSkipped
		default:
			report.Status = StatusPassed
		}
	}

	return report, nil
}

func (r *ValidationRunner) runModule(ctx context.Context, v Validator, cfg ValidationConfig) *ModuleResult {
	moduleCfg := findModuleConfig(v.Name(), cfg.Modules)

	if !moduleCfg.Enabled {
		return &ModuleResult{
			ModuleName: v.Name(),
			Status:     StatusSkipped,
		}
	}

	startedAt := time.Now()
	result := &ModuleResult{
		ModuleName: v.Name(),
		StartedAt:  startedAt,
	}

	func() {
		defer func() {
			if rec := recover(); rec != nil {
				result.Status = StatusError
				result.Error = fmt.Errorf("%w: %v", ErrModulePanic, rec)
			}
		}()

		modResult, err := v.Validate(ctx, moduleCfg)
		if err != nil {
			result.Status = StatusError
			result.Error = err
			return
		}
		if modResult != nil {
			result.Checks = modResult.Checks
			result.Status = modResult.Status
			result.Error = modResult.Error
		} else {
			result.Status = StatusPassed
		}
	}()

	result.Duration = time.Since(startedAt)

	if result.Status == "" {
		result.Status = StatusPassed
	}

	return result
}

func (r *ValidationRunner) selectModules(ctx context.Context, cfg ValidationConfig) ([]Validator, error) {
	allModules := Modules()
	if len(allModules) == 0 {
		return nil, nil
	}

	var selected []Validator

	if len(cfg.Modules) > 0 {
		configMap := make(map[string]ModuleConfig)
		for _, mc := range cfg.Modules {
			configMap[mc.Name] = mc
		}
		for _, v := range allModules {
			if _, ok := configMap[v.Name()]; ok {
				selected = append(selected, v)
			}
		}
	} else {
		detected, err := DetectModules(ctx)
		if err != nil {
			return nil, err
		}
		selected = detected
	}

	if len(r.AllowList) > 0 {
		allowSet := make(map[string]bool)
		for _, name := range r.AllowList {
			allowSet[name] = true
		}
		var filtered []Validator
		for _, v := range selected {
			if allowSet[v.Name()] {
				filtered = append(filtered, v)
			}
		}
		selected = filtered
	}

	return selected, nil
}

func findModuleConfig(name string, modules []ModuleConfig) ModuleConfig {
	for _, m := range modules {
		if m.Name == name {
			return m
		}
	}
	return ModuleConfig{Name: name, Enabled: true}
}
