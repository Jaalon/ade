package validation

import "errors"

var (
	ErrConfigInvalid    = errors.New("configuration de validation invalide")
	ErrModuleNotFound   = errors.New("module de validation non trouvé")
	ErrModulePanic      = errors.New("le module a paniqué pendant l'exécution")
	ErrValidationFailed = errors.New("la validation a échoué")
)
