package validation

import "context"

type Validator interface {
	Name() string
	Description() string
	Detect(ctx context.Context) (bool, error)
	Validate(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error)
}

var registry []Validator

func Register(v Validator) {
	registry = append(registry, v)
}

func Modules() []Validator {
	result := make([]Validator, len(registry))
	copy(result, registry)
	return result
}

func DetectModules(ctx context.Context) ([]Validator, error) {
	var detected []Validator
	for _, v := range registry {
		ok, err := v.Detect(ctx)
		if err != nil {
			continue
		}
		if ok {
			detected = append(detected, v)
		}
	}
	return detected, nil
}
