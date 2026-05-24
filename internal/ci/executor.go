package ci

import "context"

type Executor interface {
	Execute(ctx context.Context, step StepConfig) (*StepResult, error)
}

type DockerStepExecutor struct{}

func NewDockerStepExecutor() *DockerStepExecutor {
	return &DockerStepExecutor{}
}

func (e *DockerStepExecutor) Execute(ctx context.Context, step StepConfig) (*StepResult, error) {
	return nil, ErrNotImplemented
}
