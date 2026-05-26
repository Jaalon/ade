package orchestrator

import "errors"

var (
	ErrProjectNotFound      = errors.New("projet introuvable")
	ErrProjectAlreadyExists = errors.New("un projet avec ce nom existe déjà")
	ErrWorkflowNotFound     = errors.New("workflow introuvable")
)
