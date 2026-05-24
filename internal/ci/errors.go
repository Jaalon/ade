package ci

import "errors"

var (
	ErrUnknownStage   = errors.New("stage inconnu")
	ErrInvalidStep    = errors.New("étape invalide")
	ErrConfigInvalid  = errors.New("configuration invalide")
	ErrStageOrder     = errors.New("ordre des stages invalide")
	ErrNotImplemented = errors.New("non implémenté — sera disponible dans une version ultérieure")
)
