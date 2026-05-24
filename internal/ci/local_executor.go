package ci

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type LocalExecutor struct{}

func NewLocalExecutor() *LocalExecutor {
	return &LocalExecutor{}
}

func (e *LocalExecutor) Execute(ctx context.Context, step StepConfig) (*StepResult, error) {
	if len(step.Command) == 0 {
		return &StepResult{
			Name:   step.Name,
			Status: StatusFailed,
			Err:    fmt.Errorf("%w: commande vide", ErrInvalidStep),
		}, nil
	}

	start := time.Now()

	cmd := exec.CommandContext(ctx, step.Command[0], step.Command[1:]...)

	if step.WorkDir != "" {
		cmd.Dir = step.WorkDir
	}

	if len(step.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range step.Env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	duration := time.Since(start)

	if err != nil {
		select {
		case <-ctx.Done():
			return &StepResult{
				Name:     step.Name,
				Status:   StatusCancelled,
				Output:   buf.String(),
				Duration: duration,
			}, nil
		default:
		}

		return &StepResult{
			Name:     step.Name,
			Status:   StatusFailed,
			Output:   buf.String(),
			Duration: duration,
			Err:      err,
		}, nil
	}

	return &StepResult{
		Name:     step.Name,
		Status:   StatusSucceeded,
		Output:   buf.String(),
		Duration: duration,
	}, nil
}
