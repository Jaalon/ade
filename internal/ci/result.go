package ci

import "time"

type StepResult struct {
	Name     string
	Status   StageStatus
	Output   string
	Duration time.Duration
	Err      error
}

type StageResult struct {
	Type     StageType
	Name     string
	Status   StageStatus
	Steps    []StepResult
	Duration time.Duration
	Started  time.Time
}

type PipelineResult struct {
	Stages      []StageResult
	Status      StageStatus
	Duration    time.Duration
	StartedAt   time.Time
	CompletedAt time.Time
}

func (r *PipelineResult) Failed() bool {
	for _, s := range r.Stages {
		if s.Status == StatusFailed {
			return true
		}
	}
	return false
}

func (r *PipelineResult) Succeeded() bool {
	for _, s := range r.Stages {
		if s.Status == StatusFailed || s.Status == StatusCancelled {
			return false
		}
	}
	return true
}
