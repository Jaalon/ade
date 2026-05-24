package ci

import (
	"context"
	"fmt"
	"time"
)

type Pipeline interface {
	Run(ctx context.Context, config PipelineConfig) (*PipelineResult, error)
	Validate(config PipelineConfig) error
}

type PipelineRunner struct {
	executor Executor
}

func NewPipelineRunner(executor Executor) *PipelineRunner {
	return &PipelineRunner{executor: executor}
}

func (r *PipelineRunner) Validate(config PipelineConfig) error {
	return ValidateConfig(config)
}

func (r *PipelineRunner) Run(ctx context.Context, config PipelineConfig) (*PipelineResult, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	startedAt := time.Now()
	allTypes := AllStages()
	typeOrder := make(map[StageType]int)
	for i, t := range allTypes {
		typeOrder[t] = i
	}

	stageMap := make(map[StageType]StageConfig)
	for _, s := range config.Stages {
		stageMap[s.Type] = s
	}

	result := &PipelineResult{
		StartedAt: startedAt,
		Status:    StatusRunning,
	}

	anyFailed := false

	for _, t := range allTypes {
		stageCfg, exists := stageMap[t]
		if !exists || !stageCfg.Enabled {
			result.Stages = append(result.Stages, StageResult{
				Type:   t,
				Name:   stageDescriptions[t],
				Status: StatusSkipped,
			})
			continue
		}

		if anyFailed {
			result.Stages = append(result.Stages, StageResult{
				Type:   t,
				Name:   stageCfg.Name,
				Status: StatusSkipped,
			})
			continue
		}

		select {
		case <-ctx.Done():
			result.Status = StatusCancelled
			result.CompletedAt = time.Now()
			result.Duration = result.CompletedAt.Sub(result.StartedAt)

			result.Stages = append(result.Stages, StageResult{
				Type:   t,
				Name:   stageCfg.Name,
				Status: StatusCancelled,
			})
			for _, remaining := range allTypes[len(result.Stages):] {
				cfg, ok := stageMap[remaining]
				name := stageDescriptions[remaining]
				if ok {
					name = cfg.Name
				}
				result.Stages = append(result.Stages, StageResult{
					Type: remaining, Name: name, Status: StatusSkipped,
				})
			}
			return result, nil
		default:
		}

		stageStart := time.Now()
		stageResult := StageResult{
			Type:    t,
			Name:    stageCfg.Name,
			Status:  StatusRunning,
			Started: stageStart,
		}

		stageFailed := false
		for _, step := range stageCfg.Steps {
			select {
			case <-ctx.Done():
				stageResult.Status = StatusCancelled
				stageResult.Duration = time.Since(stageStart)
				result.Status = StatusCancelled
				result.CompletedAt = time.Now()
				result.Duration = result.CompletedAt.Sub(result.StartedAt)
				result.Stages = append(result.Stages, stageResult)
				return result, nil
			default:
			}

			stepStart := time.Now()
			stepResult, err := r.executor.Execute(ctx, step)
			duration := time.Since(stepStart)

			if err != nil {
				sr := StepResult{
					Name:     step.Name,
					Status:   StatusFailed,
					Duration: duration,
					Err:      err,
				}
				if stepResult != nil {
					sr.Output = stepResult.Output
				}
				stageResult.Steps = append(stageResult.Steps, sr)
				stageFailed = true
				break
			}

			if stepResult == nil {
				stageResult.Steps = append(stageResult.Steps, StepResult{
					Name:     step.Name,
					Status:   StatusFailed,
					Duration: duration,
					Err:      fmt.Errorf("executor returned nil result"),
				})
				stageFailed = true
				break
			}

			stageResult.Steps = append(stageResult.Steps, *stepResult)
			if stepResult.Status == StatusFailed || stepResult.Status == StatusCancelled {
				stageFailed = true
				break
			}
		}

		stageResult.Duration = time.Since(stageStart)
		if stageFailed {
			stageResult.Status = StatusFailed
			anyFailed = true
		} else {
			stageResult.Status = StatusSucceeded
		}
		result.Stages = append(result.Stages, stageResult)
	}

	completedAt := time.Now()
	result.CompletedAt = completedAt
	result.Duration = completedAt.Sub(result.StartedAt)

	if anyFailed {
		result.Status = StatusFailed
	} else {
		result.Status = StatusSucceeded
	}

	select {
	case <-ctx.Done():
		result.Status = StatusCancelled
	default:
	}

	return result, nil
}
